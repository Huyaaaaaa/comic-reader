#!/bin/bash

# Comic Viewer 停止脚本

set -e

echo "======================="
echo "Comic Viewer 停止脚本"
echo "========================"

# 停止后端
if [ -f "backend.pid" ]; then
    BACKEND_PID=$(cat backend.pid)
    if ps -p $BACKEND_PID > /dev/null 2>&1; then
        echo "停止后端服务 (PID: $BACKEND_PID)..."
        kill $BACKEND_PID
        rm backend.pid
        echo "✓ 后端服务已停止"
    else
        echo "后端服务未运行"
        rm backend.pid
    fi
else
    echo "未找到后端 PID 文件"
fi

# 停止前端
if [ -f "frontend.pid" ]; then
    FRONTEND_PID=$(cat frontend.pid)
    if ps -p $FRONTEND_PID > /dev/null 2>&1; then
      echo "停止前端服务 (PID: $FRONTEND_PID)..."
        kill $FRONTEND_PID
      rm frontend.pid
      echo "✓ 前端服务已停止"
    else
        echo "前端服务未运行"
        rm frontend.pid
    fi
else
    echo "未找到前端 PID 文件"
fi

echo ""
echo "==================="
echo "✓ 所有服务已停止"
echo "====================="
