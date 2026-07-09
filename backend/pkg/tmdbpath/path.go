package tmdbpath

import (
	"fmt"
	"strconv"
	"strings"
)

// Prefixes 后端原生兼容的 TMDB 代理入口，按长到短排列。
// 匹配时使用 exact 或 prefix+"/" 边界，避免 /api/tmdbxxx 之类误命中。
var Prefixes = []string{
	"/api/tmdb",
	"/api/v3",
	"/v3",
	"/3",
}

// IsProxyPath 判断 path 是否为 TMDB 代理入口。
func IsProxyPath(path string) bool {
	_, ok := Resolve(path)
	return ok
}

// Resolve 剥离兼容入口前缀，返回 TMDB 相对路径（如 /movie/550）。
// 仅命中代理入口时 ok=true；未命中时返回原 path 语义外的空串。
func Resolve(path string) (tmdbPath string, ok bool) {
	for _, prefix := range Prefixes {
		if path == prefix {
			return "", true
		}
		if strings.HasPrefix(path, prefix+"/") {
			return strings.TrimPrefix(path, prefix), true
		}
	}
	return "", false
}

// ParseMediaTarget 从任意代理 path 或已规范 path 提取 movie/tv 详情目标。
// 规则：先 Resolve（若命中入口），再取首两段 media_type + 数字 id；
// media_type 仅 movie|tv；id 解析失败或 0 则返回 ("", 0)。
// 嵌套路径如 /tv/1399/season/1 提取 tv/1399。
func ParseMediaTarget(path string) (mediaType string, tmdbID int) {
	tmdbPath := strings.TrimSpace(path)
	if tmdbPath == "" {
		return "", 0
	}

	if stripped, ok := Resolve(tmdbPath); ok {
		tmdbPath = stripped
	}

	parts := strings.Split(strings.Trim(tmdbPath, "/"), "/")
	if len(parts) < 2 {
		return "", 0
	}

	mediaType = parts[0]
	if mediaType != "movie" && mediaType != "tv" {
		return "", 0
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil || id == 0 {
		return "", 0
	}
	return mediaType, id
}

// CanonicalizeForDisplay 将可能带入口前缀的 path 规范为主展示路径。
// 命中入口则返回 Resolve 相对路径（仅前缀时为空串，与 Resolve 约定一致）；
// 未命中则原样返回（兼容已是 Canonical 或非代理 path）。
func CanonicalizeForDisplay(path string) string {
	if stripped, ok := Resolve(path); ok {
		return stripped
	}
	return path
}

// PathRegexAlternation 返回不含前导斜杠的前缀交替串，仅供路径规则测试使用。
// 业务 SQL 禁止使用 path 正则回退解析媒体目标。
// 例：api/tmdb|api/v3|v3|3
func PathRegexAlternation() string {
	parts := make([]string, 0, len(Prefixes))
	for _, prefix := range Prefixes {
		parts = append(parts, strings.TrimPrefix(prefix, "/"))
	}
	return strings.Join(parts, "|")
}

// MediaDetailPathRegex 匹配代理入口下 movie/tv 详情路径，仅供路径规则测试使用。
// 业务 SQL 禁止使用该正则恢复热门统计 path 回退。
// 例：^/(api/tmdb|api/v3|v3|3)/(movie|tv)/(-?[0-9]+)($|/.*$)
func MediaDetailPathRegex() string {
	return fmt.Sprintf(`^/(%s)/(movie|tv)/(-?[0-9]+)($|/.*$)`, PathRegexAlternation())
}

// MediaDetailPathReplaceRegex 用于测试正则捕获组，不得用于业务 SQL。
// 捕获组：1=prefix, 2=media_type, 3=tmdb_id
func MediaDetailPathReplaceRegex() string {
	return fmt.Sprintf(`^/(%s)/(movie|tv)/(-?[0-9]+).*$`, PathRegexAlternation())
}
