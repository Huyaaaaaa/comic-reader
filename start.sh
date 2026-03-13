#!/bin/bash

# Comic Viewer 启动脚本

set -e

echo "========="
echo "Comic Viewer 启动脚本"
echo "==================="

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go，请先安装 Go 1.21+"
    exit 1
fi

# 检查 Node.js 是否安装
if ! command -v node &> /dev/null; then
    echo "错误: 未找到 Node.js，请先安装 Node.js"
    exit 1
fi

# 创建必要的目录
mkdir -p data logs downloads

# 启动后端
echo ""
echo "启动后端服务..."
cd backend

# 安装 Go 依赖
if [ ! -f "go.sum" ]; then
    echo "安装 Go 依赖..."
    go mod tidy
fi

# 后台运行后端
nohup go run cmd/server/main.go > ../logs/backend.log 2>&1 &
BACKEND_PID=$!
echo "后端服务已启动 (PID: $BACKEND_PID)"
echo $BACKEND_PID > ../backend.pid

cd ..

# 等待后端启动
echo "等待后端服务启动..."
sleep 3

# 检查后端是否启动成功
if curl -s http://localhost:8080/api/health > /dev/null; then
    echo "✓ 后端服务启动成功"
else
    echo "✗ 后端服务启动失败，请查看日志: logs/backend.log"
    exit 1
fi

# 启动前端
echo ""
echo "启动前端服务..."
cd frontend

# 安装 npm 依赖
if [ ! -d "node_modules" ]; then
    echo "安装 npm 依赖..."
    npm install
fi

# 后台运行前端
nohup npm run dev > ../logs/frontend.log 2>&1 &
FRONTEND_PID=$!
echo "前端服务已启动 (PID: $FRONTEND_PID)"
echo $FRONTEND_PID > ../frontend.pid

cd ..

echo ""
echo "============"
echo "✓ 所有服务已启动"
echo "=============="
echo "后端地址: http://localhost:8080"
echo "前端地址: http://localhost:5173"
echo ""
echo "查看日志:"
echo "  后端: tail -f logs/backend.log"
echo "  前端: tail -f logs/frontend.log"
echo ""
echo "停止服务: ./stop.sh"
echo "===================="
