package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ms_tmdb/internal/model"
	"ms_tmdb/pkg/tmdbclient"
)

func genreIDs(genres []map[string]interface{}) []int {
	ids := make([]int, 0, len(genres))
	for _, genre := range genres {
		ids = append(ids, genre["id"].(int))
	}
	return ids
}

func TestBuildGenresWithOfficialIDsResolvesBuiltinNames(t *testing.T) {

	// TMDB 不可达（client 为 nil）时也应通过内置对照表还原官方 ID
	genres := buildGenresWithOfficialIDs(nil, genreMediaTypeMovie, []string{"动作", "科幻", "惊悚"}, nil)

	got := genreIDs(genres)
	want := []int{28, 878, 53}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("genre[%d] id = %d, want %d（genres=%v）", i, got[i], want[i], genres)
		}
	}
}

func TestBuildGenresWithOfficialIDsResolvesEnglishNames(t *testing.T) {

	genres := buildGenresWithOfficialIDs(nil, genreMediaTypeMovie, []string{"Science Fiction", "Drama"}, nil)
	got := genreIDs(genres)
	if got[0] != 878 || got[1] != 18 {
		t.Fatalf("ids = %v, want [878 18]", got)
	}
}

func TestBuildGenresWithOfficialIDSMediaTypeAware(t *testing.T) {

	movieGenres := buildGenresWithOfficialIDs(nil, genreMediaTypeMovie, []string{"动作冒险"}, nil)
	if len(movieGenres) != 1 || movieGenres[0]["id"].(int) >= 0 {
		t.Fatalf("电影类型表不应包含 TV 专有类型 10759，got %v", movieGenres)
	}

	tvGenres := buildGenresWithOfficialIDs(nil, genreMediaTypeTV, []string{"动作冒险"}, nil)
	if tvGenres[0]["id"].(int) != 10759 {
		t.Fatalf("TV 类型 动作冒险 应解析为 10759，got %v", tvGenres)
	}
}

func TestBuildGenresWithOfficialIDsPreservesExistingIDs(t *testing.T) {

	// 条目已带官方 ID 时，即使名称不在官方列表（如 TMDB 不可达且为其他语言）也应保留原 ID
	existing := []interface{}{
		map[string]interface{}{"id": float64(18), "name": "剧情"},
		map[string]interface{}{"id": float64(10759), "name": "Action & Adventure"},
	}
	genres := buildGenresWithOfficialIDs(nil, genreMediaTypeTV, []string{"剧情", "Action & Adventure"}, existing)

	got := genreIDs(genres)
	if got[0] != 18 || got[1] != 10759 {
		t.Fatalf("ids = %v, want [18 10759]", got)
	}
}

func TestBuildGenresWithOfficialIDsReproUnchangedGenreSave(t *testing.T) {

	// 复现报告场景：干净条目（tmdb_data/local_data 均为官方 ID）仅改标题，
	// Web UI 保存时原样回传 genre_names，genres 必须保持官方 ID 不被重编号。
	recordJSON := model.RawJSON(`{"id":497698,"title":"黑寡妇","genres":[{"id":18,"name":"剧情"},{"id":12,"name":"冒险"},{"id":35,"name":"喜剧"}]}`)
	// tmdb_data 与 local_data 同内容时合并去重靠后续同名取首个 ID
	existing := genresFromRawJSON(recordJSON, recordJSON)
	if len(existing) != 6 {
		t.Fatalf("genresFromRawJSON 应合并两份数组共 6 项，got %d", len(existing))
	}

	genres := buildGenresWithOfficialIDs(nil, genreMediaTypeMovie, []string{"剧情", "冒险", "喜剧"}, existing)
	got := genreIDs(genres)
	want := []int{18, 12, 35}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("genre[%d] id = %d, want %d（genres=%v）", i, got[i], want[i], genres)
		}
	}
}

