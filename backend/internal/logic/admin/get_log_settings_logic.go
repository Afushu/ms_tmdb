// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package admin

import (
	"context"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLogSettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLogSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLogSettingsLogic {
	return &GetLogSettingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLogSettingsLogic) GetLogSettings() (*types.AdminLogSettingsResp, error) {
	return &types.AdminLogSettingsResp{
		RetentionDays: currentLogRetentionDays(l.svcCtx),
	}, nil
}
