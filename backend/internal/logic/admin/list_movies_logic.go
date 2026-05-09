package admin

import (
	"context"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMoviesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMoviesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMoviesLogic {
	return &ListMoviesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMoviesLogic) ListMovies(req *types.LibraryListReq) (*types.MovieListResp, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	searchMode := normalizeSearchMode(req.SearchMode)

	var total int64
	query := applyKeywordFilter(
		l.svcCtx.DB.Model(&model.Movie{}),
		req.Keyword,
		searchMode,
		"tmdb_id",
		"title",
		"original_title",
	)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []model.Movie
	if err := query.Select(movieListSelect).
		Order("popularity DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, err
	}

	results := make([]types.MovieListItem, len(items))
	for i, m := range items {
		results[i] = types.MovieListItem{
			TmdbId:        m.TmdbID,
			Title:         m.Title,
			OriginalTitle: m.OriginalTitle,
			PosterPath:    m.PosterPath,
			VoteAverage:   m.VoteAverage,
			ReleaseDate:   m.ReleaseDate,
			Popularity:    m.Popularity,
			IsModified:    m.IsModified,
			GenreNames:    mergeGenreNames(genreNamesFromRaw(m.LocalData), genreNamesFromRaw(m.TmdbData)),
		}
	}

	return &types.MovieListResp{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Results:  results,
	}, nil
}
