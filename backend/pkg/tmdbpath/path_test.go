package tmdbpath

import (
	"regexp"
	"testing"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantPath string
		wantOK   bool
	}{
		{name: "api tmdb prefix only", path: "/api/tmdb", wantPath: "", wantOK: true},
		{name: "api tmdb movie detail", path: "/api/tmdb/movie/550", wantPath: "/movie/550", wantOK: true},
		{name: "api tmdb deep path", path: "/api/tmdb/a/b/c/d/e/f/g/h/i", wantPath: "/a/b/c/d/e/f/g/h/i", wantOK: true},
		{name: "api v3 movie detail", path: "/api/v3/movie/550", wantPath: "/movie/550", wantOK: true},
		{name: "api v3 prefix only", path: "/api/v3", wantPath: "", wantOK: true},
		{name: "v3 movie detail", path: "/v3/movie/550", wantPath: "/movie/550", wantOK: true},
		{name: "v3 deep path", path: "/v3/tv/1399/season/1/episode/2", wantPath: "/tv/1399/season/1/episode/2", wantOK: true},
		{name: "3 movie detail", path: "/3/movie/550", wantPath: "/movie/550", wantOK: true},
		{name: "3 tv season", path: "/3/tv/1399/season/1", wantPath: "/tv/1399/season/1", wantOK: true},
		{name: "similar prefix", path: "/api/tmdbxxx/movie/550", wantOK: false},
		{name: "similar v3 prefix", path: "/v3xxx/movie/550", wantOK: false},
		{name: "admin", path: "/api/admin/home", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotOK := Resolve(tt.path)
			if gotPath != tt.wantPath || gotOK != tt.wantOK {
				t.Fatalf("Resolve(%q) = (%q, %v), want (%q, %v)", tt.path, gotPath, gotOK, tt.wantPath, tt.wantOK)
			}
			if got := IsProxyPath(tt.path); got != tt.wantOK {
				t.Fatalf("IsProxyPath(%q) = %v, want %v", tt.path, got, tt.wantOK)
			}
		})
	}
}

func TestParseMediaTarget(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		wantMediaType string
		wantTmdbID    int
	}{
		// 多前缀 + 详情
		{name: "api tmdb movie detail", path: "/api/tmdb/movie/550", wantMediaType: "movie", wantTmdbID: 550},
		{name: "api v3 movie detail", path: "/api/v3/movie/550", wantMediaType: "movie", wantTmdbID: 550},
		{name: "v3 movie detail", path: "/v3/movie/550", wantMediaType: "movie", wantTmdbID: 550},
		{name: "3 movie detail", path: "/3/movie/299536", wantMediaType: "movie", wantTmdbID: 299536},
		// 嵌套路径
		{name: "api tmdb tv nested", path: "/api/tmdb/tv/1399/season/1", wantMediaType: "tv", wantTmdbID: 1399},
		{name: "v3 tv nested", path: "/v3/tv/1399/season/1", wantMediaType: "tv", wantTmdbID: 1399},
		{name: "3 tv nested", path: "/3/tv/1399/season/1/episode/2", wantMediaType: "tv", wantTmdbID: 1399},
		// 已是 Canonical
		{name: "canonical movie", path: "/movie/550", wantMediaType: "movie", wantTmdbID: 550},
		{name: "canonical movie images", path: "/movie/299536/images", wantMediaType: "movie", wantTmdbID: 299536},
		{name: "canonical tv nested", path: "/tv/1399/season/1", wantMediaType: "tv", wantTmdbID: 1399},
		// 负 id（本地约定）
		{name: "negative id", path: "/api/tmdb/movie/-1", wantMediaType: "movie", wantTmdbID: -1},
		// 搜索 / 非媒体详情
		{name: "search movie", path: "/api/tmdb/search/movie"},
		{name: "api v3 search", path: "/api/v3/search/movie"},
		{name: "person path", path: "/api/tmdb/person/1"},
		{name: "canonical search", path: "/search/movie"},
		// 仅前缀
		{name: "prefix only api tmdb", path: "/api/tmdb"},
		{name: "prefix only api v3", path: "/api/v3"},
		{name: "prefix only v3", path: "/v3"},
		{name: "prefix only 3", path: "/3"},
		// id 非法或 0
		{name: "invalid id latest", path: "/api/tmdb/tv/latest"},
		{name: "zero id", path: "/movie/0"},
		{name: "empty", path: ""},
		// 相似前缀误命中防护：不剥离前缀，路径首段不是 movie/tv
		{name: "similar prefix tmdbxxx", path: "/api/tmdbxxx/movie/550"},
		{name: "similar prefix v3xxx", path: "/v3xxx/movie/550"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMediaType, gotTmdbID := ParseMediaTarget(tt.path)
			if gotMediaType != tt.wantMediaType || gotTmdbID != tt.wantTmdbID {
				t.Fatalf("ParseMediaTarget(%q) = (%q, %d), want (%q, %d)",
					tt.path, gotMediaType, gotTmdbID, tt.wantMediaType, tt.wantTmdbID)
			}
		})
	}
}

