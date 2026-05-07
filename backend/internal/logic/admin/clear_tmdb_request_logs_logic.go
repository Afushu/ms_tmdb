package admin

import (
	"context"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearTmdbRequestLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClearTmdbRequestLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearTmdbRequestLogsLogic {
	return &ClearTmdbRequestLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClearTmdbRequestLogsLogic) ClearTmdbRequestLogs() (resp *types.AdminRequestLogClearResp, err error) {
	if err := truncateRequestLogTable(l.svcCtx.DB, &model.TmdbRequestLog{}); err != nil {
		return nil, err
	}

	return &types.AdminRequestLogClearResp{
		Message: "TMDB 请求日志已清空",
	}, nil
}
