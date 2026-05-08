package admin

import (
	"context"
	"errors"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTvSeasonLocalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTvSeasonLocalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTvSeasonLocalLogic {
	return &UpdateTvSeasonLocalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTvSeasonLocalLogic) UpdateTvSeasonLocal(req *types.AdminTvSeasonLocalUpdateReq) (resp *types.AdminTvSeasonLocalResp, err error) {
	if err := validateTvSeasonLocalPath(req.Id, req.SeasonNumber); err != nil {
		return nil, err
	}
	if len(req.Payload) == 0 {
		return nil, errors.New("更新内容不能为空")
	}

	data, err := l.svcCtx.ProxyService.UpdateLocalTvSeason(req.Id, req.SeasonNumber, req.Payload)
	if err != nil {
		return nil, err
	}
	return &types.AdminTvSeasonLocalResp{
		Saved:   true,
		Data:    data,
		Message: "季明细本地修改已保存",
	}, nil
}
