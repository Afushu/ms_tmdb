package admin

import (
	"context"
	"time"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHomeDashboardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHomeDashboardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHomeDashboardLogic {
	return &GetHomeDashboardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHomeDashboardLogic) GetHomeDashboard() (resp *types.AdminHomeResp, err error) {
	latest, err := l.queryLatestHomeMedia()
	if err != nil {
		return nil, err
	}

	hot, err := l.queryHotHomeMedia()
	if err != nil {
		return nil, err
	}

	return &types.AdminHomeResp{
		Latest: latest,
		Hot:    hot,
	}, nil
}

type homeMediaRow struct {
	MediaType     string
	TmdbId        int
	Title         string
	OriginalTitle string
	PosterPath    string
	VoteAverage   float64
	AirDate       string
	Popularity    float64
	VisitCount    int64
	CreatedAt     time.Time
}

// queryLatestHomeMedia 只读取本地库中最新入库的电影和剧集，不触发 TMDB popular 回源。
func (l *GetHomeDashboardLogic) queryLatestHomeMedia() ([]types.AdminHomeMediaItem, error) {
	var rows []homeMediaRow
	if err := l.svcCtx.DB.WithContext(l.ctx).Raw(`
WITH latest_movies AS (
  SELECT
    'movie' AS media_type,
    tmdb_id,
    title,
    original_title,
    poster_path,
    vote_average,
    release_date AS air_date,
    popularity,
    0::bigint AS visit_count,
    created_at
  FROM movies
  WHERE deleted_at IS NULL
  ORDER BY created_at DESC, tmdb_id DESC
  LIMIT 10
),
latest_tv AS (
  SELECT
    'tv' AS media_type,
    tmdb_id,
    name AS title,
    original_name AS original_title,
    poster_path,
    vote_average,
    first_air_date AS air_date,
    popularity,
    0::bigint AS visit_count,
    created_at
  FROM tv_series
  WHERE deleted_at IS NULL
  ORDER BY created_at DESC, tmdb_id DESC
  LIMIT 10
)
SELECT
  media_type,
  tmdb_id,
  title,
  original_title,
  poster_path,
  vote_average,
  air_date,
  popularity,
  visit_count,
  created_at
FROM (
  SELECT
    media_type,
    tmdb_id,
    title,
    original_title,
    poster_path,
    vote_average,
    air_date,
    popularity,
    visit_count,
    created_at
  FROM latest_movies
  UNION ALL
  SELECT
    media_type,
    tmdb_id,
    title,
    original_title,
    poster_path,
    vote_average,
    air_date,
    popularity,
    visit_count,
    created_at
  FROM latest_tv
) AS local_media
ORDER BY created_at DESC, tmdb_id DESC
LIMIT 10
`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return convertHomeMediaRows(rows), nil
}

// queryHotHomeMedia 从近 30 天代理访问日志中按媒体目标聚合访问次数，最终只返回已入库内容。
func (l *GetHomeDashboardLogic) queryHotHomeMedia() ([]types.AdminHomeMediaItem, error) {
	var rows []homeMediaRow
	if err := l.svcCtx.DB.WithContext(l.ctx).Raw(`
WITH hit_sources AS (
  SELECT media_type, tmdb_id
  FROM proxy_access_logs
  WHERE deleted_at IS NULL
    AND status_code >= 200
    AND status_code < 400
    AND created_at >= NOW() - INTERVAL '30 days'
    AND media_type IN ('movie', 'tv')
    AND tmdb_id <> 0
  UNION ALL
  SELECT
    regexp_replace(path, '^/(api/v3|v3|3)/(movie|tv)/(-[0-9]+|[0-9]+).*$','\2') AS media_type,
    regexp_replace(path, '^/(api/v3|v3|3)/(movie|tv)/(-[0-9]+|[0-9]+).*$','\3')::int AS tmdb_id
  FROM proxy_access_logs
  WHERE deleted_at IS NULL
    AND status_code >= 200
    AND status_code < 400
    AND created_at >= NOW() - INTERVAL '30 days'
    AND (
      media_type IS NULL
      OR media_type NOT IN ('movie', 'tv')
      OR tmdb_id IS NULL
      OR tmdb_id = 0
    )
    AND path ~ '^/(api/v3|v3|3)/(movie|tv)/(-[0-9]+|[0-9]+)($|/.*$)'
),
hit_counts AS (
  SELECT media_type, tmdb_id, COUNT(*)::bigint AS visit_count
  FROM hit_sources
  GROUP BY media_type, tmdb_id
),
home_media AS (
  SELECT
    'movie' AS media_type,
    m.tmdb_id,
    m.title,
    m.original_title,
    m.poster_path,
    m.vote_average,
    m.release_date AS air_date,
    m.popularity,
    h.visit_count,
    m.created_at
  FROM hit_counts h
  JOIN movies m ON h.media_type = 'movie' AND h.tmdb_id = m.tmdb_id
  WHERE m.deleted_at IS NULL
  UNION ALL
  SELECT
    'tv' AS media_type,
    t.tmdb_id,
    t.name AS title,
    t.original_name AS original_title,
    t.poster_path,
    t.vote_average,
    t.first_air_date AS air_date,
    t.popularity,
    h.visit_count,
    t.created_at
  FROM hit_counts h
  JOIN tv_series t ON h.media_type = 'tv' AND h.tmdb_id = t.tmdb_id
  WHERE t.deleted_at IS NULL
)
SELECT
  media_type,
  tmdb_id,
  title,
  original_title,
  poster_path,
  vote_average,
  air_date,
  popularity,
  visit_count,
  created_at
FROM home_media
ORDER BY visit_count DESC, created_at DESC, tmdb_id DESC
LIMIT 10
`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return convertHomeMediaRows(rows), nil
}

func convertHomeMediaRows(rows []homeMediaRow) []types.AdminHomeMediaItem {
	if len(rows) == 0 {
		return []types.AdminHomeMediaItem{}
	}

	results := make([]types.AdminHomeMediaItem, 0, len(rows))
	for _, row := range rows {
		results = append(results, types.AdminHomeMediaItem{
			MediaType:     row.MediaType,
			TmdbId:        row.TmdbId,
			Title:         row.Title,
			OriginalTitle: row.OriginalTitle,
			PosterPath:    row.PosterPath,
			VoteAverage:   row.VoteAverage,
			AirDate:       row.AirDate,
			Popularity:    row.Popularity,
			VisitCount:    row.VisitCount,
			CreatedAt:     formatLogTime(row.CreatedAt),
		})
	}
	return results
}
