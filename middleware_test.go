package ning2

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// ==================== 内置中间件测试 ====================

func TestRecovery_NoPanic(t *testing.T) {
	called := false
	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		called = true
		return nil
	}

	wrapped := Recovery(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{responseWriter: NewResponseWriter(w), request: req, Param: map[string]string{}}

	err := wrapped(w, req, c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestRecovery_CatchesPanic(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		panic("test panic")
	}

	wrapped := Recovery(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{responseWriter: NewResponseWriter(w), request: req, Param: map[string]string{}}

	err := wrapped(w, req, c)
	if err != nil {
		t.Errorf("Recovery should not return error for panic, got: %v", err)
	}
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestRequestID_SetsHeader(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		rid := c.Param["request_id"]
		if rid == "" {
			t.Error("request_id not set in context")
		}
		return nil
	}

	wrapped := RequestID(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{responseWriter: NewResponseWriter(w), request: req, Param: map[string]string{}}

	wrapped(w, req, c)

	rid := w.Header().Get("X-Request-ID")
	if rid == "" {
		t.Error("X-Request-ID header not set")
	}
}

// ==================== chain 工具函数测试 ====================

func TestChain_Order(t *testing.T) {
	var order []string

	mw1 := func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request, c *Context) error {
			order = append(order, "mw1-before")
			err := next(w, r, c)
			order = append(order, "mw1-after")
			return err
		}
	}
	mw2 := func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request, c *Context) error {
			order = append(order, "mw2-before")
			err := next(w, r, c)
			order = append(order, "mw2-after")
			return err
		}
	}
	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		order = append(order, "handler")
		return nil
	}

	wrapped := chain(handler, mw1, mw2)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{responseWriter: NewResponseWriter(w), request: req, Param: map[string]string{}}

	wrapped(w, req, c)

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, exp := range expected {
		if order[i] != exp {
			t.Errorf("at index %d: expected %s, got %s", i, exp, order[i])
		}
	}
}

func TestChain_Empty(t *testing.T) {
	called := false
	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		called = true
		return nil
	}

	wrapped := chain(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{responseWriter: NewResponseWriter(w), request: req, Param: map[string]string{}}

	wrapped(w, req, c)
	if !called {
		t.Error("handler was not called")
	}
}

// ==================== 中间件中断测试 ====================

func TestMiddleware_Abort(t *testing.T) {
	handlerCalled := false

	authMw := func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request, c *Context) error {
			// 不调 next，直接返回，模拟鉴权失败
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return nil
		}
	}
	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		handlerCalled = true
		return nil
	}

	wrapped := chain(handler, authMw)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{responseWriter: NewResponseWriter(w), request: req, Param: map[string]string{}}

	wrapped(w, req, c)

	if handlerCalled {
		t.Error("handler should not be called when middleware aborts")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

// ==================== Mux 集成测试 ====================

func TestMux_Middleware_Global(t *testing.T) {
	mux := NewMux()

	var order []string

	// 全局中间件
	mux.Middleware(
		func(next HandleFunc) HandleFunc {
			return func(w http.ResponseWriter, r *http.Request, c *Context) error {
				order = append(order, "g1-before")
				err := next(w, r, c)
				order = append(order, "g1-after")
				return err
			}
		},
		func(next HandleFunc) HandleFunc {
			return func(w http.ResponseWriter, r *http.Request, c *Context) error {
				order = append(order, "g2-before")
				err := next(w, r, c)
				order = append(order, "g2-after")
				return err
			}
		},
	)

	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		order = append(order, "handler")
		return c.String(200, "ok")
	}, "GET")

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	expected := []string{"g1-before", "g2-before", "handler", "g2-after", "g1-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, exp := range expected {
		if order[i] != exp {
			t.Errorf("at index %d: expected %s, got %s", i, exp, order[i])
		}
	}
}

func TestMux_Use_RouteScoped(t *testing.T) {
	mux := NewMux()

	var order []string

	authMw := func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request, c *Context) error {
			order = append(order, "auth-before")
			err := next(w, r, c)
			order = append(order, "auth-after")
			return err
		}
	}

	// 路由级中间件，只对 /admin 生效
	mux.Use(authMw).Handle("/admin", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		order = append(order, "admin-handler")
		return c.String(200, "admin")
	}, "GET")

	// 普通路由，不受路由级中间件影响
	mux.Handle("/hello", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		order = append(order, "hello-handler")
		return c.String(200, "hello")
	}, "GET")

	// 请求 /admin
	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// 请求 /hello
	req2 := httptest.NewRequest("GET", "/hello", nil)
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)

	expected := []string{"auth-before", "admin-handler", "auth-after", "hello-handler"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, exp := range expected {
		if order[i] != exp {
			t.Errorf("at index %d: expected %s, got %s", i, exp, order[i])
		}
	}
}

