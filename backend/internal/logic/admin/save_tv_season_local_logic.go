package admin

import (
	"context"
	"strings"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"
	"ms_tmdb/pkg/tmdbclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveTvSeasonLocalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveTvSeasonLocalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveTvSeasonLocalLogic {
	return &SaveTvSeasonLocalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveTvSeasonLocalLogic) SaveTvSeasonLocal(req *types.AdminTvSeasonLocalReq) (resp *types.AdminTvSeasonLocalResp, err error) {
	if err := validateTvSeasonLocalPath(req.Id, req.SeasonNumber); err != nil {
		return nil, err
	}

	opts := &tmdbclient.RequestOption{
		Context:  l.ctx,
		Language: strings.TrimSpace(req.Language),
	}
	if text := strings.TrimSpace(req.AppendToResponse); text != "" {
		opts.AppendToResponse = text
	}

	data, err := l.svcCtx.ProxyService.SaveTvSeasonToLocal(req.Id, req.SeasonNumber, opts)
	if err != nil {
		return nil, err
	}
	return &types.AdminTvSeasonLocalResp{
		Saved:   true,
		Data:    data,
		Message: "季明细已保存到本地数据库",
	}, nil
}
