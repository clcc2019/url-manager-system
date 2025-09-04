# URL Manager System 部署指南

## 环境要求

### 基础环境
- Kubernetes 1.24+
- Docker 20.10+
- Helm 3.8+

### 资源要求
- **最小配置**: 2 CPU核心, 4GB内存
- **推荐配置**: 4 CPU核心, 8GB内存
- **存储**: 20GB可用空间

### 依赖服务
- PostgreSQL 14+
- Redis 6+
- Ingress Controller (推荐 NGINX)

## 部署方式

### 方式一：使用 Helm Chart (推荐)

#### 1. 添加依赖仓库
```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

#### 2. 安装
```bash
# 使用默认配置
helm install url-manager deployments/helm/url-manager --namespace url-manager --create-namespace

# 或使用自定义配置
helm install url-manager deployments/helm/url-manager \
  --namespace url-manager \
  --create-namespace \
  --set backend.config.k8s.default_domain=your-domain.com \
  --set ingress.hosts[0].host=url-manager.your-domain.com
```

#### 3. 验证部署
```bash
kubectl get pods -n url-manager
kubectl get svc -n url-manager
kubectl get ingress -n url-manager
```

### 方式二：使用 Docker Compose (开发环境)

#### 1. 启动开发环境
```bash
make dev
```

#### 2. 查看日志
```bash
make dev-logs
```

#### 3. 停止环境
```bash
make dev-stop
```

### 方式三：手动部署

#### 1. 创建命名空间
```bash
kubectl create namespace url-manager
```

#### 2. 部署 PostgreSQL
```bash
kubectl apply -f deployments/k8s/postgresql.yaml -n url-manager
```

#### 3. 部署 Redis
```bash
kubectl apply -f deployments/k8s/redis.yaml -n url-manager
```

#### 4. 部署应用
```bash
kubectl apply -f deployments/k8s/backend.yaml -n url-manager
kubectl apply -f deployments/k8s/frontend.yaml -n url-manager
kubectl apply -f deployments/k8s/ingress.yaml -n url-manager
```

## 配置选项

### Helm Chart 主要配置

```yaml
# values.yaml 示例
backend:
  replicaCount: 2
  image:
    repository: url-manager/backend
    tag: "latest"
  
  config:
    k8s:
      namespace: "default"
      default_domain: "example.com"
      ingress_class: "nginx"
    
    security:
      allowed_images:
        - "nginx:latest"
        - "nginx:1.21"
        - "httpd:latest"
        - "registry.your-company.com/"
      max_replicas: 3
      max_ttl_seconds: 604800

frontend:
  replicaCount: 2
  image:
    repository: url-manager/frontend
    tag: "latest"

ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: url-manager.example.com
      paths:
        - path: /
          pathType: Prefix
          service: frontend
        - path: /api
          pathType: Prefix
          service: backend

postgresql:
  enabled: true
  auth:
    postgresPassword: "secure-password"
    database: "url_manager"

redis:
  enabled: true
  auth:
    enabled: false
```

### 环境变量配置

#### 后端环境变量
```bash
DEBUG=false
DATABASE_URL=postgres://user:password@host:5432/url_manager
REDIS_ADDRESS=redis:6379
K8S_IN_CLUSTER=true
K8S_NAMESPACE=default
DEFAULT_DOMAIN=example.com
```

#### 前端环境变量
```bash
VITE_API_BASE_URL=https://url-manager.example.com/api/v1
```

## 安全配置

### 1. RBAC 配置

创建必要的 ServiceAccount 和权限：

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: url-manager
  namespace: url-manager

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: url-manager
rules:
- apiGroups: [""]
  resources: ["pods", "services", "secrets"]
  verbs: ["get", "list", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "create", "update", "patch", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["get", "list", "create", "update", "patch", "delete"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: url-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: url-manager
subjects:
- kind: ServiceAccount
  name: url-manager
  namespace: url-manager
```

