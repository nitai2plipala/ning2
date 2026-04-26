package spider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ==================== Engine 基础测试 ====================

func TestNew(t *testing.T) {
	headers := map[string]string{
		"User-Agent": "test-agent",
		"Accept":     "text/html",
	}

	engine := New(headers)

	if engine == nil {
		t.Fatal("New returned nil")
	}
	if engine.HttpClient == nil {
		t.Error("HttpClient should be initialized")
	}
	if engine.Header["User-Agent"] != "test-agent" {
		t.Error("Header not set correctly")
	}
	if engine.Header["Accept"] != "text/html" {
		t.Error("Accept header not set correctly")
	}
}

func TestNew_NilHeaders(t *testing.T) {
	engine := New(nil)

	if engine == nil {
		t.Fatal("New returned nil for nil headers")
	}
	// Header 可以为 nil，这是预期行为
	_ = engine.Header
}

// ==================== Request 方法测试 ====================

func TestRequest_GET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.Method))
	}))
	defer server.Close()

	engine := New(nil)

	resp, err := engine.Request("GET", server.URL, nil)
	if err != nil {
		t.Errorf("GET request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := ResponseBody(resp)
	if string(body) != "GET" {
		t.Errorf("expected 'GET', got '%s'", string(body))
	}
}

func TestRequest_POST(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.Method))
	}))
	defer server.Close()

	engine := New(nil)

	resp, err := engine.Request("POST", server.URL, strings.NewReader("test body"))
	if err != nil {
		t.Errorf("POST request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := ResponseBody(resp)
	if string(body) != "POST" {
		t.Errorf("expected 'POST', got '%s'", string(body))
	}
}

func TestRequest_PUT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.Method))
	}))
	defer server.Close()

	engine := New(nil)

	resp, err := engine.Request("PUT", server.URL, nil)
	if err != nil {
		t.Errorf("PUT request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRequest_DELETE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.Method))
	}))
	defer server.Close()

	engine := New(nil)

	resp, err := engine.Request("DELETE", server.URL, nil)
	if err != nil {
		t.Errorf("DELETE request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRequest_InvalidURL(t *testing.T) {
	engine := New(nil)

	_, err := engine.Request("GET", "://invalid-url", nil)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

// ==================== RequestWithContext 测试 ====================

func TestRequestWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	engine := New(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := engine.RequestWithContext(ctx, "GET", server.URL, nil)
	if err != nil {
		t.Errorf("RequestWithContext failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRequestWithContext_Timeout(t *testing.T) {
	// 创建慢服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	engine := New(nil)
	engine.HttpClient.Timeout = 100 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := engine.RequestWithContext(ctx, "GET", server.URL, nil)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

// ==================== Headers 测试 ====================

func TestRequest_CustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.Header.Get("X-Custom-Header")))
	}))
	defer server.Close()

	headers := map[string]string{
		"X-Custom-Header": "custom-value",
	}
	engine := New(headers)

	resp, err := engine.Request("GET", server.URL, nil)
	if err != nil {
		t.Errorf("Request failed: %v", err)
	}

	body, _ := ResponseBody(resp)
	if string(body) != "custom-value" {
		t.Errorf("expected 'custom-value', got '%s'", string(body))
	}
}

func TestRequest_ContentType_Form(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.Header.Get("Content-Type")))
	}))
	defer server.Close()

	engine := New(nil)
	engine.BodyCode = "Form"

	resp, _ := engine.Request("POST", server.URL, strings.NewReader(""))
	body, _ := ResponseBody(resp)

	if !strings.Contains(string(body), "application/x-www-form-urlencoded") {
		t.Error("Form content-type should be set")
	}
}

func TestRequest_ContentType_Json(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.Header.Get("Content-Type")))
	}))
	defer server.Close()

	engine := New(nil)
	engine.BodyCode = "Json"

	resp, _ := engine.Request("POST", server.URL, strings.NewReader(""))
	body, _ := ResponseBody(resp)

	if !strings.Contains(string(body), "application/json") {
		t.Error("JSON content-type should be set")
	}
}

// ==================== 重定向测试 ====================

