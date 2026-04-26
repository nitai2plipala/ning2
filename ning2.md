# Ning2 Web 框架

轻量级 Go Web 框架，提供高性能路由、Middleware 支持和丰富的 Context 工具。

## 目录

- [快速开始](#快速开始)
- [路由](#路由)
- [Context](#context)
- [Middleware](#middleware)
- [User-Agent](#user-agent)
- [响应方法](#响应方法)
- [模板渲染](#模板渲染)

## 快速开始

```go
package main

import (
    "net/http"
    "ning2"
)

func main() {
    mux := ning2.NewMux()
    
    mux.Handle("/hello", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        return c.String(200, "Hello, Ning2!")
    }, "GET")
    
    http.ListenAndServe(":8080", mux)
}
```

## 路由

### 基本路由

```go
mux := ning2.NewMux()

// GET 路由
mux.Handle("/users", handler, "GET")

// POST 路由
mux.Handle("/users", handler, "POST")

// 多方法路由
mux.Handle("/api", handler, "GET", "POST", "PUT", "DELETE")

// 根路由
mux.RootPath(handler, "GET")
```

### 路由参数

```go
// 参数路由 :id
mux.Handle("/users/:id", handler, "GET")
// 访问 /users/123 -> c.Param["id"] = "123"

// 多参数
mux.Handle("/users/:userId/posts/:postId", handler, "GET")
```

### 通配符路由

```go
// 匹配剩余所有路径
mux.Handle("/api/*path", handler, "GET")
// 访问 /api/v1/users -> c.Param["path"] = "v1/users"
```

### 正则路由

```go
// 正则匹配 <name>pattern
mux.Handle("/items/<id>[0-9]+", handler, "GET")
// 匹配 /items/123 但不匹配 /items/abc
```

### 静态别名路由

```go
// 静态路由 /?alias
mux.Handle("/static/?file", handler, "GET")
// 访问 /static/js/app.js -> c.Param["file"] = "js/app.js"
```

### 路由优先级

1. 精确路由 (priority: 10)
2. Default 类型 (priority: 5)
3. Regexp 类型 (priority: 4)
4. Param 类型 (priority: 3)
5. Static 类型 (priority: 2)
6. Whole 类型 (priority: 1)

## Context

### 创建 Context

```go
// 通过 Mux 自动创建
mux.ServeHTTP(w, r)
// 在 handler 中使用 c *ning2.Context
```

### 获取参数

```go
// URL Query 参数
name := c.QueryParam("name", "url", 0)
names := c.QueryParams("name", "url")

// Form 参数
name := c.QueryParam("name", "form", 0)

// JSON 参数
name := c.QueryParam("name", "json", 0)
```

### 客户端信息

```go
// 获取客户端 IP
ip := c.ClientIP()

// 获取协议
scheme := c.Scheme() // "http" 或 "https"

// 获取完整 URL
url := c.WebSite()

// 客户端信息 (通过 User-Agent 解析)
c.Client.Device    // 设备类型: iPhone, iPad, Android, Desktop
c.Client.Browser.Name   // 浏览器: Chrome, Firefox, Safari, Edge, Opera
c.Client.OS.Name   // 操作系统: Windows, macOS, iOS, Android, Linux
c.Client.Robot     // 是否机器人
c.Client.Standard  // 客户端标准: .m (移动), .t (平板), .p (桌面)
```

## Middleware

### 创建 Middleware

```go
m := ning2.NewMidWare()

// 添加前置中间件 (请求前执行)
m.UseBefore(func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
    // 处理请求前逻辑
    return nil
})

// 添加后置中间件 (响应后执行)
m.UseAfter(func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
    // 处理响应后逻辑
    return nil
})

// 添加通用中间件
m.Use(func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
    return nil
})
```

### 使用 Middleware

```go
handler := mux.FindHandle(req, c)
wrappedHandler := m.Handler

// 执行
_ = wrappedHandler(w, req, c, handler)
```

### 内置 Middleware

```go
// 日志中间件
ning2.LoggerMiddleware

// 恢复中间件 (防止 panic)
ning2.RecoveryMiddleware
```

## User-Agent

### 解析 User-Agent

```go
ua := ning2.Parse(r.UserAgent())

// 浏览器信息
ua.Name     // Chrome, Firefox, Safari, Edge, Opera, Unknown
ua.Version  // 版本号

// 操作系统
ua.OS       // Windows, macOS, iOS, Android, Linux, Unknown
ua.OSVersion

// 设备类型
ua.Device   // iPhone, iPad, Android, Windows Phone, Desktop
ua.Mobile   // 是否移动设备
ua.Tablet   // 是否平板
ua.Desktop  // 是否桌面设备

// 机器人检测
ua.Bot      // 是否机器人 (Googlebot, Bingbot, curl, wget 等)
```

## 响应方法

### 基本响应

```go
// 返回字符串
c.String(200, "Hello")

// 返回 JSON
c.JSON(200, map[string]string{"key": "value"})

// 返回 HTML
c.HTML(200, "<html><body>Hello</body></html>")

// 无内容响应
c.NoContent(204)

// 重定向
c.Redirect(302, "/new-location")
```

### 错误响应

```go
// 返回错误 (状态码 400-600)
c.Error(500, nil)
c.Error(400, errors.New("bad request"))
```

## 模板渲染

### 基本渲染

```go
// 渲染模板文件
c.Render(200, "templates/index.html", data)

// 渲染多个模板文件
c.RenderTemplate(200, "layout.html", data, []string{"header.html", "footer.html"})
```

### 响应式模板

```go
// 根据客户端类型自动选择模板
// 移动端: index.m.html
// 平板端: index.t.html
// 桌面端: index.p.html
c.HtmlGlob(200, "templates/index.html", data)
```

### 模板函数

框架内置以下模板函数：

```go
// 类型转换
{{Uint .Value}}  // 转换为 uint
{{String .Value}} // 转换为字符串

// 算术运算
{{Add "uint" 1 2 3}}  // uint 相加
{{Add "int" 1 2 3}}   // int 相加
```

## 辅助函数

### 静态文件服务

```go
// 简单静态文件服务
mux.Handle("/static/*filepath", ning2.StripPrefix("/static", "./static"), "GET")

// 或使用标准库
mux.Handle("/files/*filepath", http.FileServer(http.Dir("./files")), "GET")
```

### HTTPS 重定向

```go
// HTTP 到 HTTPS 重定向
ning2.ListenServeAndToHttps(":80")
```

### 404 处理

```go
// 默认 404 处理器
ning2.NotFound(w, r, c)
```

## 类型定义

### HandleFunc

```go
type HandleFunc func(http.ResponseWriter, *http.Request, *Context) error
```

### Context

```go
type Context struct {
    request        *http.Request
    responseWriter http.ResponseWriter
    Pattern        string           // 匹配的路由模式
    Param          map[string]string // 路由参数
    Client         Client           // 客户端信息
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

## 性能优化

- **路由缓存**: 精确路由使用 O(1) 缓存查找
- **正则缓存**: 正则表达式编译结果缓存
- **Context 池**: 使用 sync.Pool 复用 Context 对象
- **Body 缓存**: JSON 请求体只读取一次
- **FuncMap 缓存**: 模板函数映射全局只创建一次