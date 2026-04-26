package ning2

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// ==================== Middleware 基础测试 ====================

func TestNewMidWare(t *testing.T) {
	m := NewMidWare()
	if m == nil {
		t.Fatal("NewMidWare returned nil")
	}
	if m.Before == nil {
		t.Error("Before slice not initialized")
	}
	if m.After == nil {
		t.Error("After slice not initialized")
	}
}

func TestMidWare_UseBefore(t *testing.T) {
	m := NewMidWare()

	middleware := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}

	m.UseBefore(middleware)

	if len(m.Before) != 1 {
		t.Errorf("expected 1 before middleware, got %d", len(m.Before))
	}
}

func TestMidWare_UseAfter(t *testing.T) {
	m := NewMidWare()

	middleware := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}

	m.UseAfter(middleware)

	if len(m.After) != 1 {
		t.Errorf("expected 1 after middleware, got %d", len(m.After))
	}
}

func TestMidWare_UseBefore_Multiple(t *testing.T) {
	m := NewMidWare()

	m.UseBefore(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	})
	m.UseBefore(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	})
	m.UseBefore(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	})

	if len(m.Before) != 3 {
		t.Errorf("expected 3 before middlewares, got %d", len(m.Before))
	}
}

func TestMidWare_UseAfter_Multiple(t *testing.T) {
	m := NewMidWare()

	m.UseAfter(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	})
	m.UseAfter(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	})

	if len(m.After) != 2 {
		t.Errorf("expected 2 after middlewares, got %d", len(m.After))
	}
}

// ==================== Middleware 执行顺序测试 ====================

func TestMidWare_Handler_Order(t *testing.T) {
	m := NewMidWare()

	callOrder := []string{}

	// 添加前置中间件
	m.UseBefore(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		callOrder = append(callOrder, "before1")
		return nil
	})
	m.UseBefore(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		callOrder = append(callOrder, "before2")
		return nil
	})

	// 添加后置中间件
	m.UseAfter(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		callOrder = append(callOrder, "after1")
		return nil
	})
	m.UseAfter(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		callOrder = append(callOrder, "after2")
		return nil
	})

	// 创建测试 handler
	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		callOrder = append(callOrder, "handler")
		return nil
	}

	// 创建测试请求
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{responseWriter: w, request: req}

	// 执行中间件链
	err := m.Handler(w, req, c, handler)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}

	// 验证执行顺序 - 实际实现是后进先出
	expected := []string{"before2", "before1", "handler", "after2", "after1"}
	if len(callOrder) != len(expected) {
		t.Errorf("expected call order length %d, got %d", len(expected), len(callOrder))
	}
	for i, exp := range expected {
		if i >= len(callOrder) {
			t.Errorf("missing call at index %d: expected %s", i, exp)
			continue
		}
		if callOrder[i] != exp {
			t.Errorf("at index %d: expected %s, got %s", i, exp, callOrder[i])
		}
	}
}

func TestMidWare_Handler_BeforeError(t *testing.T) {
	m := NewMidWare()

	// 前置中间件返回错误
	m.UseBefore(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return &testMiddlewareError{"before error"}
	})

	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{responseWriter: w, request: req}

	err := m.Handler(w, req, c, handler)
	if err == nil {
		t.Error("Expected error to be propagated")
	}
}

func TestMidWare_Handler_AfterError(t *testing.T) {
	m := NewMidWare()

	// 后置中间件返回错误
	m.UseAfter(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return &testMiddlewareError{"after error"}
	})

	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{responseWriter: w, request: req}

	err := m.Handler(w, req, c, handler)
	if err == nil {
		t.Error("Expected error to be propagated from after middleware")
	}
}

func TestMidWare_Handler_Context(t *testing.T) {
	m := NewMidWare()

	var capturedPattern string

	m.UseBefore(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		capturedPattern = c.Pattern
		c.Param["middleware"] = "set"
		return nil
	})

	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return c.String(200, c.Param["middleware"])
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{
		responseWriter: w,
		request:        req,
		Pattern:        "/test",
		Param:          make(map[string]string),
	}

	err := m.Handler(w, req, c, handler)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}

	if capturedPattern != "/test" {
		t.Errorf("expected pattern /test, got %s", capturedPattern)
	}
	// 验证 middleware 设置的参数被 handler 使用
	if w.Body.String() != "set" {
		t.Errorf("expected body 'set', got '%s'", w.Body.String())
	}
}

// ==================== Middleware 与 Mux 集成测试 ====================

func TestMiddleware_WithMux(t *testing.T) {
	mux := NewMux()
	middleware := NewMidWare()

	callOrder := []string{}

	// 添加中间件
	middleware.UseBefore(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		callOrder = append(callOrder, "before")
		return nil
	})

	// 注册路由
	mux.Handle("/test", func(w http.ResponseWriter, r *http.Request, c *Context) error {
		callOrder = append(callOrder, "handler")
		return c.String(200, "ok")
	}, "GET")

	// 创建测试请求
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := mux.NewContext(req, w)

	// 获取 handler 并应用中间件
	handler := mux.FindHandle(req, c)
	wrappedHandler := middleware.Handler

	// 模拟完整的请求处理流程
	_ = wrappedHandler(w, req, c, handler)

	if len(callOrder) != 2 {
		t.Errorf("expected 2 calls, got %d: %v", len(callOrder), callOrder)
	}
	if callOrder[0] != "before" {
		t.Error("before middleware should execute first")
	}
	if callOrder[1] != "handler" {
		t.Error("handler should execute after middleware")
	}
}

// ==================== 辅助类型 ====================

type testMiddlewareError struct {
	message string
}

func (e *testMiddlewareError) Error() string {
	return e.message
}

// ==================== 性能基准测试 ====================

func BenchmarkMiddleware_Single(b *testing.B) {
	m := NewMidWare()
	m.UseBefore(func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	})

	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{responseWriter: w, request: req}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Handler(w, req, c, handler)
	}
}

func BenchmarkMiddleware_Multiple(b *testing.B) {
	m := NewMidWare()
	for i := 0; i < 5; i++ {
		m.UseBefore(func(w http.ResponseWriter, r *http.Request, c *Context) error {
			return nil
		})
		m.UseAfter(func(w http.ResponseWriter, r *http.Request, c *Context) error {
			return nil
		})
	}

	handler := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c := &Context{responseWriter: w, request: req}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Handler(w, req, c, handler)
	}
}