package admin

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"ms_tmdb/internal/model"
	"ms_tmdb/pkg/tmdbclient"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	genreMediaTypeMovie = "movie"
	genreMediaTypeTV    = "tv"

	// genreLookupTimeout 限制官方类型表的在线获取时长，避免 TMDB 不可达时拖住管理端保存。
	genreLookupTimeout = 3 * time.Second
)

// buildGenresWithOfficialIDs 把类型名重建为详情页可渲染结构，ID 按以下优先级还原：
// 1. TMDB 官方类型表（在线获取，进程内缓存；失败时降级内置对照表）
// 2. 条目现有 tmdb_data / local_data 中同名类型的 ID（保持已有数据稳定）
// 3. 自定义类型名分配负数占位 ID，避免伪造成官方 ID 被下游白名单误匹配
//
// 管理端只接收 genre_names（不收 ID），若在这里按下标编号（1,2,3…），
// 下游按官方 ID 白名单分类的消费方会静默误判。
func buildGenresWithOfficialIDs(client *tmdbclient.Client, mediaType string, names []string, existing []interface{}) []map[string]interface{} {
	orderedNames := dedupeGenreNames(names)
	if len(orderedNames) == 0 {
		return []map[string]interface{}{}
	}

	official := officialGenreLookup(client, mediaType)
	existingByName := genreIDsFromExisting(existing)
	usedPlaceholders := make(map[int]struct{})
	for _, id := range existingByName {
		if id <= 0 {
			usedPlaceholders[id] = struct{}{}
		}
	}

	result := make([]map[string]interface{}, 0, len(orderedNames))
	nextPlaceholder := -1
	for _, name := range orderedNames {
		id, ok := official[normalizeGenreKey(name)]
		if !ok {
			id, ok = existingByName[normalizeGenreKey(name)]
		}
		if !ok {
			for {
				if _, clash := usedPlaceholders[nextPlaceholder]; !clash {
					break
				}
				nextPlaceholder--
			}
			id = nextPlaceholder
			usedPlaceholders[id] = struct{}{}
			nextPlaceholder--
		}
		result = append(result, map[string]interface{}{
			"id":   id,
			"name": name,
		})
	}
	return result
}

// officialGenreLookup 返回 名称→官方ID 查找表，键为 normalizeGenreKey 归一化结果。
func officialGenreLookup(client *tmdbclient.Client, mediaType string) map[string]int {
	byName := builtinGenreLookup(mediaType)
	fetched := fetchOfficialGenreList(client, mediaType)
	if len(fetched) > 0 {
		// 在线结果优先，内置表仅补位（例如其他语言的别名）。
		for key, id := range fetched {
			byName[key] = id
		}
	}
	return byName
}

// fetchOfficialGenreList 从 TMDB 拉取官方类型表；任何失败都返回 nil，由调用方降级。
func fetchOfficialGenreList(client *tmdbclient.Client, mediaType string) map[string]int {
	if client == nil {
		return nil
	}

	path := "/genre/movie/list"
	if mediaType == genreMediaTypeTV {
		path = "/genre/tv/list"
	}

	ctx, cancel := context.WithTimeout(context.Background(), genreLookupTimeout)
	defer cancel()
	raw, err := client.Get(path, &tmdbclient.RequestOption{Context: ctx})
	if err != nil {
		logx.Errorf("获取 TMDB 官方类型列表失败(%s): %v", mediaType, err)
		return nil
	}

	var payload struct {
		Genres []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"genres"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		logx.Errorf("解析 TMDB 官方类型列表失败(%s): %v", mediaType, err)
		return nil
	}

	byName := make(map[string]int, len(payload.Genres))
	for _, genre := range payload.Genres {
		name := strings.TrimSpace(genre.Name)
		if name == "" || genre.ID <= 0 {
			continue
		}
		byName[normalizeGenreKey(name)] = genre.ID
	}
	return byName
}

