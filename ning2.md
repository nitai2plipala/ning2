# Ning2 Web Framework

A lightweight Go Web framework providing high-performance routing, Middleware support, rich Context utilities, and Gzip compression.

## Table of Contents

- [Quick Start](#quick-start)
- [Routing](#routing)
- [Context](#context)
- [Middleware](#middleware)
- [User-Agent](#user-agent)
- [Response Methods](#response-methods)
- [Template Rendering](#template-rendering)

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

## Routing

### Basic Routing

```go
mux := ning2.NewMux()

// GET route
mux.Handle("/users", handler, "GET")

// POST route
mux.Handle("/users", handler, "POST")

// Multiple methods
mux.Handle("/api", handler, "GET", "POST", "PUT", "DELETE")

// Root path
mux.RootPath(handler, "GET")
```

### Route Parameters

```go
// Parameter route :id
mux.Handle("/users/:id", handler, "GET")
// Access /users/123 -> c.Param["id"] = "123"

// Multiple parameters
mux.Handle("/users/:userId/posts/:postId", handler, "GET")
```

### Wildcard Routes

```go
// Match remaining path
mux.Handle("/api/*path", handler, "GET")
// Access /api/v1/users -> c.Param["path"] = "v1/users"
```

### Regexp Routes

```go
// Regexp match <name>pattern
mux.Handle("/items/<id>[0-9]+", handler, "GET")
// Matches /items/123 but not /items/abc
```

### Static Alias Routes

```go
// Static route /?alias
mux.Handle("/static/?file", handler, "GET")
// Access /static/js/app.js -> c.Param["file"] = "js/app.js"
```

### Static File Serving

```go
// Use Resource (recommended) - two params: pattern and dirPath
mux.Resource("/static/", "./static")    // /static/* -> ./static/*
mux.Resource("/", "./public")           // /* -> ./public/*

// Or use StripPrefix directly
mux.Handle("/static/*filepath", ning2.StripPrefix("/static", "./static"), "GET")
mux.Handle("/*filepath", ning2.StripPrefix("", "./public"), "GET")
```

### Route Priority

1. Exact routes (priority: 10)
2. Default type (priority: 5)
3. Regexp type (priority: 4)
4. Param type (priority: 3)
5. Static type (priority: 2)
6. Whole type (priority: 1)

## Context

### Creating Context

```go
// Created automatically by Mux
mux.ServeHTTP(w, r)
// Use c *ning2.Context in handler
```

### Getting Parameters

```go
// URL Query parameters
name := c.QueryParam("name", "url", 0)
names := c.QueryParams("name", "url")

// Form parameters
name := c.QueryParam("name", "form", 0)

// JSON parameters
name := c.QueryParam("name", "json", 0)
```

### Client Information

```go
// Get client IP
ip := c.ClientIP()

// Get protocol
scheme := c.Scheme() // "http" or "https"

// Get full URL
url := c.WebSite()

// Client info (via User-Agent parsing)
c.Client.Device    // Device type: iPhone, iPad, Android, Desktop
c.Client.Browser.Name   // Browser: Chrome, Firefox, Safari, Edge, Opera
c.Client.OS.Name   // OS: Windows, macOS, iOS, Android, Linux
c.Client.Robot     // Is robot
c.Client.Standard  // Client standard: .m (mobile), .t (tablet), .p (desktop)
```

## Middleware

### Creating Middleware

```go
m := ning2.NewMidWare()

// Add before middleware (executes before request)
m.UseBefore(func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
    // Pre-request logic
    return nil
})

// Add after middleware (executes after response)
m.UseAfter(func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
    // Post-response logic
    return nil
})

// Add general middleware
m.Use(func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
    return nil
})
```

### Using Middleware

```go
// Add to mux
mux.Use(ning2.CompressionMiddleware)

// Or use MidWare chain
m := ning2.NewMidWare()
m.UseBefore(beforeFn)
m.UseAfter(afterFn)
```

### Built-in Middleware

```go
// Gzip compression (enable with mux.Use(ning2.CompressionMiddleware))
ning2.CompressionMiddleware

// Logger middleware
ning2.LoggerMiddleware

// Recovery middleware (prevents panic crash)
ning2.RecoveryMiddleware
```

## User-Agent

### Parsing User-Agent

```go
ua := ning2.Parse(r.UserAgent())

// Browser info
ua.Name     // Chrome, Firefox, Safari, Edge, Opera, Unknown
ua.Version  // Version

// Operating system
ua.OS       // Windows, macOS, iOS, Android, Linux, Unknown
ua.OSVersion

// Device type
ua.Device   // iPhone, iPad, Android, Windows Phone, Desktop
ua.Mobile   // Is mobile device
ua.Tablet   // Is tablet
ua.Desktop  // Is desktop device

// Bot detection
ua.Bot      // Is bot (Googlebot, Bingbot, curl, wget, etc.)
```

## Response Methods

### Basic Response

```go
// Return string
c.String(200, "Hello")

// Return JSON
c.JSON(200, map[string]string{"key": "value"})

// Return HTML
c.HTML(200, "<html><body>Hello</body></html>")

// No content response
c.NoContent(204)

// Redirect
c.Redirect(302, "/new-location")
```

### Error Response

```go
// Return error (status code 400-600)
c.Error(500, nil)
c.Error(400, errors.New("bad request"))
```

## Template Rendering

### Basic Rendering

```go
// Render template file
c.Render(200, "templates/index.html", data)

// Render multiple template files
c.RenderTemplate(200, "layout.html", data, []string{"header.html", "footer.html"})
```

### Responsive Templates

```go
// Auto-select template based on client type
// Mobile: index.m.html
// Tablet: index.t.html
// Desktop: index.p.html
c.HtmlGlob(200, "templates/index.html", data)
```

### Template Functions

Built-in template functions:

```go
// Type conversion
{{Uint .Value}}  // Convert to uint
{{String .Value}} // Convert to string

// Arithmetic
{{Add "uint" 1 2 3}}  // uint addition
{{Add "int" 1 2 3}}   // int addition
```

## Helper Functions

### Static File Serving

```go
// Simple static file serving
mux.Handle("/static/*filepath", ning2.StripPrefix("/static", "./static"), "GET")

// Or use standard library
mux.Handle("/files/*filepath", http.FileServer(http.Dir("./files")), "GET")
```

### HTTPS Redirect

```go
// HTTP to HTTPS redirect
ning2.ListenServeAndToHttps(":80")
```

### 404 Handler

```go
// Default 404 handler
ning2.NotFound(w, r, c)
```

## Type Definitions

### HandleFunc

```go
type HandleFunc func(http.ResponseWriter, *http.Request, *Context) error
```

### Context

```go
type Context struct {
    request        *http.Request
    responseWriter http.ResponseWriter
    Pattern        string           // Matched route pattern
    Param          map[string]string // Route parameters
    Client         Client           // Client information
}
```

### Client

```go
type Client struct {
    OS        struct{ Name, Version string }
    Browser   struct{ Name, Version string }
    URL       string
    Robot     bool
    Device    string
    UserAgent string
    Standard  string  // .m, .t, .p
}
```

## Performance Optimization

- **Route Cache**: Exact routes use O(1) cache lookup
- **Regexp Cache**: Regex expression compilation results cached
- **Context Pool**: Use sync.Pool to reuse Context objects
- **Body Cache**: JSON request body read only once
- **FuncMap Cache**: Template function map created globally only once