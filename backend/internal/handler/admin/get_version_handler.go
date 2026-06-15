package admin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"ms_tmdb/internal/logic/admin"
	"ms_tmdb/internal/svc"
)

func GetVersionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := admin.NewGetVersionLogic(r.Context(), svcCtx)
		resp, err := l.GetVersion()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}