package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/nitai2plipala/ning2"
)

// ==================== 自定义中间件示例 ====================

// Auth 门卫型中间件：校验 Authorization 头，失败则返回 401，不调 next 中断链路
// 位置建议：路由级外层（门卫型放外层，先拦截）
func Auth(next ning2.HandleFunc) ning2.HandleFunc {
	return func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		token := r.Header.Get("Authorization")
		if token != "Bearer secret-token" {
			// 不调 next，链路到此为止
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return nil
		}
		// 校验通过，塞数据到上下文，调用下一层
		c.Param["user"] = "admin"
		return next(w, r, c)
	}
}

// RateLimit 保护型中间件：简单限流，间隔不足 500ms 则返回 429
// 位置建议：路由级中间层
func RateLimit(next ning2.HandleFunc) ning2.HandleFunc {
	var lastTime time.Time
	return func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		now := time.Now()
		if now.Sub(lastTime) < 500*time.Millisecond {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return nil // 不调 next，直接拒绝
		}
		lastTime = now
		return next(w, r, c)
	}
}

// AdminOnly 门卫型中间件：校验 X-Role 头是否为 super
func AdminOnly(next ning2.HandleFunc) ning2.HandleFunc {
	return func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		if r.Header.Get("X-Role") != "super" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return nil
		}
		return next(w, r, c)
	}
}

// startMiddlewareExample 启动一个带中间件的示例服务，监听 8081 端口
// 与 main.go 的基础路由服务（8080）独立运行，互不干扰
func startMiddlewareExample() {
	mux := ning2.NewMux()

	// ==================== 全局中间件 ====================
	// Middleware() 注册的中间件对所有路由生效
	// 执行顺序：书写顺序从外到内（Recovery 最外，RequestID 次内）
	// 建议：兜底型(Recovery)、观察型(Logger/RequestID) 放全局
	mux.Middleware(ning2.Recovery, ning2.Logger, ning2.RequestID)

	// ==================== 基础路由（无路由级中间件）====================
	// 执行链：Recovery → Logger → RequestID → helloHandler
	mux.Handle("/hello", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.JSON(200, map[string]string{
			"msg":       "Hello, Ning2!",
			"requestId": c.Param["request_id"],
		})
	}, "GET")

	// ==================== 路由级中间件（链式 Use）====================

	// 单个路由级中间件
	// 执行链：Recovery → Logger → RequestID → Auth → adminHandler
	mux.Use(Auth).Handle("/admin", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.JSON(200, map[string]string{
			"msg":  "欢迎管理员",
			"user": c.Param["user"],
		})
	}, "GET")

	// 多个路由级中间件（书写顺序 = 从外到内执行顺序）
	// 执行链：Recovery → Logger → RequestID → Auth → RateLimit → secretHandler
	// 即：先鉴权(Auth)，通过后再限流(RateLimit)
	mux.Use(Auth, RateLimit).Handle("/api/secret", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.JSON(200, map[string]string{
			"msg":       "机密数据",
			"requestId": c.Param["request_id"],
			"user":      c.Param["user"],
		})
	}, "GET")

	// 路由级中间件分组（同一组中间件复用到多条路由）
	// adminGroup 持有 Auth + RateLimit 两个中间件
	adminGroup := mux.Use(Auth, RateLimit)
	adminGroup.Handle("/admin/users", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.JSON(200, map[string]string{"msg": "用户列表"})
	}, "GET")
	adminGroup.Handle("/admin/settings", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.JSON(200, map[string]string{"msg": "系统设置"})
	}, "GET")
	// 上面两条路由执行链相同：
	// Recovery → Logger → RequestID → Auth → RateLimit → handler

	// group 继续叠加中间件（演示 group.Use 链式扩展）
	// superAdmin 执行链：Recovery → Logger → RequestID → Auth → RateLimit → AdminOnly → handler
	superAdmin := adminGroup.Use(AdminOnly)
	superAdmin.Handle("/admin/super", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.JSON(200, map[string]string{"msg": "超级管理员区域"})
	}, "GET")

	// ==================== 变换型中间件示例（Compression 放最内层）====================
	// Compression 是变换型中间件，应放在最内层（最靠近 handler）
	// 执行链：Recovery → Logger → RequestID → Auth → Compression → handler
	// 这样 Auth 失败时不会执行 Compression，避免对错误响应做无意义压缩
	mux.Use(Auth, ning2.Compression).Handle("/api/compressed", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.JSON(200, map[string]string{
			"msg":       "这会被 gzip 压缩",
			"requestId": c.Param["request_id"],
			"user":      c.Param["user"],
		})
	}, "GET")

	addr := ":8081"
	fmt.Printf("中间件示例服务启动: http://localhost%s\n", addr)
	fmt.Println("测试建议:")
	fmt.Println("  curl http://localhost:8081/hello")
	fmt.Println("  curl http://localhost:8081/admin                              # 401 未授权")
	fmt.Println("  curl -H 'Authorization: Bearer secret-token' http://localhost:8081/admin")
	fmt.Println("  curl -H 'Authorization: Bearer secret-token' -H 'Accept-Encoding: gzip' http://localhost:8081/api/compressed")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("中间件示例服务启动失败: %v\n", err)
	}
}

func init() {
	// 后台启动中间件示例服务，与主服务（main.go 的 8080）并行运行
	go startMiddlewareExample()
}
