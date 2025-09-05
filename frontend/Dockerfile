# 多阶段构建 - 构建阶段
FROM node:18-alpine AS builder

# 设置工作目录
WORKDIR /build

# 设置构建时的环境变量（可选，主要用于开发环境）
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

# 复制package文件并安装依赖
COPY frontend/package*.json ./
RUN npm ci && npm cache clean --force

# 复制源代码
COPY frontend/ .

# 构建应用
RUN npm run build

# 生产阶段 - 使用NGINX Alpine
FROM nginx:1.25-alpine

# 安装安全更新
RUN apk --no-cache upgrade

# nginx用户已经存在，无需创建

# 复制构建产物
COPY --from=builder /build/dist /usr/share/nginx/html

# 复制nginx配置
COPY frontend/nginx.conf /etc/nginx/nginx.conf

# 设置正确的权限
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    chown -R nginx:nginx /etc/nginx/conf.d

# 修改nginx配置以支持非root用户
RUN touch /var/run/nginx.pid && \
    chown -R nginx:nginx /var/run/nginx.pid

# 使用非root用户运行
USER nginx

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/ || exit 1

# 启动NGINX
CMD ["nginx", "-g", "daemon off;"]