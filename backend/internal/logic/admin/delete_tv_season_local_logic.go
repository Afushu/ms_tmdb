package admin

import (
	"context"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTvSeasonLocalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTvSeasonLocalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTvSeasonLocalLogic {
	return &DeleteTvSeasonLocalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTvSeasonLocalLogic) DeleteTvSeasonLocal(req *types.AdminTvSeasonLocalReq) (resp *types.AdminTvSeasonLocalResp, err error) {
	if err := validateTvSeasonLocalPath(req.Id, req.SeasonNumber); err != nil {
		return nil, err
	}

	if err := l.svcCtx.ProxyService.DeleteLocalTvSeason(req.Id, req.SeasonNumber); err != nil {
		return nil, err
	}
	return &types.AdminTvSeasonLocalResp{
		Saved:   false,
		Data:    nil,
		Message: "季本地数据已删除",
	}, nil
}
