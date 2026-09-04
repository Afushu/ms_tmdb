package admin

import (
	"context"
	"runtime"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"
)

type GetVersionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVersionLogic {
	return &GetVersionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVersionLogic) GetVersion() (*types.VersionResp, error) {
	return &types.VersionResp{
		Version: Version,
		Go:      runtime.Version(),
	}, nil
}