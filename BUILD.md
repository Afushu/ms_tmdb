# ms-tmdb 本地构建指南

## 前置条件

- Go 1.26+
- Node.js 20+
- pnpm 9.15.4
- Docker & Docker Compose

## 快速开始

### 1. 使用一键构建脚本

```bash
cd /workspace
./build-local.sh
```

### 2. 构建 Docker 镜像

```bash
docker build -f docker/runtime.Dockerfile -t ms_tmdb:latest .
```

### 3. 运行

```bash
cd /workspace/docker
cp tmdb.yaml /path/to/your/config/tmdb.yaml
# 编辑配置文件，填入您的 TMDB API Key 和数据库配置
docker-compose up -d
```

## 完整手动构建步骤

### 1. 构建后端

```bash
cd /workspace/backend

# 构建 amd64 版本（常见于 x86 服务器和个人电脑）
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ../build-artifacts/backend/amd64/tmdb ./tmdb.go

# 如果需要 arm64 版本（例如树莓派、Mac M1/M2/M3 芯片）
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ../build-artifacts/backend/arm64/tmdb ./tmdb.go
```

### 2. 构建前端

```bash
cd /workspace/frontend

# 安装依赖（如果还没安装）
pnpm install

# 构建生产版本
pnpm run build

# 复制到构建产物目录
mkdir -p /workspace/build-artifacts/frontend
cp -r dist /workspace/build-artifacts/frontend/
```

### 3. 构建镜像

```bash
cd /workspace

# 单架构构建
docker build -f docker/runtime.Dockerfile -t ms_tmdb:latest .

# 多架构构建（需要 Docker Buildx）
docker buildx create --use
docker buildx build --platform linux/amd64,linux/arm64 \
  -f docker/runtime.Dockerfile \
  -t ms_tmdb:latest \
  --load .  # 使用 --load 加载到本地
```

## 配置和部署

### 配置文件

复制并编辑配置文件：

```bash
cp /workspace/docker/tmdb.yaml /your/path/tmdb.yaml
```

关键配置项：

```yaml
Postgres:
  Host: "postgres"
  Port: 5432
  User: "postgres"
  Password: "your_password"
  DBName: "ms_tmdb"
  SSLMode: "disable"

Tmdb:
  ApiKey: "your_tmdb_api_key"  # 必填，从 TMDB 官网获取
  BaseURL: "https://api.tmdb.org/3"
  DefaultLanguage: "zh-CN"
  RateLimit: 40
```

### 完整 Docker Compose（含数据库）

创建一个包含 PostgreSQL 的 docker-compose.yml：

```yaml
version: "3.8"

services:
  postgres:
    image: postgres:15-alpine
    container_name: ms_tmdb_postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: your_password
      POSTGRES_DB: ms_tmdb
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  app:
    image: ms_tmdb:latest
    container_name: ms_tmdb_app
    depends_on:
      - postgres
    environment:
      TMDB_CONFIG_FILE: /app/etc/tmdb.yaml
    ports:
      - "8080:80"
    volumes:
      - ./tmdb.yaml:/app/etc/tmdb.yaml
    restart: unless-stopped

volumes:
  postgres_data:
```

## 项目 CI 工作流程

项目在 GitHub Actions 中配置了自动构建和推送：

1. 当推送到 `main` 分支或打标签时触发
2. 分别构建前端和后端（amd64 和 arm64）
3. 构建多架构 Docker 镜像并推送到 Docker Hub
