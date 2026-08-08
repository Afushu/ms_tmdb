package service

import (
	"testing"

	"ms_tmdb/config"
)

func TestBuildProxyAccessLog(t *testing.T) {
	tests := []struct {
		name          string
		entry         ProxyAccessEntry
		wantPath      string
		wantMediaType string
		wantTmdbID    int
		wantURI       string
		wantQuery     string
		wantRequestID string
	}{
		{
			name: "canonical movie detail",
			entry: ProxyAccessEntry{
				RequestID:  "req-movie-550",
				Method:     "GET",
				Path:       "/movie/550",
				Query:      "language=zh-CN",
				RequestURI: "/api/tmdb/movie/550?language=zh-CN",
				StatusCode: 200,
			},
			wantPath:      "/movie/550",
			wantMediaType: "movie",
			wantTmdbID:    550,
			wantURI:       "/api/tmdb/movie/550?language=zh-CN",
			wantQuery:     "language=zh-CN",
			wantRequestID: "req-movie-550",
		},
		{
			name: "canonical tv nested path",
			entry: ProxyAccessEntry{
				RequestID:  "req-tv-1399",
				Path:       "/tv/1399/season/1",
				Query:      "",
				RequestURI: "/3/tv/1399/season/1",
			},
			wantPath:      "/tv/1399/season/1",
			wantMediaType: "tv",
			wantTmdbID:    1399,
			wantURI:       "/3/tv/1399/season/1",
			wantRequestID: "req-tv-1399",
		},
		{
			name: "prefixed path still parses media target",
			entry: ProxyAccessEntry{
				RequestID:  "req-prefix",
				Path:       "/api/v3/movie/299536",
				Query:      "api_key=***",
				RequestURI: "/api/v3/movie/299536?api_key=***",
			},
			wantPath:      "/api/v3/movie/299536",
			wantMediaType: "movie",
			wantTmdbID:    299536,
			wantURI:       "/api/v3/movie/299536?api_key=***",
			wantQuery:     "api_key=***",
			wantRequestID: "req-prefix",
		},
		{
			name: "search path no media target",
			entry: ProxyAccessEntry{
				RequestID:  "req-search",
				Path:       "/search/movie",
				Query:      "query=x",
				RequestURI: "/v3/search/movie?query=x",
			},
			wantPath:      "/search/movie",
			wantMediaType: "",
			wantTmdbID:    0,
			wantURI:       "/v3/search/movie?query=x",
			wantQuery:     "query=x",
			wantRequestID: "req-search",
		},
		{
			name: "empty canonical path only prefix",
			entry: ProxyAccessEntry{
				RequestID:  "req-empty",
				Path:       "",
				RequestURI: "/api/v3",
			},
			wantPath:      "",
			wantMediaType: "",
			wantTmdbID:    0,
			wantURI:       "/api/v3",
			wantRequestID: "req-empty",
		},
		{
			name: "invalid id yields empty media target",
			entry: ProxyAccessEntry{
				RequestID:  "req-invalid",
				Path:       "/tv/latest",
				RequestURI: "/api/tmdb/tv/latest",
			},
			wantPath:      "/tv/latest",
			wantMediaType: "",
			wantTmdbID:    0,
			wantURI:       "/api/tmdb/tv/latest",
			wantRequestID: "req-invalid",
		},
		{
			name: "negative id is accepted by parser",
			entry: ProxyAccessEntry{
				RequestID:  "req-neg",
				Path:       "/movie/-1",
				RequestURI: "/api/tmdb/movie/-1",
			},
			wantPath:      "/movie/-1",
			wantMediaType: "movie",
			wantTmdbID:    -1,
			wantURI:       "/api/tmdb/movie/-1",
			wantRequestID: "req-neg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildProxyAccessLog(tt.entry)
			if got.Path != tt.wantPath {
				t.Fatalf("path = %q, want %q", got.Path, tt.wantPath)
			}
			if got.MediaType != tt.wantMediaType || got.TmdbID != tt.wantTmdbID {
				t.Fatalf("media_target = (%q, %d), want (%q, %d)", got.MediaType, got.TmdbID, tt.wantMediaType, tt.wantTmdbID)
			}
			if got.RequestURI != tt.wantURI {
				t.Fatalf("request_uri = %q, want %q", got.RequestURI, tt.wantURI)
			}
			if got.Query != tt.wantQuery {
				t.Fatalf("query = %q, want %q", got.Query, tt.wantQuery)
			}
			if got.RequestID != tt.wantRequestID {
				t.Fatalf("request_id = %q, want %q", got.RequestID, tt.wantRequestID)
			}
		})
	}
}

