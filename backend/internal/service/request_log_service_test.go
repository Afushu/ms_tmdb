package service

import "testing"

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
