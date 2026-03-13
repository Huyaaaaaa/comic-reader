# Comic Viewer - Go + React 版本

基于 Go 后端 + React 前端的漫画浏览器应用，完全重构自 Python Flet 版本。

## 技术栈

### 后端
- **语言**: Go 1.21+
- **Web框架**: Gin
- **数据库**: SQLite (GORM)
- **HTTP客户端**: go-resty
- **HTML解析**: goquery
- **图片处理**: imaging
- **日志**: zap
- **配置**: viper

### 前端
- **框架**: React 18
- **构建工具**: Vite
- **状态管理**: Zustand
- **路由**: React Router v6
- **UI组件**: Ant Design / Tailwind CSS
- **HTTP客户端**: axios
- **虚拟滚动**: react-window

## 项目结构

```
comic-viewer-claude/
├── backend/          # Go 后端
│   ├── cmd/
│   │   └── server/         # 主程序入口
│   ├── internal/           # 内部包
│   │   ├── api/           # HTTP handlers
│   │   ├── service/       # 业务逻辑层
│   │   ├── repository/    # 数据访问层
│   │   ├── crawler/       # 爬虫客户端
│   │   ├── parser/        # HTML解析
│   │   ├── crypto/        # 加密解密
│   │   ├── cache/         # 缓存管理
│   │   ├── downloader/    # 下载管理
│   │   ├── middleware/    # 中间件
│   │   └── model/         # 数据模型
│   ├── pkg/           # 可复用的公共包
│   │   ├── logger/        # 日志
│   │   ├── config/        # 配置
│   │   └── utils/         # 工具函数
│   ├── migrations/      # 数据库迁移
│   └── config.yaml        # 配置文件
├── frontend/              # React 前端
│   ├── src/
│   │   ├── components/   # UI组件
│   │   ├── pages/        # 页面
│   │   ├── hooks/        # 自定义hooks
│   │   ├── store/        # 状态管理
│   │   ├── api/          # API调用
│   │   ├── utils/        # 工具函数
│   │   └── assets/       # 静态资源
│   └── public/           # 公共资源
└── docs/              # 文档

```

## 核心功能

### 已实现
- [ ] 漫画列表浏览（分页）
- [ ] 漫画详情查看
- [ ] 在线阅读器
- [ ] 标签筛选
- [ ] 分类筛选
- [ ] 搜索功能（在线/本地）
- [ ] 阅读历史
- [ ] 收藏管理
- [ ] 下载管理
- [ ] 封面缓存
- [ ] 列表缓存
- [ ] 防封禁机制

### 性能优化
- [ ] 虚拟滚动列表
- [ ] 图片懒加载
- [ ] 封面压缩（50KB）
- [ ] LRU缓存策略
- [ ] 并发控制
- [ ] 请求限流

## 快速开始

### 后端
```bash
cd backend
go mod init comic-viewer-claude
go mod tidy
go run cmd/server/main.go
```

### 前端

```bash
cd frontend
npm install
npm run dev
```

## 开发计划

### Phase 1: 核心功能 (Week 1-2)
- [x] 项目结构搭建
- [ ] 数据库设计和迁移
- [ ] 爬虫客户端 + 解析器
- [ ] 基础 API (列表、详情、阅读)
- [ ] React 基础页面

### Phase 2: 缓存和优化 (Week 3)
- [ ] 封面缓存系统
- [ ] 列表缓存
- [ ] 虚拟滚动
- [ ] 图片懒加载

### Phase 3: 高级功能 (Week 4)
- [ ] 下载管理
- [ ] 搜索功能
- [ ] 收藏和历史
- [ ] 防封禁优化

### Phase 4: 用户体验 (Week 5)
- [ ] 设置页面
- [ ] 主题切换
- [ ] 快捷键支持
- [ ] 离线模式

## 配置说明

配置文件位于 `backend/config.yaml`：

```yaml
server:
  port: 8080
  mode: debug

database:
  path: ./data/comics.db

crawler:
  base_url: https://example.com
  request_delay_min: 2
  request_delay_max: 5
  anti_ban_mode: safe

cache:
  cover_max: 2000
  cover_target_size_kb: 50
```

## License

MIT License
