#!/bin/bash

# 自动化版本发布脚本
# 使用方法: ./scripts/release.sh [版本号]
# 例如: ./scripts/release.sh v1.0.0

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查参数
if [ $# -eq 0 ]; then
    echo -e "${RED}错误: 请提供版本号${NC}"
    echo "使用方法: $0 <版本号>"
    echo "例如: $0 v1.0.0"
    exit 1
fi

VERSION=$1

# 验证版本号格式 (vX.Y.Z)
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}错误: 版本号格式不正确${NC}"
    echo "版本号应该是 vX.Y.Z 格式，例如: v1.0.0"
    exit 1
fi

echo -e "${GREEN}🚀 开始发布版本 $VERSION${NC}"

# 检查是否在main分支
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo -e "${YELLOW}警告: 当前不在main分支 (当前: $CURRENT_BRANCH)${NC}"
    read -p "是否继续? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "已取消发布"
        exit 1
    fi
fi

# 检查工作目录是否干净
if [[ -n $(git status --porcelain) ]]; then
    echo -e "${RED}错误: 工作目录不干净，请先提交或暂存更改${NC}"
    git status
    exit 1
fi

# 检查是否有远程更新
echo -e "${YELLOW}📡 检查远程更新...${NC}"
git fetch origin

LOCAL=$(git rev-parse @)
REMOTE=$(git rev-parse @{u})

if [ $LOCAL != $REMOTE ]; then
    echo -e "${RED}错误: 本地分支不是最新的，请先pull最新代码${NC}"
    exit 1
fi

# 检查版本号是否已存在
if git tag -l | grep -q "^$VERSION$"; then
    echo -e "${RED}错误: 版本 $VERSION 已存在${NC}"
    exit 1
fi

# 运行测试
echo -e "${YELLOW}🧪 运行测试...${NC}"
cd backend && go test ./... && cd ..
cd frontend && npm test -- --run && cd ..

# 更新版本信息
echo -e "${YELLOW}📝 更新版本信息...${NC}"

# 更新Helm Chart版本
if [ -f "deployments/helm/url-manager/Chart.yaml" ]; then
    # 移除v前缀用于Helm版本
    CHART_VERSION=${VERSION#v}
    sed -i.bak "s/^version:.*/version: $CHART_VERSION/" deployments/helm/url-manager/Chart.yaml
    sed -i.bak "s/^appVersion:.*/appVersion: \"$VERSION\"/" deployments/helm/url-manager/Chart.yaml
    rm deployments/helm/url-manager/Chart.yaml.bak
    git add deployments/helm/url-manager/Chart.yaml
fi

# 提交版本更新
if [[ -n $(git status --porcelain) ]]; then
    git commit -m "chore: bump version to $VERSION"
fi

# 创建标签
echo -e "${YELLOW}🏷️  创建标签 $VERSION...${NC}"
git tag -a $VERSION -m "Release $VERSION"

# 推送到远程
echo -e "${YELLOW}📤 推送到远程仓库...${NC}"
git push origin main
git push origin $VERSION

echo -e "${GREEN}✅ 版本 $VERSION 发布成功!${NC}"
echo -e "${GREEN}GitHub Actions 将自动构建和发布Docker镜像${NC}"
echo -e "${GREEN}查看发布状态: https://github.com/clcc2019/url-manager-system/actions${NC}"

# 显示部署命令
echo -e "\n${YELLOW}🚀 部署命令:${NC}"
echo "helm upgrade --install url-manager ./deployments/helm/url-manager \\"
echo "  --namespace url-manager \\"
echo "  --create-namespace \\"
echo "  --set backend.image.tag=$VERSION \\"
echo "  --set frontend.image.tag=$VERSION"