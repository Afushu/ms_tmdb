package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"ms_tmdb/pkg/tmdbclient"
)

func TestTmdbProxyRouter(t *testing.T) {
	proxied := 0
	delegated := false
	router := NewTmdbProxyRouter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxied++
		w.WriteHeader(http.StatusAccepted)
	}))
	if err := router.Handle(http.MethodGet, "/api/admin/home", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delegated = true
		w.WriteHeader(http.StatusNoContent)
	})); err != nil {
		t.Fatalf("注册委托路由失败: %v", err)
	}

	proxyPaths := []string{
		"/api/tmdb/a/b/c/d/e/f/g/h/i",
		"/api/v3/movie/550",
		"/v3/tv/1399",
		"/3/person/1",
	}
	for _, path := range proxyPaths {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusAccepted {
			t.Fatalf("%s 未进入代理，code = %d", path, recorder.Code)
		}
	}
	if proxied != len(proxyPaths) {
		t.Fatalf("代理命中次数 = %d, want %d", proxied, len(proxyPaths))
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/admin/home", nil))
	if recorder.Code != http.StatusNoContent || !delegated {
		t.Fatalf("/api/admin 未进入委托路由，code = %d, delegated = %v", recorder.Code, delegated)
	}
}

// TestTmdbProxyMiddlewareMultiPrefixSmoke 覆盖真实 Proxy/Router 链路的多前缀入口行为。
// 不构造完整 ProxyService/Client，只用 dispatch 前的拦截点验证入口分流。
func TestTmdbProxyMiddlewareMultiPrefixSmoke(t *testing.T) {
	proxy := NewTmdbProxyMiddleware(nil, nil)
	router := NewTmdbProxyRouter(http.HandlerFunc(proxy.Proxy))
	if err := router.Handle(http.MethodGet, "/api/admin/home", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})); err != nil {
		t.Fatalf("注册委托路由失败: %v", err)
	}

	proxyPaths := []string{
		"/api/tmdb/movie/550",
		"/api/v3/movie/550",
		"/v3/tv/1399/season/1",
		"/3/person/1",
	}
	for _, path := range proxyPaths {
		// 非 GET：在 resolve/dispatch 前被拦截，证明多前缀已进入代理入口。
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, path, nil))
		if recorder.Code != http.StatusMethodNotAllowed {
			t.Fatalf("POST %s 应返回 405，got %d body=%s", path, recorder.Code, recorder.Body.String())
		}
	}

	// 直接调用 Proxy：非代理路径应返回“路径不存在”，而不是放行业务路由。
	recorder := httptest.NewRecorder()
	proxy.Proxy(recorder, httptest.NewRequest(http.MethodGet, "/api/admin/home", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("Proxy(/api/admin/home) 应返回 404，got %d body=%s", recorder.Code, recorder.Body.String())
	}

	// 经 Router 时，非代理路径仍走委托路由。
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/admin/home", nil))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("/api/admin/home 未进入委托路由，code = %d", recorder.Code)
	}

	// Handle 对代理路径接管，对非代理路径放行 next。
	nextCalled := false
	handler := proxy.Handle(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusTeapot)
	})

	recorder = httptest.NewRecorder()
	handler(recorder, httptest.NewRequest(http.MethodPost, "/api/v3/movie/550", nil))
	if nextCalled {
		t.Fatal("代理路径不应放行 next")
	}
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Handle(POST /api/v3/...) 应返回 405，got %d", recorder.Code)
	}

	nextCalled = false
	recorder = httptest.NewRecorder()
	handler(recorder, httptest.NewRequest(http.MethodGet, "/api/admin/home", nil))
	if !nextCalled || recorder.Code != http.StatusTeapot {
		t.Fatalf("非代理路径未放行 next，nextCalled=%v code=%d", nextCalled, recorder.Code)
	}
}

func TestProxyErrorMapping(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantCode    int
		wantMessage string
	}{
		{
			name:        "rate limited",
			err:         &tmdbclient.APIError{StatusCode: http.StatusTooManyRequests, Body: `{"status_message":"too many requests"}`},
			wantCode:    http.StatusTooManyRequests,
			wantMessage: "TMDB 请求触发限流，请稍后重试",
		},
		{
			name:        "upstream status message",
			err:         &tmdbclient.APIError{StatusCode: http.StatusNotFound, Body: `{"status_message":"The resource you requested could not be found."}`},
			wantCode:    http.StatusNotFound,
			wantMessage: "The resource you requested could not be found.",
		},
		{
			name:        "non tmdb error",
			err:         errors.New("dial tcp timeout"),
			wantCode:    http.StatusBadGateway,
			wantMessage: "TMDB 代理请求失败，请稍后重试",
		},
		{
			name:        "invalid status falls back",
			err:         &tmdbclient.APIError{StatusCode: 0, Body: `{}`},
			wantCode:    http.StatusBadGateway,
			wantMessage: "TMDB 返回错误状态码 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := proxyErrorStatus(tt.err)
			if code != tt.wantCode {
				t.Fatalf("code = %d, want %d", code, tt.wantCode)
			}
			body := buildProxyErrorBody(code, proxyErrorMessage(tt.err))
			if got := body["status_message"]; got != tt.wantMessage {
				t.Fatalf("status_message = %v, want %q", got, tt.wantMessage)
			}
			if got := body["status_code"]; got != tt.wantCode {
				t.Fatalf("status_code = %v, want %d", got, tt.wantCode)
			}
			if got := body["success"]; got != false {
				t.Fatalf("success = %v, want false", got)
			}
		})
	}
}
