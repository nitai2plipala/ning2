package ning2

import (
	"net/http"
)

// MidWareFunc 中间件函数类型
type MidWareFunc func(HandleFunc) HandleFunc

// NewMidWare 创建新的中间件管理器
func NewMidWare() *MidWare {
	return &MidWare{
		Before: make([]HandleFunc, 0),
		Middle: make([]HandleFunc, 0),
		After:  make([]HandleFunc, 0),
	}
}

// UseBefore 添加请求前中间件
func (m *MidWare) UseBefore(fn HandleFunc) {
	m.Before = append(m.Before, fn)
}

// UseMiddle 添加通用中间件
func (m *MidWare) UseMiddle(fn HandleFunc) {
	m.Middle = append(m.Middle, fn)
}

// UseAfter 添加请求后中间件
func (m *MidWare) UseAfter(fn HandleFunc) {
	m.After = append(m.After, fn)
}

// Use 添加中间件（默认添加到 Middle）
func (m *MidWare) Use(fn HandleFunc) {
	m.Middle = append(m.Middle, fn)
}

// Handler 执行中间件链
func (m *MidWare) Handler(w http.ResponseWriter, r *http.Request, c *Context, final HandleFunc) error {
	// 构建处理链
	handler := final

	// 后置中间件反向执行
	for i := len(m.After) - 1; i >= 0; i-- {
		handler = m.wrapAfter(m.After[i], handler)
	}

	// 前置中间件正向执行
	for _, fn := range m.Before {
		handler = m.wrapBefore(fn, handler)
	}

	// 通用中间件
	for _, fn := range m.Middle {
		handler = m.wrapMiddle(fn, handler)
	}

	return handler(w, r, c)
}

// wrapBefore 包装前置中间件
func (m *MidWare) wrapBefore(fn HandleFunc, next HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		if err := fn(w, r, c); err != nil {
			return err
		}
		return next(w, r, c)
	}
}

// wrapMiddle 包装通用中间件
func (m *MidWare) wrapMiddle(fn HandleFunc, next HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		// 前置处理
		if err := fn(w, r, c); err != nil {
			return err
		}
		// 执行下一个
		if err := next(w, r, c); err != nil {
			return err
		}
		// 后置处理
		return nil
	}
}

// wrapAfter 包装后置中间件
func (m *MidWare) wrapAfter(fn HandleFunc, next HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		err := next(w, r, c)
		if err != nil {
			return err
		}
		return fn(w, r, c)
	}
}

// Chain 中间件链式调用辅助函数
func Chain(fns ...HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		for _, fn := range fns {
			if err := fn(w, r, c); err != nil {
				return err
			}
		}
		return nil
	}
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware(w http.ResponseWriter, r *http.Request, c *Context) error {
	// 这里可以添加日志逻辑
	return nil
}

// RecoveryMiddleware 恢复中间件，防止 panic 导致崩溃
func RecoveryMiddleware(w http.ResponseWriter, r *http.Request, c *Context) error {
	defer func() {
		if err := recover(); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()
	return nil
}