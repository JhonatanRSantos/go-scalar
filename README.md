# Go Scalar

A Go API documentation generator that uses [Scalar](https://github.com/scalar/scalar) to create elegant and interactive documentation interfaces from OpenAPI/Swagger specifications.

## Features

- ✨ Modern and responsive interface using [Scalar](https://github.com/scalar/scalar)
- 📁 Support for loading specifications from local files
- 🌐 Support for loading specifications via HTTP/HTTPS
- 📊 Integration with `swag.Spec` [Swag](https://github.com/swaggo/swag)
- 🔧 Flexible configuration with builder pattern
- ⚡ Embedded templates for simple distribution

## Installation

```bash
go get github.com/JhonatanRSantos/goscalar
```

## Basic Usage

### Loading from a file

```go
package main

import (
    "os"
    "github.com/JhonatanRSantos/goscalar"
)

func main() {
    scalar, err := goscalar.FromFile("./docs/swagger.json")
    if err != nil {
        panic(err)
    }

    if err := scalar.RenderDocs(os.Stdout); err != nil {
        panic(err)
    }
}
```

### Loading from a URL

```go
package main

import (
    "os"
    "github.com/JhonatanRSantos/goscalar"
)

func main() {
    scalar, err := goscalar.FromURL("https://cdn.jsdelivr.net/npm/@scalar/galaxy/dist/latest.json")
    if err != nil {
        panic(err)
    }

    if err := scalar.RenderDocs(os.Stdout); err != nil {
        panic(err)
    }
}
```

### Using with Swag

```go
package main

import (
    "os"
    "github.com/JhonatanRSantos/goscalar"
    "github.com/swaggo/swag"
)

func main() {
    spec := &swag.Spec{
        // your specification here
    }

    scalar, err := goscalar.FromSpec(spec)
    if err != nil {
        panic(err)
    }

    if err := scalar.RenderDocs(os.Stdout); err != nil {
        panic(err)
    }
}
```

## Advanced Configuration

### Using the Builder Pattern

```go
package main

import (
    "net/http"
    "os"
    "time"
    "github.com/JhonatanRSantos/goscalar"
)

func main() {
    client := &http.Client{
        Timeout: 60 * time.Second,
    }

    scalar, err := goscalar.NewBuilder().
        Title("My API Documentation").
        Language("en-US").
        URL("https://cdn.jsdelivr.net/npm/@scalar/galaxy/dist/latest.json").
        HTTPClient(client).
        Build()

    if err != nil {
        panic(err)
    }

    if err := scalar.RenderDocs(os.Stdout); err != nil {
        panic(err)
    }
}
```

### Using Options

```go
package main

import (
    "os"
    "github.com/JhonatanRSantos/goscalar"
)

func main() {
    scalar, err := goscalar.NewScalar(
		goscalar.WithTitle("My API Documentation"),
		goscalar.WithLanguage("en-US"),
		goscalar.WithURL("https://cdn.jsdelivr.net/npm/@scalar/galaxy/dist/latest.json"),
		goscalar.WithHTTPClient(client),
	)

    if err != nil {
        panic(err)
    }

    if err := scalar.RenderDocs(os.Stdout); err != nil {
        panic(err)
    }
}
```

## HTTP Server Integration

### Gin Framework

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/JhonatanRSantos/goscalar"
)

func main() {
    r := gin.Default()

    scalar, err := goscalar.FromFile("./docs/swagger.json", goscalar.WithTitle("My API Documentation"))
    if err != nil {
        panic(err)
    }

    r.GET("/docs", func(c *gin.Context) {
        c.Header("Content-Type", "text/html")
        if err := scalar.RenderDocs(c.Writer); err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
        }
    })

    r.Run(":8080")
}
```

### Fiber Framework

```go
package main

import (
	"github.com/JhonatanRSantos/goscalar"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	scalar, err := goscalar.FromSpec(docs.SwaggerInfo, goscalar.WithTitle("My API Documentation"))
	if err != nil {
		panic(err)
	}

	app.Get("/docs", func(c *fiber.Ctx) error {
		if err := scalar.RenderDocs(c.Response().BodyWriter()); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error rendering docs: " + err.Error())
		}
		c.Response().Header.Set("Content-Type", "text/html; charset=utf-8")
		c.Response().Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Response().Header.Set("Pragma", "no-cache")
		c.Response().Header.Set("Expires", "0")
		return nil
	})
	app.Listen(":8080")
}
```

### Standard HTTP

```go
package main

import (
    "net/http"
    "log"
    "github.com/JhonatanRSantos/goscalar"
)

func main() {
    scalar, err := goscalar.FromURL("https://cdn.jsdelivr.net/npm/@scalar/galaxy/dist/latest.json")
    if err != nil {
        log.Fatal(err)
    }

    http.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        if err := scalar.RenderDocs(w); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })

    log.Println("Server running at http://localhost:8080/docs")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Loading Methods

### 1. Local File

```go
// Relative path
scalar, err := goscalar.FromFile("./docs/api.json")

// Absolute path
scalar, err := goscalar.FromFile("/path/to/swagger.yaml")

// file:// URL
scalar, err := goscalar.FromFile("file:///absolute/path/to/spec.json")
```

### 2. HTTP/HTTPS URL

```go
// Public URL
scalar, err := goscalar.FromURL("https://api.example.com/swagger.json")

// With custom HTTP client
client := &http.Client{Timeout: 30 * time.Second}
scalar, err := goscalar.FromURL("https://api.example.com/swagger.json", goscalar.WithHTTPClient(client))
```

### 3. Direct Content

```go
// JSON as string
jsonSpec := `{"openapi": "3.0.0", "info": {"title": "API", "version": "1.0.0"}}`
scalar, err := goscalar.FromContent(jsonSpec)
```

### 4. Swag Spec

```go
// Using with swag
import "github.com/swaggo/swag"

spec := &swag.Spec{
    // specification configuration
}
scalar, err := goscalar.FromSpec(spec)
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithTitle(string)` | Sets the documentation title | "Scalar API Reference" |
| `WithLanguage(string)` | Sets the interface language | "en-US" |
| `WithFile(string)` | Loads spec from file | - |
| `WithURL(string)` | Loads spec from URL | - |
| `WithSpec(*swag.Spec)` | Loads spec from swag | - |
| `WithSpecContent(string)` | Loads spec from string | - |
| `WithHTTPClient(*http.Client)` | Custom HTTP client | 30s timeout |

## Error Handling

The package defines specific errors that can be checked:

```go
scalar, err := goscalar.FromFile("nonexistent.json")
if err != nil {
    switch {
    case errors.Is(err, goscalar.ErrInvalidTitle):
        // Invalid title
    case errors.Is(err, goscalar.ErrInvalidSpec):
        // Invalid specification
    case errors.Is(err, goscalar.ErrSpecRequired):
        // Specification required
    case errors.Is(err, goscalar.ErrUnsupportedScheme):
        // Unsupported URL scheme
    default:
        // Other error
    }
}
```
## Requirements

- Go 1.18+
- Valid OpenAPI/Swagger specification

## Contributing

Contributions are welcome! Please open an issue or pull request.

## License

This project is licensed under the MIT License. See the LICENSE file for details.