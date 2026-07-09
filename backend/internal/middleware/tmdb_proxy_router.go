package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/router"
)

// TmdbProxyRouter 在 go-zero 路由匹配前接管 TMDB 代理入口，避免固定深度兜底路由遗漏路径。
type TmdbProxyRouter struct {
	delegate httpx.Router
	proxy    http.Handler
}

func NewTmdbProxyRouter(proxy http.Handler) httpx.Router {
	return &TmdbProxyRouter{
		delegate: router.NewRouter(),
		proxy:    proxy,
	}
}

func (r *TmdbProxyRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r != nil && r.proxy != nil && isTmdbProxyPath(req.URL.Path) {
		r.proxy.ServeHTTP(w, req)
		return
	}
	r.delegate.ServeHTTP(w, req)
}

func (r *TmdbProxyRouter) Handle(method, path string, handler http.Handler) error {
	return r.delegate.Handle(method, path, handler)
}

func (r *TmdbProxyRouter) SetNotFoundHandler(handler http.Handler) {
	r.delegate.SetNotFoundHandler(handler)
}

func (r *TmdbProxyRouter) SetNotAllowedHandler(handler http.Handler) {
	r.delegate.SetNotAllowedHandler(handler)
}
