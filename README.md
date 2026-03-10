# 合欢阅读器

合欢阅读器是一个基于 Go + React 的本地优先漫画阅读器项目，目标是提供：
- 本地优先加载
- 离线可浏览已缓存内容
- 多源站切换与容错
- L1 / L2 / L3 分层缓存
- 后续可扩展为 Electron 一体化桌面应用

当前仓库已经不是纯脚手架，而是进入了“可运行、可同步真实源站”的阶段。

## 当前能力

### 已完成
- Go 后端基础服务
  - Gin API
  - SQLite 本地库
  - GORM 数据访问层
  - SSE 事件流
- React 前端基础页面
  - 总览页
  - 漫画列表页
  - 漫画详情页
  - 阅读页
  - 收藏页
  - 历史页
  - 源站管理页
  - 设置页
- 本地业务链路
  - 漫画列表读取
  - 漫画详情读取
  - 正文图片索引读取
  - 收藏
  - 阅读历史
  - 搜索历史
- 真实源站同步链路
  - 头部列表同步
  - 详情补全
  - 阅读页正文图片索引补全
  - 源站超时重试
  - 源站失败切换
  - 页面密文解密与解析

### 当前验证通过
- `go build ./...`
- `npm run build`
- 真实镜像站头部同步
- 单漫画详情补全
- 正文图片索引入库

## 当前架构

### 后端
- Go 1.21
- Gin
- SQLite
- GORM

### 前端
- React
- TypeScript
- Vite

### 部署准备
- Docker
- Nginx
- systemd

## 目录结构

```text
comic-go-codex/
├── backend/    # Go 后端
├── frontend/   # React 前端
├── deploy/     # Docker / Nginx / systemd
├── docs/       # 需求与技术文档
├── scripts/    # 开发辅助脚本
└── README.md
```

## 现有 API

### 基础
- `GET /api/health`
- `GET /api/events/stream`

### 设置
- `GET /api/settings`
- `PUT /api/settings/:key`

### 源站管理
- `GET /api/sources`
- `POST /api/sources`
- `PUT /api/sources/:id`
- `DELETE /api/sources/:id`
- `POST /api/sources/:id/check`

### 同步
- `POST /api/sync/head`
- `POST /api/sync/comics/:id/detail`

### 漫画数据
- `GET /api/comics`
- `GET /api/comics/:id`
- `GET /api/comics/:id/images`
- `GET /api/search`
- `GET /api/tags`
- `GET /api/categories`
- `GET /api/authors`

### 用户本地数据
- `POST /api/favorites`
- `DELETE /api/favorites/:comic_id`
- `GET /api/favorites`
- `POST /api/history`
- `GET /api/history`
- `POST /api/search/history`
- `GET /api/search/history`
- `DELETE /api/search/history`

## 启动方式

### 1. 启动后端

```bash
cd backend
go mod tidy
go run ./cmd/server
```

默认监听端口：`8080`

### 2. 启动前端

```bash
cd frontend
npm install
npm run dev
```

默认监听端口：`5173`

### 3. 一键开发

```bash
./scripts/dev.sh
```

## 环境变量

### 后端
- `HEHUAN_HTTP_ADDR` 默认 `:8080`
- `HEHUAN_DATA_DIR` 默认 `./data`
- `HEHUAN_DB_PATH` 默认 `<data>/hehuan.db`
- `HEHUAN_APP_NAME` 默认 `合欢阅读器`
- `HEHUAN_APP_VERSION` 默认 `0.1.0`
- `HEHUAN_ALLOWED_ORIGINS` 默认空，由部署者自行填写允许跨域来源

### 前端
- `VITE_API_BASE_URL` 默认空，留空时前端使用同源 `/api`

## 当前实现边界

当前这版已经能做真实同步，但还没完成以下核心能力：
- L3 正文图片被动缓存策略细化
- 缓存上限、淘汰策略与空间不足处理
- 下载队列接管批量图片落盘
- 下载队列与空间预检查
- 导入导出
- 应用更新中心
- 导航页自动导入源站
- 多源解析器抽象层

## 路线图

### Phase A
- 稳定真实源站同步链路
- 补齐本地优先详情与阅读体验
- 做好 L1 / L2 数据完整性

### Phase B
- 接入 L3 图片代理与缓存复用
- 下载管理
- 磁盘空间预检查

### Phase C
- 导入导出
- 更新中心
- 多源导航页导入

## 文档

仓库内已包含设计文档：
- `docs/合欢阅读器需求说明书.md`
- `docs/合欢阅读器技术执行文档.md`

## 说明

- 当前数据库首次启动会写入少量示例数据，便于联调。
- 当前源站实现基于现有镜像站结构，后续会继续抽象成可插拔解析器。
- 项目最终形态目标是桌面端本地优先阅读器，而不是单纯网页采集器。
