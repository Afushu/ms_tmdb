package tmdbpath

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Feature: proxy-path-hot-stats-log-fix, Property 1: 前缀剥离幂等与一致性
//
// Property 1: 前缀剥离幂等与一致性
// *For any* 由已声明 Proxy_Entry_Prefix 与任意相对后缀（含空）拼接而成的合法代理 path，
// Resolve 返回的 Canonical 路径在再次 Resolve 时要么不命中入口（ok=false），要么得到与首次相同的相对路径语义；
// 且对全部前缀使用同一边界规则（path == prefix 或 HasPrefix(path, prefix+"/")），相似前缀不得命中。
//
// **Validates: Requirements 1.1, 1.4, 5.1, 5.4**
func TestProperty1_PrefixStripIdempotentAndConsistent(t *testing.T) {
	const iterations = 100
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	if len(Prefixes) == 0 {
		t.Fatal("Prefixes 为空，无法验证前缀剥离")
	}

	for i := 0; i < iterations; i++ {
		// 覆盖全部 Prefixes：轮转 + 随机，保证每个前缀都参与
		prefix := Prefixes[i%len(Prefixes)]
		if rng.Intn(2) == 0 {
			prefix = Prefixes[rng.Intn(len(Prefixes))]
		}

		// 相对后缀：空（exact 边界）或 TMDB 相对路径（prefix+"/" 边界）
		// 约束：后缀本身不得再是代理入口，以保证 Canonical 语义稳定
		suffix := genCanonicalRelativeSuffix(rng)
		var full string
		var wantCanonical string
		if suffix == "" {
			full = prefix // exact 边界
			wantCanonical = ""
		} else {
			full = prefix + suffix // suffix 以 / 开头 → 满足 prefix+"/" 边界
			wantCanonical = suffix
		}

		got, ok := Resolve(full)
		if !ok {
			t.Fatalf("iter %d: Resolve(%q) ok=false, want true (prefix=%q suffix=%q)", i, full, prefix, suffix)
		}
		if got != wantCanonical {
			t.Fatalf("iter %d: Resolve(%q)=%q, want %q", i, full, got, wantCanonical)
		}

		// 二次 Resolve：合法 Canonical 相对路径不应再命中入口；
		// 若命中（极端情形），结果须与首次 Canonical 相同
		got2, ok2 := Resolve(got)
		if ok2 {
			if got2 != got {
				t.Fatalf("iter %d: second Resolve(%q)=(%q, true), want same as first or ok=false", i, got, got2)
			}
		} else if got2 != "" {
			t.Fatalf("iter %d: second Resolve(%q)=(%q, false), empty path expected when ok=false", i, got, got2)
		}

		// 边界规则一致性：对每个前缀，仅 exact 或 prefix+"/" 命中
		for _, p := range Prefixes {
			assertBoundaryRule(t, i, p, full)
		}

		// 相似前缀不得命中
		similar := genSimilarNonMatchingPath(rng, prefix, suffix)
		if sPath, sOK := Resolve(similar); sOK {
			t.Fatalf("iter %d: similar path %q 误命中 Resolve → (%q, true)", i, similar, sPath)
		}
		if IsProxyPath(similar) {
			t.Fatalf("iter %d: similar path %q IsProxyPath=true, want false", i, similar)
		}
	}

	// 确定性边界：每个 Prefix 的 exact 与 prefix+"/x" 及相似串
	for _, prefix := range Prefixes {
		if got, ok := Resolve(prefix); !ok || got != "" {
			t.Fatalf("exact Resolve(%q)=(%q,%v), want (\"\", true)", prefix, got, ok)
		}
		// 二次：空串不命中
		if got2, ok2 := Resolve(""); ok2 || got2 != "" {
			t.Fatalf("second Resolve(\"\")=(%q,%v), want (\"\", false)", got2, ok2)
		}

		withSlash := prefix + "/movie/550"
		if got, ok := Resolve(withSlash); !ok || got != "/movie/550" {
			t.Fatalf("Resolve(%q)=(%q,%v), want (/movie/550, true)", withSlash, got, ok)
		}
		if got2, ok2 := Resolve("/movie/550"); ok2 {
			t.Fatalf("second Resolve(/movie/550) should not match entry, got (%q, true)", got2)
		}

		// 无分隔边界的相似前缀
		similarExact := prefix + "xxx"
		if _, ok := Resolve(similarExact); ok {
			t.Fatalf("similar %q should not match", similarExact)
		}
		similarWithPath := prefix + "xxx/movie/550"
		if _, ok := Resolve(similarWithPath); ok {
			t.Fatalf("similar %q should not match", similarWithPath)
		}
	}
}

