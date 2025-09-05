# 多阶段构建 - 构建阶段
FROM golang:1.21-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 首先复制go mod文件（用于缓存依赖）
COPY ../go.mod ../go.sum ./

# 下载依赖
RUN go mod download

# 复制后端源代码
COPY . ./

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/server

# 运行阶段 - 使用distroless镜像
FROM gcr.io/distroless/static:nonroot

# 从构建阶段复制二进制文件
COPY --from=builder /build/main /
COPY --from=builder /build/config.yaml /
COPY --from=builder /build/internal/db/migrations ./internal/db/migrations

# 复制时区信息
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 使用非root用户
USER nonroot:nonroot

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/main", "-health-check"]

# 启动应用
ENTRYPOINT ["/main"]