func TestBuildGenresWithOfficialIDsHealsPoisonedIDs(t *testing.T) {

	// 复现报告场景：已被重编号为 1,2,3 的条目再次保存时，官方 ID 应优先于现有坏 ID
	recordJSON := model.RawJSON(`{"id":497698,"title":"黑寡妇","genres":[{"id":1,"name":"动作"},{"id":2,"name":"冒险"},{"id":3,"name":"科幻"}]}`)
	existing := genresFromRawJSON(recordJSON, recordJSON)

	genres := buildGenresWithOfficialIDs(nil, genreMediaTypeMovie, []string{"动作", "冒险", "科幻"}, existing)
	got := genreIDs(genres)
	want := []int{28, 12, 878}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("genre[%d] id = %d, want %d（genres=%v）", i, got[i], want[i], genres)
		}
	}
}

func TestBuildGenresWithOfficialIDsAssignsNegativePlaceholderForCustomNames(t *testing.T) {

	genres := buildGenresWithOfficialIDs(nil, genreMediaTypeMovie, []string{"自定义风格", "王家卫式"}, nil)
	got := genreIDs(genres)
	if got[0] >= 0 || got[1] >= 0 {
		t.Fatalf("自定义类型名应获得负数占位 ID，got %v", got)
	}
	if got[0] == got[1] {
		t.Fatalf("不同自定义类型名不应共享占位 ID，got %v", got)
	}
}

func TestBuildGenresWithOfficialIDsDedupesNames(t *testing.T) {

	// 同名不同大小写/空白应去重；不同语言名称保留，但都应解析到官方 ID
	genres := buildGenresWithOfficialIDs(nil, genreMediaTypeMovie, []string{"动作", " 动作 ", "ACTION"}, nil)
	if len(genres) != 2 {
		t.Fatalf("同名重复应去重为 2 项，got %v", genres)
	}
	for _, genre := range genres {
		if genre["id"].(int) != 28 {
			t.Fatalf("id = %v, want 28", genre["id"])
		}
	}
}

func TestBuildGenresWithOfficialIDsEmptyNames(t *testing.T) {

	genres := buildGenresWithOfficialIDs(nil, genreMediaTypeMovie, []string{"", "   "}, nil)
	if len(genres) != 0 {
		t.Fatalf("空名称应产出空数组，got %v", genres)
	}
}

func TestBuildGenresWithOfficialIDsFetchesOfficialList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/genre/movie/list" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"genres": []map[string]interface{}{
				{"id": 28, "name": "动作"},
				{"id": 878, "name": "科幻"},
			},
		})
	}))
	defer server.Close()

	client := tmdbclient.NewClient("test-key", server.URL, "zh-CN", 1000, "")
	genres := buildGenresWithOfficialIDs(client, genreMediaTypeMovie, []string{"动作", "科幻"}, nil)
	got := genreIDs(genres)
	if got[0] != 28 || got[1] != 878 {
		t.Fatalf("ids = %v, want [28 878]", got)
	}
}

func TestBuildGenresWithOfficialIDSFallsBackWhenUpstreamFails(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := tmdbclient.NewClient("test-key", server.URL, "zh-CN", 1000, "")
	genres := buildGenresWithOfficialIDs(client, genreMediaTypeMovie, []string{"动作", "自定义"}, nil)
	got := genreIDs(genres)
	if got[0] != 28 {
		t.Fatalf("上游失败时应降级内置对照表，id = %d, want 28", got[0])
	}
	if got[1] >= 0 {
		t.Fatalf("未识别名称应为负数占位 ID，got %d", got[1])
	}
}

func TestNormalizeGenreKey(t *testing.T) {
	cases := []struct {
		left, right string
	}{
		{"Sci-Fi & Fantasy", "sci-fi & fantasy"},
		{"科幻&奇幻", "科幻奇幻"},
		{"Action", "action"},
		{"TV Movie", "tvmovie"},
	}
	for _, c := range cases {
		if normalizeGenreKey(c.left) != normalizeGenreKey(c.right) {
			t.Fatalf("normalizeGenreKey(%q) != normalizeGenreKey(%q)", c.left, c.right)
		}
	}
}