// assertBoundaryRule 验证 path 对给定 prefix 的命中仅当 exact 或 prefix+"/"。
func assertBoundaryRule(t *testing.T, iter int, prefix, path string) {
	t.Helper()
	matchesBoundary := path == prefix || strings.HasPrefix(path, prefix+"/")
	// 反向：若 path 以 prefix 开头但不是 exact 也不是 prefix+"/"，则该 prefix 不得单独导致命中
	if strings.HasPrefix(path, prefix) && !matchesBoundary {
		// 例如 /api/tmdbxxx — 不得因本 prefix 命中
		// 仍可能被更短前缀命中（极少），但对当前 Prefixes 列表不会
		got, ok := Resolve(path)
		if ok {
			// 若命中，必须是其它合法前缀的边界匹配，而非本相似前缀
			matchedOther := false
			for _, p := range Prefixes {
				if path == p || strings.HasPrefix(path, p+"/") {
					matchedOther = true
					break
				}
			}
			if !matchedOther {
				t.Fatalf("iter %d: path %q matched Resolve(%q) without boundary rule", iter, path, got)
			}
		}
	}
}

// genCanonicalRelativeSuffix 生成不会再被 Resolve 识别为入口的相对后缀（含空）。
func genCanonicalRelativeSuffix(rng *rand.Rand) string {
	for attempt := 0; attempt < 16; attempt++ {
		suffix := genRelativeResourcePath(rng)
		if suffix == "" {
			return ""
		}
		// 拒绝自身可被剥离的后缀（如偶然以 /3/... 开头）
		if _, ok := Resolve(suffix); ok {
			continue
		}
		// 确保拼接后满足 prefix+"/" 形态
		if !strings.HasPrefix(suffix, "/") {
			suffix = "/" + suffix
		}
		if _, ok := Resolve(suffix); ok {
			continue
		}
		return suffix
	}
	// 回退：稳定的非入口相对路径
	return fmt.Sprintf("/movie/%d", 1+rng.Intn(10000))
}

// genSimilarNonMatchingPath 构造「以 prefix 为字符串前缀但缺少 / 边界」的相似 path。
func genSimilarNonMatchingPath(rng *rand.Rand, prefix, suffix string) string {
	glue := pick(rng, "xxx", "api", "x", "2", "_extra")
	if suffix == "" {
		return prefix + glue
	}
	// 保持相对资源形态，但破坏边界
	return prefix + glue + suffix
}

