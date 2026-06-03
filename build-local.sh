#!/bin/bash

# ms-tmdb 本地构建脚本

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
BUILD_DIR="${PROJECT_ROOT}/build-artifacts"

echo "=== 开始构建 ms-tmdb ==="

# 1. 清理旧构建产物
echo "清理旧构建产物..."
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}/backend/amd64" "${BUILD_DIR}/frontend/dist"

# 2. 构建后端
echo ""
echo "构建后端..."
cd "${PROJECT_ROOT}/backend"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "${BUILD_DIR}/backend/amd64/tmdb" ./tmdb.go

# 3. 构建前端
echo ""
echo "构建前端..."
cd "${PROJECT_ROOT}/frontend"

if [ ! -d "node_modules" ]; then
  echo "安装前端依赖..."
  pnpm install
fi

pnpm run build

# 复制前端构建产物
cp -r "${PROJECT_ROOT}/frontend/dist"/* "${BUILD_DIR}/frontend/dist/"

# 4. 复制配置文件
echo ""
echo "复制配置文件..."
mkdir -p "${BUILD_DIR}/etc"
cp -r "${PROJECT_ROOT}/backend/etc" "${BUILD_DIR}/"

echo ""
echo "=== 构建完成 ==="
echo "构建产物目录: ${BUILD_DIR}"
echo ""
echo "接下来可以构建 Docker 镜像:"
echo "docker build -f docker/runtime.Dockerfile -t ms_tmdb:latest ."
