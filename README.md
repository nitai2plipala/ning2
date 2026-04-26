# Ning2

轻量级 Go Web 框架 + 爬虫工具库

## 特性

- **Web 框架** - 高性能路由、Middleware 支持、Context 封装、Gzip 压缩
- **爬虫工具** - HTTP 客户端、HTML 解析、Cookie 管理
- **User-Agent 解析** - 浏览器/OS/设备/机器人检测

## 快速开始

```go
package main

import (
    "net/http"
    "github.com/nitai2plipala/ning2"
)

func main() {
    mux := ning2.NewMux()
    
    // 开启 Gzip 压缩
    mux.Use(ning2.CompressionMiddleware)
    
    mux.Handle("/hello", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
        return c.String(200, "Hello, Ning2!")
    }, "GET")
    
    http.ListenAndServe(":8080", mux)
}
```

## 模块

| 模块 | 说明 |
|------|------|
| `ning2` | Web 框架核心 |
| `ning2/spider` | 爬虫工具 |

## 文档

- [Ning2 框架文档](./ning2.md)
- [spider 爬虫文档](./spider.md)

## 测试

```bash
go test ./... -v
```

## 许可证

Apache License 2.0 - see [LICENSE](./LICENSE) file