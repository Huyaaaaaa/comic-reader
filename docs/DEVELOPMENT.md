# Comic Viewer - 开发指南

## 项目概述

这是一个基于 Go + React 的漫画浏览器应用，完全重构自 Python Flet 版本。

## 技术栈

### 后端
- Go 1.21+
- Gin (Web框架)
- GORM (ORM)
- SQLite (数据库)
- Zap (日志)
- Viper (配置管理)

### 前端
- React 18
- Vite (构建工具)
- Ant Design (UI组件库)
- Zustand (状态管理)
- React Router (路由)
- Axios (HTTP客户端)

## 快速开始

### 后端启动

```bash
cd backend

# 初始化 Go 模块
go mod tidy

# 运行服务器
go run cmd/server/main.go
```

服务器将在 `http://localhost:8080` 启动

### 前端启动

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端将在 `http://localhost:5173` 启动

## 项目结构

```
comic-viewer-claude/
├── backend/          # Go 后端
│   ├── cmd/
│   │   └── server/     # 主程序入口
│   │       └── main.go
│   ├── internal/           # 内部包
│   │   ├── api/           # HTTP handlers
│   │   │   ├── comic_handler.go
│   │   └── router.go
│   │   ├── service/       # 业务逻辑层
│   │   │   └── comic_service.go
│   │   ├── repository/    # 数据访问层
│   │   │   └── repository.go
│   │   ├── crawler/     # 爬虫客户端
│   │   │   └── client.go
│   │   ├── parser/        # HTML解析
│   │   │   └── parser.go
│   │   ├── crypto/     # 加密解密
│   │   │   └── crypto.go
│   │   ├── cache/         # 缓存管理
│   │   │   └── cover_cache.go
│   │   ├── middleware/    # 中间件
│   │   │   └── middleware.go
│   │   └── model/         # 数据模型
│   │       ├── comic.go
│   │       └── dto.go
│   ├── pkg/         # 可复用的公共包
│   │   ├── logger/        # 日志
│   │   ├── config/        # 配置
│   │   └── utils/         # 工具函数
│   ├── config.yaml        # 配置文件
│   └── go.mod
├── frontend/              # React 前端
│   ├── src/
│   │   ├── components/   # UI组件
│   │   │   └── Sidebar.jsx
│   │   ├── pages/        # 页面
│   │   │   ├── Dashboard.jsx
│   │   │   ├── ComicList.jsx
│   │   │   ├── ComicDetail.jsx
│   │   │   ├── Reader.jsx
│   │   │   ├── Search.jsx
│   │   │   ├── Favorites.jsx
│   │   │   ├── History.jsx
│   │   │   └── Settings.jsx
│   │   ├── store/        # 状态管理
│   │   │   └── comicStore.js
│   │   ├── api/          # API调用
│   │   │   └── index.js
│   │   ├── App.jsx
│   │   └── main.jsx
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
└── README.md
```

## API 接口

### 漫画相关

- `GET /api/comics` - 获取漫画列表
  - 参数: `page` (页码), `use_cache` (是否使用缓存)

- `GET /api/comics/:id` - 获取漫画详情

- `GET /api/comics/:id/images` - 获取阅读器图片列表

- `GET /api/comics/search` - 搜索漫画
  - 参数: `keyword` (关键词), `page` (页码), `mode` (搜索模式)

- `GET /api/comics/filter` - 筛选漫画
  - 参数: `tag_id`, `category_id`, `author_id`, `page`

- `POST /api/comics/:id/history` - 添加阅读历史

- `POST /api/comics/:id/favorite` - 切换收藏状态

### 仪表盘

- `GET /api/dashboard/stats` - 获取统计数据

### 健康检查

- `GET /api/health` - 健康检查

## 配置说明

配置文件位于 `backend/config.yaml`：

```yaml
server:
  port: 8080              # 服务器端口
  mode: debug          # 运行模式 (debug/release)
  cors_origins:           # CORS 允许的源
    - "http://localhost:5173"

database:
  path: ./data/comics.db  # 数据库路径

crawler:
  base_urls:              # 目标站点URL列表
    - "https://example.com"
  request_delay_min: 2.0  # 请求最小延迟(秒)
  request_delay_max: 5.0  # 请求最大延迟(秒)
  anti_ban_mode: safe     # 防封禁模式 (standard/safe/extreme)

cache:
  cover_max: 2000         # 最大封面缓存数
  cover_target_size_kb: 50 # 封面目标大小(KB)
  cover_strategy: passive  # 封面缓存策略 (passive/auto)

log:
  level: info             # 日志级别
  file: ./logs/app.log    # 日志文件路径
```

## 开发注意事项

### 后端

1. **错误处理**: Go 的显式错误处理，每个错误都要检查
2. **并发控制**: 使用 goroutine 和 channel，注意避免泄漏
3. **数据库事务**: GORM 的事务处理
4. **日志记录**: 使用 zap 记录结构化日志

### 前端

1. **状态管理**: 使用 Zustand 管理全局状态
2. **路由**: React Router v6 的新 API
3. **API 调用**: 统一使用 axios 实例
4. **组件优化**: 使用 React.memo 避免不必要的重渲染

## 待实现功能

- [ ] 在线搜索
- [ ] 标签筛选
- [ ] 分类筛选
- [ ] 作者作品浏览
- [ ] 下载管理
- [ ] 收藏列表
- [ ] 历史记录列表
- [ ] 设置页面
- [ ] 虚拟滚动优化
- [ ] 图片懒加载
- [ ] 主题切换
- [ ] 快捷键支持

## 调试技巧

### 后端调试

```bash
# 查看日志
tail -f logs/app.log

# 使用 delve 调试
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug cmd/server/main.go
```

### 前端调试

- 使用 Chrome DevTools
- React Developer Tools 扩展
- 查看 Network 面板的 API 请求

## 性能优化建议

1. **后端**:
   - 使用连接池
   - 添加缓存层 (Redis)
   - 数据库索引优化
   - 并发请求控制

2. **前端**:
   - 虚拟滚动 (react-window)
   - 图片懒加载
   - 代码分割
   - CDN 加速

## 部署

### 后端部署

```bash
# 编译
cd backend
go build -o comic-viewer cmd/server/main.go

# 运行
./comic-viewer
```

### 前端部署

```bash
# 构建
cd frontend
npm run build

# 部署 dist 目录到静态服务器
```

## 常见问题

### 1. 跨域问题
确保后端配置了正确的 CORS 源

### 2. 数据库锁定
SQLite 并发写入限制，考虑使用 PostgreSQL

### 3. 图片加载慢
使用封面缓存和 CDN

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## License

MIT License
