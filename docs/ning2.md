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
mux.Handle("/api/*", handler, "GET")
// Access /api/v1/users -> c.Param["path"] = "v1/users"
// "path" is the internal fixed key name, hidden from users
```

### Regexp Routes

```go
// Regexp match <name>pattern
mux.Handle("/items/<id>[0-9]+", handler, "GET")
// Matches /items/123 but not /items/abc
```

### Static Alias Routes

```go
// Static route /?alias - matches the entire remaining path
mux.Handle("/static/?file", handler, "GET")
// Access /static/js/app.js -> c.Param["file"] = "js/app.js"
//
// Note: static routes have priority 2 (lower than param/regexp/default)
// so they are typically used for catch-all alias scenarios
```

### Static File Serving

```go
// Use Resource (recommended) - two params: pattern and dirPath
mux.Resource("/static/", "./static")    // /static/* -> ./static/*
mux.Resource("/", "./public")           // /* -> ./public/*

// Or use StripPrefix directly
mux.Handle("/static/*", ning2.StripPrefix("/static", "./static"), "GET")
mux.Handle("/*", ning2.StripPrefix("", "./public"), "GET")
```

### Route Priority

When multiple routes match the same URL, the router selects the handler with the highest priority:

| Priority | Type | Value | Description |
|----------|------|-------|-------------|
| 1 (highest) | `default` | 5 | Exact path segment, e.g. `/users` |
| 2 | `regexp` | 4 | Regex match segment, e.g. `/<id>[0-9]+` |
| 3 | `param` | 3 | Named parameter, e.g. `/:id` |
| 4 | `static` | 2 | Alias for remaining path, e.g. `/?file` |
| 5 (lowest) | `whole` | 1 | Wildcard for remaining path, e.g. `/*` |

**Example**: Given routes `/files/*` and `/files/static`, a request to `/files/static` matches the `default` route (priority 5) over the `whole` route (priority 1).

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
name := c.Query("name")

// JSON body binding (strongly typed)
var req struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Age      int    `json:"age"`
    Tags     []string `json:"tags"`
}
if err := c.Bind(&req); err != nil {
    return c.Error(400, err)
}
// req.Username, req.Password, req.Age, req.Tags

// Form parameters (use standard library)
r.ParseForm()
username := r.PostForm.Get("username")
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

Ning2 采用洋葱模型（Onion Model）中间件，`HandleFunc` 签名不变，中间件通过包装函数 `Middleware` 实现。

### Middleware Type

```go
// Middleware 包装一个 handler，返回新的 handler
type Middleware func(next HandleFunc) HandleFunc
```

每个中间件可在 `next` 调用前后执行逻辑，也可不调 `next` 中断链路。

### Global Middleware

`mux.Middleware(mw...)` 注册全局中间件，对所有路由生效：

```go
// 全局中间件，执行顺序：书写顺序从外到内
mux.Middleware(ning2.Recovery, ning2.Logger, ning2.RequestID)
```

### Route-Scoped Middleware

`mux.Use(mw...)` 返回 `*MiddlewareGroup`，通过 `.Handle()` 注册的路由带上组内中间件，只对选中路由生效：

```go
// 单个路由级中间件
mux.Use(Auth).Handle("/admin", handler, "GET")

// 多个路由级中间件（书写顺序 = 从外到内执行顺序）
mux.Use(Auth, RateLimit).Handle("/api/secret", handler, "GET")

// 分组复用：同一组中间件应用到多条路由
adminGroup := mux.Use(Auth, RateLimit)
adminGroup.Handle("/admin/users", handler1, "GET")
adminGroup.Handle("/admin/settings", handler2, "GET")

// group 继续叠加中间件
superAdmin := adminGroup.Use(AdminOnly)
superAdmin.Handle("/admin/super", handler3, "GET")
```

### Execution Order (Onion Model)

全局中间件在最外层，路由级中间件在中间，handler 在核心：

```
请求 → Recovery(外) → Logger → RequestID → Auth → RateLimit → handler
                                                          ↓
响应 ← Recovery(内) ← Logger ← RequestID ← Auth ← RateLimit ← 
```

### Writing Custom Middleware

```go
// 洋葱型：前置 + 后置
func Logger(next ning2.HandleFunc) ning2.HandleFunc {
    return func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        start := time.Now()
        // 前半
        err := next(w, r, c)  // 调用下一层
        // 后半
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
        return err
    }
}

// 门卫型：不调 next 中断
func Auth(next ning2.HandleFunc) ning2.HandleFunc {
    return func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        if r.Header.Get("Authorization") != "Bearer token" {
            http.Error(w, "Unauthorized", 401)
            return nil  // 不调 next，链路中断
        }
        return next(w, r, c)
    }
}

// 变换型：替换 writer
func Compression(next ning2.HandleFunc) ning2.HandleFunc {
    return func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        gw := gzip.NewWriter(w)
        c.responseWriter = &CompressWriter{Writer: gw, ResponseWriter: w}
        err := next(w, r, c)  // handler 写数据进 gw
        gw.Close()            // 后半：刷出压缩数据
        return err
    }
}
```

### Middleware Ordering Guide

| Layer | Type | Example | Position |
|-------|------|---------|----------|
| Outermost | Rescue | Recovery | Global |
| Outer | Observer | Logger, RequestID | Global |
| Middle | Guard | Auth, RateLimit, CORS | Route-scoped |
| Innermost | Transformer | Compression | Route-scoped, last |

**Key**: Transformer middleware (Compression) should be placed innermost (closest to handler), so error responses (e.g. 401 from Auth) are not needlessly compressed.

### Built-in Middleware

```go
ning2.Recovery     // 兑底型：捕获 panic
ning2.Logger        // 观察型：记录请求日志和耗时
ning2.RequestID     // 观察型：生成请求唯一ID
ning2.Compression   // 变换型：gzip 压缩（放路由级最内层）
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
mux.Handle("/static/*", ning2.StripPrefix("/static", "./static"), "GET")

// Or use standard library
mux.Handle("/files/*", http.FileServer(http.Dir("./files")), "GET")
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
- **Body Cache**: Request body read only once per request (cached in Context.body, auto-released with pool)
- **FuncMap Cache**: Template function map created globally only once