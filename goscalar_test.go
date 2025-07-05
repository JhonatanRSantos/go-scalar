package goscalar

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/swaggo/swag"
)

func Test_WithTitle(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid title",
			title:       "My API Documentation",
			expectError: false,
		},
		{
			name:        "empty title",
			title:       "",
			expectError: true,
			expectedErr: ErrInvalidTitle,
		},
		{
			name:        "whitespace only title",
			title:       "   ",
			expectError: true,
			expectedErr: ErrInvalidTitle,
		},
		{
			name:        "title with spaces",
			title:       "  My API  ",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scalar := &Scalar{
				config: Config{
					Title:    defaultTitle,
					Language: defaultLanguage,
				},
			}

			err := WithTitle(tt.title)(scalar)

			if tt.expectError {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, strings.TrimSpace(tt.title), scalar.config.Title)
			}
		})
	}
}

func Test_WithLanguage(t *testing.T) {
	tests := []struct {
		name             string
		language         string
		expectedLanguage string
	}{
		{
			name:             "valid language",
			language:         "pt-BR",
			expectedLanguage: "pt-BR",
		},
		{
			name:             "empty language uses default",
			language:         "",
			expectedLanguage: defaultLanguage,
		},
		{
			name:             "whitespace only language uses default",
			language:         "   ",
			expectedLanguage: defaultLanguage,
		},
		{
			name:             "language with spaces",
			language:         "  en-US  ",
			expectedLanguage: "en-US",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scalar := &Scalar{
				config: Config{
					Title:    defaultTitle,
					Language: defaultLanguage,
				},
			}

			err := WithLanguage(tt.language)(scalar)

			require.NoError(t, err)
			require.Equal(t, tt.expectedLanguage, scalar.config.Language)
		})
	}
}

func Test_WithHTTPClient(t *testing.T) {
	tests := []struct {
		name         string
		client       *http.Client
		expectNonNil bool
	}{
		{
			name:         "valid client",
			client:       &http.Client{Timeout: 10 * time.Second},
			expectNonNil: true,
		},
		{
			name:         "nil client creates default",
			client:       nil,
			expectNonNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scalar := &Scalar{
				config: Config{
					Title:    defaultTitle,
					Language: defaultLanguage,
				},
			}

			err := WithHTTPClient(tt.client)(scalar)

			require.NoError(t, err)
			if tt.expectNonNil {
				require.NotNil(t, scalar.config.HTTPClient)
			}
		})
	}
}

func Test_WithSpec(t *testing.T) {
	validSpec := &swag.Spec{SwaggerTemplate: `{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`}
	emptySpec := &swag.Spec{SwaggerTemplate: ""}

	tests := []struct {
		name        string
		spec        *swag.Spec
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid spec",
			spec:        validSpec,
			expectError: false,
		},
		{
			name:        "nil spec",
			spec:        nil,
			expectError: true,
			expectedErr: ErrInvalidSpec,
		},
		{
			name:        "empty spec content",
			spec:        emptySpec,
			expectError: true,
			expectedErr: ErrInvalidSpec,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scalar := &Scalar{
				config: Config{
					Title:    defaultTitle,
					Language: defaultLanguage,
				},
			}

			err := WithSpec(tt.spec)(scalar)

			if tt.expectError {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, scalar.config.Content)
			}
		})
	}
}

func Test_WithSpecContent(t *testing.T) {
	validJSON := `{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`

	tests := []struct {
		name        string
		content     string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid JSON content",
			content:     validJSON,
			expectError: false,
		},
		{
			name:        "empty content",
			content:     "",
			expectError: true,
			expectedErr: ErrInvalidSpec,
		},
		{
			name:        "whitespace only content",
			content:     "   ",
			expectError: true,
			expectedErr: ErrInvalidSpec,
		},
		{
			name:        "content with spaces",
			content:     "  " + validJSON + "  ",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scalar := &Scalar{
				config: Config{
					Title:    defaultTitle,
					Language: defaultLanguage,
				},
			}

			err := WithSpecContent(tt.content)(scalar)

			if tt.expectError {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, scalar.config.Content)
			}
		})
	}
}

func Test_WithFile(t *testing.T) {
	tempDir := t.TempDir()

	validFile := filepath.Join(tempDir, "valid.json")
	validContent := `{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`
	require.NoError(t, os.WriteFile(validFile, []byte(validContent), 0644))

	nonExistentFile := filepath.Join(tempDir, "nonexistent.json")

	tests := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{
			name:        "valid file",
			filePath:    validFile,
			expectError: false,
		},
		{
			name:        "non-existent file",
			filePath:    nonExistentFile,
			expectError: true,
		},
		{
			name:        "empty file path",
			filePath:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scalar := &Scalar{
				config: Config{
					Title:    defaultTitle,
					Language: defaultLanguage,
				},
			}

			err := WithFile(tt.filePath)(scalar)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, scalar.config.Content)
			}
		})
	}
}

