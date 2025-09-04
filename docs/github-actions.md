# GitHub Actions 配置指南

## 概述

本项目使用GitHub Actions进行CI/CD，自动构建Docker镜像并推送到GitHub Packages (ghcr.io)。

## 工作流说明

### 1. CI/CD Pipeline (`.github/workflows/ci-cd.yml`)

**触发条件:**
- 推送到 `main` 或 `develop` 分支
- 创建Pull Request到 `main` 或 `develop` 分支
- 推送标签（`v*`格式）

**主要步骤:**
1. **代码质量检查** - 运行linting和测试
2. **构建和推送镜像** - 构建Docker镜像并推送到ghcr.io
3. **安全扫描** - 使用Trivy扫描镜像漏洞
4. **部署到staging** - 自动部署到staging环境（develop分支）
5. **创建Release** - 为标签创建GitHub Release

### 2. Release Pipeline (`.github/workflows/release.yml`)

**触发条件:**
- 推送版本标签（`v*`格式，如 `v1.0.0`）

**功能:**
- 构建生产版本镜像
- 生成变更日志
- 创建GitHub Release
- 提供部署命令

### 3. Build Images (`.github/workflows/build-images.yml`)

**触发条件:**
- 手动触发（workflow_dispatch）

**功能:**
- 按需构建指定版本的镜像
- 支持选择构建平台

### 4. Code Quality (`.github/workflows/code-quality.yml`)

**触发条件:**
- 推送到主要分支
- Pull Request

**功能:**
- Go代码质量检查（golangci-lint, staticcheck）
- 前端代码检查（ESLint, TypeScript）
- 安全扫描（Gosec, Nancy）
- 代码覆盖率报告

## 镜像标签策略

### 自动标签

GitHub Actions会根据触发条件自动生成以下标签：

| 触发条件 | 标签示例 |
|---------|---------|
| main分支推送 | `latest`, `main-abc123` |
| develop分支推送 | `develop`, `develop-abc123` |
| Pull Request | `pr-123` |
| 版本标签 | `v1.0.0`, `1.0`, `stable` |
| Git SHA | `sha-abc123` |

### 镜像命名

- **后端镜像**: `ghcr.io/clcc2019/url-manager-system/backend:tag`
- **前端镜像**: `ghcr.io/clcc2019/url-manager-system/frontend:tag`

## 环境变量和密钥

### 自动提供的密钥

GitHub Actions自动提供以下密钥，无需手动配置：

- `GITHUB_TOKEN` - 用于访问GitHub API和Packages
- `github.actor` - 当前用户名
- `github.repository` - 仓库名称

### 可选配置

如需部署到外部环境，可在Repository Settings > Secrets添加：

```bash
# Kubernetes配置（用于自动部署）
KUBE_CONFIG_DATA=<base64编码的kubeconfig>

# Docker Registry（如使用其他registry）
DOCKER_REGISTRY=your-registry.com
DOCKER_USERNAME=username
DOCKER_PASSWORD=password

# 通知配置
SLACK_WEBHOOK=https://hooks.slack.com/...
DISCORD_WEBHOOK=https://discord.com/api/webhooks/...
```

## 使用指南

### 1. 开发流程

```bash
# 1. 创建功能分支
git checkout -b feature/new-feature

# 2. 开发和测试
git add .
git commit -m "feat: add new feature"

# 3. 推送分支
git push origin feature/new-feature

# 4. 创建Pull Request
# GitHub Actions会自动运行测试和质量检查

# 5. 合并到develop分支
# 自动部署到staging环境

# 6. 合并到main分支
# 构建latest镜像
```

### 2. 版本发布

```bash
# 使用自动化脚本
./scripts/release.sh v1.0.0

# 或手动创建标签
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### 3. 手动构建镜像

1. 访问 GitHub Actions 页面
2. 选择 "Build Docker Images" 工作流
3. 点击 "Run workflow"
4. 输入标签和平台信息
5. 点击 "Run workflow"

### 4. 本地测试

```bash
# 构建和测试本地镜像
make build-and-test

# 或使用脚本
./scripts/build-local.sh all
```

## 故障排除

### 1. 构建失败

**检查构建日志:**
1. 访问GitHub Actions页面
2. 点击失败的工作流
3. 查看详细日志

**常见问题:**
- 测试失败 → 修复代码错误
- Docker构建失败 → 检查Dockerfile
- 权限问题 → 检查GITHUB_TOKEN权限

### 2. 推送到Packages失败

**可能原因:**
- 权限不足 → 检查Repository设置中的Actions权限
- 镜像标签冲突 → 使用不同的标签

**解决方法:**
1. 确保Repository > Settings > Actions > General中启用了"Read and write permissions"
2. 检查Package设置中的访问权限

### 3. 部署失败

**检查项目:**
- Kubernetes配置正确性
- 镜像标签是否存在
- 网络连接问题

## 监控和通知

### 1. 构建状态徽章

在README中添加构建状态徽章：

```markdown
![CI/CD](https://github.com/clcc2019/url-manager-system/workflows/CI/CD%20Pipeline/badge.svg)
![Security](https://github.com/clcc2019/url-manager-system/workflows/Code%20Quality/badge.svg)
```

### 2. 通知设置

可以配置以下通知：
- Slack通知（构建结果）
- 邮件通知（发布新版本）
- Discord通知（部署状态）

### 3. 监控仪表板

- **GitHub Actions**: 查看工作流历史
- **GitHub Packages**: 查看镜像统计
- **Dependabot**: 依赖安全更新

## 最佳实践

### 1. 分支策略

- `main` - 生产分支，触发生产部署
- `develop` - 开发分支，触发staging部署
- `feature/*` - 功能分支，触发测试
- `hotfix/*` - 热修复分支

### 2. 提交信息规范

使用约定式提交：
```
feat: 新功能
fix: 修复bug
docs: 文档更新
style: 代码格式
refactor: 重构
test: 测试
chore: 构建/工具
```

### 3. 版本号规范

使用语义化版本：
- `v1.0.0` - 主版本.次版本.修订版本
- `v1.0.0-alpha.1` - 预发布版本
- `v1.0.0+20230101` - 构建元数据

### 4. 安全考虑

- 定期更新依赖
- 启用安全扫描
- 使用最小权限原则
- 定期审查访问权限

## 参考链接

- [GitHub Actions文档](https://docs.github.com/en/actions)
- [GitHub Packages文档](https://docs.github.com/en/packages)
- [Docker构建最佳实践](https://docs.docker.com/develop/dev-best-practices/)
- [语义化版本](https://semver.org/)