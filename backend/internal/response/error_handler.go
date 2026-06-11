package response

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"ms_tmdb/pkg/tmdbclient"
	"ms_tmdb/xerr"

	"github.com/zeromicro/go-zero/rest/httpx"
)

const clientClosedRequestStatus = 499

var publicBusinessMessages = map[string]struct{}{
	"无效电影 ID":                      {},
	"无效剧集 ID":                      {},
	"无效人物 ID":                      {},
	"无效季号":                         {},
	"无效电影 TMDB 同步 ID":              {},
	"无效剧集 TMDB 同步 ID":              {},
	"电影不存在或已删除":                    {},
	"剧集不存在或已删除":                    {},
	"电影标题不能为空":                     {},
	"剧集名称不能为空":                     {},
	"文件不能为空":                       {},
	"图片大小不能超过 10MB":                {},
	"仅支持 jpg/jpeg/png/webp/gif 图片": {},
	"仅支持图片文件上传":                    {},
	"文件内容不是受支持的图片格式":               {},
	"本地电影不存在或已删除，请重新创建":            {},
	"本地剧集不存在或已删除，请重新创建":            {},
	"电影未同步到本地，请刷新详情页后重试":           {},
	"剧集未同步到本地，请刷新详情页后重试":           {},
	"当前季未保存到本地数据库":                 {},
	"tmdb_id 必须大于 0":               {},
	"目标 tmdb_id 已存在，请使用其他 ID":      {},
	"没有可更新的电影字段":                   {},
	"没有可更新的剧集字段":                   {},
	"更新内容不能为空":                     {},
	"日志 ID 不合法":                    {},
	"日志不存在":                        {},
	"自动同步调度器未初始化":                  {},
	"配置文件路径为空":                     {},
	"配置文件缺少 Tmdb 段":                {},
	"文件名不能为空":                      {},
	"非法文件名":                        {},
	"创建本地电影失败，请重试":                 {},
	"创建本地剧集失败，请重试":                 {},
}

// ErrorBody 统一错误响应体。
type ErrorBody struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// RegisterErrorHandler 注册 go-zero 全局错误处理器。
func RegisterErrorHandler() {
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, any) {
		statusCode, body := BuildErrorResponse(ctx, err)
		return statusCode, body
	})
}

// BuildErrorResponse 将业务错误、上游错误和系统错误归一成前端稳定消费的格式。
func BuildErrorResponse(ctx context.Context, err error) (int, ErrorBody) {
	if err == nil {
		return http.StatusInternalServerError, systemErrorBody()
	}

	var codeErr *xerr.CodeError
	if errors.As(err, &codeErr) {
		return httpStatusByCode(codeErr.Code), ErrorBody{
			Code: codeErr.Code,
			Msg:  safeMessage(codeErr.Msg, "请求处理失败"),
		}
	}

	var apiErr *tmdbclient.APIError
	if errors.As(err, &apiErr) {
		return normalizeHTTPStatus(apiErr.StatusCode), ErrorBody{
			Code: xerr.TmdbRequestError,
			Msg:  tmdbErrorMessage(apiErr),
		}
	}

	if errors.Is(err, context.Canceled) || (ctx != nil && errors.Is(ctx.Err(), context.Canceled)) {
		return clientClosedRequestStatus, ErrorBody{
			Code: xerr.ServerError,
			Msg:  "请求已取消",
		}
	}
	if errors.Is(err, context.DeadlineExceeded) || (ctx != nil && errors.Is(ctx.Err(), context.DeadlineExceeded)) {
		return http.StatusGatewayTimeout, ErrorBody{
			Code: xerr.ServerError,
			Msg:  "请求超时，请稍后重试",
		}
	}

	if msg := businessMessage(err); msg != "" {
		return http.StatusBadRequest, ErrorBody{
			Code: xerr.ParamError,
			Msg:  msg,
		}
	}

	return http.StatusInternalServerError, systemErrorBody()
}

func systemErrorBody() ErrorBody {
	return ErrorBody{
		Code: xerr.ServerError,
		Msg:  "服务器异常，请稍后再试",
	}
}

func httpStatusByCode(code int) int {
	switch code {
	case xerr.OK:
		return http.StatusOK
	case xerr.ParamError:
		return http.StatusBadRequest
	case xerr.NotFound:
		return http.StatusNotFound
	case xerr.TmdbRequestError:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

func normalizeHTTPStatus(code int) int {
	if code < 100 || code > 599 {
		return http.StatusBadGateway
	}
	return code
}

func tmdbErrorMessage(err *tmdbclient.APIError) string {
	if err == nil {
		return "TMDB 请求失败，请稍后重试"
	}
	if err.StatusCode == http.StatusTooManyRequests {
		return "TMDB 请求触发限流，请稍后重试"
	}
	if msg := strings.TrimSpace(err.StatusMessage()); msg != "" {
		return msg
	}
	return "TMDB 请求失败，请稍后重试"
}

func businessMessage(err error) string {
	msg := strings.TrimSpace(err.Error())
	if msg == "" || errors.Unwrap(err) != nil || containsSensitiveMarker(msg) {
		return ""
	}
	if _, ok := publicBusinessMessages[msg]; ok {
		return msg
	}
	return ""
}

func containsSensitiveMarker(text string) bool {
	normalized := strings.ToLower(text)
	markers := []string{
		"sqlstate",
		"duplicate key",
		"gorm",
		"pq:",
		"pgconn",
		"password",
		"api_key",
		"://",
		"dial tcp",
		"connection refused",
		"no such host",
		"tls",
		"x509",
		"i/o timeout",
		"context deadline exceeded",
	}
	for _, marker := range markers {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func safeMessage(msg string, fallback string) string {
	if text := strings.TrimSpace(msg); text != "" {
		return text
	}
	return fallback
}
