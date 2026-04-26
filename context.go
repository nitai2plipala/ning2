package ning2

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
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


// jsonBodyCache 用于缓存 JSON 请求体，避免重复读取
var jsonBodyCache = make(map[*http.Request][]byte)

func (c *Context) QueryParam(name string, flag string, index int) string {
	switch flag {
	case "url":
		values := c.request.URL.Query()[name]
		if values != nil && len(values) > index {
			return values[index]
		}

	case "form":
		_ = c.request.ParseForm()
		values := c.request.Form[name]
		if len(values) > index {
			return values[index]
		}

	case "json":
		// 尝试从缓存获取 body
		var body []byte
		if cached, ok := jsonBodyCache[c.request]; ok {
			body = cached
		} else {
			var err error
			body, err = io.ReadAll(c.request.Body)
			if err != nil {
				log.Println("Read body error:", err)
				return ""
			}
			// 缓存 body 并重新填充
			jsonBodyCache[c.request] = body
			c.request.Body = io.NopCloser(strings.NewReader(string(body)))
		}

		var jsonData map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			log.Println("JSON parse error:", err)
			return ""
		}
		if val, ok := jsonData[name]; ok {
			return fmt.Sprintf("%v", val)
		}

	default:
		log.Println("QueryParam: unknown flag", flag)
	}

	return ""
}

func (c *Context) QueryParams(name string, flag string) []string {
	switch flag {
	case "url":
		values := c.request.URL.Query()[name]
		if values != nil {
			return values
		}

	case "form":
		_ = c.request.ParseForm()
		values := c.request.PostForm[name]
		if values != nil {
			return values
		}

	case "json":
		// 尝试从缓存获取 body
		var body []byte
		if cached, ok := jsonBodyCache[c.request]; ok {
			body = cached
		} else {
			var err error
			body, err = io.ReadAll(c.request.Body)
			if err != nil {
				log.Println("Read body error:", err)
				return nil
			}
			// 缓存 body 并重新填充
			jsonBodyCache[c.request] = body
			c.request.Body = io.NopCloser(strings.NewReader(string(body)))
		}

		var jsonData map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			log.Println("JSON parse error:", err)
			return nil
		}
		if val, ok := jsonData[name]; ok {
			if arr, ok := val.([]interface{}); ok {
				result := make([]string, len(arr))
				for i, v := range arr {
					result[i] = fmt.Sprintf("%v", v)
				}
				return result
			}
			return []string{fmt.Sprintf("%v", val)}
		}

	default:
		log.Println("QueryParams: unknown flag", flag)
	}

	return nil
}



