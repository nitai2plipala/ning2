package ning2

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

// debugMode 控制是否输出路由匹配调试信息
// 通过环境变量 NING2_DEBUG=1 开启
var debugMode = os.Getenv("NING2_DEBUG") == "1"

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
			if handler, ok := root.methodHandle[method]; ok {
				c.Pattern = "/"
				return handler
			}
		}
		// 如果根路径没有注册 handler，继续走 matchNode
		// 让 /?static 等节点有机会匹配根路径
		arbiter := matchNode(root, "/", method)
		if arbiter.handleFunc != nil {
			c.Param = arbiter.mimicry
			c.Pattern = arbiter.pattern
			return arbiter.handleFunc
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

	// 2. 递归遍历路由树查找最佳匹配
	arbiter := matchNode(root, urlPath, method)

	if debugMode {
		log.Printf("[DEBUG] FindHandleByPath: urlPath=%q method=%q -> priority=%d pattern=%q mimicry=%v handleFunc=%v",
			urlPath, method, arbiter.priority, arbiter.pattern, arbiter.mimicry, arbiter.handleFunc != nil)
	}

	c.Param = arbiter.mimicry
	c.Pattern = arbiter.pattern

	if arbiter.handleFunc == nil {
		return NotFound
	}
	return arbiter.handleFunc
}

// priorityValue 返回节点类型的优先级权重
// default(5) > regexp(4) > param(3) > static(2) > whole(1) > root(0)
func priorityValue(nType NodeType) uint {
	switch nType {
	case "default":
		return 5
	case "regexp":
		return 4
	case "param":
		return 3
	case "static":
		return 2
	case "whole":
		return 1
	default:
		return 0
	}
}

// matchNode 从 node 开始递归匹配 urlPath，返回最佳匹配结果
func matchNode(node *Node, urlPath, method string) Arbiter {
	pathRegex := getPathRegex()
	best := Arbiter{
		mimicry:    make(map[string]string),
		handleFunc: nil,
		priority:   0,
	}

	// 遍历子节点，按优先级从高到低尝试匹配
	for _, child := range node.children {
		var (
			matched    bool
			paramValue string
			restPath   string // 剩余路径
		)

		seg := pathRegex.FindString(urlPath)

		switch child.nType {
		case "default":
			if len(seg) == 0 {
				continue
			}
			if seg == child.pattern {
				matched = true
				restPath = strings.TrimPrefix(urlPath, seg)
			}

		case "regexp":
			if len(seg) == 0 {
				continue
			}
			rew := getCompiledRegexp("^/" + child.regexp + "$")
			if rew.MatchString(seg) {
				matched = true
				paramValue = seg[1:]
				restPath = strings.TrimPrefix(urlPath, seg)
			}

		case "param":
			if len(seg) == 0 {
				continue
			}
			matched = true
			paramValue = seg[1:]
			restPath = strings.TrimPrefix(urlPath, seg)

		case "static":
			// static 匹配整个剩余路径
			if len(urlPath) == 0 {
				continue
			}
			matched = true
			paramValue = urlPath[1:]
			restPath = ""

		case "whole":
			// whole 匹配整个剩余路径，固定用 "path" 作为 key
			matched = true
			rest := urlPath
			if len(rest) > 0 && rest[0] == '/' {
				rest = rest[1:]
			}
			paramValue = rest
			restPath = ""
		}

		if !matched {
			continue
		}

		// 递归匹配剩余路径
		var sub Arbiter
		if restPath == "" {
			// 路径已消费完，检查当前节点是否有 handler
			sub = Arbiter{
				mimicry:    make(map[string]string),
				pattern:    child.pattern,
				priority:   priorityValue(child.nType),
				handleFunc: nil,
			}
			if child.methodHandle != nil {
				if h, ok := child.methodHandle[method]; ok {
					sub.handleFunc = h
				}
			}
		} else if child.nType == "whole" || child.nType == "static" {
			// whole/static 已消费所有路径，不应再有剩余
			continue
		} else {
			// 递归匹配子节点的子节点
			sub = matchNode(child, restPath, method)
			if sub.handleFunc == nil {
				continue
			}
			// 累加 pattern 和 priority
			sub.pattern = child.pattern + sub.pattern
			sub.priority += priorityValue(child.nType)
		}

		// 存储参数
		if paramValue != "" || child.alias != "" {
			if child.alias != "" {
				sub.mimicry[child.alias] = paramValue
			}
		}

		// 选择 priority 最高的匹配
		if sub.handleFunc != nil && sub.priority > best.priority {
			best = sub
			// 精确匹配（default 类型且路径完全消费）立即返回
			if child.nType == "default" && restPath == "" {
				break
			}
		}
	}

	return best
}
