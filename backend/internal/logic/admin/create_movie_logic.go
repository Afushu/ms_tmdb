package admin

import (
	"context"
	"errors"
	"strings"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateMovieLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateMovieLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMovieLogic {
	return &CreateMovieLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateMovieLogic) CreateMovie(req *types.AdminCreateMovieReq) (*types.AdminCreateResp, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, errors.New("电影标题不能为空")
	}

	originalTitle := strings.TrimSpace(req.OriginalTitle)
	overview := strings.TrimSpace(req.Overview)
	tagline := strings.TrimSpace(req.Tagline)
	releaseDate := strings.TrimSpace(req.ReleaseDate)
	status := strings.TrimSpace(req.Status)
	originalLanguage := strings.TrimSpace(req.OriginalLanguage)
	homepage := strings.TrimSpace(req.Homepage)
	posterPath := strings.TrimSpace(req.PosterPath)
	backdropPath := strings.TrimSpace(req.BackdropPath)
	genres := buildGenresFromNames(req.GenreNames, []interface{}{})

	runtime := 0
	if req.Runtime != nil {
		runtime = *req.Runtime
	}
	voteAverage := 0.0
	if req.VoteAverage != nil {
		voteAverage = *req.VoteAverage
	}
	popularity := 0.0
	if req.Popularity != nil {
		popularity = *req.Popularity
	}

	for attempt := 0; attempt < 5; attempt++ {
		tmdbID, err := nextLocalTmdbID(l.svcCtx.DB, &model.Movie{})
		if err != nil {
			return nil, err
		}

		tmdbDataPayload := map[string]interface{}{
			"id":                tmdbID,
			"title":             title,
			"original_title":    originalTitle,
			"overview":          overview,
			"tagline":           tagline,
			"release_date":      releaseDate,
			"status":            status,
			"runtime":           runtime,
			"original_language": originalLanguage,
			"homepage":          homepage,
			"poster_path":       posterPath,
			"backdrop_path":     backdropPath,
			"vote_average":      voteAverage,
			"popularity":        popularity,
			"genres":            genres,
		}
		localDataPayload := map[string]interface{}{
			"title":             title,
			"original_title":    originalTitle,
			"overview":          overview,
			"tagline":           tagline,
			"release_date":      releaseDate,
			"status":            status,
			"runtime":           runtime,
			"original_language": originalLanguage,
			"homepage":          homepage,
			"poster_path":       posterPath,
			"backdrop_path":     backdropPath,
			"vote_average":      voteAverage,
			"popularity":        popularity,
			"genres":            genres,
		}

		tmdbData, err := toRawJSON(tmdbDataPayload)
		if err != nil {
			return nil, err
		}
		localData, err := toRawJSON(localDataPayload)
		if err != nil {
			return nil, err
		}

		record := &model.Movie{
			TmdbID:           tmdbID,
			SyncTmdbID:       tmdbID,
			Title:            title,
			OriginalTitle:    originalTitle,
			Overview:         overview,
			ReleaseDate:      releaseDate,
			Popularity:       popularity,
			VoteAverage:      voteAverage,
			PosterPath:       posterPath,
			BackdropPath:     backdropPath,
			OriginalLanguage: originalLanguage,
			Status:           status,
			Runtime:          runtime,
			Tagline:          tagline,
			Homepage:         homepage,
			TmdbData:         tmdbData,
			LocalData:        localData,
			IsModified:       true,
		}

		if err := l.svcCtx.DB.Create(record).Error; err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
				continue
			}
			return nil, err
		}

		return &types.AdminCreateResp{
			TmdbId:     tmdbID,
			SyncTmdbId: tmdbID,
			IsLocal:    true,
			Message:    "已创建本地电影条目",
		}, nil
	}

	return nil, errors.New("创建本地电影失败，请重试")
}
