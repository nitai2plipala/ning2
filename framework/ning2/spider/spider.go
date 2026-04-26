package spider

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

func New(header map[string]string) *Engine {
	client := new(http.Client)
	client.Timeout = 30 * time.Second

	// 初始化 CookieJar，失败时使用 nil（无 cookie 功能）
	cookie, err := cookiejar.New(nil)
	if err != nil {
		log.Println("cookiejar.New:", err)
		cookie = nil
	}

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	client.Jar = cookie

	return &Engine{
		HttpClient: client,
		Header:     header,
	}
}

func (e *Engine) Request (method, uri string, body io.Reader) (*http.Response, error) {

	return e.RequestWithContext(context.Background(), method, uri, body)
}

func (e *Engine) RequestWithContext (ctx context.Context, method, uri string, body io.Reader) (*http.Response, error) {

	request, err := http.NewRequestWithContext(ctx, method, uri, body)

	if err != nil {

		return nil, err
	}

	for name, value := range e.Header {

		request.Header.Set(name, value)
	}

	switch e.BodyCode {

	case "Form":

		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	case "Json":

		request.Header.Set("Content-Type", "application/json")

	default:

	}

	response, err := e.HttpClient.Do(request)

	if err != nil {

		log.Println("Spider crawls:", err)

		return nil, err
	}

	return response, err

}

// ErrUnsupportedContentType 非文本内容类型错误
var ErrUnsupportedContentType = errors.New("unsupported content type")

func ResponseBody(response *http.Response) ([]byte, error) {
	if response == nil {
		return nil, nil
	}

	contentType := strings.ToLower(response.Header.Get("Content-Type"))
	if contentType != "" && !strings.Contains(contentType, "text/") && 
	   !strings.Contains(contentType, "json") && !strings.Contains(contentType, "xml") {
		log.Println("ResponseBody aborted: unsupported content-type:", contentType)
		return nil, ErrUnsupportedContentType
	}

	// 处理空响应体
	if response.Body == nil {
		return []byte{}, nil
	}

	limitReader := io.LimitReader(response.Body, 10*1024*1024)
	body, err := io.ReadAll(limitReader)

	if err != nil {
		log.Println("Response Body: ", err)
		return nil, err
	}

	// 关闭原 Body 并替换为可重复读取的 NopCloser
	response.Body.Close()
	response.Body = io.NopCloser(bytes.NewReader(body))

	return body, nil
}

// ResponseBodyString 返回响应体字符串（兼容旧 API）
func ResponseBodyString(response *http.Response) string {
	body, err := ResponseBody(response)
	if err != nil {
		return ""
	}
	return string(body)
}

