package ning2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

// 全局模板函数映射（只创建一次）
var globalFuncMap = template.FuncMap{
	"Uint": func(P1 interface{}) uint {
		switch value := P1.(type) {
		case int:
			return uint(value)
		case int8:
			return uint(value)
		case int16:
			return uint(value)
		case int32:
			return uint(value)
		case int64:
			return uint(value)
		case uint:
			return value
		case uint8:
			return uint(value)
		case uint16:
			return uint(value)
		case uint32:
			return uint(value)
		case uint64:
			return uint(value)
		case float32:
			return uint(value)
		case float64:
			return uint(value)
		case string:
			if v, err := strconv.ParseUint(value, 10, 64); err == nil {
				return uint(v)
			}
		default:
			log.Println("html template Uint: unexpected type", fmt.Sprintf("%T", P1))
		}
		return 0
	},
	"String": func(P1 interface{}) string {
		return fmt.Sprint(P1)
	},
	"Add": func(P1 string, PE ...interface{}) interface{} {
		switch P1 {
		case "uint":
			var uintValue uint
			for _, v := range PE {
				uintValue = uintValue + v.(uint)
			}
			return uintValue
		case "int":
			var intValue int
			for _, v := range PE {
				intValue = intValue + v.(int)
			}
			return intValue
		}
		return ""
	},
}

func (c *Context) HTML (code int, html string) error {

	c.responseWriter.Header().Set("Content-Type", "text/html; charset=UTF-8")

	c.responseWriter.WriteHeader(code)

	c.responseWriter.Write([]byte(html))

	return nil
}

func (c *Context) JSON (code int, data interface{}) error {

	c.responseWriter.Header().Set("Content-Type", "application/json; charset=UTF-8")

	c.responseWriter.WriteHeader(code)

	err := json.NewEncoder(c.responseWriter).Encode(data)

	if  err != nil {

		return err
	}

	return nil
}

func (c *Context) Error (code int, err error) error {

	if code < 400 || code > 600 {

		return errors.New("invalid error status code")

	}

	c.responseWriter.WriteHeader(code)

	c.responseWriter.Write([]byte(err.Error()))

	return nil
}

func (c *Context) String (code int, str string) error {

	c.responseWriter.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	c.responseWriter.WriteHeader(code)

	c.responseWriter.Write([]byte(str))

	return nil
}

func (c *Context) NoContent (code int) error {

	c.responseWriter.WriteHeader(code)

	return nil
}

func (c *Context) Redirect (code int, url string) error {

	if code < 300 || code > 308 {

		return errors.New("invalid redirect status code")

	}

	c.responseWriter.Header().Set("Location", url)

	c.responseWriter.WriteHeader(code)

	return nil
}

func (c *Context) Render(code int, filename string, data interface{}) error {
	// 读取文件
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		log.Println(err)
		return err
	}

	c.responseWriter.WriteHeader(code)

	tpl := template.New(filename).Funcs(globalFuncMap)

	tpl, err = tpl.Parse(string(fileContent))
	if err != nil {
		log.Println(err)
		return err
	}

	return tpl.Execute(c.responseWriter, data)
}

func (c *Context) RenderTemplate(code int, filename string, data interface{}, files []string) error {
	if len(files) == 0 {
		files = []string{filename}
	}
	c.responseWriter.WriteHeader(code)
	tpl, err := template.ParseFiles(files...)
	if err != nil {
		log.Println("RenderTemplate parse error:", err)
		return err
	}
	return tpl.Execute(c.responseWriter, data)
}

func (c *Context) HtmlGlob (code int, filename string, data interface{}) error {

	fileDirPath, fileFullName := path.Split(filename)

	fileSuffix := path.Ext(fileFullName)

	fileName := strings.TrimSuffix(fileFullName, fileSuffix)

	filename = strings.Join([]string{fileDirPath, fileName, c.Client.Standard, fileSuffix}, "")

	return c.Render(code, filename, data)

}


func (c *Context) Scheme () string {

	if c.request.TLS != nil {

		return "https"
	}

	return "http"
}

func (c *Context) WebSite () string {

	scheme := "http://"

	if c.request.TLS != nil {

		scheme = "https://"
	}

	return strings.Join([]string{ scheme, c.request.Host, c.request.RequestURI }, "")
}

func (c *Context) ClientIP() string {
	// 处理 IPv6 格式: [::1]:8080
	addr := c.request.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx > 0 {
		// 检查是否是 IPv6
		if strings.HasPrefix(addr, "[") {
			if endIdx := strings.LastIndex(addr, "]:"); endIdx > 0 {
				return addr[1:endIdx]
			}
		}
		return addr[:idx]
	}
	return addr
}


// Bind 将请求 JSON body 解析到任意结构体指针
// 自包含实现：读 Body → 缓存 → 解析，不依赖其他函数
// 同一次请求内多次调用只读一次 Body（后续走 c.body 缓存）
// 例：var req LoginReq; if err := c.Bind(&req); err != nil { ... }
func (c *Context) Bind(ptr interface{}) error {
	// 有缓存直接用，不重复读 Body
	if c.body != nil {
		return json.Unmarshal(c.body, ptr)
	}
	body, err := io.ReadAll(c.request.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("empty body")
	}
	// 缓存原始字节并回填 Body
	c.body = body
	c.request.Body = io.NopCloser(bytes.NewReader(body))
	return json.Unmarshal(body, ptr)
}

// Query 获取 URL 查询参数（单值，返回第一个）
// 底层调用 c.request.URL.Query()，Go 标准库已做缓存
// 不存在时返回空字符串
// 例：q := c.Query("q")
func (c *Context) Query(name string) string {
	return c.request.URL.Query().Get(name)
}



