# URL Manager System API 文档

## 概述

URL Manager System 提供了一套完整的 RESTful API，用于管理项目和临时 URL。

## 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: 暂无（后续可扩展）
- **数据格式**: JSON
- **字符编码**: UTF-8

## 通用响应格式

### 成功响应
```json
{
  "data": {
    // 具体数据
  }
}
```

### 错误响应
```json
{
  "error": "错误描述信息"
}
```

## 项目管理 API

### 1. 创建项目

**请求**
```
POST /projects
Content-Type: application/json

{
  "name": "项目名称",
  "description": "项目描述（可选）"
}
```

**响应**
```json
{
  "id": "uuid",
  "name": "项目名称",
  "description": "项目描述",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z"
}
```

### 2. 获取项目列表

**请求**
```
GET /projects?limit=20&offset=0
```

**响应**
```json
{
  "projects": [
    {
      "id": "uuid",
      "name": "项目名称",
      "description": "项目描述",
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

### 3. 获取单个项目

**请求**
```
GET /projects/{id}
```

**响应**
```json
{
  "id": "uuid",
  "name": "项目名称",
  "description": "项目描述",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z"
}
```

### 4. 更新项目

**请求**
```
PUT /projects/{id}
Content-Type: application/json

{
  "name": "新项目名称",
  "description": "新项目描述"
}
```

**响应**
```json
{
  "id": "uuid",
  "name": "新项目名称",
  "description": "新项目描述",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T12:00:00Z"
}
```

### 5. 删除项目

**请求**
```
DELETE /projects/{id}
```

**响应**
```
204 No Content
```

**注意**: 只有当项目下没有活跃的 URL 时才能删除项目。

## URL 管理 API

### 1. 创建临时 URL

**请求**
```
POST /projects/{project_id}/urls
Content-Type: application/json

{
  "image": "nginx:latest",
  "ttl_seconds": 3600,
  "replicas": 1,
  "env": [
    {
      "name": "ENV_VAR_NAME",
      "value": "env_var_value"
    }
  ],
  "resources": {
    "requests": {
      "cpu": "100m",
      "memory": "128Mi"
    },
    "limits": {
      "cpu": "500m",
      "memory": "512Mi"
    }
  }
}
```

**响应**
```json
{
  "url": "https://example.com/abc123",
  "id": "uuid"
}
```

### 2. 获取项目的 URL 列表

**请求**
```
GET /projects/{project_id}/urls?limit=20&offset=0
```

**响应**
```json
{
  "urls": [
    {
      "id": "uuid",
      "project_id": "uuid",
      "path": "/abc123",
      "image": "nginx:latest",
      "env": [],
      "replicas": 1,
      "resources": {
        "requests": {
          "cpu": "100m",
          "memory": "128Mi"
        },
        "limits": {
          "cpu": "500m",
          "memory": "512Mi"
        }
      },
      "status": "active",
      "k8s_deployment_name": "ephemeral-abc123",
      "k8s_service_name": "svc-ephemeral-abc123",
      "k8s_secret_name": null,
      "error_message": null,
      "expire_at": "2023-01-01T01:00:00Z",
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:05:00Z"
    }
  ],
  "total": 1
}
```

### 3. 获取单个 URL

**请求**
```
GET /urls/{id}
```

**响应**
```json
{
  "id": "uuid",
  "project_id": "uuid",
  "path": "/abc123",
  "image": "nginx:latest",
  "env": [],
  "replicas": 1,
  "resources": {
    "requests": {
      "cpu": "100m",
      "memory": "128Mi"
    },
    "limits": {
      "cpu": "500m",
      "memory": "512Mi"
    }
  },
  "status": "active",
  "k8s_deployment_name": "ephemeral-abc123",
  "k8s_service_name": "svc-ephemeral-abc123",
  "k8s_secret_name": null,
  "error_message": null,
  "expire_at": "2023-01-01T01:00:00Z",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:05:00Z",
  "project": {
    "id": "uuid",
    "name": "项目名称",
    "description": "项目描述",
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

### 4. 删除 URL

**请求**
```
DELETE /urls/{id}
```

**响应**
```
204 No Content
```

## 状态码说明

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 204 | 删除成功（无内容返回） |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 409 | 资源冲突（如删除有活跃URL的项目） |
| 500 | 服务器内部错误 |

## URL 状态说明

| 状态 | 说明 |
|------|------|
| creating | 正在创建相关资源 |
| active | 运行中，可以正常访问 |
| deleting | 正在删除资源 |
| deleted | 已删除 |
| failed | 创建或运行失败 |

## 限制说明

### 镜像限制
- 只能使用白名单中的镜像
- 默认允许: `nginx:latest`, `nginx:1.21`, `httpd:latest`, `httpd:2.4`

### 资源限制
- 最大副本数: 3
- 最大 TTL: 7 天 (604800 秒)
- 默认 CPU 限制: 500m
- 默认内存限制: 512Mi

### 命名规范
- 项目名称: 1-100 个字符，只能包含字母、数字、连字符和下划线
- 环境变量名: 只能包含字母、数字和下划线，不能以数字开头

## 健康检查

**请求**
```
GET /health
```

**响应**
```json
{
  "status": "ok",
  "service": "url-manager-system"
}
```