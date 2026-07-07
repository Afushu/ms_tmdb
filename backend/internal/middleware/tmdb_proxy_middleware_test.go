package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"ms_tmdb/pkg/tmdbclient"
)

func TestResolveTmdbPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantPath string
		wantOK   bool
	}{
		{name: "prefix only", path: "/api/tmdb", wantPath: "", wantOK: true},
		{name: "movie detail", path: "/api/tmdb/movie/550", wantPath: "/movie/550", wantOK: true},
		{name: "deep path", path: "/api/tmdb/a/b/c/d/e/f/g/h/i", wantPath: "/a/b/c/d/e/f/g/h/i", wantOK: true},
		{name: "similar prefix", path: "/api/tmdbxxx/movie/550", wantOK: false},
		{name: "old api v3", path: "/api/v3/movie/550", wantOK: false},
		{name: "old v3", path: "/v3/movie/550", wantOK: false},
		{name: "admin", path: "/api/admin/home", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotOK := resolveTmdbPath(tt.path)
			if gotPath != tt.wantPath || gotOK != tt.wantOK {
				t.Fatalf("resolveTmdbPath(%q) = (%q, %v), want (%q, %v)", tt.path, gotPath, gotOK, tt.wantPath, tt.wantOK)
			}
			if got := isTmdbProxyPath(tt.path); got != tt.wantOK {
				t.Fatalf("isTmdbProxyPath(%q) = %v, want %v", tt.path, got, tt.wantOK)
			}
		})
	}
}

func TestTmdbProxyRouter(t *testing.T) {
	proxied := false
	delegated := false
	router := NewTmdbProxyRouter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxied = true
		w.WriteHeader(http.StatusAccepted)
	}))
	if err := router.Handle(http.MethodGet, "/api/admin/home", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delegated = true
		w.WriteHeader(http.StatusNoContent)
	})); err != nil {
		t.Fatalf("注册委托路由失败: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/tmdb/a/b/c/d/e/f/g/h/i", nil))
	if recorder.Code != http.StatusAccepted || !proxied {
		t.Fatalf("/api/tmdb 未进入代理，code = %d, proxied = %v", recorder.Code, proxied)
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/admin/home", nil))
	if recorder.Code != http.StatusNoContent || !delegated {
		t.Fatalf("/api/admin 未进入委托路由，code = %d, delegated = %v", recorder.Code, delegated)
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
