# 部署指南

## 生产环境部署

### 方式一：Docker 部署（推荐）

#### 1. 创建 Dockerfile

**后端 Dockerfile** (`backend/Dockerfile`):

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o comic-viewer cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

COPY --from=builder /app/comic-viewer .
COPY --from=builder /app/config.yaml .

RUN mkdir -p data logs downloads

EXPOSE 8080

CMD ["./comic-viewer"]
```

**前端 Dockerfile** (`frontend/Dockerfile`):

```dockerfile
FROM node:18-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

#### 2. 创建 docker-compose.yml

```yaml
version: '3.8'

services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    volumes:
      - ./data:/root/data
      - ./logs:/root/logs
      - ./downloads:/root/downloads
    environment:
      - GIN_MODE=release
    restart: unless-stopped

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "80:80"
    depends_on:
    - backend
    restart: unless-stopped
```

#### 3. 启动服务

```bash
docker-compose up -d
```

### 方式二：传统部署

#### 后端部署

1. **编译**

```bash
cd backend
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o comic-viewer cmd/server/main.go
```

2. **配置 systemd 服务**

创建 `/etc/systemd/system/comic-viewer.service`:

```ini
[Unit]
Description=Comic Viewer Backend
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/comic-viewer
ExecStart=/opt/comic-viewer/comic-viewer
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

3. **启动服务**

```bash
sudo systemctl daemon-reload
sudo systemctl enable comic-viewer
sudo systemctl start comic-viewer
```

#### 前端部署

1. **构建**

```bash
cd frontend
npm run build
```

2. **Nginx 配置**

创建 `/etc/nginx/sites-available/comic-viewer`:

```nginx
server {
    listen 80;
    server_name your-domain.com;

    root /var/www/comic-viewer;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection 'upgrade';
      proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

3. **启用站点**

```bash
sudo ln -s /etc/nginx/sites-available/comic-viewer /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 方式三：使用 PM2 部署

#### 1. 安装 PM2

```bash
npm install -g pm2
```

#### 2. 创建 ecosystem.config.js

```javascript
module.exports = {
  apps: [
    {
   name: 'comic-viewer-backend',
      cwd: './backend',
   script: 'go',
      args: 'run cmd/server/main.go',
      env: {
      GIN_MODE: 'release'
      }
    },
    {
      name: 'comic-viewer-frontend',
      cwd: './frontend',
      script: 'npm',
      args: 'run preview'
    }
  ]
};
```

#### 3. 启动服务

```bash
pm2 start ecosystem.config.js
pm2 save
pm2 startup
```

## 性能优化

### 后端优化

1. **启用 Gzip 压缩**

```go
import "github.com/gin-contrib/gzip"

engine.Use(gzip.Gzip(gzip.DefaultCompression))
```

2. **添加缓存层**

使用 Redis 缓存热点数据：

```bash
docker run -d -p 6379:6379 redis:alpine
```

3. **数据库优化**

- 添加索引
- 使用连接池
- 考虑迁移到 PostgreSQL

### 前端优化

1. **CDN 加速**

将静态资源上传到 CDN

2. **启用 HTTP/2**

Nginx 配置：

```nginx
listen 443 ssl http2;
```

3. **压缩资源**

```bash
npm install -D vite-plugin-compression
```

## 监控和日志

### 日志管理

使用 Logrotate 管理日志：

```bash
/opt/comic-viewer/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
}
```

### 监控

使用 Prometheus + Grafana 监控：

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'comic-viewer'
    static_configs:
      - targets: ['localhost:8080']
```

## 备份策略

### 数据库备份

```bash
#!/bin/bash
# backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/comic-viewer"

mkdir -p $BACKUP_DIR

# 备份数据库
cp /opt/comic-viewer/data/comics.db $BACKUP_DIR/comics_$DATE.db

# 保留最近 7 天的备份
find $BACKUP_DIR -name "comics_*.db" -mtime +7 -delete
```

设置定时任务：

```bash
crontab -e
# 每天凌晨 2 点备份
0 2 * * /opt/comic-viewer/backup.sh
```

## 安全建议

1. **使用 HTTPS**

```bash
sudo certbot --nginx -d your-domain.com
```

2. **防火墙配置**

```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

3. **限制 API 访问频率**

使用 Gin 的限流中间件

4. **定期更新依赖**

```bash
go get -u ./...
npm update
```

## 故障排查

### 常见问题

1. **端口被占用**

```bash
lsof -i :8080
kill -9 <PID>
```

2. **数据库锁定**

```bash
fuser -k data/comics.db
```

3. **内存不足**

增加 swap 空间或升级服务器

### 查看日志

```bash
# 后端日志
tail -f logs/app.log

# Nginx 日志
tail -f /var/log/nginx/error.log

# systemd 日志
journalctl -u comic-viewer -f
```

## 回滚策略

1. **保留旧版本**

```bash
cp comic-viewer comic-viewer.backup
```

2. **快速回滚**

```bash
systemctl stop comic-viewer
cp comic-viewer.backup comic-viewer
systemctl start comic-viewer
```

## 扩展性

### 水平扩展

使用 Nginx 负载均衡：

```nginx
upstream backend {
    server backend1:8080;
    server backend2:8080;
    server backend3:8080;
}

server {
    location /api {
        proxy_pass http://backend;
    }
}
```

### 数据库分离

将 SQLite 迁移到 PostgreSQL 或 MySQL

## 维护计划

- **每日**: 检查日志、监控指标
- **每周**: 备份验证、性能分析
- **每月**: 依赖更新、安全审计
- **每季度**: 容量规划、架构优化
