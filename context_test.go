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
	// responseWriter 现在是包装后的 ResponseWriter
	if c.responseWriter == nil {
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

// ==================== Query / Bind 参数测试 ====================

func TestContext_Query(t *testing.T) {
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "GET")

	req := httptest.NewRequest("GET", "/test?name=john&age=25&name=jane", nil)
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	// 单值查询（返回第一个）
	if c.Query("name") != "john" {
		t.Error("Query name failed")
	}
	if c.Query("age") != "25" {
		t.Error("Query age failed")
	}
	// 不存在的 key 返回空字符串
	if c.Query("nonexist") != "" {
		t.Error("nonexist key should return empty")
	}
}

func TestContext_Bind(t *testing.T) {
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "POST")

	jsonBody := `{"username":"bob","password":"secret","age":35,"tags":["a","b"]}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	var data struct {
		Username string   `json:"username"`
		Password string   `json:"password"`
		Age      int      `json:"age"`
		Tags     []string `json:"tags"`
	}
	if err := c.Bind(&data); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}
	if data.Username != "bob" {
		t.Errorf("expected username 'bob', got '%s'", data.Username)
	}
	if data.Password != "secret" {
		t.Errorf("expected password 'secret', got '%s'", data.Password)
	}
	if data.Age != 35 {
		t.Errorf("expected age 35, got %d", data.Age)
	}
	if len(data.Tags) != 2 || data.Tags[0] != "a" || data.Tags[1] != "b" {
		t.Errorf("expected tags [a b], got %v", data.Tags)
	}
}

func TestContext_Bind_EmptyBody(t *testing.T) {
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "POST")

	req := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	var data struct{ Name string }
	if err := c.Bind(&data); err == nil {
		t.Error("Bind with empty body should return error")
	}
}

func TestContext_Bind_InvalidJSON(t *testing.T) {
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "POST")

	req := httptest.NewRequest("POST", "/test", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	var data struct{ Name string }
	if err := c.Bind(&data); err == nil {
		t.Error("Bind with invalid JSON should return error")
	}
}

func TestContext_Bind_Cache(t *testing.T) {
	// 测试 body 缓存 - 多次 Bind 只读一次 Body
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}, "POST")

	jsonBody := `{"name":"test"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	var d1 struct{ Name string }
	var d2 struct{ Name string }
	if err := c.Bind(&d1); err != nil {
		t.Fatalf("first Bind failed: %v", err)
	}
	if err := c.Bind(&d2); err != nil {
		t.Fatalf("second Bind failed: %v", err)
	}
	if d1.Name != d2.Name {
		t.Error("body cache not working - values differ")
	}
}

func TestContext_Bind_Concurrent(t *testing.T) {
	// 验证并发场景下无数据竞争
	mux := NewMux()
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		var data struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		_ = c.Bind(&data)
		return nil
	}, "POST")

	done := make(chan bool, 200)
	for i := 0; i < 200; i++ {
		go func() {
			jsonBody := `{"name":"bob","age":35}`
			req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			done <- true
		}()
	}
	for i := 0; i < 200; i++ {
		<-done
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