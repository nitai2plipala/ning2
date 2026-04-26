package ning2

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// ==================== Context 基础测试 ====================

func TestContext_New(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	c := mux.NewContext(req, w)
	if c == nil {
		t.Fatal("NewContext returned nil")
	}
	if c.request != req {
		t.Error("request not set correctly")
	}
	if c.responseWriter != w {
		t.Error("responseWriter not set correctly")
	}
	if c.Param == nil {
		t.Error("Param map not initialized")
	}
}

func TestContext_Response_String(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	err := c.String(200, "hello world")
	if err != nil {
		t.Errorf("String() returned error: %v", err)
	}
	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "hello world" {
		t.Errorf("expected body 'hello world', got '%s'", w.Body.String())
	}
	if w.Header().Get("Content-Type") != "text/plain; charset=UTF-8" {
		t.Error("Content-Type header not set correctly")
	}
}

func TestContext_Response_JSON(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	data := map[string]string{"key": "value", "num": "123"}
	err := c.JSON(201, data)
	if err != nil {
		t.Errorf("JSON() returned error: %v", err)
	}
	if w.Code != 201 {
		t.Errorf("expected status 201, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "key") {
		t.Error("JSON body should contain 'key'")
	}
	if w.Header().Get("Content-Type") != "application/json; charset=UTF-8" {
		t.Error("Content-Type header not set correctly")
	}
}

func TestContext_Response_HTML(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	err := c.HTML(200, "<html><body>test</body></html>")
	if err != nil {
		t.Errorf("HTML() returned error: %v", err)
	}
	if w.Header().Get("Content-Type") != "text/html; charset=UTF-8" {
		t.Error("Content-Type header not set correctly")
	}
}

func TestContext_Response_NoContent(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	err := c.NoContent(204)
	if err != nil {
		t.Errorf("NoContent() returned error: %v", err)
	}
	if w.Code != 204 {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

func TestContext_Response_Redirect(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	err := c.Redirect(302, "/new-location")
	if err != nil {
		t.Errorf("Redirect() returned error: %v", err)
	}
	if w.Header().Get("Location") != "/new-location" {
		t.Errorf("expected Location '/new-location', got '%s'", w.Header().Get("Location"))
	}
}

func TestContext_Response_Redirect_InvalidCode(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	err := c.Redirect(200, "/")
	if err == nil {
		t.Error("Redirect with invalid code should return error")
	}
}

func TestContext_Response_Error(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	// 无效状态码应该返回错误
	err := c.Error(200, nil)
	if err == nil {
		t.Error("Error with invalid code should return error")
	}
}

func TestContext_Response_Error_InvalidCode(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	err := c.Error(200, nil)
	if err == nil {
		t.Error("Error with invalid code should return error")
	}
}

// ==================== Query 参数测试 ====================

func TestContext_QueryParam_URL(t *testing.T) {
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "GET")

	req := httptest.NewRequest("GET", "/test?name=john&age=25&name=jane", nil)
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	// 单值查询
	if c.QueryParam("name", "url", 0) != "john" {
		t.Error("QueryParam url failed")
	}
	// 索引越界
	if c.QueryParam("name", "url", 10) != "" {
		t.Error("QueryParam out of bounds should return empty")
	}
	// 多值查询
	if c.QueryParam("name", "url", 1) != "jane" {
		t.Error("QueryParam index 1 failed")
	}
	// 不存在的 key
	if c.QueryParam("nonexist", "url", 0) != "" {
		t.Error("nonexist key should return empty")
	}
}

func TestContext_QueryParams_URL(t *testing.T) {
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "GET")

	req := httptest.NewRequest("GET", "/test?foo=bar&foo=baz&num=123", nil)
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	params := c.QueryParams("foo", "url")
	if len(params) != 2 {
		t.Errorf("expected 2 params, got %d", len(params))
	}
	if params[0] != "bar" || params[1] != "baz" {
		t.Errorf("expected [bar baz], got %v", params)
	}

	// 不存在的 key
	if c.QueryParams("nonexist", "url") != nil {
		t.Error("nonexist key should return nil")
	}
}

func TestContext_QueryParam_Form(t *testing.T) {
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "POST")

	req := httptest.NewRequest("POST", "/test", strings.NewReader("name=jane&age=30"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	if c.QueryParam("name", "form", 0) != "jane" {
		t.Error("QueryParam form failed")
	}
	if c.QueryParam("age", "form", 0) != "30" {
		t.Error("QueryParam form age failed")
	}
}

func TestContext_QueryParams_Form(t *testing.T) {
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "POST")

	req := httptest.NewRequest("POST", "/test", strings.NewReader("tags=a&tags=b&tags=c"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	params := c.QueryParams("tags", "form")
	if len(params) != 3 {
		t.Errorf("expected 3 params, got %d", len(params))
	}
}

func TestContext_QueryParam_Json(t *testing.T) {
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "POST")

	jsonBody := `{"name":"bob","age":35,"tags":["a","b"]}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	if c.QueryParam("name", "json", 0) != "bob" {
		t.Error("QueryParam json failed")
	}
	if c.QueryParam("age", "json", 0) != "35" {
		t.Error("QueryParam json age failed")
	}
}

func TestContext_QueryParams_Json(t *testing.T) {
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "POST")

	jsonBody := `{"tags":["x","y","z"]}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	params := c.QueryParams("tags", "json")
	if len(params) != 3 {
		t.Errorf("expected 3 params, got %d", len(params))
	}
	if params[0] != "x" || params[1] != "y" || params[2] != "z" {
		t.Errorf("expected [x y z], got %v", params)
	}
}

func TestContext_QueryParam_Json_Cache(t *testing.T) {
	// 测试 JSON body 缓存 - 多次读取应该使用缓存
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "POST")

	jsonBody := `{"name":"test"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	// 第一次读取
	val1 := c.QueryParam("name", "json", 0)
	// 第二次读取（应该使用缓存）
	val2 := c.QueryParam("name", "json", 0)

	if val1 != val2 {
		t.Error("JSON body cache not working - values differ")
	}
}

// ==================== Client 信息测试 ====================

func TestContext_ClientIP(t *testing.T) {
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "GET")

	tests := []struct {
		addr     string
		expected string
	}{
		{"192.168.1.1:8080", "192.168.1.1"},
		{"10.0.0.1:12345", "10.0.0.1"},
		{"127.0.0.1", "127.0.0.1"},
		{"[::1]:8080", "::1"},
		{"[fe80::1]:8080", "fe80::1"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = tt.addr
		c := mux.NewContext(req, httptest.NewRecorder())
		if c.ClientIP() != tt.expected {
			t.Errorf("RemoteAddr %s: expected %s, got %s", tt.addr, tt.expected, c.ClientIP())
		}
	}
}

func TestContext_Scheme(t *testing.T) {
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "GET")

	// HTTP
	req := httptest.NewRequest("GET", "/test", nil)
	c := mux.NewContext(req, httptest.NewRecorder())
	if c.Scheme() != "http" {
		t.Error("HTTP scheme failed")
	}

	// HTTPS
	req = httptest.NewRequest("GET", "/test", nil)
	req.TLS = &tls.ConnectionState{}
	c = mux.NewContext(req, httptest.NewRecorder())
	if c.Scheme() != "https" {
		t.Error("HTTPS scheme failed")
	}
}

func TestContext_WebSite(t *testing.T) {
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "GET")

	req := httptest.NewRequest("GET", "/test?foo=bar", nil)
	req.Host = "example.com"
	c := mux.NewContext(req, httptest.NewRecorder())

	website := c.WebSite()
	if !strings.Contains(website, "example.com") {
		t.Error("Website should contain host")
	}
	if !strings.Contains(website, "/test") {
		t.Error("Website should contain path")
	}
	if !strings.Contains(website, "foo=bar") {
		t.Error("Website should contain query")
	}
}

// ==================== Context 池化测试 ====================

func TestContextPool_Get(t *testing.T) {
	mux := NewMux()

	c := mux.NewContext(nil, nil)
	if c == nil {
		t.Fatal("NewContext returned nil")
	}
	// 验证池化时 Param 被重置
	if c.Param == nil {
		t.Error("Param should be initialized")
	}
}

func TestContextPool_Reuse(t *testing.T) {
	mux := NewMux()

	// 获取并放回
	c1 := mux.NewContext(nil, nil)
	c1.Param["test"] = "value"
	mux.syncPool.Put(c1)

	// 重新获取
	c2 := mux.NewContext(nil, nil)
	if c2.Param["test"] != "" {
		t.Error("Param should be reset after pool reuse")
	}
}

func TestContextPool_Concurrent(t *testing.T) {
	mux := NewMux()
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			c := mux.NewContext(req, w)
			mux.syncPool.Put(c)
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

// ==================== Render 测试 ====================

func TestContext_Render(t *testing.T) {
	// 创建临时模板文件
	tmpDir := t.TempDir()
	templateFile := tmpDir + "/test.html"
	err := os.WriteFile(templateFile, []byte("<html>{{.Name}}</html>"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return c.Render(200, templateFile, map[string]string{"Name": "World"})
	}, "GET")

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "World") {
		t.Error("Render should output template data")
	}
}

func TestContext_Render_FuncMap(t *testing.T) {
	// 测试模板函数
	tmpDir := t.TempDir()
	templateFile := tmpDir + "/test.html"
	err := os.WriteFile(templateFile, []byte("<html>{{Uint .Num}}</html>"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return c.Render(200, templateFile, map[string]interface{}{"Num": 42})
	}, "GET")

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "42") {
		t.Error("Uint function should work in template")
	}
}