package ning2

import (
	"html/template"
	"net/http"
	"path"
	"strings"
	"sync"
)


func NewMux() *RouteMux {
	mux := &RouteMux{
		treeNode: &Node{
			pattern: "/",
			nType:   "root",
		},
	}
	// 初始化 syncPool
	mux.syncPool = sync.Pool{
		New: func() interface{} {
			return &Context{
				Param: make(map[string]string),
			}
		},
	}
	return mux
}

func (mux *RouteMux) NewContext(r *http.Request, w http.ResponseWriter) *Context {
	c := mux.syncPool.Get().(*Context)
	c.request = r
	c.responseWriter = NewResponseWriter(w)
	c.Pattern = ""
	c.Param = make(map[string]string)
	c.body = nil
	return c
}

func (mux *RouteMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := mux.NewContext(r, w)

	Hunter := mux.FindHandle(r, c)

	userAgent := Parse(r.UserAgent())

	c.Client.UserAgent = userAgent.String
	c.Client.URL = userAgent.URL
	c.Client.Device = userAgent.Device
	c.Client.Browser.Name = userAgent.Name
	c.Client.Browser.Version = userAgent.Version
	c.Client.OS.Name = userAgent.OS
	c.Client.OS.Version = userAgent.OSVersion
	c.Client.Robot = userAgent.Bot

	switch {
	case userAgent.Mobile:
		c.Client.Standard = ".m"
	case userAgent.Tablet:
		c.Client.Standard = ".t"
	default:
		c.Client.Standard = ".p"
	}

	// 拼接全局中间件洋葱：全局中间件在最外层，包住 handler
	// 执行顺序：全局外层 → ... → 全局内层 → handler（handler 可能已被路由级中间件包裹）
	final := chain(Hunter, mux.middlewares...)
	final(c.responseWriter, r, c)

	mux.syncPool.Put(c)
}

// Middleware 注册全局中间件，对所有路由生效
// 执行顺序：书写顺序从外到内（先注册的在最外层）
// 建议：兜底型(Recovery)、观察型(Logger) 放全局
func (mux *RouteMux) Middleware(mw ...Middleware) {
	mux.middlewares = append(mux.middlewares, mw...)
}

// Use 注册路由级中间件，返回 MiddlewareGroup 供链式调用
// 中间件只对 group.Handle 注册的路由生效
// 执行顺序：书写顺序从外到内，变换型(Gzip) 应放最后（最靠近 handler）
func (mux *RouteMux) Use(mw ...Middleware) *MiddlewareGroup {
	return &MiddlewareGroup{
		mux:         mux,
		middlewares: mw,
	}
}

// Use 在 group 上继续叠加中间件，返回新的 group
func (g *MiddlewareGroup) Use(mw ...Middleware) *MiddlewareGroup {
	return &MiddlewareGroup{
		mux:         g.mux,
		middlewares: append(append([]Middleware{}, g.middlewares...), mw...),
	}
}

// Handle 在 group 下注册路由，自动带上 group 的中间件
func (g *MiddlewareGroup) Handle(pattern string, handle HandleFunc, methods ...string) error {
	// 路由级中间件从外到内叠到 handler 上
	wrapped := chain(handle, g.middlewares...)
	return g.mux.Handle(pattern, wrapped, methods...)
}

// Resource 在 group 下注册静态资源，自动带上 group 的中间件
func (g *MiddlewareGroup) Resource(pattern, dirPath string) error {
	wrapped := chain(StripPrefix(pattern, dirPath), g.middlewares...)
	return g.mux.Handle(pattern+"?static", wrapped, "GET")
}

// chain 把一组 Middleware 从外到内叠到 handler 上，返回最终 handler
// chain(h, A, B, C) → A(B(C(h)))，调用时 A 最先执行（最外层）
func chain(handler HandleFunc, mws ...Middleware) HandleFunc {
	h := handler
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}


func NotFound(w http.ResponseWriter, r *http.Request, c *Context) error {

	http.Error(w, "404 this page not found : " + r.URL.RequestURI(), http.StatusNotFound)

	return nil
}

func StripPrefix(prefix, dirPath string) HandleFunc {
	if prefix == "" {
		return func(w http.ResponseWriter, r *http.Request, c *Context) error {
			handle := http.FileServer(http.Dir(dirPath))
			handle.ServeHTTP(w, r)
			return nil
		}
	}

	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
			// 正确修改 URL 路径，空路径补 "/" 让 FileServer 返回 index.html
			if p == "" {
				p = "/"
			}
			r.URL.Path = p
			handle := http.FileServer(http.Dir(dirPath))
			handle.ServeHTTP(w, r)
		} else {
			NotFound(w, r, c)
		}
		return nil
	}
}

func (t *Template) Render (code int, files ...string) error {

	t.Writer.WriteHeader(code)

	tpl, err := template.ParseFiles(files...)

	if err != nil {

		return err
	}

	tpl.ExecuteTemplate(t.Writer, t.Name, t.Data)

	return nil
}

func (t *Template) HtmlGlobs (code int, files ...string) error {

	fileDirPath, fileFullName := path.Split(files[0])

	fileSuffix := path.Ext(fileFullName)

	fileName := strings.TrimSuffix(fileFullName, fileSuffix)

	files[0] = strings.Join([]string{fileDirPath, fileName, ".", strings.ToLower(t.ClientStandard), fileSuffix}, "")

	t.Writer.WriteHeader(code)

	tpl, err := template.ParseFiles(files...)

	if err != nil {

		return err
	}

	tpl.ExecuteTemplate(t.Writer, t.Name, t.Data)

	return nil
}

func ListenServeAndToHttps(origin string)  {

	go http.ListenAndServe(origin, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		host := strings.Split(r.Host, ":")

		url := strings.Join([]string{"https://", host[0], r.RequestURI}, "")

		http.Redirect(w, r, url, http.StatusMovedPermanently)

		return

	}))
}
