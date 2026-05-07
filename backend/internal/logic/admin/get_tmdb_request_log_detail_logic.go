package admin

import (
	"context"
	"errors"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetTmdbRequestLogDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTmdbRequestLogDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTmdbRequestLogDetailLogic {
	return &GetTmdbRequestLogDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTmdbRequestLogDetailLogic) GetTmdbRequestLogDetail(req *types.AdminTmdbRequestLogDetailReq) (resp *types.AdminTmdbRequestLogDetailResp, err error) {
	if req.Id <= 0 {
		return nil, errors.New("日志 ID 不合法")
	}

	var record model.TmdbRequestLog
	if err := l.svcCtx.DB.First(&record, req.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("日志不存在")
		}
		return nil, err
	}

	return tmdbRequestLogDetail(record), nil
}
