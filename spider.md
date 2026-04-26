# Spider Crawler Toolkit

A lightweight Go crawler toolkit providing HTTP client, HTML parsing and content extraction.

## Table of Contents

- [Quick Start](#quick-start)
- [Engine (HTTP Client)](#engine-http-client)
- [ResponseBody (Response Handling)](#responsebody-response-handling)
- [HTML Parsing](#html-parsing)
- [QueryNode (Node Query)](#querynode-node-query)
- [HtmNode Methods](#htmnode-methods)

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/nitai2plipala/ning2/spider"
)

func main() {
    // Create crawler engine
    engine := spider.New(map[string]string{
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    })
    
    // Send request
    resp, err := engine.Request("GET", "https://example.com", nil)
    if err != nil {
        panic(err)
    }
    
    // Read response body
    body, err := spider.ResponseBody(resp)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(string(body))
}
```

## Engine (HTTP Client)

### Creating Engine

```go
// Basic creation
engine := spider.New(nil)

// Custom headers
engine := spider.New(map[string]string{
    "User-Agent": "MyBot/1.0",
    "Accept":     "text/html",
})
```

### Engine Fields

```go
type Engine struct {
    HttpClient *http.Client  // HTTP client
    Header     map[string]string  // Default headers
    BodyCode   string       // Content-Type: Form/Json
}
```

### Sending Requests

```go
// Basic request
resp, err := engine.Request("GET", url, nil)
resp, err := engine.Request("POST", url, strings.NewReader("body"))

// Request with Context (supports timeout/cancel)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
resp, err := engine.RequestWithContext(ctx, "GET", url, nil)

// Set Content-Type
engine.BodyCode = "Form"    // application/x-www-form-urlencoded
engine.BodyCode = "Json"    // application/json
```

### Request Methods

```go
engine.Request("GET", url, nil)
engine.Request("POST", url, body)
engine.Request("PUT", url, body)
engine.Request("DELETE", url, nil)
```

### Features

- **Auto Cookie Management**: Uses cookiejar for persistent cookies
- **No Redirect Follow**: Does not follow HTTP redirects by default
- **Default Timeout**: 30 second timeout
- **Reusable Response Body**: Automatically replaced with NopCloser after reading

## ResponseBody (Response Handling)

### Basic Usage

```go
// Read response body
body, err := spider.ResponseBody(resp)
if err != nil {
    // Handle error
}

// Read as string
str := spider.ResponseBodyString(resp)
```

### Content Type Handling

```go
// Supported types: text/*, application/json, application/xml
// Unsupported types (e.g., images): Returns ErrUnsupportedContentType error
```

### Response Body Reuse

```go
// ResponseBody automatically replaces Body with reusable NopCloser
body1, _ := spider.ResponseBody(resp)
body2, _ := spider.ResponseBody(resp) // Still usable
```

### Size Limit

- Maximum 10MB, excess is truncated

### Error Types

```go
var ErrUnsupportedContentType = errors.New("unsupported content type")
```

## HTML Parsing

### Parsing HTML

```go
// Parse from Reader
root, err := spider.ParseHtml(strings.NewReader(htmlString))

// Parse from response
resp, _ := engine.Request("GET", url, nil)
body, _ := spider.ResponseBody(resp)
root, err := spider.ParseHtml(strings.NewReader(string(body)))
```

### HtmNode Structure

```go
type HtmNode struct {
    Parent      *HtmNode
    FirstChild  *HtmNode
    LastChild   *HtmNode
    PrevSibling *HtmNode
    NextSibling *HtmNode
    
    Type        html.NodeType  // Element, Text, Comment, etc.
    DataAtom    atom.Atom
    Data        string         // Tag name or text content
    Namespace   string
    Attr        []html.Attribute
}
```

## QueryNode (Node Query)

### Query Syntax

Supports CSS-like selector syntax:

| Syntax | Description | Example |
|--------|-------------|---------|
| `tag` | Tag name | `div`, `span`, `a` |
| `#id` | ID selector | `#main`, `#header` |
| `.class` | Class selector | `.container`, `.active` |
| `[attr]` | Has attribute | `[href]`, `[type]` |
| `[attr=value]` | Attribute equals | `[type=text]`, `[href="/link"]` |
| `>` | Child selector | `div > span` |
| ` ` (space) | Descendant selector | `div span` |
| `+` | Sibling selector | `li + li` |

### Basic Query

```go
// Query all div elements
nodes := root.QueryNode([]string{"div"})

// Query element with id main
nodes := root.QueryNode([]string{"#main"})

// Query elements with class active
nodes := root.QueryNode([]string{".active"})
```

### Attribute Query

```go
// Query a tags with href attribute
nodes := root.QueryNode([]string{"a", "[href]"})

// Query input with type=text
nodes := root.QueryNode([]string{"input", "[type=text]"})
```

### Combined Query

```go
// Descendant query: all spans under div
nodes := root.QueryNode([]string{"div", "span"})

// Child query: direct child spans of div
nodes := root.QueryNode([]string{"div", ">", "span"})

// Combined: a.link under div.container
nodes := root.QueryNode([]string{"div", ".container", "a", ".link"})
```

### Nested Query

```go
// Multiple queries
divs := root.QueryNode([]string{"div"})
for _, div := range divs {
    spans := div.QueryNode([]string{"span"})
}
```

## HtmNode Methods

### Attribute

Get element attribute value:

```go
href := node.Attribute("href")
title := node.Attribute("title")
// Non-existent attributes return empty string
```

### HasClass

Check if element has a class:

```go
if node.HasClass("active") {
    // Handle
}
```

## Complete Examples

### Crawl Page and Extract Links

```go
package main

import (
    "fmt"
    "github.com/nitai2plipala/ning2/spider"
    "strings"
)

func main() {
    engine := spider.New(map[string]string{
        "User-Agent": "Mozilla/5.0 (compatible; MyBot/1.0)",
    })
    
    // Get page
    resp, err := engine.Request("GET", "https://example.com", nil)
    if err != nil {
        panic(err)
    }
    
    body, err := spider.ResponseBody(resp)
    if err != nil {
        panic(err)
    }
    
    // Parse HTML
    root, err := spider.ParseHtml(strings.NewReader(string(body)))
    if err != nil {
        panic(err)
    }
    
    // Extract all links
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

### Extract Specific Content

```go
// Extract all paragraphs under div with class article
root, _ := spider.ParseHtml(strings.NewReader(html))
articles := root.QueryNode([]string{".article", "p"})

for _, p := range articles {
    if p.FirstChild != nil {
        fmt.Println(p.FirstChild.Data)
    }
}
```

## Performance Tips

1. **Reuse Engine**: Create Engine once and reuse
2. **Set Timeout**: Use `RequestWithContext` to avoid hanging requests
3. **Limit Response Size**: ResponseBody defaults to 10MB limit
4. **Reuse Response Body**: Can be read multiple times