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
	// if pattern == "/" {
	// 	return mux.Handle("/*", StripPrefix("", dirPath), "GET")
	// }

	fullPattern := pattern + "?static"
	return mux.Handle(fullPattern, StripPrefix(pattern, dirPath), "GET")
}

// segment 表示从 pattern 中拆分出的一段路径节点描述
type segment struct {
	pattern string // 节点 pattern，如 "/user"、"/:id"、"/<id>[0-9]+"、"/*"
	nType   NodeType
	alias   string
	regexp  string
}

// parsePattern 把完整 pattern 拆解成有序的 segment 列表
func parsePattern(pattern string) []segment {
	var segs []segment
	for pattern != "" {
		switch {
		case strings.HasPrefix(pattern, "/:"):
			path := paramRegex.FindString(pattern)
			segs = append(segs, segment{
				pattern: path,
				nType:   "param",
				alias:   path[2:],
			})
			pattern = strings.TrimPrefix(pattern, path)

		case strings.HasPrefix(pattern, "/<"):
			path := regexpRegex.FindString(pattern)
			alias := aliasRegex.FindString(path)
			segs = append(segs, segment{
				pattern: path,
				nType:   "regexp",
				alias:   alias[2 : len(alias)-1],
				regexp:  strings.TrimPrefix(path, alias),
			})
			pattern = strings.TrimPrefix(pattern, path)

		case strings.HasPrefix(pattern, "/?"):
			segs = append(segs, segment{
				pattern: pattern,
				nType:   "static",
				alias:   pattern[2:],
			})
			pattern = "" // static 匹配整个剩余路径

		case strings.HasPrefix(pattern, "/*"):
			segs = append(segs, segment{
				pattern: pattern,
				nType:   "whole",
				alias:   "path",
			})
			pattern = "" // whole 匹配整个剩余路径

		default:
			path := defaultRegex.FindString(pattern)
			segs = append(segs, segment{
				pattern: path,
				nType:   "default",
			})
			pattern = strings.TrimPrefix(pattern, path)
		}
	}
	return segs
}

// matchNode 判断一个 segment 是否与已存在的 Node 描述同一段路径（可复用）
func (s segment) matchNode(n *Node) bool {
	if s.nType != n.nType || s.pattern != n.pattern {
		return false
	}
	if s.nType == "regexp" && s.regexp != n.regexp {
		return false
	}
	return true
}

func (mux *RouteMux) Register(pattern string, methodHandle map[string]HandleFunc) error {
	if pattern == "" {
		return errors.New("ning2: pattern cannot be empty")
	}

	// 根路径处理
	if pattern == "/" {
		mux.treeNode.methodHandle = methodHandle
		mux.addStaticRoute("/", methodHandle)
		return nil
	}

	segs := parsePattern(pattern)

	// 精确路由缓存（只缓存 default 类型，且只有一段）
	isStaticRoute := true
	for _, s := range segs {
		if s.nType != "default" {
			isStaticRoute = false
			break
		}
	}

	// 逐段挂载到树中，相同前缀共享节点
	current := mux.treeNode
	for i, seg := range segs {
		// 在 current.children 中查找可复用的节点
		var matched *Node
		for _, child := range current.children {
			if seg.matchNode(child) {
				matched = child
				break
			}
		}

		isLast := i == len(segs)-1

		if matched != nil {
			// 复用已有节点
			if isLast {
				// 最后一段：合并 methodHandle
				if matched.methodHandle == nil {
					matched.methodHandle = make(map[string]HandleFunc)
				}
				for m, h := range methodHandle {
					matched.methodHandle[m] = h
				}
			}
			current = matched
		} else {
			// 创建新节点
			node := &Node{
				pattern: seg.pattern,
				nType:  seg.nType,
				alias:  seg.alias,
				regexp: seg.regexp,
			}
			if isLast {
				node.methodHandle = make(map[string]HandleFunc)
				for m, h := range methodHandle {
					node.methodHandle[m] = h
				}
			}
			current.children = append(current.children, node)
			current = node
		}
	}

	// 缓存精确路由（只缓存单段 default 路由）
	if isStaticRoute && len(segs) == 1 {
		mux.addStaticRoute(segs[0].pattern, methodHandle)
	}

	return nil
}

// 添加到静态路由缓存（按 method 合并，不覆盖已有 method）
func (mux *RouteMux) addStaticRoute(path string, methodHandle map[string]HandleFunc) {
	if mux.staticRoutes == nil {
		mux.staticRoutes = make(map[string]map[string]HandleFunc)
	}
	if mux.staticRoutes[path] == nil {
		mux.staticRoutes[path] = make(map[string]HandleFunc)
	}
	for m, h := range methodHandle {
		mux.staticRoutes[path][m] = h
	}
}
