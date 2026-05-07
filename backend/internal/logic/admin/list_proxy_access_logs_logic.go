package admin

import (
	"context"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProxyAccessLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListProxyAccessLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProxyAccessLogsLogic {
	return &ListProxyAccessLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProxyAccessLogsLogic) ListProxyAccessLogs(req *types.AdminProxyAccessLogListReq) (resp *types.AdminProxyAccessLogListResp, err error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	query := applyRequestLogStatusFilter(l.svcCtx.DB.Model(&model.ProxyAccessLog{}), req.Status)
	query = applyRequestLogKeywordFilter(
		query,
		req.Keyword,
		"method",
		"path",
		"query",
		"request_uri",
		"client_ip",
		"user_agent",
		"error_message",
	)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []model.ProxyAccessLog
	if err := query.
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, err
	}

	results := make([]types.AdminProxyAccessLogItem, 0, len(records))
	for _, record := range records {
		results = append(results, proxyAccessLogItem(record))
	}

	return &types.AdminProxyAccessLogListResp{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Results:  results,
	}, nil
}
