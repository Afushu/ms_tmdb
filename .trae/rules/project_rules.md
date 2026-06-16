# ms_tmdb 项目规则

## 上游同步注意事项

同步上游代码（`git reset --hard upstream/main`）后，必须重新应用以下本地修改。使用 `git cherry-pick` 或手动重新应用，**不要用 merge**（两个分支历史 unrelated）。

---

### 1. 版本号注入链路（3 个文件 + 2 个新文件）

**backend/tmdb.go** — 添加 Version 变量、--version flag、启动时同步版本号
```go
var (
    Version    = "dev"
    configFile = flag.String("f", "etc/tmdb.yaml", "the config file")
    showVer    = flag.Bool("version", false, "显示版本号并退出")
)

func main() {
    flag.Parse()
    if *showVer { ... }
    adminlogic.AppVersion = Version
    logx.Infof("服务启动: %s:%d version=%s", c.Host, c.Port, Version)
}
```

**backend/internal/logic/admin/get_version_logic.go** — 新建文件，版本 API logic
**backend/internal/handler/admin/get_version_handler.go** — 新建文件，版本 API handler

**backend/internal/handler/routes.go** — 在 admin 路由组中添加：
```go
{
    Method:  http.MethodGet,
    Path:    "/version",
    Handler: admin.GetVersionHandler(serverCtx),
},
```

**backend/internal/types/types.go** — 文件末尾追加：
```go
type VersionResp struct {
    Version string `json:"version"`
    Go      string `json:"go"`
    Os      string `json:"os"`
    Arch    string `json:"arch"`
}
```

---

### 2. Genre ID 保留修复（5 个文件）

**backend/internal/logic/admin/helper.go** — `buildGenresFromNames` 添加 `originalGenres` 参数，新增 `extractGenresFromTmdbData` 函数
- 从 TMDB 原始数据提取 name→ID 映射，保留原始 ID（如动画=16）
- 新 genre 用负数 ID

**backend/internal/logic/admin/update_movie_logic.go** — 调用时传 `extractGenresFromTmdbData(movie.TmdbData)`
**backend/internal/logic/admin/create_movie_logic.go** — 调用时传 `nil`
**backend/internal/logic/admin/update_tv_series_logic.go** — 调用时传 `extractGenresFromTmdbData(tv.TmdbData)`
**backend/internal/logic/admin/create_tv_series_logic.go** — 调用时传 `nil`

---

### 3. 前端版本显示（2 个文件）

**frontend/src/api/admin.ts** — 文件末尾追加：
```typescript
export type AdminVersionResp = { version: string; go?: string; os?: string; arch?: string };
export function getVersion() { return http.get<AdminVersionResp>("/api/admin/version"); }
```

**frontend/src/pages/SystemSettingsPage.vue** — 添加后端版本卡片
- import `getVersion`
- 添加 `backendVersion`, `backendGo` 变量
- `loadSettings()` 中并行请求 `/api/admin/version`
- 模板添加后端版本卡片

---

### 4. Docker & CI/CD（3 个文件）

**docker/runtime.Dockerfile** — 添加 ARG VERSION 和 image version label
```dockerfile
ARG VERSION=dev
LABEL org.opencontainers.image.version="${VERSION}"
```

**.github/workflows/build-and-push.yml** — 关键修改点：
- `frontend` job 添加 `outputs: app_version: ${{ steps.app-version.outputs.APP_VERSION }}`
- `backend` job 添加 `needs: [frontend]`（**必须！否则无法读取 outputs**）
- `backend` 构建步骤添加 `-X main.Version=${APP_VERSION}`
- `build` job 添加 `build-args: VERSION=...` 和 version label

**Dockerfile**（根目录）— 本地多阶段构建，使用 `golang:1.24-alpine` + `GOTOOLCHAIN=auto`

---

### 5. 验证命令

```bash
# Go 编译
cd backend && go build -o /dev/null ./tmdb.go

# 前端类型检查
cd frontend && npx vue-tsc --noEmit
```

---

### 恢复步骤（上游同步时）

```bash
git checkout main
git fetch upstream
git reset --hard upstream/main
# 然后逐一重新应用上述所有修改
git add <所有修改文件>
git commit -m "feat: 重新应用本地修改（版本号注入 + Genre ID 修复）"
git push origin main --force
```