package ning2

import (
	"io"
	"net/http"
	"sync"
)

type Context struct {

	request  *http.Request

	responseWriter  http.ResponseWriter

	Pattern string

	Param map[string]string

	Client  Client

	// body 缓存当前请求的请求体原始字节（per-request，请求结束即释放）
	// 避免 r.Body 流只能读一次的问题，供 Bind 复用
	body []byte
}

type RouteMux struct {
	treeNode      *Node
	syncPool      sync.Pool
	staticRoutes  map[string]map[string]HandleFunc // 精确路由缓存: path -> method -> handler
	middlewares   []Middleware                     // 全局中间件（对所有路由生效）
}


// HandleFunc 业务处理函数签名（用户 handler 和中间件 next 的统一签名）
type HandleFunc func(http.ResponseWriter, *http.Request, *Context) error

// Middleware 中间件类型：接收 next handler，返回包装后的 handler
// 洋葱模型：每个中间件可在 next 调用前后执行逻辑，也可不调 next 中断链路
type Middleware func(next HandleFunc) HandleFunc

// MiddlewareGroup 路由级中间件组，由 mux.Use() 返回
// 通过 .Handle() 注册的路由会自动带上组内中间件
type MiddlewareGroup struct {
	mux         *RouteMux
	middlewares []Middleware
}


type H map[string]interface{}


//HTTP METHOD  "GET", "POST", "PUT", "DELETE", "HEAD", "PATCH", "OPTIONS", "CONNECT", "TRACE"


type (

	Node struct {

		pattern string //注册路由

		regexp  string //正则表达式

		alias  string //模拟匹配别名

		nType  NodeType //路径类型

		children  Children //路径子节点

		methodHandle  map[string]HandleFunc

	}

	NodeType   string  // 节点类型: root(0) whole(1) static(2) param(3) regexp(4) default(5)

	Children  []*Node

)

type Arbiter struct {
	priority  uint
	pattern   string
	mimicry   map[string]string
	handleFunc  HandleFunc
}

type Template struct {
	Name string
	Data interface{}
	Writer http.ResponseWriter
	ClientStandard string
}

type Client struct {
	OS  struct{
		Name   string
		Version  string
	}
	Browser  struct{
		Name   string
		Version  string
	}
	URL       string
	Robot     bool
	Device    string
	UserAgent  string
	Standard   string
}

// ResponseWriter 包装 http.ResponseWriter，支持状态码和 Header 追踪
type ResponseWriter struct {
	http.ResponseWriter
	Code    int
	Written bool
}

// NewResponseWriter 创建新的 ResponseWriter 包装器
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		Code:           200,
		Written:        false,
	}
}

// WriteHeader 追踪状态码
func (r *ResponseWriter) WriteHeader(code int) {
	r.Code = code
	r.Written = true
	r.ResponseWriter.WriteHeader(code)
}

// Write 追踪写入状态
func (r *ResponseWriter) Write(p []byte) (int, error) {
	if !r.Written {
		r.WriteHeader(200)
	}
	return r.ResponseWriter.Write(p)
}

// CompressWriter 压缩响应包装器
type CompressWriter struct {
	io.Writer
	http.ResponseWriter
}

// Write 写入压缩数据
func (c *CompressWriter) Write(p []byte) (int, error) {
	return c.Writer.Write(p)
}

// WriteHeader 设置响应头
func (c *CompressWriter) WriteHeader(code int) {
	c.ResponseWriter.Header().Del("Content-Length")
	c.ResponseWriter.WriteHeader(code)
}




