package goscalar

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/swaggo/swag"

	"github.com/JhonatanRSantos/goscalar/utils"
)

const (
	// Template paths
	templatePattern = "templates/index.html"
	defaultTitle    = "Scalar API Reference"
	defaultLanguage = "en-US"

	// URL schemes
	fileScheme  = "file"
	httpScheme  = "http"
	httpsScheme = "https"
	filePrefix  = "file://"

	// HTTP client settings
	defaultTimeout = 30 * time.Second
)

var (
	//go:embed templates/scripts/api_reference.js
	embedScript string
	//go:embed templates/*.html
	embedTemplates embed.FS

	// Errors
	ErrInvalidTitle      = errors.New("title cannot be empty")
	ErrInvalidSpec       = errors.New("spec cannot be empty")
	ErrInvalidURL        = errors.New("invalid URL provided")
	ErrSpecRequired      = errors.New("spec content is required, use WithFile(), WithURL(), or WithSpec()")
	ErrUnsupportedScheme = errors.New("unsupported URL scheme, only file://, http://, and https:// are supported")
	ErrHTTPRequest       = errors.New("HTTP request failed")
	ErrEmptyResponse     = errors.New("received empty response from URL")
)

// Scalar represents the API documentation generator
type Scalar struct {
	config Config
}

// Config holds the template configuration
type Config struct {
	Title      string
	Language   string
	Script     template.JS
	Content    string
	HTTPClient *http.Client // Optional HTTP client for URL requests
}

// Option defines a configuration option for Scalar
type Option func(s *Scalar) error

// WithTitle sets the documentation title
func WithTitle(title string) Option {
	return func(s *Scalar) error {
		title = strings.TrimSpace(title)
		if title == "" {
			return ErrInvalidTitle
		}
		s.config.Title = title
		return nil
	}
}

// WithLanguage sets the documentation language
func WithLanguage(language string) Option {
	return func(s *Scalar) error {
		language = strings.TrimSpace(language)
		if language == "" {
			language = defaultLanguage
		}
		s.config.Language = language
		return nil
	}
}

// WithFile loads specification from a file path
func WithFile(filePath string) Option {
	return func(s *Scalar) error {
		content, err := loadSpecFromFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to load spec from file: %w", err)
		}
		s.config.Content = escapeJSString(content)
		return nil
	}
}

// WithURL loads specification from a URL (HTTP/HTTPS)
func WithURL(specURL string) Option {
	return func(s *Scalar) error {
		content, err := loadSpecFromURL(specURL, s.config.HTTPClient)
		if err != nil {
			return fmt.Errorf("failed to load spec from URL: %w", err)
		}
		s.config.Content = escapeJSString(content)
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client for URL requests
func WithHTTPClient(client *http.Client) Option {
	return func(s *Scalar) error {
		if client == nil {
			client = &http.Client{Timeout: defaultTimeout}
		}
		s.config.HTTPClient = client
		return nil
	}
}

// WithSpec loads specification from swag.Spec object
func WithSpec(spec *swag.Spec) Option {
	return func(s *Scalar) error {
		if spec == nil {
			return ErrInvalidSpec
		}
		content := spec.ReadDoc()
		if content == "" {
			return ErrInvalidSpec
		}
		s.config.Content = escapeJSString(normalizeSpecContent(content))
		return nil
	}
}

// WithSpecContent loads specification from raw content
func WithSpecContent(content string) Option {
	return func(s *Scalar) error {
		content = strings.TrimSpace(content)
		if content == "" {
			return ErrInvalidSpec
		}
		s.config.Content = escapeJSString(normalizeSpecContent(content))
		return nil
	}
}

// NewScalar creates a new Scalar instance with the given options
func NewScalar(options ...Option) (*Scalar, error) {
	scalar := &Scalar{
		config: Config{
			Title:      defaultTitle,
			Language:   defaultLanguage,
			Script:     template.JS(fmt.Sprintf("<script>%s</script>", embedScript)),
			HTTPClient: &http.Client{Timeout: defaultTimeout},
		},
	}

	for _, opt := range options {
		if err := opt(scalar); err != nil {
			return nil, err
		}
	}

	if scalar.config.Content == "" {
		return nil, ErrSpecRequired
	}

	return scalar, nil
}

// RenderDocs renders the API documentation to the provided writer
func (s *Scalar) RenderDocs(writer io.Writer) error {
	if writer == nil {
		return errors.New("writer cannot be nil")
	}

	tmpl, err := utils.ParseTemplateFromFS(embedTemplates, templatePattern)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	if err := tmpl.Execute(writer, s.config); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	return nil
}

// loadSpecFromFile loads specification content from a file
func loadSpecFromFile(filePath string) (string, error) {
	fileURL, err := normalizeFileURL(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to normalize file URL: %w", err)
	}

	content, err := readFileFromURL(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return normalizeSpecContent(string(content)), nil
}

// loadSpecFromURL loads specification content from a URL
func loadSpecFromURL(specURL string, client *http.Client) (string, error) {
	if err := validateURL(specURL); err != nil {
		return "", err
	}

	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}

	content, err := fetchFromURL(specURL, client)
	if err != nil {
		return "", fmt.Errorf("failed to fetch from URL: %w", err)
	}

	return normalizeSpecContent(string(content)), nil
}

// validateURL validates if the URL is properly formatted and uses supported scheme
func validateURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return ErrInvalidURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidURL, err.Error())
	}

	switch parsedURL.Scheme {
	case httpScheme, httpsScheme:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedScheme, parsedURL.Scheme)
	}
}

