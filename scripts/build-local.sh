#!/bin/bash

# 本地开发脚本
# 用于构建和测试Docker镜像

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

BACKEND_IMAGE="url-manager/backend:local"
FRONTEND_IMAGE="url-manager/frontend:local"

function build_backend() {
    echo -e "${YELLOW}🔨 构建后端镜像...${NC}"
    cd backend
    docker build -t $BACKEND_IMAGE .
    cd ..
    echo -e "${GREEN}✅ 后端镜像构建完成${NC}"
}

function build_frontend() {
    echo -e "${YELLOW}🔨 构建前端镜像...${NC}"
    cd frontend
    docker build -t $FRONTEND_IMAGE .
    cd ..
    echo -e "${GREEN}✅ 前端镜像构建完成${NC}"
}

function test_images() {
    echo -e "${YELLOW}🧪 测试镜像...${NC}"
    
    # 测试后端镜像
    echo "测试后端镜像..."
    docker run --rm --name backend-test $BACKEND_IMAGE &
    BACKEND_PID=$!
    sleep 5
    if kill -0 $BACKEND_PID 2>/dev/null; then
        echo -e "${GREEN}✅ 后端镜像运行正常${NC}"
        docker stop backend-test
    else
        echo -e "${RED}❌ 后端镜像启动失败${NC}"
    fi
    
    # 测试前端镜像
    echo "测试前端镜像..."
    docker run --rm -d --name frontend-test -p 18080:8080 $FRONTEND_IMAGE
    sleep 5
    if curl -f http://localhost:18080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 前端镜像运行正常${NC}"
    else
        echo -e "${RED}❌ 前端镜像健康检查失败${NC}"
    fi
    docker stop frontend-test
}

function scan_images() {
    echo -e "${YELLOW}🔍 扫描镜像安全漏洞...${NC}"
    
    if command -v trivy &> /dev/null; then
        echo "扫描后端镜像..."
        trivy image $BACKEND_IMAGE
        
        echo "扫描前端镜像..."
        trivy image $FRONTEND_IMAGE
    else
        echo -e "${YELLOW}⚠️ 未安装Trivy，跳过安全扫描${NC}"
        echo "安装Trivy: https://aquasecurity.github.io/trivy/latest/getting-started/installation/"
    fi
}

function push_to_local_registry() {
    LOCAL_REGISTRY="localhost:5000"
    
    echo -e "${YELLOW}📤 推送到本地仓库 $LOCAL_REGISTRY...${NC}"
    
    # 重新标记镜像
    docker tag $BACKEND_IMAGE $LOCAL_REGISTRY/url-manager/backend:local
    docker tag $FRONTEND_IMAGE $LOCAL_REGISTRY/url-manager/frontend:local
    
    # 推送镜像
    docker push $LOCAL_REGISTRY/url-manager/backend:local
    docker push $LOCAL_REGISTRY/url-manager/frontend:local
    
    echo -e "${GREEN}✅ 镜像已推送到本地仓库${NC}"
}

function show_help() {
    echo "URL Manager System - 本地开发脚本"
    echo
    echo "使用方法:"
    echo "  $0 [命令]"
    echo
    echo "命令:"
    echo "  build     构建所有镜像"
    echo "  backend   仅构建后端镜像"
    echo "  frontend  仅构建前端镜像"
    echo "  test      测试镜像"
    echo "  scan      安全扫描"
    echo "  push      推送到本地仓库"
    echo "  all       执行所有步骤 (构建+测试+扫描)"
    echo "  help      显示帮助信息"
    echo
    echo "示例:"
    echo "  $0 build     # 构建所有镜像"
    echo "  $0 backend   # 仅构建后端"
    echo "  $0 all       # 完整流程"
}

# 主逻辑
case "${1:-help}" in
    "build")
        build_backend
        build_frontend
        ;;
    "backend")
        build_backend
        ;;
    "frontend")
        build_frontend
        ;;
    "test")
        test_images
        ;;
    "scan")
        scan_images
        ;;
    "push")
        push_to_local_registry
        ;;
    "all")
        build_backend
        build_frontend
        test_images
        scan_images
        ;;
    "help"|*)
        show_help
        ;;
esac