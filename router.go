package ning2

import (
	"net/http"
	"regexp"
	"strings"
	"sync"
)

// 正则表达式缓存
var (
	regexpCache   = &sync.Map{}
	pathRegexCache *regexp.Regexp
)

func getPathRegex() *regexp.Regexp {
	if pathRegexCache == nil {
		pathRegexCache = regexp.MustCompile(`/[^/]+`)
	}
	return pathRegexCache
}

func getCompiledRegexp(pattern string) *regexp.Regexp {
	if re, ok := regexpCache.Load(pattern); ok {
		return re.(*regexp.Regexp)
	}
	re := regexp.MustCompile(pattern)
	regexpCache.Store(pattern, re)
	return re
}

func (mux *RouteMux) FindHandle(r *http.Request, c *Context) HandleFunc {
	return mux.FindHandleByPath(r.URL.Path, r.Method, c)
}

func (mux *RouteMux) FindHandleByPath(urlPath, method string, c *Context) HandleFunc {
	root := mux.treeNode

	// 根路径处理
	if urlPath == "/" || urlPath == "" {
		if root.methodHandle != nil {
			c.Pattern = "/"
			return root.methodHandle[method]
		}
		return NotFound
	}

	// 1. 优先尝试精确匹配 (O(1))
	if mux.staticRoutes != nil {
		if methodMap, ok := mux.staticRoutes[urlPath]; ok {
			if handler, ok := methodMap[method]; ok {
				c.Pattern = urlPath
				return handler
			}
		}
	}

	// 2. 遍历子节点查找最佳匹配
	var matchedArbiter Arbiter
	matchedArbiter.handleFunc = NotFound
	matchedArbiter.mimicry = make(map[string]string)

	for _, child := range root.children {
		arbiter := child.findHandleWithMethod(urlPath, method)
		if arbiter.handleFunc != nil && arbiter.priority > matchedArbiter.priority {
			matchedArbiter = arbiter
		}
		// 精确匹配立即返回（完全匹配 priority >= 10）
		if arbiter.priority >= 10 && arbiter.handleFunc != nil {
			break
		}
	}

	c.Param = matchedArbiter.mimicry
	c.Pattern = matchedArbiter.pattern

	return matchedArbiter.handleFunc
}

func (node *Node) findHandleWithMethod(urlPath, method string) Arbiter {
	arbiter := Arbiter{
		mimicry:    make(map[string]string),
		handleFunc: nil,
	}

	// 使用缓存的路径正则
	pathRegex := getPathRegex()

	// 遍历子节点
	for _, child := range node.children {
		path := pathRegex.FindString(urlPath)

		switch child.nType {
		case "default":
			if path == child.pattern {
				urlPath = strings.TrimPrefix(urlPath, path)
				arbiter.pattern += child.pattern
				arbiter.priority += 5
			} else {
				continue
			}

		case "regexp":
			rew := getCompiledRegexp("^/" + child.regexp + "$")
			if rew.MatchString(path) {
				arbiter.mimicry[child.alias] = path[1:]
				urlPath = strings.TrimPrefix(urlPath, path)
				arbiter.pattern += child.pattern
				arbiter.priority += 4
			} else {
				continue
			}

		case "param":
			arbiter.mimicry[child.alias] = path[1:]
			urlPath = strings.TrimPrefix(urlPath, path)
			arbiter.pattern += child.pattern
			arbiter.priority += 3

		case "static":
			arbiter.mimicry[child.alias] = urlPath[1:]
			urlPath = ""
			arbiter.pattern += child.pattern
			arbiter.priority += 2

		case "whole":
			// whole 匹配剩余所有路径
			// 例如 /api/*path 匹配 /api/v1/users，提取 v1/users
			prefix := strings.TrimPrefix(child.pattern, "/*")
			// 使用完整 urlPath 而非只取第一段
			if strings.HasPrefix(urlPath, "/"+prefix) {
				// 提取剩余所有路径
				rest := strings.TrimPrefix(urlPath, "/"+prefix)
				arbiter.mimicry[child.alias] = rest
				urlPath = ""
				arbiter.pattern += child.pattern
				arbiter.priority += 1
			} else {
				continue
			}
		}
	}

	// 处理当前节点
	switch node.nType {
	case "default":
		path := pathRegex.FindString(urlPath)
		if path == node.pattern {
			urlPath = strings.TrimPrefix(urlPath, path)
			arbiter.priority += 5
		}

	case "regexp":
		path := pathRegex.FindString(urlPath)
		rew := getCompiledRegexp("^/" + node.regexp + "$")
		if rew.MatchString(path) {
			arbiter.mimicry[node.alias] = path[1:]
			urlPath = strings.TrimPrefix(urlPath, path)
			arbiter.priority += 4
		}

	case "param":
		path := pathRegex.FindString(urlPath)
		arbiter.mimicry[node.alias] = path[1:]
		urlPath = strings.TrimPrefix(urlPath, path)
		arbiter.priority += 3

	case "static":
		arbiter.mimicry[node.alias] = urlPath[1:]
		urlPath = ""
		arbiter.priority += 2

	case "whole":
		// whole 匹配剩余所有路径
		prefix := strings.TrimPrefix(node.pattern, "/*")
		// 使用完整 urlPath 而非只取第一段
		if strings.HasPrefix(urlPath, "/"+prefix) {
			rest := strings.TrimPrefix(urlPath, "/"+prefix)
			arbiter.mimicry[node.alias] = rest
			urlPath = ""
			arbiter.priority += 1
		}
	}

	// 检查是否完全匹配
	if urlPath == "" {
		arbiter.pattern += node.pattern
		if node.methodHandle != nil {
			arbiter.handleFunc = node.methodHandle[method]
		}
	}

	// 如果没有匹配，返回 NotFound
	if arbiter.handleFunc == nil {
		arbiter.handleFunc = NotFound
	}

	return arbiter
}
