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

	var records []model.ProxyAccessLog
	if err := query.
		Select(
			"id",
			"request_id",
			"method",
			"path",
			"query",
			"request_uri",
			"client_ip",
			"user_agent",
			"status_code",
			"duration_ms",
			"error_message",
			"request_body_bytes",
			"request_body_truncated",
			"response_body_bytes",
			"response_body_truncated",
			"created_at",
		).
		Order(requestLogListOrder).
		Offset((page - 1) * pageSize).
		Limit(requestLogPageLimit(pageSize)).
		Find(&records).Error; err != nil {
		return nil, err
	}

	total := requestLogWindowTotal(page, pageSize, len(records))
	visibleCount := requestLogVisibleCount(pageSize, len(records))
	results := make([]types.AdminProxyAccessLogItem, 0, visibleCount)
	for _, record := range records[:visibleCount] {
		results = append(results, proxyAccessLogItem(record))
	}

	return &types.AdminProxyAccessLogListResp{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Results:  results,
	}, nil
}
