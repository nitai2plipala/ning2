package ning2

import (
	"net/http"
	"sync"
)

type Context struct {

	request  *http.Request

	responseWriter  http.ResponseWriter

	Pattern string

	Param map[string]string

	Client  Client

}

type RouteMux struct {
	treeNode      *Node
	syncPool      sync.Pool
	staticRoutes  map[string]map[string]HandleFunc // 精确路由缓存: path -> method -> handler
}


type HandleFunc func(http.ResponseWriter, *http.Request, *Context) error


type H map[string]interface{}


type MidWare struct {
	Before    []HandleFunc
	Middle    []HandleFunc
	After     []HandleFunc
}


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

	NodeType   string  //root=>0 whole=>1 static=>2 param=>3 regexp=>4 default=>5

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




