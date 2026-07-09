package tmdbpath

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// Feature: proxy-path-hot-stats-log-fix, Property 3: 展示规范化不引入错误前缀
//
// Property 3: 展示规范化不引入错误前缀
// *For any* path 字符串：若其命中代理入口，则 CanonicalizeForDisplay(path) 等于 Resolve 的相对路径
// （或约定的空串）；若未命中，则输出等于输入。因而两条仅入口前缀不同、相对路径相同的原始 path，
// 规范化后主展示文本相同。
//
// **Validates: Requirements 3.3, 6.3**
func TestProperty3_CanonicalizeForDisplayNoWrongPrefix(t *testing.T) {
	const iterations = 100
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < iterations; i++ {
		// --- 分支 A：命中入口 → 等于 Resolve 结果；仅前缀不同则展示相同 ---
		relative := genRelativeResourcePath(rng)
		var displays []string
		for _, prefix := range Prefixes {
			var full string
			if relative == "" {
				full = prefix
			} else {
				full = prefix + relative
			}

			stripped, ok := Resolve(full)
			if !ok {
				t.Fatalf("iter %d: Resolve(%q) should hit entry prefix %q", i, full, prefix)
			}
			got := CanonicalizeForDisplay(full)
			if got != stripped {
				t.Fatalf("iter %d: hit CanonicalizeForDisplay(%q)=%q, want Resolve result %q",
					i, full, got, stripped)
			}
			// 命中后不得残留任何已声明入口前缀
			for _, p := range Prefixes {
				if got == p || (len(got) > len(p) && got[:len(p)+1] == p+"/") {
					t.Fatalf("iter %d: display %q still carries entry prefix %q (from %q)",
						i, got, p, full)
				}
			}
			displays = append(displays, got)
		}
		// 同一相对路径 + 不同前缀 → 主展示相同
		for j := 1; j < len(displays); j++ {
			if displays[j] != displays[0] {
				t.Fatalf("iter %d: same relative %q normalized differently: %q vs %q",
					i, relative, displays[0], displays[j])
			}
		}
		// 期望展示为 relative 本身（仅前缀时 relative 为空串）
		if displays[0] != relative {
			t.Fatalf("iter %d: normalized display %q != relative %q", i, displays[0], relative)
		}

		// --- 分支 B：未命中入口 → 恒等（原样返回）---
		miss := genMissPath(rng)
		if _, ok := Resolve(miss); ok {
			// 生成器偶发与前缀边界重合时跳过恒等断言，避免假失败
			continue
		}
		if got := CanonicalizeForDisplay(miss); got != miss {
			t.Fatalf("iter %d: miss CanonicalizeForDisplay(%q)=%q, want identity", i, miss, got)
		}
	}
}

// genMissPath 生成应未命中代理入口的 path（相似前缀、admin、已是 Canonical、杂项）。
func genMissPath(rng *rand.Rand) string {
	switch rng.Intn(5) {
	case 0:
		// 相似前缀误命中防护
		return pick(rng,
			"/api/tmdbxxx/movie/550",
			"/api/tmdbxxx",
			"/v3xxx/movie/550",
			"/3xxx/tv/1",
			"/api/v3xxx/search/movie",
		)
	case 1:
		// 非代理业务 path
		return pick(rng,
			"/api/admin/home",
			"/api/admin/logs",
			"/health",
			"/metrics",
			"/",
			"",
		)
	case 2:
		// 已是 Canonical 相对路径（无入口前缀，Resolve 不命中）
		return genRelativeResourcePath(rng)
	case 3:
		// 带 query 形态的字符串（path 组件本身不应含 query，作杂项恒等）
		return pick(rng,
			"/movie/550?language=zh-CN",
			"not-a-path",
			"movie/550",
			"//movie/550",
		)
	default:
		// 随机杂项段
		return fmt.Sprintf("/%s/%s",
			pick(rng, "admin", "internal", "webhook", "static", "assets"),
			randomSegment(rng),
		)
	}
}
