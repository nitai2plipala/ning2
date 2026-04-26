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
    
    // Basic route - GET
    mux.Handle("/hello", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        return c.String(200, "Hello, Ning2!")
    }, "GET")
    
    // Route with parameters - /user/123
    mux.Handle("/user/:id", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        userID := c.Param["id"]
        return c.JSON(200, map[string]string{"user_id": userID})
    }, "GET")
    
    // Multiple HTTP methods
    mux.Handle("/api/data", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        return c.JSON(200, map[string]string{"status": "ok"})
    }, "GET", "POST", "PUT", "DELETE")
    
    // Wildcard route - /files/*
    mux.Handle("/files/*path", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        return c.String(200, "File: "+c.Param["path"])
    }, "GET")
    
    // Query parameters
    mux.Handle("/search", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        query := c.QueryParam("q", "url", 0)
        return c.JSON(200, map[string]string{"query": query})
    }, "GET")
    
    // HTML response
    mux.Handle("/page", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        return c.HTML(200, "<html><body><h1>Hello World</h1></body></html>")
    }, "GET")
    
    // Static file serving
    mux.Resource("/static/", "./public")  // /static/* -> ./public/*
    mux.Resource("/", "./public")         // root path static files
    
    http.ListenAndServe(":8080", mux)
}
```

### Core Features

| Feature | Description |
|---------|-------------|
| Routing | Static, parameter, wildcard, regexp routes |
| Response | String, JSON, HTML, File, Redirect |
| Middleware | Before, After, Middle hooks |
| Compression | Gzip response compression |
| Context | Query, Form, JSON params |

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