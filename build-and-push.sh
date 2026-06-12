#!/usr/bin/env bash
set -euo pipefail

###############################################################################
# ms_tmdb 一站式构建 & 推送脚本
#
# 支持两种构建模式：
#   1. 自包含模式 (默认) - 使用根目录 Dockerfile，在 Docker 构建阶段完成前端/后端编译
#   2. 预构建模式         - 先在本地构建前端/后端，再用 docker/runtime.Dockerfile 打包
#
# 使用示例:
#   # 自包含模式（不需要本地安装 Node/Go）:
#   DOCKERHUB_USERNAME=your-user ./build-and-push.sh
#
#   # 仅本地构建（多架构需要 buildx）:
#   ./build-and-push.sh --no-push
#
#   # 预构建模式（本地有 Node/Go 环境）:
#   ./build-and-push.sh --pre-build
#
#   # 指定自定义镜像名和平台:
#   DOCKERHUB_USERNAME=your-user DOCKERHUB_IMAGE_NAME=custom \
#     PLATFORMS=linux/amd64,linux/arm64 ./build-and-push.sh
###############################################################################

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_ROOT"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $1"; }

# 参数解析
PUSH=true
USE_SELF_CONTAINED=true
for arg in "$@"; do
    case "$arg" in
        --no-push)         PUSH=false ;;
        --pre-build)       USE_SELF_CONTAINED=false ;;
        -h|--help)
            echo "用法: DOCKERHUB_USERNAME=xxx ./build-and-push.sh [options]"
            echo ""
            echo "选项:"
            echo "  --no-push        仅构建本地镜像，不推送到 Docker Hub"
            echo "  --pre-build      先在本地构建前端/后端，再用 runtime.Dockerfile 打包"
            echo "                   （默认使用自包含 Dockerfile，无需本地 Node/Go 环境）"
            exit 0
            ;;
        *)
            log_error "未知参数: $arg"
            exit 1
            ;;
    esac
done

# 生成版本号（和 CI 一致: YYYYMMDDHHMM）
APP_VERSION="$(TZ=Asia/Shanghai date +%Y%m%d%H%M)"
IMAGE_NAME="${DOCKERHUB_IMAGE_NAME:-ms_tmdb}"
PLATFORMS="${PLATFORMS:-linux/amd64}"

echo "========================================"
echo "  ms_tmdb 构建 & 推送"
echo "========================================"
echo "  版本号: ${APP_VERSION}"
echo "  镜像名: ${DOCKERHUB_USERNAME:-<not-set>}/${IMAGE_NAME}"
echo "  平台:   ${PLATFORMS}"
echo "  推送:   ${PUSH}"
echo "  模式:   $([ "$USE_SELF_CONTAINED" = true ] && echo "自包含 (Dockerfile)" || echo "预构建 (docker/runtime.Dockerfile)")"
echo "========================================"
echo ""

# 检查依赖
if [ "$USE_SELF_CONTAINED" = false ]; then
    log_info "检查本地 Node/Go 依赖..."
    command -v node >/dev/null 2>&1 || { log_error "需要安装 Node.js"; exit 1; }
    command -v pnpm >/dev/null 2>&1 || { log_error "需要安装 pnpm"; exit 1; }
    command -v go >/dev/null 2>&1 || { log_error "需要安装 Go"; exit 1; }
    log_success "依赖检查通过"
fi

command -v docker >/dev/null 2>&1 || { log_error "需要安装 Docker"; exit 1; }

if [ "$PUSH" = true ] && [ -z "${DOCKERHUB_USERNAME:-}" ]; then
    log_error "推送镜像需要设置 DOCKERHUB_USERNAME 环境变量"
    exit 1
fi

