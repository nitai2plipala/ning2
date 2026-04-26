# Ning2

A lightweight Go Web framework + Crawler toolkit

## Features

- **Web Framework** - High-performance routing, Middleware support, Context utilities, Gzip compression
- **Crawler Toolkit** - HTTP client, HTML parsing, Cookie management
- **User-Agent Parsing** - Browser/OS/Device/Bot detection

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/nitai2plipala/ning2"
)

func main() {
    mux := ning2.NewMux()
    
    // Enable Gzip compression
    mux.Use(ning2.CompressionMiddleware)
    
    mux.Handle("/hello", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        return c.String(200, "Hello, Ning2!")
    }, "GET")
    
    http.ListenAndServe(":8080", mux)
}
```

## Modules

| Module | Description |
|--------|-------------|
| `ning2` | Web framework core |
| `ning2/spider` | Crawler toolkit |

## Documentation

- [Ning2 Framework Docs](./ning2.md)
- [Spider Crawler Docs](./spider.md)

## Testing

```bash
go test ./... -v
```

## License

Apache License 2.0 - see [LICENSE](./LICENSE) file