func TestMux_Use_Group(t *testing.T) {
	mux := NewMux()

	var order []string

	authMw := func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request, c *Context) error {
			order = append(order, "auth")
			return next(w, r, c)
		}
	}
	logMw := func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request, c *Context) error {
			order = append(order, "log")
			return next(w, r, c)
		}
	}

	// group 复用中间件
	group := mux.Use(authMw, logMw)
	group.Handle("/api/v1/users", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		order = append(order, "users-handler")
		return c.String(200, "users")
	}, "GET")
	group.Handle("/api/v1/settings", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		order = append(order, "settings-handler")
		return c.String(200, "settings")
	}, "GET")

	// 请求第一条路由
	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// 请求第二条路由
	req2 := httptest.NewRequest("GET", "/api/v1/settings", nil)
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)

	expected := []string{"auth", "log", "users-handler", "auth", "log", "settings-handler"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, exp := range expected {
		if order[i] != exp {
			t.Errorf("at index %d: expected %s, got %s", i, exp, order[i])
		}
	}
}

func TestMux_Use_ChainOnGroup(t *testing.T) {
	mux := NewMux()

	var order []string

	baseMw := func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request, c *Context) error {
			order = append(order, "base")
			return next(w, r, c)
		}
	}
	extraMw := func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request, c *Context) error {
			order = append(order, "extra")
			return next(w, r, c)
		}
	}

	// group 叠加中间件
	base := mux.Use(baseMw)
	super := base.Use(extraMw)
	super.Handle("/super", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		order = append(order, "handler")
		return c.String(200, "super")
	}, "GET")

	req := httptest.NewRequest("GET", "/super", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	expected := []string{"base", "extra", "handler"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, exp := range expected {
		if order[i] != exp {
			t.Errorf("at index %d: expected %s, got %s", i, exp, order[i])
		}
	}
}

// ==================== Resource 静态文件服务测试 ====================

func TestResource_Basic(t *testing.T) {
	mux := NewMux()

	// 创建临时目录和文件
	tmpDir := t.TempDir()
	os.WriteFile(tmpDir+"/test.txt", []byte("hello"), 0644)

	// 测试 Resource 注册
	err := mux.Resource("/static/", tmpDir)
	if err != nil {
		t.Errorf("Resource failed: %v", err)
	}

	// 测试请求
	req := httptest.NewRequest("GET", "/static/test.txt", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestResource_RootPath(t *testing.T) {
	// 根路径静态文件服务测试
	tmpDir := t.TempDir()
	os.WriteFile(tmpDir+"/test.txt", []byte("hello root"), 0644)

	mux := NewMux()
	err := mux.Resource("/", tmpDir)
	if err != nil {
		t.Fatalf("Resource failed: %v", err)
	}

	// 访问根路径下的文件
	req := httptest.NewRequest("GET", "/test.txt", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "hello root" {
		t.Errorf("expected body 'hello root', got '%s'", w.Body.String())
	}
}

func TestResource_InvalidPattern(t *testing.T) {
	mux := NewMux()

	// 测试无效模式（不以 / 结尾）
	err := mux.Resource("/static", "./public")
	if err == nil {
		t.Error("expected error for invalid pattern")
	}
}

// ==================== 性能基准测试 ====================

func BenchmarkMiddleware_Single(b *testing.B) {
	mw := func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request, c *Context) error {
			return next(w, r, c)
		}
	}
	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}
	wrapped := chain(handler, mw)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{responseWriter: NewResponseWriter(w), request: req, Param: map[string]string{}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapped(w, req, c)
	}
}

func BenchmarkMiddleware_Multiple(b *testing.B) {
	mws := make([]Middleware, 0, 10)
	for i := 0; i < 5; i++ {
		mws = append(mws, func(next HandleFunc) HandleFunc {
			return func(w http.ResponseWriter, r *http.Request, c *Context) error {
				return next(w, r, c)
			}
		})
	}
	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}
	wrapped := chain(handler, mws...)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{responseWriter: NewResponseWriter(w), request: req, Param: map[string]string{}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapped(w, req, c)
	}
}
