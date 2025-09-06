# 模板功能重构指南

## 概述

本次重构对模板功能进行了全面升级，现在支持：

1. **YAML导入和解析** - 自动解析Kubernetes YAML配置
2. **结构化字段提取** - 提取镜像、环境变量、端口、资源配置等关键信息
3. **可视化编辑** - 通过表单界面编辑解析后的参数
4. **云原生风格** - 完全兼容Kubernetes原生配置

## 新功能特性

### 1. YAML解析功能

系统现在能够自动解析Deployment资源中的以下字段：

- **镜像地址** (`image`)
- **环境变量** (`env`)
- **启动命令** (`command`)
- **启动参数** (`args`)
- **容器端口** (`ports`)
- **资源配额** (`resources`)
- **卷挂载** (`volumeMounts`)
- **存活检测** (`livenessProbe`)
- **就绪检测** (`readinessProbe`)
- **工作目录** (`workingDir`)
- **安全上下文** (`securityContext`)

### 2. 模板数据结构

```typescript
interface TemplateSpec {
  image: string;
  env?: EnvironmentVar[];
  command?: string[];
  args?: string[];
  ports?: ContainerPort[];
  resources?: ResourceLimits;
  volume_mounts?: VolumeMount[];
  config_maps?: ConfigMap[];
  secrets?: Secret[];
  liveness_probe?: Probe;
  readiness_probe?: Probe;
  working_dir?: string;
  security_context?: SecurityContext;
}
```

### 3. URL创建优化

基于模板创建URL时，现在会：
- 自动从模板解析规格中获取镜像信息
- 继承模板的环境变量配置
- 使用模板的资源限制
- 应用模板的工作目录和命令配置

## 使用示例

### 1. 创建模板

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${DEPLOYMENT_NAME}
spec:
  template:
    spec:
      containers:
      - name: app
        image: nginx:1.20
        ports:
        - containerPort: 80
        env:
        - name: PROJECT_NAME
          value: ${PROJECT_NAME}
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 256Mi
```

### 2. 系统会自动解析为：

```json
{
  "image": "nginx:1.20",
  "env": [
    {
      "name": "PROJECT_NAME",
      "value": "${PROJECT_NAME}"
    }
  ],
  "ports": [
    {
      "container_port": 80,
      "protocol": "TCP"
    }
  ],
  "resources": {
    "requests": {
      "cpu": "100m",
      "memory": "128Mi"
    },
    "limits": {
      "cpu": "500m",
      "memory": "256Mi"
    }
  }
}
```

## 数据库变更

### 新增字段

- `app_templates.parsed_spec` - JSON格式存储解析后的结构化数据
- 支持完整的Kubernetes配置解析和存储

### 兼容性

- 向后兼容现有模板
- 旧模板会自动解析并填充parsed_spec字段
- API接口保持兼容

## 前端更新

### 新增类型定义

```typescript
// 新增的完整类型系统
interface TemplateSpec {...}
interface ContainerPort {...}
interface VolumeMount {...}
// ... 其他类型
```

### 表单组件

- 支持YAML文本编辑
- 自动解析显示结构化字段
- 可视化编辑各个配置项
- 实时预览和验证

## 技术实现

### 1. YAML解析器 (`utils/yaml_parser.go`)

```go
func ParseYAMLToTemplateSpec(yamlContent string) (*models.TemplateSpec, error) {
    // 解析YAML并提取结构化数据
}
```

### 2. 数据模型扩展

- 新增TemplateSpec结构
- 支持JSON序列化存储
- 完整的类型安全

### 3. 服务层更新

- 模板创建时自动解析YAML
- URL创建时使用解析后的规格
- 保持向后兼容性

## 优势

1. **用户友好** - 无需手动填写复杂配置
2. **类型安全** - 完整的TypeScript类型支持
3. **云原生** - 完全兼容Kubernetes标准
4. **可扩展** - 支持未来添加更多配置项
5. **向后兼容** - 不破坏现有功能

## 迁移指南

### 现有模板

- 无需手动迁移，系统会自动解析
- 可以在编辑时查看解析结果
- 支持逐步升级到新格式

### 新模板

- 推荐使用标准Kubernetes YAML格式
- 系统会自动提取和验证配置
- 支持模板变量替换

## 后续计划

1. **UI优化** - 改进模板编辑界面
2. **验证增强** - 添加更严格的YAML验证
3. **模板市场** - 支持模板分享和下载
4. **高级功能** - 支持Helm Chart集成
