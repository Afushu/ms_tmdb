package admin

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Feature: proxy-path-hot-stats-log-fix, Property 4: 热门聚合忽略无效 Media_Target
//
// Property 4: 热门聚合忽略无效 Media_Target
// *For any* 访问日志记录集合，在应用热门过滤（近 30 天、状态码 [200,400)、media_type∈{movie,tv}、tmdb_id≠0）后，
// 聚合键仅为 (media_type, tmdb_id)，同一键的访问次数等于满足条件的记录条数；
// media_type/tmdb_id 无效的记录不贡献计数（无论 path 是否可被正则解析）。
//
// **Validates: Requirements 2.1, 2.2, 2.3, 2.6, 2.7, 6.2, 6.4**
func TestProperty4_HotAggregationIgnoresInvalidMediaTarget(t *testing.T) {
	const iterations = 100
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	now := time.Date(2025, 1, 31, 12, 0, 0, 0, time.UTC)

	for i := 0; i < iterations; i++ {
		records := genHotMediaLogRecords(rng, now)
		got := aggregateValidHotMediaTargets(records, now)
		want := expectedHotMediaCounts(records, now)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("iter %d: aggregateValidHotMediaTargets()=%v, want %v, records=%#v", i, got, want, records)
		}
	}

	// 确定性覆盖：path 可解析但 Media_Target 字段无效时，仍不计入热门聚合。
	records := []hotMediaLogRecord{
		{StatusCode: 200, CreatedAt: now, MediaType: "movie", TmdbID: 550, Path: "/movie/550"},
		{StatusCode: 201, CreatedAt: now, MediaType: "movie", TmdbID: 550, Path: "/movie/550"},
		{StatusCode: 200, CreatedAt: now, MediaType: "", TmdbID: 0, Path: "/movie/550"},
		{StatusCode: 200, CreatedAt: now, MediaType: "person", TmdbID: 550, Path: "/movie/550"},
		{StatusCode: 200, CreatedAt: now, MediaType: "tv", TmdbID: 0, Path: "/tv/1399"},
	}
	got := aggregateValidHotMediaTargets(records, now)
	want := map[hotMediaKey]int64{{MediaType: "movie", TmdbID: 550}: 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("path 可解析但字段无效的记录不应计入，got %v, want %v", got, want)
	}
}

func genHotMediaLogRecords(rng *rand.Rand, now time.Time) []hotMediaLogRecord {
	size := 1 + rng.Intn(80)
	records := make([]hotMediaLogRecord, 0, size)
	for i := 0; i < size; i++ {
		mediaType, tmdbID := genMediaTargetFields(rng)
		records = append(records, hotMediaLogRecord{
			Deleted:    rng.Intn(10) == 0,
			StatusCode: genStatusCode(rng),
			CreatedAt:  genCreatedAt(rng, now),
			MediaType:  mediaType,
			TmdbID:     tmdbID,
			Path:       genHotMediaPath(rng),
		})
	}
	return records
}

func genMediaTargetFields(rng *rand.Rand) (string, int) {
	switch rng.Intn(6) {
	case 0, 1:
		return "movie", genNonZeroTmdbID(rng)
	case 2, 3:
		return "tv", genNonZeroTmdbID(rng)
	case 4:
		return pickHotMediaString(rng, "", "person", "search", "MOVIE", "Movie"), genNonZeroTmdbID(rng)
	default:
		return pickHotMediaString(rng, "movie", "tv", "", "person"), 0
	}
}

func genStatusCode(rng *rand.Rand) int {
	return []int{199, 200, 201, 204, 302, 399, 400, 404, 500}[rng.Intn(9)]
}

func genCreatedAt(rng *rand.Rand, now time.Time) time.Time {
	daysAgo := rng.Intn(46)
	return now.Add(-time.Duration(daysAgo)*24*time.Hour - time.Duration(rng.Intn(86400))*time.Second)
}

func genHotMediaPath(rng *rand.Rand) string {
	switch rng.Intn(5) {
	case 0:
		return fmt.Sprintf("/movie/%d", genNonZeroTmdbID(rng))
	case 1:
		return fmt.Sprintf("/tv/%d/season/1", genNonZeroTmdbID(rng))
	case 2:
		return fmt.Sprintf("/api/tmdb/movie/%d", genNonZeroTmdbID(rng))
	case 3:
		return "/search/movie"
	default:
		return "/person/1"
	}
}

func expectedHotMediaCounts(records []hotMediaLogRecord, now time.Time) map[hotMediaKey]int64 {
	counts := make(map[hotMediaKey]int64)
	cutoff := now.AddDate(0, 0, -30)
	for _, record := range records {
		if record.Deleted || record.StatusCode < 200 || record.StatusCode >= 400 || record.CreatedAt.Before(cutoff) {
			continue
		}
		if (record.MediaType == "movie" || record.MediaType == "tv") && record.TmdbID != 0 {
			counts[hotMediaKey{MediaType: record.MediaType, TmdbID: record.TmdbID}]++
		}
	}
	return counts
}

func genNonZeroTmdbID(rng *rand.Rand) int {
	id := rng.Intn(2_000_000) - 500_000
	if id == 0 {
		id = 1
	}
	return id
}

func pickHotMediaString(rng *rand.Rand, opts ...string) string {
	return strings.TrimSpace(opts[rng.Intn(len(opts))])
}
