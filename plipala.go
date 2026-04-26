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

	// 使用包装后的 ResponseWriter，支持 middleware 替换
	Hunter(c.responseWriter, r, c)

	mux.syncPool.Put(c)
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
			// 正确修改 URL 路径
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
