package admin

import (
	"context"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTmdbRequestLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTmdbRequestLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTmdbRequestLogsLogic {
	return &ListTmdbRequestLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTmdbRequestLogsLogic) ListTmdbRequestLogs(req *types.AdminTmdbRequestLogListReq) (resp *types.AdminTmdbRequestLogListResp, err error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	query := applyRequestLogStatusFilter(l.svcCtx.DB.Model(&model.TmdbRequestLog{}), req.Status)
	query = applyRequestLogKeywordFilter(
		query,
		req.Keyword,
		"method",
		"path",
		"url",
		"error_message",
	)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []model.TmdbRequestLog
	if err := query.
		Select(
			"id",
			"request_id",
			"method",
			"path",
			"url",
			"status_code",
			"duration_ms",
			"error_message",
			"request_body_bytes",
			"request_body_truncated",
			"response_body_bytes",
			"response_body_truncated",
			"created_at",
		).
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, err
	}

	results := make([]types.AdminTmdbRequestLogItem, 0, len(records))
	for _, record := range records {
		results = append(results, tmdbRequestLogItem(record))
	}

	return &types.AdminTmdbRequestLogListResp{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Results:  results,
	}, nil
}
