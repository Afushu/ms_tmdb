package middleware

import (
	"errors"
	"net/http"
	"testing"

	"ms_tmdb/pkg/tmdbclient"
)

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
