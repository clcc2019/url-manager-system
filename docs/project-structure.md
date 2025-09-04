url-manager-system/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go                 # 主程序入口
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handlers/              # HTTP处理器
│   │   │   ├── middleware/            # 中间件
│   │   │   └── routes/                # 路由定义
│   │   ├── config/                    # 配置管理
│   │   ├── db/
│   │   │   ├── migrations/            # 数据库迁移
│   │   │   └── models/                # 数据模型
│   │   ├── k8s/                       # Kubernetes客户端
│   │   ├── services/                  # 业务逻辑服务
│   │   └── utils/                     # 工具函数
│   ├── pkg/                           # 公共包
│   ├── deployments/                   # 部署相关配置
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── frontend/
│   ├── public/
│   ├── src/
│   │   ├── components/                # 公共组件
│   │   ├── pages/                     # 页面组件
│   │   ├── services/                  # API服务
│   │   ├── hooks/                     # 自定义Hook
│   │   ├── types/                     # TypeScript类型定义
│   │   ├── utils/                     # 工具函数
│   │   └── App.tsx
│   ├── package.json
│   ├── vite.config.ts
│   └── Dockerfile
├── deployments/
│   ├── helm/                          # Helm Charts
│   │   ├── url-manager/
│   │   │   ├── templates/
│   │   │   ├── values.yaml
│   │   │   └── Chart.yaml
│   │   └── dependencies/              # 依赖服务Charts
│   ├── k8s/                          # 原生K8s YAML
│   └── docker-compose.yml            # 本地开发环境
├── docs/
│   ├── api.md                        # API文档
│   ├── deployment.md                 # 部署指南
│   ├── architecture.md               # 架构文档
│   └── development.md                # 开发指南
├── scripts/                          # 构建和部署脚本
├── .gitignore
├── Makefile
└── README.md