func Test_WithURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/valid":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`))
		case "/empty":
			w.WriteHeader(http.StatusOK)
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "valid URL",
			url:         server.URL + "/valid",
			expectError: false,
		},
		{
			name:        "empty response",
			url:         server.URL + "/empty",
			expectError: true,
		},
		{
			name:        "server error",
			url:         server.URL + "/error",
			expectError: true,
		},
		{
			name:        "invalid URL",
			url:         "not-a-url",
			expectError: true,
		},
		{
			name:        "unsupported scheme",
			url:         "ftp://example.com",
			expectError: true,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scalar := &Scalar{
				config: Config{
					Title:      defaultTitle,
					Language:   defaultLanguage,
					HTTPClient: &http.Client{Timeout: 5 * time.Second},
				},
			}

			err := WithURL(tt.url)(scalar)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, scalar.config.Content)
			}
		})
	}
}

func Test_NewScalar(t *testing.T) {
	validContent := `{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`

	tests := []struct {
		name        string
		options     []Option
		expectError bool
		expectedErr error
	}{
		{
			name: "valid options",
			options: []Option{
				WithTitle("Test API"),
				WithLanguage("pt-BR"),
				WithSpecContent(validContent),
			},
			expectError: false,
		},
		{
			name:        "no spec content",
			options:     []Option{WithTitle("Test API")},
			expectError: true,
			expectedErr: ErrSpecRequired,
		},
		{
			name: "invalid title",
			options: []Option{
				WithTitle(""),
				WithSpecContent(validContent),
			},
			expectError: true,
			expectedErr: ErrInvalidTitle,
		},
		{
			name: "invalid spec content",
			options: []Option{
				WithTitle("Test API"),
				WithSpecContent(""),
			},
			expectError: true,
			expectedErr: ErrInvalidSpec,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scalar, err := NewScalar(tt.options...)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, scalar)
				if tt.expectedErr != nil {
					require.Equal(t, tt.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, scalar)
				require.NotEmpty(t, scalar.config.Content)
			}
		})
	}
}

func Test_RenderDocs(t *testing.T) {
	validContent := `{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`

	tests := []struct {
		name        string
		writer      io.Writer
		expectError bool
	}{
		{
			name:        "valid writer",
			writer:      &bytes.Buffer{},
			expectError: false,
		},
		{
			name:        "nil writer",
			writer:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scalar, err := NewScalar(WithSpecContent(validContent))
			require.NoError(t, err)

			err = scalar.RenderDocs(tt.writer)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// Check if content was written
				if buf, ok := tt.writer.(*bytes.Buffer); ok {
					require.NotEmpty(t, buf.String())
				}
			}
		})
	}
}

func Test_ValidateURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid HTTP URL",
			url:         "http://example.com",
			expectError: false,
		},
		{
			name:        "valid HTTPS URL",
			url:         "https://example.com",
			expectError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
			expectedErr: ErrInvalidURL,
		},
		{
			name:        "whitespace only URL",
			url:         "   ",
			expectError: true,
			expectedErr: ErrInvalidURL,
		},
		{
			name:        "unsupported scheme",
			url:         "ftp://example.com",
			expectError: true,
			expectedErr: ErrUnsupportedScheme,
		},
		{
			name:        "invalid URL format",
			url:         "not-a-url",
			expectError: true,
			expectedErr: ErrUnsupportedScheme,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)

			if tt.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_FetchFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/valid":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"test": "content"}`))
		case "/empty":
			w.WriteHeader(http.StatusOK)
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
		case "/timeout":
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"test": "content"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name        string
		url         string
		client      *http.Client
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid response",
			url:         server.URL + "/valid",
			client:      &http.Client{Timeout: 5 * time.Second},
			expectError: false,
		},
		{
			name:        "empty response",
			url:         server.URL + "/empty",
			client:      &http.Client{Timeout: 5 * time.Second},
			expectError: true,
			expectedErr: ErrEmptyResponse,
		},
		{
			name:        "server error",
			url:         server.URL + "/error",
			client:      &http.Client{Timeout: 5 * time.Second},
			expectError: true,
			expectedErr: ErrHTTPRequest,
		},
		{
			name:        "timeout",
			url:         server.URL + "/timeout",
			client:      &http.Client{Timeout: 10 * time.Millisecond},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := fetchFromURL(tt.url, tt.client)

			if tt.expectError {
				require.Error(t, err)
				if tt.expectedErr != nil {
					require.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, content)
			}
		})
	}
}

