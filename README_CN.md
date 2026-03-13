# 漫画浏览器 - Comic Viewer

> 基于 Go + React 的现代化漫画浏览器应用

[English](README.md) | 简体中文

## 项目简介

这是一个完全重构的漫画浏览器应用，从 Python Flet 迁移到 Go + React 架构，提供更好的性能和用户体验。

## 主要特性

- 🚀 **高性能**: Go 后端 + React 前端
- 📱 **响应式设计**: 适配各种屏幕尺寸
- 🎨 **现代化 UI**: 基于 Ant Design
- 💾 **智能缓存**: 封面缓存、列表缓存
- 🔒 **安全**: AES-CBC 解密支持
- 🛡️ **防封禁**: 多档位防封禁机制
- 📊 **数据统计**: 完整的仪表盘

## 技术栈

### 后端
- Go 1.21+
- Gin (Web 框架)
- GORM (ORM)
- SQLite (数据库)
- Zap (日志)
- Viper (配置)

### 前端
- React 18
- Vite
- Ant Design
- Zustand (状态管理)
- React Router v6
- Axios

## 快速开始

### 前置要求

- Go 1.21+
- Node.js 18+
- npm 或 yarn

### 安装

```bash
# 克隆项目
cd ~/project/comic-viewer-claude

# 安装依赖
make install
```

### 运行

```bash
# 启动开发服务器
./start.sh

# 或使用 make
make dev
```

### 访问

- 前端: http://localhost:5173
- 后端: http://localhost:8080

## 项目结构

```
comic-viewer-claude/
├── backend/          # Go 后端
├── frontend/         # React 前端
├── docs/            # 文档
├── start.sh         # 启动脚本
├── stop.sh          # 停止脚本
└── Makefile         # Make 命令
```

## 功能列表
- [x] 漫画列表浏览
- [x] 漫画详情查看
- [x] 在线阅读器
- [x] 本地搜索
- [x] 收藏功能
- [x] 阅读历史
- [x] 仪表盘统计
- [ ] 在线搜索
- [ ] 标签筛选
- [ ] 下载管理

## 文档

- [API 文档](docs/API.md)
- [数据库设计](docs/DATABASE.md)
- [开发指南](docs/DEVELOPMENT.md)
- [部署指南](docs/DEPLOYMENT.md)

## 开发

```bash
# 只启动后端
make backend

# 只启动前端
make frontend

# 运行测试
make test

# 构建生产版本
make build
```

## 部署

详见 [部署指南](docs/DEPLOYMENT.md)

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 致谢

感谢原 Python Flet 版本的开发经验，为本项目提供了功能参考。
