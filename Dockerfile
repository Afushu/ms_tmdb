# syntax=docker/dockerfile:1.6
# 多阶段构建：无需本地安装 Node/Go 环境
# docker build --build-arg VERSION=$(date +%Y%m%d%H%M) -t ms_tmdb:latest .
# docker buildx build --platform linux/amd64,linux/arm64 --build-arg VERSION=$(date +%Y%m%d%H%M) -t user/ms_tmdb:latest --push .

ARG VERSION=dev
ARG NODE_VERSION=22
ARG GO_VERSION=1.26
ARG NGINX_VERSION=1.27-alpine

# --- Stage 1: 构建前端 ---
FROM node:${NODE_VERSION}-alpine AS frontend-builder

ARG VERSION

WORKDIR /app

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY frontend/package.json frontend/pnpm-lock.yaml* ./
RUN SKIP_SIMPLE_GIT_HOOKS=1 pnpm install --frozen-lockfile 2>/dev/null || \
    SKIP_SIMPLE_GIT_HOOKS=1 pnpm install

COPY frontend/ ./
RUN VITE_APP_VERSION="${VERSION}" pnpm run build

# --- Stage 2: 构建后端 ---
FROM golang:${GO_VERSION}-alpine AS backend-builder

ARG VERSION
ARG TARGETARCH

WORKDIR /app

COPY backend/ ./
RUN GOARCH=${TARGETARCH} CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w -X main.Version=${VERSION}" \
    -o /out/tmdb ./tmdb.go

# --- Stage 3: 运行时镜像 ---
FROM nginx:${NGINX_VERSION}

ARG VERSION

LABEL org.opencontainers.image.title="ms_tmdb"
LABEL org.opencontainers.image.description="TMDB proxy and local enhancement platform"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.source="https://github.com/GateCross/ms_tmdb"

WORKDIR /app

RUN apk add --no-cache ca-certificates curl tzdata

COPY --from=frontend-builder /app/dist /usr/share/nginx/html
COPY --from=backend-builder /out/tmdb /app/tmdb

RUN chmod +x /app/tmdb

COPY backend/etc /app/etc
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh && mkdir -p /app/uploads

EXPOSE 80

ENV TMDB_CONFIG_FILE=/app/etc/tmdb.yaml
ENTRYPOINT ["/entrypoint.sh"]