###############################################################################
# 预构建模式: 本地编译前端 + 后端
###############################################################################
if [ "$USE_SELF_CONTAINED" = false ]; then
    echo ""
    log_info "[1/3] 构建前端 (VITE_APP_VERSION=${APP_VERSION})..."
    cd "$PROJECT_ROOT/frontend"
    if ! SKIP_SIMPLE_GIT_HOOKS=1 pnpm install --frozen-lockfile 2>&1 | tail -5; then
        log_warn "frozen-lockfile 失败，尝试普通安装..."
        SKIP_SIMPLE_GIT_HOOKS=1 pnpm install 2>&1 | tail -10
    fi
    VITE_APP_VERSION="${APP_VERSION}" pnpm run build 2>&1 | tail -10
    log_success "前端构建完成"

    echo ""
    log_info "[2/3] 构建后端 (main.Version=${APP_VERSION})..."
    cd "$PROJECT_ROOT/backend"
    mkdir -p "$PROJECT_ROOT/build-artifacts/backend/amd64"
    mkdir -p "$PROJECT_ROOT/build-artifacts/backend/arm64"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go build -trimpath -ldflags="-s -w -X main.Version=${APP_VERSION}" \
        -o "$PROJECT_ROOT/build-artifacts/backend/amd64/tmdb" ./tmdb.go
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
        go build -trimpath -ldflags="-s -w -X main.Version=${APP_VERSION}" \
        -o "$PROJECT_ROOT/build-artifacts/backend/arm64/tmdb" ./tmdb.go
    log_success "后端构建完成"

    echo ""
    log_info "[3/3] 准备构建产物..."
    rm -rf "$PROJECT_ROOT/build-artifacts/frontend/dist"
    mkdir -p "$PROJECT_ROOT/build-artifacts/frontend/dist"
    cp -r "$PROJECT_ROOT/frontend/dist/"* "$PROJECT_ROOT/build-artifacts/frontend/dist/"
    log_success "构建产物准备完成"

    DOCKERFILE="docker/runtime.Dockerfile"
else
    DOCKERFILE="Dockerfile"
fi

###############################################################################
# 构建 Docker 镜像
###############################################################################
echo ""
log_info "构建 Docker 镜像 (file=${DOCKERFILE}, VERSION=${APP_VERSION})..."

cd "$PROJECT_ROOT"

# 设置 Docker buildx（用于多架构）
if echo "$PLATFORMS" | grep -q ","; then
    log_info "  初始化 Docker buildx（多架构）..."
    docker buildx create --name ms-tmdb-builder --use 2>/dev/null || true
    docker buildx use ms-tmdb-builder 2>/dev/null || true
fi

FULL_IMAGE="${DOCKERHUB_USERNAME}/${IMAGE_NAME}"

if [ "$PUSH" = true ]; then
    log_info "  登录 Docker Hub..."
    if ! docker login --username "$DOCKERHUB_USERNAME" 2>&1 | tail -3; then
        log_error "Docker Hub 登录失败"
        exit 1
    fi

    log_info "  构建并推送镜像..."
    docker buildx build \
        --platform "${PLATFORMS}" \
        --file "$DOCKERFILE" \
        --tag "${FULL_IMAGE}:latest" \
        --tag "${FULL_IMAGE}:${APP_VERSION}" \
        --build-arg VERSION="${APP_VERSION}" \
        --output "type=image,push=true" \
        --provenance=false \
        --sbom=false \
        . 2>&1 | tail -30

    log_success "镜像推送完成:"
    echo "    ${FULL_IMAGE}:latest"
    echo "    ${FULL_IMAGE}:${APP_VERSION}"
else
    log_info "  构建本地镜像..."
    docker build \
        --file "$DOCKERFILE" \
        --tag "${IMAGE_NAME}:latest" \
        --tag "${IMAGE_NAME}:${APP_VERSION}" \
        --build-arg VERSION="${APP_VERSION}" \
        . 2>&1 | tail -20

    log_success "镜像构建完成:"
    echo "    ${IMAGE_NAME}:latest"
    echo "    ${IMAGE_NAME}:${APP_VERSION}"
fi

###############################################################################
# 验证版本号
###############################################################################
echo ""
log_info "验证版本号..."

# 后端二进制版本（仅预构建模式）
if [ "$USE_SELF_CONTAINED" = false ]; then
    BACKEND_VER="$("$PROJECT_ROOT/build-artifacts/backend/amd64/tmdb" --version 2>&1)"
    echo "  后端二进制: ${BACKEND_VER}"
fi

# 镜像标签元数据
if command -v docker >/dev/null 2>&1; then
    INSPECT_VER="$(docker inspect --format='{{ index .Config.Labels "org.opencontainers.image.version"}}' "${IMAGE_NAME}:latest" 2>/dev/null || echo "N/A")"
    echo "  镜像标签版本: ${INSPECT_VER}"
fi

# 输出构建目录中的 dist 版本号（前端注入的 __APP_VERSION__）
if [ -d "$PROJECT_ROOT/frontend/dist" ]; then
    FE_VER=$(grep -oP '"[0-9]{8,12}"' "$PROJECT_ROOT/frontend/dist/assets"/*.js 2>/dev/null | head -1 | tr -d '"' || true)
    if [ -n "$FE_VER" ]; then
        echo "  前端 dist 内版本号: ${FE_VER}"
    fi
fi

echo ""
log_success "全部完成！"
echo "========================================"
echo "  版本: ${APP_VERSION}"
echo "========================================"
