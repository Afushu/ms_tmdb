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

type GetProxyAccessLogDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProxyAccessLogDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProxyAccessLogDetailLogic {
	return &GetProxyAccessLogDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProxyAccessLogDetailLogic) GetProxyAccessLogDetail(req *types.AdminProxyAccessLogDetailReq) (resp *types.AdminProxyAccessLogDetailResp, err error) {
	if req.Id <= 0 {
		return nil, errors.New("日志 ID 不合法")
	}

	var record model.ProxyAccessLog
	if err := l.svcCtx.DB.First(&record, req.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("日志不存在")
		}
		return nil, err
	}

	return proxyAccessLogDetail(record), nil
}
