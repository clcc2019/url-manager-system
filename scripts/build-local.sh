#!/bin/bash

# 本地构建脚本
# 用于测试Docker镜像构建

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🏗️  开始本地构建测试${NC}"

# 检查Docker是否可用
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker未安装或不可用${NC}"
    exit 1
fi

# 构建后端镜像
echo -e "${YELLOW}📦 构建后端镜像...${NC}"
docker build -f backend/Dockerfile -t url-manager-backend:local .
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ 后端镜像构建成功${NC}"
else
    echo -e "${RED}❌ 后端镜像构建失败${NC}"
    exit 1
fi

# 构建前端镜像
echo -e "${YELLOW}🎨 构建前端镜像...${NC}"
docker build -f frontend/Dockerfile -t url-manager-frontend:local .
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ 前端镜像构建成功${NC}"
else
    echo -e "${RED}❌ 前端镜像构建失败${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 所有镜像构建完成!${NC}"
echo -e "${GREEN}后端镜像: url-manager-backend:local${NC}"
echo -e "${GREEN}前端镜像: url-manager-frontend:local${NC}"

# 显示镜像信息
echo -e "\n${YELLOW}📋 镜像信息:${NC}"
docker images | grep url-manager