func Test_NormalizeFileURL(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{
			name:        "absolute path",
			filePath:    "/path/to/file.json",
			expectError: false,
		},
		{
			name:        "relative path",
			filePath:    "file.json",
			expectError: false,
		},
		{
			name:        "file URL with absolute path",
			filePath:    "file:///path/to/file.json",
			expectError: false,
		},
		{
			name:        "file URL with relative path",
			filePath:    "file://file.json",
			expectError: false,
		},
		{
			name:        "empty path",
			filePath:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeFileURL(tt.filePath)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.True(t, strings.HasPrefix(result, "file://"))
			}
		})
	}
}

func Test_NormalizeSpecContent(t *testing.T) {
	validJSON := `{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`
	validMap := map[string]any{
		"openapi": "3.0.0",
		"info": map[string]any{
			"title":   "Test API",
			"version": "1.0.0",
		},
	}

	tests := []struct {
		name     string
		content  any
		expected string
	}{
		{
			name:     "valid JSON string",
			content:  validJSON,
			expected: validJSON,
		},
		{
			name:     "valid map",
			content:  validMap,
			expected: `{"info":{"title":"Test API","version":"1.0.0"},"openapi":"3.0.0"}`,
		},
		{
			name: "function returning map",
			content: func() map[string]any {
				return validMap
			},
			expected: `{"info":{"title":"Test API","version":"1.0.0"},"openapi":"3.0.0"}`,
		},
		{
			name:     "invalid JSON string",
			content:  "not json",
			expected: "",
		},
		{
			name:     "empty string",
			content:  "",
			expected: "",
		},
		{
			name:     "whitespace string",
			content:  "   ",
			expected: "",
		},
		{
			name:     "nil content",
			content:  nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeSpecContent(tt.content)

			if tt.expected == "" {
				require.Empty(t, result)
			} else {
				require.NotEmpty(t, result)
				// For JSON comparison, we need to normalize both strings
				if tt.expected != "" {
					var expectedJSON, resultJSON any
					require.NoError(t, json.Unmarshal([]byte(tt.expected), &expectedJSON))
					require.NoError(t, json.Unmarshal([]byte(result), &resultJSON))
					require.Equal(t, expectedJSON, resultJSON)
				}
			}
		})
	}
}

func Test_IsValidJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid JSON object",
			input:    `{"key": "value"}`,
			expected: true,
		},
		{
			name:     "valid JSON array",
			input:    `[1, 2, 3]`,
			expected: true,
		},
		{
			name:     "valid JSON string",
			input:    `"hello"`,
			expected: true,
		},
		{
			name:     "valid JSON number",
			input:    `42`,
			expected: true,
		},
		{
			name:     "valid JSON boolean",
			input:    `true`,
			expected: true,
		},
		{
			name:     "valid JSON null",
			input:    `null`,
			expected: true,
		},
		{
			name:     "invalid JSON",
			input:    `{key: value}`,
			expected: false,
		},
		{
			name:     "empty string",
			input:    ``,
			expected: false,
		},
		{
			name:     "invalid syntax",
			input:    `{"key": }`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidJSON(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_EscapeJSString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no special characters",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "backtick",
			input:    "hello `world`",
			expected: "hello \\`world\\`",
		},
		{
			name:     "double quote",
			input:    `hello "world"`,
			expected: `hello \"world\"`,
		},
		{
			name:     "backslash",
			input:    `hello \world`,
			expected: `hello \\world`,
		},
		{
			name:     "newline",
			input:    "hello\nworld",
			expected: "hello\\nworld",
		},
		{
			name:     "carriage return",
			input:    "hello\rworld",
			expected: "hello\\rworld",
		},
		{
			name:     "tab",
			input:    "hello\tworld",
			expected: "hello\\tworld",
		},
		{
			name:     "form feed",
			input:    "hello\fworld",
			expected: "hello\\fworld",
		},
		{
			name:     "backspace",
			input:    "hello\bworld",
			expected: "hello\\bworld",
		},
		{
			name:     "vertical tab",
			input:    "hello\vworld",
			expected: "hello\\vworld",
		},
		{
			name:     "null character",
			input:    "hello\u0000world",
			expected: "hello\\u0000world",
		},
		{
			name:     "control character",
			input:    "hello\u0001world",
			expected: "hello\\u0001world",
		},
		{
			name:     "delete character",
			input:    "hello\u007fworld",
			expected: "hello\\u007fworld",
		},
		{
			name:     "mixed special characters",
			input:    "hello\n\t\"world`",
			expected: "hello\\n\\t\\\"world\\`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeJSString(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_Builder(t *testing.T) {
	validContent := `{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`

	tests := []struct {
		name        string
		buildFunc   func(*Builder) *Builder
		expectError bool
	}{
		{
			name: "valid builder with content",
			buildFunc: func(b *Builder) *Builder {
				return b.Title("Test API").Language("pt-BR").Content(validContent)
			},
			expectError: false,
		},
		{
			name: "builder without content",
			buildFunc: func(b *Builder) *Builder {
				return b.Title("Test API").Language("pt-BR")
			},
			expectError: true,
		},
		{
			name: "builder with invalid title",
			buildFunc: func(b *Builder) *Builder {
				return b.Title("").Content(validContent)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder()
			builder = tt.buildFunc(builder)

			scalar, err := builder.Build()

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, scalar)
			} else {
				require.NoError(t, err)
				require.NotNil(t, scalar)
			}
		})
	}
}

