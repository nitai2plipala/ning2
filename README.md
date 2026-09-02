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
    
    // Global middleware (applies to all routes)
    mux.Middleware(ning2.Recovery, ning2.Logger, ning2.RequestID)
    
    // Basic route - GET
    mux.Handle("/hello", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        return c.String(200, "Hello, Ning2!")
    }, "GET")
    
    // Route with parameters - /user/123
    mux.Handle("/user/:id", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        userID := c.Param["id"]
        return c.JSON(200, map[string]string{"user_id": userID})
    }, "GET")
    
    // Multi-level parameters - /user/123/post/456
    mux.Handle("/user/:userId/post/:postId", handler, "GET")
    
    // Regexp route - only matches numbers
    mux.Handle("/items/<id>[0-9]+", handler, "GET")
    // Matches /items/123 but not /items/abc
    
    // Multiple HTTP methods
    mux.Handle("/api/data", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        return c.JSON(200, map[string]string{"status": "ok"})
    }, "GET", "POST", "PUT", "DELETE")
    
    // Wildcard route - /files/*
    mux.Handle("/files/*", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        return c.String(200, "File: "+c.Param["path"])
    }, "GET")
    
    // Query parameters
    mux.Handle("/search", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        query := c.Query("q")
        return c.JSON(200, map[string]string{"query": query})
    }, "GET")
    
    // JSON body binding
    mux.Handle("/login", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        var req struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }
        if err := c.Bind(&req); err != nil {
            return c.Error(400, err)
        }
        return c.JSON(200, map[string]string{"username": req.Username})
    }, "POST")
    
    // HTML response
    mux.Handle("/page", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        return c.HTML(200, "<html><body><h1>Hello World</h1></body></html>")
    }, "GET")
    
    // Route-scoped middleware (only applies to matched routes)
    mux.Use(AuthMiddleware).Handle("/admin", handler, "GET")
    // Compression should be placed innermost (closest to handler)
    mux.Use(AuthMiddleware, ning2.Compression).Handle("/api/data", handler, "GET")
    
    // Static file serving
    mux.Resource("/static/", "./public")  // /static/* -> ./public/*
    mux.Resource("/", "./public")         // root path static files
    
    http.ListenAndServe(":8080", mux)
}
```

### Core Features

| Feature | Description |
|---------|-------------|
| Routing | Prefix-shared tree, Static, parameter, wildcard, regexp routes with priority |
| Response | String, JSON, HTML, File, Redirect |
| Middleware | Onion model, Global + route-scoped middleware with priority |
| Compression | Gzip response compression |
| Context | Query (URL params), Bind (JSON body to struct) |

## Modules

| Module | Description |
|--------|-------------|
| `ning2` | Web framework core |
| `ning2/spider` | Crawler toolkit |

## Documentation

- [Ning2 Framework Docs](./docs/ning2.md)
- [Spider Crawler Docs](./docs/spider.md)

## Testing

```bash
go test ./... -v
```

## License

Apache License 2.0 - see [LICENSE](./LICENSE) file