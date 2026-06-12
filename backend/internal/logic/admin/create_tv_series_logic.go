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

type CreateTvSeriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTvSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTvSeriesLogic {
	return &CreateTvSeriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTvSeriesLogic) CreateTvSeries(req *types.AdminCreateTvReq) (*types.AdminCreateResp, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("剧集名称不能为空")
	}

	originalName := strings.TrimSpace(req.OriginalName)
	overview := strings.TrimSpace(req.Overview)
	tagline := strings.TrimSpace(req.Tagline)
	firstAirDate := strings.TrimSpace(req.FirstAirDate)
	status := strings.TrimSpace(req.Status)
	originalLanguage := strings.TrimSpace(req.OriginalLanguage)
	homepage := strings.TrimSpace(req.Homepage)
	posterPath := strings.TrimSpace(req.PosterPath)
	backdropPath := strings.TrimSpace(req.BackdropPath)
	seriesType := strings.TrimSpace(req.Type)
	genres := buildGenresFromNames(req.GenreNames, nil)

	numberOfSeasons := 0
	if req.NumberOfSeasons != nil {
		numberOfSeasons = *req.NumberOfSeasons
	}
	numberOfEpisodes := 0
	if req.NumberOfEpisodes != nil {
		numberOfEpisodes = *req.NumberOfEpisodes
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
		tmdbID, err := nextLocalTmdbID(l.svcCtx.DB, &model.TVSeries{})
		if err != nil {
			return nil, err
		}

		tmdbDataPayload := map[string]interface{}{
			"id":                 tmdbID,
			"name":               name,
			"original_name":      originalName,
			"overview":           overview,
			"tagline":            tagline,
			"first_air_date":     firstAirDate,
			"status":             status,
			"number_of_seasons":  numberOfSeasons,
			"number_of_episodes": numberOfEpisodes,
			"original_language":  originalLanguage,
			"homepage":           homepage,
			"poster_path":        posterPath,
			"backdrop_path":      backdropPath,
			"vote_average":       voteAverage,
			"popularity":         popularity,
			"type":               seriesType,
			"genres":             genres,
		}
		localDataPayload := map[string]interface{}{
			"name":               name,
			"original_name":      originalName,
			"overview":           overview,
			"tagline":            tagline,
			"first_air_date":     firstAirDate,
			"status":             status,
			"number_of_seasons":  numberOfSeasons,
			"number_of_episodes": numberOfEpisodes,
			"original_language":  originalLanguage,
			"homepage":           homepage,
			"poster_path":        posterPath,
			"backdrop_path":      backdropPath,
			"vote_average":       voteAverage,
			"popularity":         popularity,
			"type":               seriesType,
			"genres":             genres,
		}

		tmdbData, err := toRawJSON(tmdbDataPayload)
		if err != nil {
			return nil, err
		}
		localData, err := toRawJSON(localDataPayload)
		if err != nil {
			return nil, err
		}

		record := &model.TVSeries{
			TmdbID:           tmdbID,
			SyncTmdbID:       tmdbID,
			Name:             name,
			OriginalName:     originalName,
			Overview:         overview,
			FirstAirDate:     firstAirDate,
			Popularity:       popularity,
			VoteAverage:      voteAverage,
			PosterPath:       posterPath,
			BackdropPath:     backdropPath,
			OriginalLanguage: originalLanguage,
			Status:           status,
			Type:             seriesType,
			NumberOfSeasons:  numberOfSeasons,
			NumberOfEpisodes: numberOfEpisodes,
			Homepage:         homepage,
			Tagline:          tagline,
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
			Message:    "已创建本地剧集条目",
		}, nil
	}

	return nil, errors.New("创建本地剧集失败，请重试")
}
