package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"ms_tmdb/internal/logic/proxy"
	"ms_tmdb/pkg/tmdbclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// TmdbProxyMiddleware TMDB API 代理中间件
// 拦截所有 /api/v3/*、/v3/*、/3/* 请求，直接代理到 TMDB 并返回原始 JSON
type TmdbProxyMiddleware struct {
	Client       *tmdbclient.Client
	ProxyService *proxy.ProxyService
	dispatcher   *tmdbRouteDispatcher
}

func NewTmdbProxyMiddleware(client *tmdbclient.Client, proxyService *proxy.ProxyService) *TmdbProxyMiddleware {
	return &TmdbProxyMiddleware{
		Client:       client,
		ProxyService: proxyService,
		dispatcher:   newTmdbRouteDispatcher(client, proxyService),
	}
}

func (m *TmdbProxyMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 提取版本前缀后的 TMDB 路径
		tmdbPath, ok := resolveTmdbPath(r.URL.Path)
		if !ok {
			next(w, r)
			return
		}

		data, err := m.dispatcher.dispatch(tmdbPath, parseRequestOptions(r), r)
		if err != nil {
			logx.Errorf("TMDB 代理请求失败: %s, 错误: %v", tmdbPath, err)
			writeProxyError(w, proxyErrorStatus(err), proxyErrorMessage(err))
			return
		}

		writeJSONResponse(w, data)
	}
}

func resolveTmdbPath(path string) (string, bool) {
	for _, prefix := range []string{"/api/v3", "/v3", "/3"} {
		if strings.HasPrefix(path, prefix) {
			return strings.TrimPrefix(path, prefix), true
		}
	}
	return "", false
}

func proxyErrorStatus(err error) int {
	var apiErr *tmdbclient.APIError
	if errors.As(err, &apiErr) {
		return normalizeHTTPStatus(apiErr.StatusCode)
	}
	return http.StatusBadGateway
}

func proxyErrorMessage(err error) string {
	var apiErr *tmdbclient.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusTooManyRequests {
		return "TMDB 请求触发限流，请稍后重试"
	}
	if errors.As(err, &apiErr) {
		if message := apiErr.StatusMessage(); message != "" {
			return message
		}
		return fmt.Sprintf("TMDB 返回错误状态码 %d", apiErr.StatusCode)
	}
	return "TMDB 代理请求失败，请稍后重试"
}

func normalizeHTTPStatus(code int) int {
	if code < 100 || code > 599 {
		return http.StatusBadGateway
	}
	return code
}
