package response

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"ms_tmdb/pkg/tmdbclient"
	"ms_tmdb/xerr"
)

func TestBuildErrorResponseWithCodeError(t *testing.T) {
	statusCode, body := BuildErrorResponse(context.Background(), xerr.NewNotFoundError("电影不存在"))

	if statusCode != http.StatusNotFound {
		t.Fatalf("statusCode = %d, want %d", statusCode, http.StatusNotFound)
	}
	if body.Code != xerr.NotFound || body.Msg != "电影不存在" {
		t.Fatalf("body = %#v", body)
	}
}

func TestBuildErrorResponseWithPlainBusinessError(t *testing.T) {
	statusCode, body := BuildErrorResponse(context.Background(), errors.New("目标 tmdb_id 已存在，请使用其他 ID"))

	if statusCode != http.StatusBadRequest {
		t.Fatalf("statusCode = %d, want %d", statusCode, http.StatusBadRequest)
	}
	if body.Code != xerr.ParamError || body.Msg != "目标 tmdb_id 已存在，请使用其他 ID" {
		t.Fatalf("body = %#v", body)
	}
}

func TestBuildErrorResponseHidesWrappedSystemError(t *testing.T) {
	root := errors.New(`Get "https://api.themoviedb.org/3/movie/1?api_key=secret": i/o timeout`)
	statusCode, body := BuildErrorResponse(context.Background(), fmt.Errorf("TMDB 请求失败: %w", root))

	if statusCode != http.StatusInternalServerError {
		t.Fatalf("statusCode = %d, want %d", statusCode, http.StatusInternalServerError)
	}
	if body.Code != xerr.ServerError || body.Msg != "服务器异常，请稍后再试" {
		t.Fatalf("body = %#v", body)
	}
	if strings.Contains(body.Msg, "secret") || strings.Contains(body.Msg, "api_key") {
		t.Fatalf("sensitive message leaked: %q", body.Msg)
	}
}

func TestBuildErrorResponseHidesUnlistedPlainError(t *testing.T) {
	statusCode, body := BuildErrorResponse(context.Background(), errors.New("未知中文系统错误"))

	if statusCode != http.StatusInternalServerError {
		t.Fatalf("statusCode = %d, want %d", statusCode, http.StatusInternalServerError)
	}
	if body.Code != xerr.ServerError || body.Msg != "服务器异常，请稍后再试" {
		t.Fatalf("body = %#v", body)
	}
}

func TestBuildErrorResponseWithTmdbAPIError(t *testing.T) {
	statusCode, body := BuildErrorResponse(context.Background(), &tmdbclient.APIError{
		StatusCode: http.StatusTooManyRequests,
		Body:       `{"status_message":"too many requests"}`,
	})

	if statusCode != http.StatusTooManyRequests {
		t.Fatalf("statusCode = %d, want %d", statusCode, http.StatusTooManyRequests)
	}
	if body.Code != xerr.TmdbRequestError || body.Msg != "TMDB 请求触发限流，请稍后重试" {
		t.Fatalf("body = %#v", body)
	}
}
