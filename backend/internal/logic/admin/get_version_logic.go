package admin

import (
	"context"
	"runtime"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

var AppVersion = "dev"

type GetVersionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVersionLogic {
	return &GetVersionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVersionLogic) GetVersion() (resp *types.VersionResp, err error) {
	return &types.VersionResp{
		Version: AppVersion,
		Go:      runtime.Version(),
		Os:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}, nil
}