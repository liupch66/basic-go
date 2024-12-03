package accesslog

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
)

type AccessLog struct {
	Method     string `json:"method"`
	Url        string `json:"url"`
	Duration   string `json:"duration"`
	ReqBody    string `json:"req_body"`
	RespBody   string `json:"resp_body"`
	StatusCode int    `json:"status_code"`
}

type MiddlewareBuilder struct {
	// 考虑动态开关,配置 viper 监听,小心并发安全,所以这里使用原子操作
	allowReqBody  *atomic.Bool
	allowRespBody *atomic.Bool
	logLevelFunc  func(ctx context.Context, al *AccessLog)
}

func NewMiddlewareBuilder(fn func(ctx context.Context, al *AccessLog)) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logLevelFunc: fn,
		// *atomic.Bool 记得初始化,不然是 nil,会 panic; bool 就不用初始化,默认值就是 false
		allowReqBody:  atomic.NewBool(false),
		allowRespBody: atomic.NewBool(false),
	}
}

func (b *MiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		url := ctx.Request.URL.String()
		// if len(url) > 1024 {
		// 	url = url[:1024]
		// }
		al := &AccessLog{
			Method: ctx.Request.Method,
			Url:    url,
		}
		if b.allowReqBody.Load() && ctx.Request.Body != nil {
			// 这一步已经读到了 ReqBody
			// body, _ := io.ReadAll(ctx.Request.Body)
			body, _ := ctx.GetRawData()
			// 在 HTTP 请求中，Request.Body 是一个 流式读取的对象，这意味着它只能被读取一次。
			// 在默认情况下，如果直接从 ctx.Request.Body 读取了请求体的数据（例如通过 io.ReadAll() 或 ctx.GetRawData()），
			// 那么流会被消耗掉，此时如果后续的中间件或处理函数需要再次访问请求体，可能会发现它已经为空。
			// 通过重新包装请求体数据，使用 io.NopCloser(bytes.NewReader(body)) 将其转换为一个可读的 io.ReadCloser，
			// 就能够在不丢失数据的情况下多次读取请求体，并且保持请求体在后续处理中可用。
			reader := io.NopCloser(bytes.NewReader(body))
			ctx.Request.Body = reader
			al.ReqBody = string(body)
		}
		if b.allowRespBody.Load() {
			ctx.Writer = &responseWriter{
				al:             al,
				ResponseWriter: ctx.Writer,
			}
		}
		defer func() {
			al.Duration = time.Since(start).String()
			b.logLevelFunc(ctx, al)
		}()
		ctx.Next()
	}
}

func (b *MiddlewareBuilder) AllowReqBody(ok bool) *MiddlewareBuilder {
	b.allowReqBody.Store(ok)
	return b
}

func (b *MiddlewareBuilder) AllowRespBody(ok bool) *MiddlewareBuilder {
	b.allowRespBody.Store(ok)
	return b
}

// Gin 的 ctx 没有暴露响应,读不到响应.但是暴露了 ResponseWriter,所以我们可以换一个实现帮我们记录响应
// 使用组合实现部分方法即可, writer gin.ResponseWriter 需要实现所有方法
type responseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (w *responseWriter) WriteHeader(code int) {
	w.al.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.al.RespBody = s
	return w.ResponseWriter.WriteString(s)
}
