package admin

import (
	"context"
	"errors"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTvSeasonLocalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTvSeasonLocalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTvSeasonLocalLogic {
	return &GetTvSeasonLocalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTvSeasonLocalLogic) GetTvSeasonLocal(req *types.AdminTvSeasonLocalReq) (resp *types.AdminTvSeasonLocalResp, err error) {
	if err := validateTvSeasonLocalPath(req.Id, req.SeasonNumber); err != nil {
		return nil, err
	}

	data, saved, err := l.svcCtx.ProxyService.GetLocalTvSeason(req.Id, req.SeasonNumber)
	if err != nil {
		return nil, err
	}
	return &types.AdminTvSeasonLocalResp{
		Saved: saved,
		Data:  data,
	}, nil
}

func validateTvSeasonLocalPath(seriesID, seasonNumber int) error {
	if seriesID <= 0 {
		return errors.New("无效剧集 ID")
	}
	if seasonNumber < 0 {
		return errors.New("无效季号")
	}
	return nil
}