func Test_Miscellaneous(t *testing.T) {
	validContent := `{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`
	validSpec := &swag.Spec{SwaggerTemplate: validContent}

	// Create temp file for testing
	tempDir := t.TempDir()
	validFile := filepath.Join(tempDir, "valid.json")
	require.NoError(t, os.WriteFile(validFile, []byte(validContent), 0644))

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validContent))
	}))
	defer server.Close()

	tests := []struct {
		name        string
		createFunc  func() (*Scalar, error)
		expectError bool
	}{
		{
			name: "FromFile",
			createFunc: func() (*Scalar, error) {
				return FromFile(validFile)
			},
			expectError: false,
		},
		{
			name: "FromURL",
			createFunc: func() (*Scalar, error) {
				return FromURL(server.URL)
			},
			expectError: false,
		},
		{
			name: "FromSpec",
			createFunc: func() (*Scalar, error) {
				return FromSpec(validSpec)
			},
			expectError: false,
		},
		{
			name: "FromContent",
			createFunc: func() (*Scalar, error) {
				return FromContent(validContent)
			},
			expectError: false,
		},
		{
			name: "FromFile with options",
			createFunc: func() (*Scalar, error) {
				return FromFile(validFile, WithTitle("Custom Title"))
			},
			expectError: false,
		},
		{
			name: "FromFile non-existent",
			createFunc: func() (*Scalar, error) {
				return FromFile("/non/existent/file.json")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scalar, err := tt.createFunc()

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, scalar)
			} else {
				require.NoError(t, err)
				require.NotNil(t, scalar)
				require.NotEmpty(t, scalar.config.Content)
			}
		})
	}
}

func Test_ReadFileFromURL(t *testing.T) {
	tempDir := t.TempDir()
	validFile := filepath.Join(tempDir, "valid.json")
	validContent := `{"test": "content"}`
	require.NoError(t, os.WriteFile(validFile, []byte(validContent), 0644))

	tests := []struct {
		name        string
		fileURL     string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid file URL",
			fileURL:     "file://" + validFile,
			expectError: false,
		},
		{
			name:        "non-existent file",
			fileURL:     "file:///non/existent/file.json",
			expectError: true,
		},
		{
			name:        "invalid scheme",
			fileURL:     "http://example.com/file.json",
			expectError: true,
			expectedErr: ErrUnsupportedScheme,
		},
		{
			name:        "invalid URL format",
			fileURL:     "not-a-url",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := readFileFromURL(tt.fileURL)

			if tt.expectError {
				require.Error(t, err)
				if tt.expectedErr != nil {
					require.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, content)
				require.Equal(t, validContent, string(content))
			}
		})
	}
}

func Test_LoadSpecFromFile(t *testing.T) {
	tempDir := t.TempDir()

	validFile := filepath.Join(tempDir, "valid.json")
	validContent := `{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`
	require.NoError(t, os.WriteFile(validFile, []byte(validContent), 0644))

	invalidJSONFile := filepath.Join(tempDir, "invalid.json")
	require.NoError(t, os.WriteFile(invalidJSONFile, []byte("not json"), 0644))

	tests := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{
			name:        "valid JSON file",
			filePath:    validFile,
			expectError: false,
		},
		{
			name:        "non-existent file",
			filePath:    "/non/existent/file.json",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := loadSpecFromFile(tt.filePath)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, content)
			}
		})
	}
}

