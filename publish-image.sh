#!/bin/bash

# ms-tmdb Docker 镜像发布脚本

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 默认配置
IMAGE_NAME="ms_tmdb"
IMAGE_TAG="latest"
REGISTRY=""  # 留空表示 Docker Hub
MULTI_ARCH=false
PLATFORMS="linux/amd64,linux/arm64"

function print_usage() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -u, --username <用户名>    Docker Hub 用户名"
    echo "  -n, --name <镜像名>     镜像名称 (默认: ms_tmdb)"
    echo "  -t, --tag <标签>       镜像标签 (默认: latest)"
    echo "  -r, --registry <仓库>    私有仓库地址 (例如: registry.example.com)"
    echo "  -m, --multi-arch         构建并发布多架构镜像"
    echo "  -p, --push              只推送到仓库"
    echo "  -h, --help             显示帮助"
    echo ""
    echo "示例:"
    echo "  $0 -u yourname --push"
    echo "  $0 -u yourname -t v1.0.0 --multi-arch --push"
    echo "  $0 -r registry.example.com -u yourname --push"
}

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--username)
            DOCKER_USERNAME="$2"
            shift; shift
            ;;
        -n|--name)
            IMAGE_NAME="$2"
            shift; shift
            ;;
        -t|--tag)
            IMAGE_TAG="$2"
            shift; shift
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift; shift
            ;;
        -m|--multi-arch)
            MULTI_ARCH=true
            shift
            ;;
        -p|--push)
            PUSH=true
            shift
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        *)
            echo -e "${RED}未知选项: $1${NC}"
            print_usage
            exit 1
            ;;
    esac
done

# 验证必要参数
if [ -z "$DOCKER_USERNAME" ]; then
    echo -e "${RED}错误: 请指定 Docker Hub 用户名${NC}"
    print_usage
    exit 1
fi

# 构建完整镜像名称
if [ -z "$REGISTRY" ]; then
    FULL_IMAGE_NAME="${DOCKER_USERNAME}/${IMAGE_NAME}"
else
    FULL_IMAGE_NAME="${REGISTRY}/${DOCKER_USERNAME}/${IMAGE_NAME}"
fi

echo -e "${GREEN}=== ms-tmdb 镜像发布脚本 ===${NC}"
echo "镜像: ${FULL_IMAGE_NAME}:${IMAGE_TAG}"

# 检查构建产物
if [ ! -d "${PROJECT_ROOT}/build-artifacts" ]; then
    echo -e "${YELLOW}警告: 未找到构建产物，开始构建...${NC}"
    "${PROJECT_ROOT}/build-local.sh"
fi

# 登录 Docker
if [ "$PUSH" = true ]; then
    echo ""
    echo -e "${GREEN}登录 Docker...${NC}"
    if [ -z "$REGISTRY" ]; then
        docker login
    else
        docker login "$REGISTRY"
    fi
fi

# 构建镜像
echo ""
if [ "$MULTI_ARCH" = true ]; then
    echo -e "${GREEN}构建多架构镜像...${NC}"
    
    # 检查 Buildx
    if ! docker buildx version &> /dev/null; then
        echo -e "${YELLOW}Docker Buildx 不可用，回退到单架构${NC}"
        MULTI_ARCH=false
    else
        # 创建并使用 builder
        docker buildx create --name ms-tmdb-builder --use 2>/dev/null || true
        docker buildx inspect --bootstrap
        
        # 构建命令
        BUILD_CMD="docker buildx build"
        BUILD_CMD="$BUILD_CMD -f \"${PROJECT_ROOT}/docker/runtime.Dockerfile\""
        BUILD_CMD="$BUILD_CMD -t \"${FULL_IMAGE_NAME}:${IMAGE_TAG}\""
        
        if [ "$PUSH" = true ]; then
            BUILD_CMD="$BUILD_CMD --push"
        else
            BUILD_CMD="$BUILD_CMD --load"
        fi
        
        BUILD_CMD="$BUILD_CMD --platform \"$PLATFORMS\""
        BUILD_CMD="$BUILD_CMD \"${PROJECT_ROOT}\""
        
        echo "执行: $BUILD_CMD"
        eval "$BUILD_CMD"
    fi
fi

if [ "$MULTI_ARCH" != true ]; then
    echo -e "${GREEN}构建单架构镜像...${NC}"
    
    # 构建
    docker build -f "${PROJECT_ROOT}/docker/runtime.Dockerfile" \
        -t "${FULL_IMAGE_NAME}:${IMAGE_TAG}" \
        "${PROJECT_ROOT}"
    
    # 推送
    if [ "$PUSH" = true ]; then
        echo ""
        echo -e "${GREEN}推送镜像...${NC}"
        docker push "${FULL_IMAGE_NAME}:${IMAGE_TAG}"
    fi
fi

echo ""
echo -e "${GREEN}=== 完成 ===${NC}"
if [ "$PUSH" = true ]; then
    echo "镜像已发布: ${FULL_IMAGE_NAME}:${IMAGE_TAG}"
else
    echo "镜像已构建: ${FULL_IMAGE_NAME}:${IMAGE_TAG}"
    echo "使用 'docker push ${FULL_IMAGE_NAME}:${IMAGE_TAG} 推送到仓库"
fi