func TestNewRequestLogServiceDefaultRetentionDays(t *testing.T) {
	service := NewRequestLogService(nil, config.TmdbLogConf{})
	if service.RetentionDays() != 7 {
		t.Fatalf("RetentionDays() = %d, want 7", service.RetentionDays())
	}
}

func TestNewRequestLogServiceKeepsConfiguredRetentionDays(t *testing.T) {
	service := NewRequestLogService(nil, config.TmdbLogConf{RetentionDays: 30})
	if service.RetentionDays() != 30 {
		t.Fatalf("RetentionDays() = %d, want 30", service.RetentionDays())
	}
}

func TestSetRetentionDaysNormalizesInvalidValue(t *testing.T) {
	service := NewRequestLogService(nil, config.TmdbLogConf{RetentionDays: 30})
	service.SetRetentionDays(0)
	if service.RetentionDays() != 7 {
		t.Fatalf("after SetRetentionDays(0) RetentionDays() = %d, want 7", service.RetentionDays())
	}

	service.SetRetentionDays(21)
	if service.RetentionDays() != 21 {
		t.Fatalf("after SetRetentionDays(21) RetentionDays() = %d, want 21", service.RetentionDays())
	}
}

func TestCleanupExpiredLogTables(t *testing.T) {
	tables := cleanupExpiredLogTables()
	want := []string{
		"proxy_access_logs",
		"tmdb_request_logs",
		"auto_sync_execution_logs",
	}
	if len(tables) != len(want) {
		t.Fatalf("cleanupExpiredLogTables() len = %d, want %d (%v)", len(tables), len(want), tables)
	}
	for i, name := range want {
		if tables[i] != name {
			t.Fatalf("cleanupExpiredLogTables()[%d] = %q, want %q", i, tables[i], name)
		}
	}
}

func TestIsAllowedCleanupTable(t *testing.T) {
	if !isAllowedCleanupTable("tmdb_request_logs") {
		t.Fatal("tmdb_request_logs should be allowed")
	}
	if isAllowedCleanupTable("movies") {
		t.Fatal("movies must not be allowed")
	}
}

func TestFormatBytes(t *testing.T) {
	if got := formatBytes(512); got != "512 B" {
		t.Fatalf("formatBytes(512) = %q, want 512 B", got)
	}
	if got := formatBytes(1536); got != "1.5 KB" {
		t.Fatalf("formatBytes(1536) = %q, want 1.5 KB", got)
	}
}

func TestReclaimSpaceEnabledFromConfig(t *testing.T) {
	enabled := NewRequestLogService(nil, config.TmdbLogConf{ReclaimSpace: true})
	if !enabled.ReclaimSpaceEnabled() {
		t.Fatal("ReclaimSpaceEnabled() = false, want true")
	}

	disabled := NewRequestLogService(nil, config.TmdbLogConf{ReclaimSpace: false})
	if disabled.ReclaimSpaceEnabled() {
		t.Fatal("ReclaimSpaceEnabled() = true, want false")
	}

	disabled.SetReclaimSpace(true)
	if !disabled.ReclaimSpaceEnabled() {
		t.Fatal("after SetReclaimSpace(true) want true")
	}
}
