package ning2

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ==================== 内置中间件（Middleware 类型，洋葱模型）===================

// Recovery 兜底型中间件：捕获 panic，防止进程崩溃
// 位置建议：全局最外层
func Recovery(next HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		return next(w, r, c)
	}
}

// Logger 观察型中间件：记录请求日志和耗时
// 位置建议：全局次外层
func Logger(next HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		start := time.Now()
		fmt.Printf("[Logger] --> %s %s\n", r.Method, r.URL.Path)
		err := next(w, r, c)
		fmt.Printf("[Logger] <-- %s %s %v\n", r.Method, r.URL.Path, time.Since(start))
		return err
	}
}

// RequestID 观察型中间件：生成请求唯一ID，存入 c.Param["request_id"]
// 位置建议：全局
func RequestID(next HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		rid := fmt.Sprintf("req-%d", time.Now().UnixNano())
		c.Param["request_id"] = rid
		w.Header().Set("X-Request-ID", rid)
		return next(w, r, c)
	}
}

// GzipWriterPool gzip 写入器池
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

// Compression 变换型中间件：对响应做 gzip 压缩
// 位置建议：路由级最内层（最靠近 handler），这样只在 handler 确实执行时才压缩
func Compression(next HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		// 检查客户端是否支持 gzip
		encoding := r.Header.Get("Accept-Encoding")
		if !strings.Contains(encoding, "gzip") {
			return next(w, r, c) // 不支持就放行
		}

		// 前半：替换 writer
		gw := gzipWriterPool.Get().(*gzip.Writer)
		gw.Reset(w)

		c.responseWriter.Header().Set("Content-Encoding", "gzip")
		c.responseWriter.Header().Del("Content-Length")
		c.responseWriter = &CompressWriter{
			Writer:         gw,
			ResponseWriter: c.responseWriter,
		}

		// 调 next：handler 写数据进 gw
		err := next(w, r, c)

		// 后半：刷出压缩数据
		gw.Close()
		gzipWriterPool.Put(gw)
		return err
	}
}

// shouldCompress 判断是否应该压缩
func shouldCompress(contentType string) bool {
	if contentType == "" {
		return true
	}
	noCompress := []string{
		"image/", "video/", "audio/", "application/zip",
		"application/x-gzip", "application/gzip",
		"application/octet-stream",
	}
	for _, v := range noCompress {
		if strings.Contains(contentType, v) {
			return false
		}
	}
	return true
}
