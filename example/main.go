package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/nitai2plipala/ning2"
)

func main() {
	mux := ning2.NewMux()

	// ==================== 基础路由 ====================

	// 基础路由 - 返回字符串
	mux.Handle("/hello", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.String(200, "Hello, Ning2!")
	}, "GET")

	// 带参数路由 - /user/123
	mux.Handle("/user/:id", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		userID := c.Param["id"]
		return c.JSON(200, map[string]string{"user_id": userID})
	}, "GET")

	// 多 HTTP 方法
	mux.Handle("/api/data", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
			"method": r.Method,
		})
	}, "GET", "POST", "PUT", "DELETE")

	// 多级参数路由 - /user/123/post/456
	mux.Handle("/user/:userId/post/:postId", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.JSON(200, map[string]string{
			"user_id": c.Param["userId"],
			"post_id": c.Param["postId"],
		})
	}, "GET")

	// 正则路由 - 只匹配数字
	mux.Handle("/items/<id>[0-9]+", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.JSON(200, map[string]string{"item_id": c.Param["id"]})
	}, "GET")

	// Query 参数
	mux.Handle("/search", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		query := c.Query("q")
		page := c.Query("page")
		return c.JSON(200, map[string]string{
			"query": query,
			"page":  page,
		})
	}, "GET")

	// POST JSON 请求体绑定
	// 用 c.Bind 将 JSON body 解析到结构体，类型安全
	mux.Handle("/login", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.Bind(&req); err != nil {
			return c.Error(400, err)
		}
		return c.JSON(200, map[string]string{
			"username": req.Username,
			"password": req.Password,
		})
	}, "POST")

	// HTML 响应
	mux.Handle("/page", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.HTML(200, "<html><body><h1>Hello World</h1></body></html>")
	}, "GET")

	// 重定向
	mux.Handle("/old", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.Redirect(302, "/hello")
	}, "GET")

	// ==================== 通配符路由 ====================

	// 通配符路由 - 匹配 /files 下所有路径
	// 对外表达: /files/*  内部用 c.Param["path"] 取值
	mux.Handle("/files/*", func(w http.ResponseWriter, r *http.Request, c *ning2.Context) error {
		return c.JSON(200, map[string]interface{}{
			"message": "通配符路由匹配成功",
			"path":    c.Param["path"],
			"pattern": c.Pattern,
		})
	}, "GET")

	// ==================== 静态文件路由 ====================

	// 静态文件服务 - 把 ./public 目录映射到 /static/ 前缀
	// 访问 http://localhost:8080/static/             -> public/index.html
	// 访问 http://localhost:8080/static/css/style.css -> public/css/style.css
	// 访问 http://localhost:8080/static/js/app.js      -> public/js/app.js
	mux.Resource("/static/", "./public")

	// 静态文件服务 - 根路径映射，把 ./public 目录映射到 / 前缀
	// 访问 http://localhost:8080/             -> public/index.html
	// 访问 http://localhost:8080/css/style.css -> public/css/style.css
	// 注意：根路径静态服务优先级最低，不会覆盖上面已注册的 /hello、/user/:id 等路由
	mux.Resource("/", "./public")

	addr := ":8080"
	fmt.Printf("Ning2 v%s 示例服务启动中...\n", ning2.Version)
	fmt.Printf("访问: http://localhost%s\n", addr)
	fmt.Println("按 Ctrl+C 停止服务")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
	}
}
