package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"ms_tmdb/internal/logic/proxy"
	"ms_tmdb/pkg/tmdbclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// TmdbProxyMiddleware TMDB API 代理中间件。
// 原生兼容 /api/tmdb、/api/v3、/v3、/3 入口，统一剥离前缀后走 dispatcher。
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
		if !isTmdbProxyPath(r.URL.Path) {
			next(w, r)
			return
		}
		m.Proxy(w, r)
	}
}

func (m *TmdbProxyMiddleware) Proxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeProxyError(w, http.StatusMethodNotAllowed, "TMDB 代理暂不支持该请求方法")
		return
	}

	// 剥离兼容入口前缀后的 TMDB 路径。
	tmdbPath, ok := resolveTmdbPath(r.URL.Path)
	if !ok {
		writeProxyError(w, http.StatusNotFound, "TMDB 代理路径不存在")
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
