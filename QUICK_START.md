# 🚀 Comic Viewer 快速启动指南

## 一分钟快速启动

```bash
# 1. 进入项目目录
cd ~/project/comic-viewer-claude

# 2. 安装依赖（首次运行）
make install

# 3. 启动服务
./start.sh

# 4. 打开浏览器访问
# 前端: http://localhost:5173
# 后端: http://localhost:8080
```

## 停止服务

```bash
./stop.sh
```

## 常用命令

```bash
# 查看帮助
make help

# 只启动后端
make backend

# 只启动前端
make frontend

# 运行测试
make test

# 构建生产版本
make build

# 清理构建文件
make clean
```

## 目录说明

- `backend/` - Go 后端代码
- `frontend/` - React 前端代码
- `docs/` - 详细文档
- `data/` - 数据库文件（自动创建）
- `logs/` - 日志文件（自动创建）
- `downloads/` - 下载目录（自动创建）

## 配置修改

编辑 `backend/config.yaml` 修改配置：

```yaml
server:
  port: 8080              # 后端端口
  
crawler:
  base_urls:          # 目标站点
    - "https://example.com"
  request_delay_min: 2.0  # 请求延迟
  
cache:
  cover_max: 2000         # 最大封面缓存数
```

## 常见问题

### 端口被占用

```bash
# 查看端口占用
lsof -i :8080
lsof -i :5173

# 杀死进程
kill -9 <PID>
```

### 依赖安装失败

```bash
# Go 依赖
cd backend
go clean -modcache
go mod tidy

# npm 依赖
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### 数据库问题

```bash
# 删除数据库重新初始化
rm -rf data/comics.db
# 重启服务会自动创建新数据库
```

## 查看日志

```bash
# 后端日志
tail -f logs/backend.log

# 前端日志
tail -f logs/frontend.log

# 应用日志
tail -f logs/app.log
```

## 更多帮助

- 📖 [完整文档](README.md)
- 🔧 [开发指南](docs/DEVELOPMENT.md)
- 🚀 [部署指南](docs/DEPLOYMENT.md)
- 📡 [API 文档](docs/API.md)

---

**祝你使用愉快！** ✨