// genreIDsFromExisting 从条目现有 genres 数组提取 名称→ID，同名取首个出现的 ID。
func genreIDsFromExisting(existing []interface{}) map[string]int {
	byName := make(map[string]int, len(existing))
	for _, item := range existing {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name := strings.TrimSpace(mapString(entry, "name"))
		id := mapInt(entry, "id")
		if name == "" || id == 0 {
			continue
		}
		key := normalizeGenreKey(name)
		if _, exists := byName[key]; !exists {
			byName[key] = id
		}
	}
	return byName
}

// genresFromRawJSON 汇总条目 tmdb_data / local_data 中的 genres 数组。
func genresFromRawJSON(tmdbData, localData model.RawJSON) []interface{} {
	merged := make([]interface{}, 0)
	for _, raw := range []model.RawJSON{tmdbData, localData} {
		if len(raw) == 0 {
			continue
		}
		payload := make(map[string]interface{})
		if err := json.Unmarshal(raw, &payload); err != nil {
			continue
		}
		if genres, ok := payload["genres"].([]interface{}); ok {
			merged = append(merged, genres...)
		}
	}
	return merged
}

// dedupeGenreNames 去空格、去空项、按名称大小写不敏感去重，保持首次出现顺序。
func dedupeGenreNames(names []string) []string {
	result := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, name)
	}
	return result
}

// normalizeGenreKey 归一化类型名用于查找：忽略大小写、空白与连接符差异，
// 使 "Sci-Fi & Fantasy"、"科幻&奇幻"、"科幻 奇幻" 互相匹配。
func normalizeGenreKey(name string) string {
	key := strings.ToLower(strings.TrimSpace(name))
	key = strings.ReplaceAll(key, "&", "")
	key = strings.ReplaceAll(key, " ", "")
	return key
}

// builtinGenreNames TMDB 官方 genre ID 与常见名称（zh-CN / en-US）的对照，
// 供 TMDB 不可达（离线本地编辑、上游故障）时兜底。键为 normalizeGenreKey 归一化结果。
var builtinGenreNames = map[int][]string{
	28:    {"动作", "action"},
	12:    {"冒险", "adventure"},
	16:    {"动画", "animation"},
	35:    {"喜剧", "comedy"},
	80:    {"犯罪", "crime"},
	99:    {"纪录片", "documentary"},
	18:    {"剧情", "drama"},
	10751: {"家庭", "family"},
	14:    {"奇幻", "fantasy"},
	36:    {"历史", "history"},
	27:    {"恐怖", "horror"},
	10402: {"音乐", "music"},
	9648:  {"悬疑", "mystery"},
	10749: {"爱情", "romance"},
	878:   {"科幻", "science fiction", "sciencefiction"},
	10770: {"电视电影", "tv movie"},
	53:    {"惊悚", "thriller"},
	10752: {"战争", "war"},
	37:    {"西部片", "西部", "western"},
	10759: {"动作冒险", "动作冒险动画", "action adventure", "actionadventure"},
	10762: {"儿童", "kids"},
	10763: {"新闻", "news"},
	10764: {"真人秀", "reality"},
	10765: {"科幻奇幻", "sci-fifantasy", "scififantasy"},
	10766: {"肥皂剧", "soap"},
	10767: {"脱口秀", "talk"},
	10768: {"战争政治", "war politics"},
}

// builtinGenreIDs 各媒体类型的官方 genre ID 集合，与 TMDB /genre/{movie,tv}/list 一致。
var builtinGenreIDs = map[string][]int{
	genreMediaTypeMovie: {28, 12, 16, 35, 80, 99, 18, 10751, 14, 36, 27, 10402, 9648, 10749, 878, 10770, 53, 10752, 37},
	genreMediaTypeTV:    {10759, 16, 35, 80, 99, 18, 10751, 10762, 9648, 10763, 10764, 10765, 10766, 10767, 10768, 37},
}

// builtinGenreLookup 返回指定媒体类型的 名称→官方ID 对照表。
func builtinGenreLookup(mediaType string) map[string]int {
	byName := make(map[string]int, len(builtinGenreNames))
	for _, id := range builtinGenreIDs[mediaType] {
		for _, alias := range builtinGenreNames[id] {
			byName[normalizeGenreKey(alias)] = id
		}
	}
	return byName
}
