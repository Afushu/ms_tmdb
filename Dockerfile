# syntax=docker/dockerfile:1.6
# ============================================================
#  ms_tmdb 多阶段构建 Dockerfile
#  本地构建: docker build -t ms_tmdb:latest .
#  多架构:   docker buildx build --platform linux/amd64,linux/arm64 \
#             --build-arg VERSION=$(date +%Y%m%d%H%M) \
#             -t your-user/ms_tmdb:latest --push .
# ============================================================

ARG VERSION=dev
ARG NODE_VERSION=20
ARG GO_VERSION=1.23
ARG NGINX_VERSION=1.27-alpine

# ============================================================
# Stage 1: 构建前端
# ============================================================
FROM node:${NODE_VERSION}-alpine AS frontend-builder

ARG VERSION

WORKDIR /app

# 安装 pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# 复制依赖定义，利用 Docker layer 缓存
COPY frontend/package.json frontend/pnpm-lock.yaml* ./

# 安装依赖（跳过 git hooks）
RUN SKIP_SIMPLE_GIT_HOOKS=1 pnpm install --frozen-lockfile 2>/dev/null || \
    SKIP_SIMPLE_GIT_HOOKS=1 pnpm install

# 复制源代码
COPY frontend/ ./

# 构建前端（注入版本号）
RUN VITE_APP_VERSION="${VERSION}" pnpm run build

# ============================================================
# Stage 2: 构建后端（根据 TARGETARCH 自动选择架构）
# ============================================================
FROM golang:${GO_VERSION}-alpine AS backend-builder

ARG VERSION
ARG TARGETARCH

WORKDIR /app

# 复制后端源码
COPY backend/ ./

# 编译对应架构的二进制
RUN GOARCH=${TARGETARCH} CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w -X main.Version=${VERSION}" \
    -o /out/tmdb ./tmdb.go

# ============================================================
# Stage 3: 运行时镜像
# ============================================================
FROM nginx:${NGINX_VERSION}

ARG VERSION

# 元数据
LABEL org.opencontainers.image.title="ms_tmdb"
LABEL org.opencontainers.image.description="TMDB proxy and local enhancement platform"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.source="https://github.com/GateCross/ms_tmdb"

WORKDIR /app

# 安装运行时依赖（alpine 使用 apk）
RUN apk add --no-cache ca-certificates curl tzdata

# 前端 dist
COPY --from=frontend-builder /app/dist /usr/share/nginx/html

# 后端二进制
COPY --from=backend-builder /out/tmdb /app/tmdb
RUN chmod +x /app/tmdb

# 配置文件和脚本
COPY backend/etc /app/etc
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# 确保 uploads 目录存在
RUN mkdir -p /app/uploads

EXPOSE 80

# 默认配置文件路径
ENV TMDB_CONFIG_FILE=/app/etc/tmdb.yaml

ENTRYPOINT ["/entrypoint.sh"]