### 2. 网络策略

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: url-manager-netpol
  namespace: url-manager
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: url-manager
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 5432  # PostgreSQL
    - protocol: TCP
      port: 6379  # Redis
  - to: []
    ports:
    - protocol: TCP
      port: 443   # Kubernetes API
```

### 3. Pod 安全策略

```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: url-manager-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
```

## 监控和日志

### 1. Prometheus 监控

```yaml
# ServiceMonitor 示例
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: url-manager
  namespace: url-manager
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: url-manager
  endpoints:
  - port: http
    path: /metrics
```

### 2. 日志收集

推荐使用 ELK Stack 或 Loki + Grafana 收集日志：

```yaml
# Fluentd 配置示例
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/containers/url-manager*.log
      pos_file /var/log/fluentd-containers.log.pos
      tag kubernetes.*
      read_from_head true
      <parse>
        @type json
      </parse>
    </source>
    
    <match kubernetes.**>
      @type elasticsearch
      host elasticsearch.logging.svc.cluster.local
      port 9200
      index_name url-manager-logs
    </match>
```

## 备份和恢复

### 1. 数据库备份

```bash
# 创建备份任务
kubectl create job --from=cronjob/postgres-backup postgres-backup-manual -n url-manager

# 手动备份
kubectl exec -it postgres-pod -n url-manager -- pg_dump -U postgres url_manager > backup.sql
```

### 2. 配置备份

```bash
# 备份 Helm Release
helm get values url-manager -n url-manager > url-manager-values-backup.yaml

# 备份 Kubernetes 资源
kubectl get all -n url-manager -o yaml > url-manager-resources-backup.yaml
```

## 故障排除

### 1. 常见问题

#### Pod 无法启动
```bash
# 查看 Pod 状态
kubectl get pods -n url-manager

# 查看 Pod 日志
kubectl logs -f deployment/url-manager-backend -n url-manager

# 查看 Pod 事件
kubectl describe pod <pod-name> -n url-manager
```

#### 数据库连接失败
```bash
# 检查数据库服务
kubectl get svc -n url-manager | grep postgres

# 测试数据库连接
kubectl exec -it postgres-pod -n url-manager -- psql -U postgres -d url_manager -c "SELECT 1"
```

#### Ingress 路由问题
```bash
# 检查 Ingress 状态
kubectl get ingress -n url-manager

# 检查 Ingress Controller 日志
kubectl logs -f deployment/ingress-nginx-controller -n ingress-nginx
```

### 2. 调试技巧

#### 启用调试模式
```bash
# 通过 Helm 启用调试
helm upgrade url-manager deployments/helm/url-manager \
  --namespace url-manager \
  --set backend.env.DEBUG=true
```

#### 访问应用日志
```bash
# 实时查看后端日志
kubectl logs -f deployment/url-manager-backend -n url-manager

# 查看最近的错误日志
kubectl logs --tail=100 deployment/url-manager-backend -n url-manager | grep ERROR
```

## 性能优化

### 1. 资源配置优化

```yaml
backend:
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 2000m
      memory: 2Gi

frontend:
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
```

### 2. 自动扩容配置

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

## 升级指南

### 1. 滚动升级

```bash
# 升级到新版本
helm upgrade url-manager deployments/helm/url-manager \
  --namespace url-manager \
  --set backend.image.tag=v1.1.0 \
  --set frontend.image.tag=v1.1.0
```

### 2. 回滚

```bash
# 查看发布历史
helm history url-manager -n url-manager

# 回滚到上一个版本
helm rollback url-manager -n url-manager

# 回滚到指定版本
helm rollback url-manager 2 -n url-manager
```

## 维护任务

### 1. 清理过期数据

```bash
# 手动清理过期 URL（已集成在系统中自动执行）
kubectl exec -it deployment/url-manager-backend -n url-manager -- \
  /app/main -cleanup
```

### 2. 数据库维护

```bash
# 执行数据库清理
kubectl exec -it postgres-pod -n url-manager -- \
  psql -U postgres -d url_manager -c "VACUUM ANALYZE;"
```