func Test_LoadSpecFromURL(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/valid":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`))
		case "/invalid":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name        string
		specURL     string
		client      *http.Client
		expectError bool
	}{
		{
			name:        "valid URL with client",
			specURL:     server.URL + "/valid",
			client:      &http.Client{Timeout: 5 * time.Second},
			expectError: false,
		},
		{
			name:        "valid URL with nil client",
			specURL:     server.URL + "/valid",
			client:      nil,
			expectError: false,
		},
		{
			name:        "invalid URL response",
			specURL:     server.URL + "/invalid",
			client:      &http.Client{Timeout: 5 * time.Second},
			expectError: true,
		},
		{
			name:        "invalid URL format",
			specURL:     "not-a-url",
			client:      &http.Client{Timeout: 5 * time.Second},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := loadSpecFromURL(tt.specURL, tt.client)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, content)
			}
		})
	}
}

// Integration tests
func Test_ScalarWorkflow(t *testing.T) {
	// Create a temporary file with valid OpenAPI spec
	tempDir := t.TempDir()
	specFile := filepath.Join(tempDir, "openapi.json")
	specContent := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0",
			"description": "A test API for integration testing"
		},
		"paths": {
			"/users": {
				"get": {
					"summary": "Get users",
					"responses": {
						"200": {
							"description": "Successful response"
						}
					}
				}
			}
		}
	}`
	require.NoError(t, os.WriteFile(specFile, []byte(specContent), 0644))

	// Test complete workflow: create scalar from file, render docs
	scalar, err := NewScalar(
		WithTitle("Integration Test API"),
		WithLanguage("en-US"),
		WithFile(specFile),
	)
	require.NoError(t, err)
	require.NotNil(t, scalar)

	// Verify configuration
	require.Equal(t, "Integration Test API", scalar.config.Title)
	require.Equal(t, "en-US", scalar.config.Language)
	require.NotEmpty(t, scalar.config.Content)

	// Render documentation
	var buf bytes.Buffer
	err = scalar.RenderDocs(&buf)
	require.NoError(t, err)

	rendered := buf.String()
	require.NotEmpty(t, rendered)
	require.Contains(t, rendered, "Integration Test API")
	require.Contains(t, rendered, "en-US")
}

func Test_BuilderWorkflow(t *testing.T) {
	specContent := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Builder Test API",
			"version": "2.0.0"
		},
		"paths": {}
	}`

	// Test builder pattern
	scalar, err := NewBuilder().
		Title("Builder Pattern API").
		Language("pt-BR").
		Content(specContent).
		HTTPClient(&http.Client{Timeout: 10 * time.Second}).
		Build()

	require.NoError(t, err)
	require.NotNil(t, scalar)

	require.Equal(t, "Builder Pattern API", scalar.config.Title)
	require.Equal(t, "pt-BR", scalar.config.Language)
	require.NotEmpty(t, scalar.config.Content)
	require.Equal(t, 10*time.Second, scalar.config.HTTPClient.Timeout)
}

func Test_HTTPServerWorkflow(t *testing.T) {
	specContent := `{
		"openapi": "3.0.0",
		"info": {
			"title": "HTTP Server Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/test": {
				"get": {
					"summary": "Test endpoint",
					"responses": {
						"200": {
							"description": "Success"
						}
					}
				}
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(specContent))
	}))
	defer server.Close()

	// Test loading spec from HTTP server
	scalar, err := FromURL(server.URL, WithTitle("HTTP Loaded API"))
	require.NoError(t, err)
	require.NotNil(t, scalar)

	require.Equal(t, "HTTP Loaded API", scalar.config.Title)
	require.NotEmpty(t, scalar.config.Content)

	// Render and verify
	var buf bytes.Buffer
	err = scalar.RenderDocs(&buf)
	require.NoError(t, err)
	require.NotEmpty(t, buf.String())
}

func Test_SwagSpecWorkflow(t *testing.T) {
	mockSpec := &swag.Spec{
		SwaggerTemplate: `{
			"openapi": "3.0.0",
			"info": {
				"title": "Swag Test API",
				"version": "1.0.0"
			},
			"paths": {
				"/swagger": {
					"get": {
						"summary": "Swagger endpoint",
						"responses": {
							"200": {
								"description": "Success"
							}
						}
					}
				}
			}
		}`,
	}

	// Test loading from swag.Spec
	scalar, err := FromSpec(mockSpec, WithTitle("Swag Integration Test"))
	require.NoError(t, err)
	require.NotNil(t, scalar)

	require.Equal(t, "Swag Integration Test", scalar.config.Title)
	require.NotEmpty(t, scalar.config.Content)

	// Verify the content contains expected structure
	require.Contains(t, scalar.config.Content, "Swag Test API")
	require.Contains(t, scalar.config.Content, "/swagger")
}
