package admin

import (
	"context"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearProxyAccessLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClearProxyAccessLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearProxyAccessLogsLogic {
	return &ClearProxyAccessLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClearProxyAccessLogsLogic) ClearProxyAccessLogs() (resp *types.AdminRequestLogClearResp, err error) {
	if err := truncateRequestLogTable(l.svcCtx.DB, &model.ProxyAccessLog{}); err != nil {
		return nil, err
	}

	return &types.AdminRequestLogClearResp{
		Message: "代理访问日志已清空",
	}, nil
}