// fetchFromURL fetches content from HTTP/HTTPS URL
func fetchFromURL(specURL string, client *http.Client) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, specURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json, application/yaml, text/yaml, */*")
	req.Header.Set("User-Agent", "go-scalar/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrHTTPRequest, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: HTTP %d %s", ErrHTTPRequest, resp.StatusCode, resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if len(content) == 0 {
		return nil, ErrEmptyResponse
	}
	return content, nil
}

// normalizeFileURL ensures the file path is a proper file:// URL
func normalizeFileURL(filePath string) (string, error) {
	// Already a file URL
	if strings.HasPrefix(filePath, filePrefix) {
		path := strings.TrimPrefix(filePath, filePrefix)
		if !filepath.IsAbs(path) {
			// Convert relative path to absolute
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("failed to get absolute path: %w", err)
			}
			return filePrefix + absPath, nil
		}
		return filePath, nil
	}

	// Convert to absolute path if relative
	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path: %w", err)
		}
		filePath = absPath
	}

	return filePrefix + filePath, nil
}

// readFileFromURL reads file content from a file:// URL
func readFileFromURL(fileURL string) ([]byte, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsedURL.Scheme != fileScheme {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedScheme, parsedURL.Scheme)
	}

	content, err := os.ReadFile(parsedURL.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", parsedURL.Path, err)
	}

	return content, nil
}

// normalizeSpecContent normalizes specification content to JSON string
func normalizeSpecContent(specContent any) string {
	switch spec := specContent.(type) {
	case func() map[string]any:
		// Function that returns map
		result := spec()
		if jsonData, err := json.Marshal(result); err == nil {
			return string(jsonData)
		}
	case map[string]any:
		// Direct map
		if jsonData, err := json.Marshal(spec); err == nil {
			return string(jsonData)
		}
	case string:
		// String content - validate if it's JSON
		spec = strings.TrimSpace(spec)
		if isValidJSON(spec) {
			return spec
		}
	}
	return ""
}

// isValidJSON checks if a string is valid JSON
func isValidJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

// Builder provides a fluent interface for creating Scalar instances
type Builder struct {
	options []Option
}

// NewBuilder creates a new Scalar builder
func NewBuilder() *Builder {
	return &Builder{}
}

// Title sets the documentation title
func (b *Builder) Title(title string) *Builder {
	b.options = append(b.options, WithTitle(title))
	return b
}

// Language sets the documentation language
func (b *Builder) Language(language string) *Builder {
	b.options = append(b.options, WithLanguage(language))
	return b
}

// File loads specification from file
func (b *Builder) File(filePath string) *Builder {
	b.options = append(b.options, WithFile(filePath))
	return b
}

// Spec loads specification from swag.Spec
func (b *Builder) Spec(spec *swag.Spec) *Builder {
	b.options = append(b.options, WithSpec(spec))
	return b
}

// Content loads specification from raw content
func (b *Builder) Content(content string) *Builder {
	b.options = append(b.options, WithSpecContent(content))
	return b
}

// URL loads specification from URL
func (b *Builder) URL(specURL string) *Builder {
	b.options = append(b.options, WithURL(specURL))
	return b
}

// HTTPClient sets custom HTTP client
func (b *Builder) HTTPClient(client *http.Client) *Builder {
	b.options = append(b.options, WithHTTPClient(client))
	return b
}

// Build creates the Scalar instance
func (b *Builder) Build() (*Scalar, error) {
	return NewScalar(b.options...)
}

// Miscellaneous

// FromFile creates a Scalar instance from a file path
func FromFile(filePath string, options ...Option) (*Scalar, error) {
	opts := append([]Option{WithFile(filePath)}, options...)
	return NewScalar(opts...)
}

// FromURL creates a Scalar instance from a URL
func FromURL(specURL string, options ...Option) (*Scalar, error) {
	opts := append([]Option{WithURL(specURL)}, options...)
	return NewScalar(opts...)
}

// FromSpec creates a Scalar instance from a swag.Spec
func FromSpec(spec *swag.Spec, options ...Option) (*Scalar, error) {
	opts := append([]Option{WithSpec(spec)}, options...)
	return NewScalar(opts...)
}

// FromContent creates a Scalar instance from raw content
func FromContent(content string, options ...Option) (*Scalar, error) {
	opts := append([]Option{WithSpecContent(content)}, options...)
	return NewScalar(opts...)
}

func escapeJSString(raw string) string {
	if raw == "" {
		return raw
	}

	var builder strings.Builder
	builder.Grow(len(raw) + len(raw)/10)

	for _, r := range raw {
		switch r {
		case '`':
			builder.WriteString("\\`")
		case '"':
			builder.WriteString(`\"`)
		case '\\':
			builder.WriteString(`\\`)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		case '\f':
			builder.WriteString(`\f`)
		case '\b':
			builder.WriteString(`\b`)
		case '\v':
			builder.WriteString(`\v`)
		case '\u0000':
			builder.WriteString(`\u0000`)
		default:
			if r < 32 || r == 127 {
				builder.WriteString(`\u`)
				hex := "0123456789abcdef"
				builder.WriteByte(hex[(r>>12)&0xf])
				builder.WriteByte(hex[(r>>8)&0xf])
				builder.WriteByte(hex[(r>>4)&0xf])
				builder.WriteByte(hex[r&0xf])
			} else {
				builder.WriteRune(r)
			}
		}
	}

	return builder.String()
}
