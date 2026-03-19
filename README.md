# Comic Reader

基于 Go 后端和 React 前端的本地漫画浏览器，围绕在线抓取、多源站管理、分层缓存、离线阅读和下载管理构建。

## 当前特性

- 多源站管理与可切换活动源站
- 直连 / 代理 / 回退代理访问策略
- L1 / L2 / L3 分层缓存与被动缓存
- 列表、标签搜索、详情页、阅读页、下载管理
- 阅读页渐进加载、临时图片复用、阅读优先级预取
- 本地数据库与缓存持久化

## 技术栈

- 后端：Go、Gin、GORM、SQLite
- 前端：React、TypeScript、Vite、Tailwind CSS、Recharts
- 抓取与解析：go-resty、goquery

## 项目结构

```text
comic/
├── backend/     # Go 后端
├── frontend/    # React 前端
├── start.sh     # 启动脚本
├── stop.sh      # 停止脚本
└── data/        # 本地数据库与缓存数据
```

## 本地开发

```bash
# 安装依赖
cd backend && go mod tidy
cd ../frontend && npm install

# 回到项目根目录
cd ..

# 启动前后端
./start.sh
```

也可以分别启动：

```bash
# 后端
cd backend
go run cmd/server/main.go

# 前端
cd frontend
npm run dev
```

## 说明

- 根目录只保留这个 `README.md` 作为 GitHub 项目说明。
- 其他设计稿、阶段性总结、测试记录和中文补充文档统一收纳在本地 `docs/` 目录。
- `docs/` 默认加入忽略，不作为主要版本产物提交。
