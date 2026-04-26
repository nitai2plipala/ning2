# Spider 爬虫工具

轻量级 Go 爬虫工具库，提供 HTTP 客户端、HTML 解析和内容提取功能。

## 目录

- [快速开始](#快速开始)
- [Engine (HTTP 客户端)](#engine-http-客户端)
- [ResponseBody (响应处理)](#responsebody-响应处理)
- [HTML 解析](#html-解析)
- [QueryNode (节点查询)](#querynode-节点查询)
- [HtmNode 方法](#htmnode-方法)

## 快速开始

```go
package main

import (
    "fmt"
    "ning2/spider"
)

func main() {
    // 创建爬虫引擎
    engine := spider.New(map[string]string{
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    })
    
    // 发送请求
    resp, err := engine.Request("GET", "https://example.com", nil)
    if err != nil {
        panic(err)
    }
    
    // 读取响应体
    body, err := spider.ResponseBody(resp)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(string(body))
}
```

## Engine (HTTP 客户端)

### 创建 Engine

```go
// 基础创建
engine := spider.New(nil)

// 自定义请求头
engine := spider.New(map[string]string{
    "User-Agent": "MyBot/1.0",
    "Accept":     "text/html",
})
```

### Engine 字段

```go
type Engine struct {
    HttpClient *http.Client  // HTTP 客户端
    Header     map[string]string  // 默认请求头
    BodyCode   string       // Content-Type: Form/Json
}
```

### 发送请求

```go
// 基础请求
resp, err := engine.Request("GET", url, nil)
resp, err := engine.Request("POST", url, strings.NewReader("body"))

// 带 Context 的请求 (支持超时/取消)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
resp, err := engine.RequestWithContext(ctx, "GET", url, nil)

// 设置 Content-Type
engine.BodyCode = "Form"    // application/x-www-form-urlencoded
engine.BodyCode = "Json"    // application/json
```

### 请求方法

```go
engine.Request("GET", url, nil)
engine.Request("POST", url, body)
engine.Request("PUT", url, body)
engine.Request("DELETE", url, nil)
```

### 特性

- **自动 Cookie 管理**: 使用 cookiejar 持久化 Cookie
- **不跟随重定向**: 默认不跟随 HTTP 重定向
- **默认超时**: 30 秒超时
- **可复用的响应体**: 读取后自动替换为 NopCloser

## ResponseBody (响应处理)

### 基本用法

```go
// 读取响应体
body, err := spider.ResponseBody(resp)
if err != nil {
    // 处理错误
}

// 读取为字符串
str := spider.ResponseBodyString(resp)
```

### 内容类型处理

```go
// 支持的类型: text/*, application/json, application/xml
// 不支持的类型 (如图片): 返回 ErrUnsupportedContentType 错误
```

### 响应体复用

```go
// ResponseBody 会自动将 Body 替换为可重复读取的 NopCloser
body1, _ := spider.ResponseBody(resp)
body2, _ := spider.ResponseBody(resp) // 仍然可用
```

### 大小限制

- 最大读取 10MB，超出部分被截断

### 错误类型

```go
var ErrUnsupportedContentType = errors.New("unsupported content type")
```

## HTML 解析

### 解析 HTML

```go
// 从 Reader 解析
root, err := spider.ParseHtml(strings.NewReader(htmlString))

// 从响应解析
resp, _ := engine.Request("GET", url, nil)
body, _ := spider.ResponseBody(resp)
root, err := spider.ParseHtml(strings.NewReader(string(body)))
```

### HtmNode 结构

```go
type HtmNode struct {
    Parent      *HtmNode
    FirstChild  *HtmNode
    LastChild   *HtmNode
    PrevSibling *HtmNode
    NextSibling *HtmNode
    
    Type        html.NodeType  // Element, Text, Comment 等
    DataAtom    atom.Atom
    Data        string         // 标签名或文本内容
    Namespace   string
    Attr        []html.Attribute
}
```

## QueryNode (节点查询)

### 查询语法

支持类似 CSS 选择器的查询语法：

| 语法 | 说明 | 示例 |
|------|------|------|
| `tag` | 标签名 | `div`, `span`, `a` |
| `#id` | ID 选择器 | `#main`, `#header` |
| `.class` | Class 选择器 | `.container`, `.active` |
| `[attr]` | 有此属性 | `[href]`, `[type]` |
| `[attr=value]` | 属性等于 | `[type=text]`, `[href="/link"]` |
| `>` | 子代选择器 | `div > span` |
| ` ` (空格) | 后代选择器 | `div span` |
| `+` | 兄弟选择器 | `li + li` |

### 基本查询

```go
// 查询所有 div 元素
nodes := root.QueryNode([]string{"div"})

// 查询 id 为 main 的元素
nodes := root.QueryNode([]string{"#main"})

// 查询 class 包含 active 的元素
nodes := root.QueryNode([]string{".active"})
```

### 属性查询

```go
// 查询有 href 属性的 a 标签
nodes := root.QueryNode([]string{"a", "[href]"})

// 查询 type 属性为 text 的 input
nodes := root.QueryNode([]string{"input", "[type=text]"})
```

### 组合查询

```go
// 后代查询: div 下的所有 span
nodes := root.QueryNode([]string{"div", "span"})

// 子代查询: div 的直接子元素 span
nodes := root.QueryNode([]string{"div", ">", "span"})

// 组合: div.container 下的 a.link
nodes := root.QueryNode([]string{"div", ".container", "a", ".link"})
```

### 嵌套查询

```go
// 多次查询
divs := root.QueryNode([]string{"div"})
for _, div := range divs {
    spans := div.QueryNode([]string{"span"})
}
```

## HtmNode 方法

### Attribute

获取元素属性值：

```go
href := node.Attribute("href")
title := node.Attribute("title")
// 不存在的属性返回空字符串
```

### HasClass

检查元素是否包含某个 class：

```go
if node.HasClass("active") {
    // 处理
}
```

## 完整示例

### 爬取网页并提取链接

```go
package main

import (
    "fmt"
    "ning2/spider"
    "strings"
)

func main() {
    engine := spider.New(map[string]string{
        "User-Agent": "Mozilla/5.0 (compatible; MyBot/1.0)",
    })
    
    // 获取页面
    resp, err := engine.Request("GET", "https://example.com", nil)
    if err != nil {
        panic(err)
    }
    
    body, err := spider.ResponseBody(resp)
    if err != nil {
        panic(err)
    }
    
    // 解析 HTML
    root, err := spider.ParseHtml(strings.NewReader(string(body)))
    if err != nil {
        panic(err)
    }
    
    // 提取所有链接
    links := root.QueryNode([]string{"a", "[href]"})
    for _, link := range links {
        href := link.Attribute("href")
        text := ""
        if link.FirstChild != nil {
            text = link.FirstChild.Data
        }
        fmt.Printf("Link: %s - %s\n", href, text)
    }
}
```

### 提取特定内容

```go
// 提取 class 为 article 的 div 下的所有段落
root, _ := spider.ParseHtml(strings.NewReader(html))
articles := root.QueryNode([]string{".article", "p"})

for _, p := range articles {
    if p.FirstChild != nil {
        fmt.Println(p.FirstChild.Data)
    }
}
```

## 性能提示

1. **复用 Engine**: 创建一次 Engine 并重复使用
2. **设置超时**: 使用 `RequestWithContext` 避免请求卡死
3. **限制响应大小**: ResponseBody 默认限制 10MB
4. **复用响应体**: 读取后仍可再次读取