// Feature: proxy-path-hot-stats-log-fix, Property 2: Media_Target 提取与 Canonical 无关入口
//
// Property 2: Media_Target 提取与 Canonical 无关入口
// *For any* 同一 TMDB 相对资源路径与任一已声明入口前缀的组合，
// ParseMediaTarget 得到的 (media_type, tmdb_id) 与对 Canonical 路径直接解析的结果相同；
// 当路径不是 movie/tv 详情目标时，结果恒为 ("", 0)。
//
// **Validates: Requirements 1.2, 1.3, 5.2, 5.4**
func TestProperty2_MediaTargetIndependentOfEntryPrefix(t *testing.T) {
	const iterations = 100
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		canonical := genRelativeResourcePath(rng)
		wantMedia, wantID := ParseMediaTarget(canonical)

		// 非 movie/tv 详情目标：Canonical 与带前缀均应为 ("", 0)
		if !isMovieOrTVDetail(canonical) {
			if wantMedia != "" || wantID != 0 {
				t.Fatalf("iter %d: non-detail canonical %q => (%q, %d), want (\"\", 0)",
					i, canonical, wantMedia, wantID)
			}
		}

		// 直接解析 Canonical 与「任意前缀 + 同一相对资源」结果一致
		for _, prefix := range Prefixes {
			var full string
			if canonical == "" {
				full = prefix
			} else {
				full = prefix + canonical
			}
			gotMedia, gotID := ParseMediaTarget(full)
			if gotMedia != wantMedia || gotID != wantID {
				t.Fatalf("iter %d: ParseMediaTarget(%q)=(%q,%d) != ParseMediaTarget(%q)=(%q,%d)",
					i, full, gotMedia, gotID, canonical, wantMedia, wantID)
			}
		}

		// 无前缀（已是 Canonical）再断言一次，保证基线稳定
		gotMedia, gotID := ParseMediaTarget(canonical)
		if gotMedia != wantMedia || gotID != wantID {
			t.Fatalf("iter %d: unstable ParseMediaTarget(%q)", i, canonical)
		}
	}
}

// isMovieOrTVDetail 判断 Canonical 相对路径是否为 movie/tv 详情目标
// （首段 movie|tv 且第二段为非 0 整数）。
func isMovieOrTVDetail(canonical string) bool {
	parts := strings.Split(strings.Trim(canonical, "/"), "/")
	if len(parts) < 2 {
		return false
	}
	if parts[0] != "movie" && parts[0] != "tv" {
		return false
	}
	id, err := strconv.Atoi(parts[1])
	return err == nil && id != 0
}

// genRelativeResourcePath 生成 TMDB 相对资源路径（可为空串表示仅前缀场景）。
func genRelativeResourcePath(rng *rand.Rand) string {
	switch rng.Intn(8) {
	case 0:
		// 空：仅入口前缀
		return ""
	case 1:
		// movie 详情
		return fmt.Sprintf("/movie/%d", genNonZeroID(rng))
	case 2:
		// tv 详情
		return fmt.Sprintf("/tv/%d", genNonZeroID(rng))
	case 3:
		// 嵌套 movie/tv
		media := pick(rng, "movie", "tv")
		id := genNonZeroID(rng)
		suffix := pick(rng,
			"/images",
			"/credits",
			"/season/1",
			"/season/1/episode/2",
			"/videos",
		)
		return fmt.Sprintf("/%s/%d%s", media, id, suffix)
	case 4:
		// 搜索 / 非媒体
		return pick(rng,
			"/search/movie",
			"/search/tv",
			"/person/1",
			"/genre/movie/list",
			"/configuration",
			"/discover/movie",
		)
	case 5:
		// 非法 id / 0
		media := pick(rng, "movie", "tv")
		return pick(rng,
			fmt.Sprintf("/%s/0", media),
			fmt.Sprintf("/%s/latest", media),
			fmt.Sprintf("/%s/abc", media),
			fmt.Sprintf("/%s/", media),
		)
	case 6:
		// 仅单段
		return pick(rng, "/movie", "/tv", "/search", "/trending")
	default:
		// 随机深度路径（可能或不可能是详情）
		depth := 1 + rng.Intn(4)
		segs := make([]string, 0, depth)
		for j := 0; j < depth; j++ {
			segs = append(segs, randomSegment(rng))
		}
		return "/" + strings.Join(segs, "/")
	}
}

func genNonZeroID(rng *rand.Rand) int {
	// 覆盖正/负 id；排除 0
	id := rng.Intn(2_000_000) - 500_000
	if id == 0 {
		id = 1
	}
	return id
}

func randomSegment(rng *rand.Rand) string {
	return pick(rng,
		"movie", "tv", "search", "person", "genre", "discover",
		"images", "credits", "season", "episode", "latest",
		strconv.Itoa(rng.Intn(10_000)),
		fmt.Sprintf("seg%d", rng.Intn(100)),
	)
}

func pick(rng *rand.Rand, opts ...string) string {
	return opts[rng.Intn(len(opts))]
}