func TestCanonicalizeForDisplay(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		// 多前缀剥离
		{name: "api tmdb movie", path: "/api/tmdb/movie/550", want: "/movie/550"},
		{name: "api v3 movie", path: "/api/v3/movie/550", want: "/movie/550"},
		{name: "v3 tv nested", path: "/v3/tv/1399/season/1", want: "/tv/1399/season/1"},
		{name: "3 search", path: "/3/search/movie", want: "/search/movie"},
		// 仅前缀 → 空串
		{name: "prefix only api tmdb", path: "/api/tmdb", want: ""},
		{name: "prefix only 3", path: "/3", want: ""},
		// 已是 Canonical 原样返回
		{name: "canonical movie", path: "/movie/550", want: "/movie/550"},
		{name: "canonical empty-ish", path: "/search/movie", want: "/search/movie"},
		// 未命中原样返回
		{name: "admin", path: "/api/admin/home", want: "/api/admin/home"},
		// 相似前缀误命中防护
		{name: "similar prefix tmdbxxx", path: "/api/tmdbxxx/movie/550", want: "/api/tmdbxxx/movie/550"},
		{name: "similar prefix v3xxx", path: "/v3xxx/movie/550", want: "/v3xxx/movie/550"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanonicalizeForDisplay(tt.path); got != tt.want {
				t.Fatalf("CanonicalizeForDisplay(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestPathRegexAlternation(t *testing.T) {
	got := PathRegexAlternation()
	want := "api/tmdb|api/v3|v3|3"
	if got != want {
		t.Fatalf("PathRegexAlternation() = %q, want %q", got, want)
	}
}

func TestMediaDetailPathRegex(t *testing.T) {
	re, err := regexp.Compile(MediaDetailPathRegex())
	if err != nil {
		t.Fatalf("编译 MediaDetailPathRegex 失败: %v", err)
	}

	matchCases := []string{
		"/api/tmdb/movie/550",
		"/api/v3/tv/1399/season/1",
		"/v3/movie/-1",
		"/3/tv/279446",
	}
	for _, path := range matchCases {
		if !re.MatchString(path) {
			t.Fatalf("MediaDetailPathRegex 未匹配 %q", path)
		}
	}

	missCases := []string{
		"/api/admin/home",
		"/api/tmdb/search/movie",
		"/api/v3/person/1",
		"/v3xxx/movie/550",
	}
	for _, path := range missCases {
		if re.MatchString(path) {
			t.Fatalf("MediaDetailPathRegex 误匹配 %q", path)
		}
	}
}

func TestMediaDetailPathReplaceRegex(t *testing.T) {
	re, err := regexp.Compile(MediaDetailPathReplaceRegex())
	if err != nil {
		t.Fatalf("编译 MediaDetailPathReplaceRegex 失败: %v", err)
	}

	matches := re.FindStringSubmatch("/api/v3/tv/1399/season/1")
	if len(matches) < 4 {
		t.Fatalf("捕获组不足: %#v", matches)
	}
	if matches[2] != "tv" || matches[3] != "1399" {
		t.Fatalf("提取结果 = media_type=%q id=%q, want tv/1399", matches[2], matches[3])
	}
}
