package ning2

import (
	"errors"
	"regexp"
	"strings"
)

// 预编译的正则表达式缓存
var (
	paramRegex   = regexp.MustCompile(`^/:[^/]+`)
	regexpRegex  = regexp.MustCompile(`^/<[^/]+>[^/]+`)
	aliasRegex   = regexp.MustCompile(`^/<[^/]+>`)
	staticRegex  = regexp.MustCompile(`^/[^/]+`)
	defaultRegex = regexp.MustCompile(`/[^/]+`)
)

func (mux *RouteMux) Handle(pattern string, handle HandleFunc, methods ...string) error {
	if handle == nil {
		return errors.New("ning2: nil handleFunc")
	}

	if len(methods) == 0 {
		return errors.New("ning2: at least one method is required")
	}

	methodHandle := make(map[string]HandleFunc)
	for _, method := range methods {
		method = strings.ToUpper(method)
		methodHandle[method] = handle
	}

	if mux.treeNode == nil {
		mux.treeNode = &Node{
			pattern: "/",
			nType:   "root",
		}
	}

	mux.Register(pattern, methodHandle)
	return nil
}

func (mux *RouteMux) RootPath(handle HandleFunc, methods ...string) error {
	if handle == nil {
		return errors.New("ning2: nil handleFunc")
	}
	if len(methods) == 0 {
		return errors.New("ning2: at least one method is required")
	}
	return mux.Handle("/", handle, methods...)
}

// Resource 注册静态文件服务
// pattern: 路由前缀，如 "/static/" 或 "/"
// dirPath: 静态文件目录，如 "./public"
func (mux *RouteMux) Resource(pattern, dirPath string) error {

	if !strings.HasSuffix(pattern, "/") {
		return errors.New("ning2: Resources pattern can only end with '/'")
	}

	// 根路径特殊处理
	if pattern == "/" {
		return mux.Handle("/*filepath", StripPrefix("", dirPath), "GET")
	}

	fullPattern := pattern + "?static"
	return mux.Handle(fullPattern, StripPrefix(pattern, dirPath), "GET")
}

func (mux *RouteMux) Register(pattern string, methodHandle map[string]HandleFunc) error {
	if pattern == "" {
		return errors.New("ning2: pattern cannot be empty")
	}

	// 根路径处理
	if pattern == "/" {
		mux.treeNode.methodHandle = methodHandle
		// 缓存到 staticRoutes
		mux.addStaticRoute("/", methodHandle)
		return nil
	}

	// 精确路由缓存（只缓存 default 类型）
	isStaticRoute := !strings.Contains(pattern, "/:") && 
	                 !strings.Contains(pattern, "/<") && 
	                 !strings.HasPrefix(pattern, "/*")

	children := make([]*Node, 0)

	for pattern != "" {
		node := new(Node)

		switch {
		case strings.HasPrefix(pattern, "/:"):
			// param: /:id
			path := paramRegex.FindString(pattern)
			node.pattern = path
			node.nType = "param"
			node.alias = path[2:]
			pattern = strings.TrimPrefix(pattern, path)
			isStaticRoute = false

		case strings.HasPrefix(pattern, "/<"):
			// regexp: /<name>pattern
			path := regexpRegex.FindString(pattern)
			node.pattern = path
			node.nType = "regexp"
			alias := aliasRegex.FindString(path)
			node.alias = alias[2 : len(alias)-1]
			node.regexp = strings.TrimPrefix(path, alias)
			pattern = strings.TrimPrefix(pattern, path)
			isStaticRoute = false

		case strings.HasPrefix(pattern, "/?"):
			// static: /?alias
			node.pattern = pattern
			node.nType = "static"
			node.alias = pattern[2:]
			pattern = "" // static 匹配整个剩余路径

		case strings.HasPrefix(pattern, "/*"):
			// whole: /*wildcard
			node.pattern = pattern
			node.nType = "whole"
			node.alias = pattern[2:] // 提取通配符名称，如 "path"
			pattern = "" // whole 匹配整个剩余路径
			isStaticRoute = false

		default:
			// default: 普通路径
			path := defaultRegex.FindString(pattern)
			node.pattern = path
			node.nType = "default"
			pattern = strings.TrimPrefix(pattern, path)
		}

		if pattern == "" {
			node.children = children
			node.methodHandle = methodHandle
			mux.treeNode.children = append(mux.treeNode.children, node)
			
			// 缓存精确路由
			if isStaticRoute {
				mux.addStaticRoute(node.pattern, methodHandle)
			}
		} else {
			children = append(children, node)
		}
	}
	
	return nil
}

// 添加到静态路由缓存
func (mux *RouteMux) addStaticRoute(path string, methodHandle map[string]HandleFunc) {
	if mux.staticRoutes == nil {
		mux.staticRoutes = make(map[string]map[string]HandleFunc)
	}
	mux.staticRoutes[path] = methodHandle
}