func TestRequest_NoRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/target", http.StatusMovedPermanently)
	}))
	defer server.Close()

	engine := New(nil)
	resp, _ := engine.Request("GET", server.URL, nil)

	// 应该不跟随重定向
	if resp.StatusCode == http.StatusOK {
		t.Error("Should not follow redirect by default")
	}
}

// ==================== ResponseBody 测试 ====================

func TestResponseBody_Nil(t *testing.T) {
	body, err := ResponseBody(nil)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if body != nil {
		t.Error("Expected nil body for nil response")
	}
}

func TestResponseBody_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// 不写入任何内容
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL)
	body, err := ResponseBody(resp)
	if err != nil {
		t.Errorf("Error reading empty body: %v", err)
	}
	// 空响应体应该返回空切片而非 nil
	if body == nil {
		t.Error("Empty body should return empty slice, not nil")
	}
}

func TestResponseBody_TextContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html>test</html>"))
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL)
	body, err := ResponseBody(resp)
	if err != nil {
		t.Errorf("Error reading body: %v", err)
	}
	if string(body) != "<html>test</html>" {
		t.Errorf("expected body '<html>test</html>', got '%s'", string(body))
	}
}

func TestResponseBody_JsonContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL)
	body, err := ResponseBody(resp)
	if err != nil {
		t.Errorf("Error reading body: %v", err)
	}
	if !strings.Contains(string(body), "ok") {
		t.Error("JSON body should be readable")
	}
}

func TestResponseBody_XmlContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<root>test</root>"))
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL)
	body, err := ResponseBody(resp)
	if err != nil {
		t.Errorf("Error reading body: %v", err)
	}
	if !strings.Contains(string(body), "test") {
		t.Error("XML body should be readable")
	}
}

func TestResponseBody_UnsupportedContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake image data"))
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL)
	body, err := ResponseBody(resp)

	// 应该返回错误
	if err != ErrUnsupportedContentType {
		t.Errorf("Expected ErrUnsupportedContentType, got %v", err)
	}
	if body != nil {
		t.Error("Body should be nil for unsupported content type")
	}
}

func TestResponseBody_ImageContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake jpeg"))
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL)
	_, err := ResponseBody(resp)

	if err != ErrUnsupportedContentType {
		t.Error("JPEG content should be rejected")
	}
}

func TestResponseBody_LargeContent(t *testing.T) {
	// 创建大于 10MB 的响应
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		// 写入 11MB 数据
		data := strings.Repeat("x", 11*1024*1024)
		io.WriteString(w, data)
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL)
	body, err := ResponseBody(resp)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// 应该被截断到 10MB
	if len(body) > 10*1024*1024 {
		t.Error("Body should be limited to 10MB")
	}
}

func TestResponseBody_Reusable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test content"))
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL)

	// 第一次读取
	body1, _ := ResponseBody(resp)
	// 第二次读取（应该仍然可用）
	body2, _ := ResponseBody(resp)

	if string(body1) != string(body2) {
		t.Error("Body should be reusable after ResponseBody")
	}
}

func TestResponseBodyString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test string"))
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL)
	str := ResponseBodyString(resp)

	if str != "test string" {
		t.Errorf("expected 'test string', got '%s'", str)
	}
}

// ==================== Cookie 测试 ====================

func TestCookiePersistence(t *testing.T) {
	var cookieValue string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := r.Cookies()
		if len(cookie) > 0 {
			cookieValue = cookie[0].Value
		}
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "test-session-id"})
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	engine := New(nil)

	// 第一次请求（获取 cookie）
	resp1, _ := engine.Request("GET", server.URL, nil)
	ResponseBody(resp1)

	// 第二次请求（应该携带 cookie）
	resp2, _ := engine.Request("GET", server.URL, nil)
	ResponseBody(resp2)

	if cookieValue == "" {
		t.Error("Cookie should be persisted between requests")
	}
}

// ==================== 性能基准测试 ====================

func BenchmarkEngine_Request(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	engine := New(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, _ := engine.Request("GET", server.URL, nil)
		ResponseBody(resp)
	}
}