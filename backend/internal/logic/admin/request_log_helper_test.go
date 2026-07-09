package admin

import (
	"testing"

	"ms_tmdb/internal/model"
)

func TestProxyAccessLogItemCanonicalizesPathForDisplay(t *testing.T) {
	record := model.ProxyAccessLog{
		RequestID:  "req-1",
		Path:       "/api/tmdb/movie/550",
		Query:      "language=zh-CN",
		RequestURI: "/api/tmdb/movie/550?language=zh-CN",
	}

	item := proxyAccessLogItem(record)

	if item.Path != "/movie/550" {
		t.Fatalf("Path = %q, want %q", item.Path, "/movie/550")
	}
	if item.Query != record.Query || item.RequestUri != record.RequestURI || item.RequestId != record.RequestID {
		t.Fatalf("追溯字段丢失：query=%q request_uri=%q request_id=%q", item.Query, item.RequestUri, item.RequestId)
	}
}

func TestProxyAccessLogDetailCanonicalizesPathForDisplay(t *testing.T) {
	record := model.ProxyAccessLog{
		RequestID:  "req-2",
		Path:       "/3/tv/1399/season/1",
		Query:      "append_to_response=credits",
		RequestURI: "/3/tv/1399/season/1?append_to_response=credits",
	}

	detail := proxyAccessLogDetail(record)

	if detail.Path != "/tv/1399/season/1" {
		t.Fatalf("Path = %q, want %q", detail.Path, "/tv/1399/season/1")
	}
	if detail.Query != record.Query || detail.RequestUri != record.RequestURI || detail.RequestId != record.RequestID {
		t.Fatalf("追溯字段丢失：query=%q request_uri=%q request_id=%q", detail.Query, detail.RequestUri, detail.RequestId)
	}
}

func TestTmdbRequestLogItemKeepsBackendPath(t *testing.T) {
	record := model.TmdbRequestLog{
		RequestID: "req-tmdb-1",
		Path:      "/movie/550",
		URL:       "https://api.themoviedb.org/3/movie/550?api_key=***",
	}

	item := tmdbRequestLogItem(record)

	if item.Path != record.Path {
		t.Fatalf("Path = %q, want backend path %q", item.Path, record.Path)
	}
	if item.Url != record.URL || item.RequestId != record.RequestID {
		t.Fatalf("TMDB 日志字段丢失：url=%q request_id=%q", item.Url, item.RequestId)
	}
}

func TestTmdbRequestLogDetailKeepsBackendPath(t *testing.T) {
	record := model.TmdbRequestLog{
		RequestID: "req-tmdb-2",
		Path:      "/tv/1399/season/1",
		URL:       "https://api.themoviedb.org/3/tv/1399/season/1?api_key=***",
	}

	detail := tmdbRequestLogDetail(record)

	if detail.Path != record.Path {
		t.Fatalf("Path = %q, want backend path %q", detail.Path, record.Path)
	}
	if detail.Url != record.URL || detail.RequestId != record.RequestID {
		t.Fatalf("TMDB 日志字段丢失：url=%q request_id=%q", detail.Url, detail.RequestId)
	}
}
