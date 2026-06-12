#!/usr/bin/env bash
set -euo pipefail

###############################################################################
# ms_tmdb 一键构建并推送 Docker 镜像（:latest 版）
# 使用:
#   DOCKERHUB_USERNAME=xxx ./build-and-push.sh          # 构建并推送 :latest
#   DOCKERHUB_USERNAME=xxx PLATFORMS=linux/amd64,linux/arm64 ./build-and-push.sh
#   ./build-and-push.sh --no-push                       # 仅本地构建，不推送
###############################################################################

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_ROOT"

# 参数解析
PUSH=true
for arg in "$@"; do
    case "$arg" in
        --no-push) PUSH=false ;;
        -h|--help)
            echo "用法: DOCKERHUB_USERNAME=xxx ./build-and-push.sh [--no-push]"
            echo ""
            echo "可选环境变量:"
            echo "  DOCKERHUB_USERNAME   Docker Hub 用户名（推送时必须）"
            echo "  DOCKERHUB_IMAGE_NAME 镜像名称（默认: ms_tmdb）"
            echo "  PLATFORMS           目标平台（默认: linux/amd64）"
            echo "                      多架构示例: linux/amd64,linux/arm64"
            exit 0
            ;;
        *) echo "未知参数: $arg"; exit 1 ;;
    esac
done

# 版本号 = 构建时的时间戳（唯一标识本次构建产物）
APP_VERSION="$(TZ=Asia/Shanghai date +%Y%m%d%H%M)"
IMAGE_NAME="${DOCKERHUB_IMAGE_NAME:-ms_tmdb}"
PLATFORMS="${PLATFORMS:-linux/amd64}"

echo "========================================"
echo "  ms_tmdb 镜像构建"
echo "========================================"
echo "  版本号: ${APP_VERSION}"
echo "  镜像:   ${DOCKERHUB_USERNAME:-<local>}/${IMAGE_NAME}:latest"
echo "  平台:   ${PLATFORMS}"
echo "  推送:   ${PUSH}"
echo "========================================"
echo ""

# 依赖检查
command -v docker >/dev/null 2>&1 || { echo "错误: 需要安装 docker"; exit 1; }
if [ "$PUSH" = true ] && [ -z "${DOCKERHUB_USERNAME:-}" ]; then
    echo "错误: 推送需要设置 DOCKERHUB_USERNAME"; exit 1
fi

# 多架构时初始化 buildx
if echo "$PLATFORMS" | grep -q ","; then
    echo "[1/2] 初始化 Docker buildx（多架构）..."
    docker buildx create --name ms-tmdb-builder --use 2>/dev/null || true
    docker buildx use ms-tmdb-builder 2>/dev/null || true
fi

echo ""
echo "[构建] 编译镜像并注入版本号 ${APP_VERSION}..."

FULL_IMAGE="${DOCKERHUB_USERNAME}/${IMAGE_NAME}"

if [ "$PUSH" = true ]; then
    echo "→ 登录 Docker Hub..."
    docker login --username "$DOCKERHUB_USERNAME"

    echo "→ 构建并推送 ${FULL_IMAGE}:latest"
    docker buildx build \
        --platform "${PLATFORMS}" \
        --file Dockerfile \
        --tag "${FULL_IMAGE}:latest" \
        --build-arg VERSION="${APP_VERSION}" \
        --output "type=image,push=true" \
        --provenance=false \
        --sbom=false \
        .
else
    echo "→ 本地构建 ${IMAGE_NAME}:latest"
    docker build \
        --file Dockerfile \
        --tag "${IMAGE_NAME}:latest" \
        --build-arg VERSION="${APP_VERSION}" \
        .
fi

echo ""
echo "========================================"
echo "  完成！版本: ${APP_VERSION}"
echo "  镜像 tag: latest"
echo "========================================"
