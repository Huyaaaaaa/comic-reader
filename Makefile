# Makefile for Comic Viewer

.PHONY: help install dev build clean test

# 默认目标
help:
	@echo "Comic Viewer - Makefile"
	@echo ""
	@echo "可用命令:"
	@echo "  make install    - 安装所有依赖"
	@echo "  make dev        - 启动开发服务器"
	@echo "  make build      - 构建生产版本"
	@echo "  make clean      - 清理构建文件"
	@echo "  make test       - 运行测试"
	@echo "  make backend    - 只启动后端"
	@echo "  make frontend   - 只启动前端"

# 安装依赖
install:
	@echo "安装后端依赖..."
	cd backend && go mod tidy
	@echo "安装前端依赖..."
	cd frontend && npm install
	@echo "✓ 依赖安装完成"

# 启动开发服务器
dev:
	@echo "启动开发服务器..."
	./start.sh

# 只启动后端
backend:
	@echo "启动后端服务..."
	cd backend && go run cmd/server/main.go

# 只启动前端
frontend:
	@echo "启动前端服务..."
	cd frontend && npm run dev

# 构建生产版本
build:
	@echo "构建后端..."
	cd backend && go build -o ../build/comic-viewer cmd/server/main.go
	@echo "构建前端..."
	cd frontend && npm run build
	@echo "✓ 构建完成"
	@echo "  后端: build/comic-viewer"
	@echo "  前端: frontend/dist"

# 清理
clean:
	@echo "清理构建文件..."
	rm -rf build
	rm -rf frontend/dist
	rm -rf backend/tmp
	rm -f backend.pid frontend.pid
	@echo "✓ 清理完成"

# 运行测试
test:
	@echo "运行后端测试..."
	cd backend && go test ./...
	@echo "✓ 测试完成"

# 停止服务
stop:
	@echo "停止服务..."
	./stop.sh
