@echo off
chcp 65001 >nul
title go-stock 前端开发服务器

echo ========================================
echo   go-stock 前端开发服务器
echo ========================================
echo.

REM 检查Node.js
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到Node.js，请先安装Node.js
    pause
    exit /b 1
)

REM 检查前端依赖
if not exist "node_modules" (
    echo 前端依赖未安装，正在安装...
    call npm install
    if %errorlevel% neq 0 (
        echo 错误: 前端依赖安装失败
        pause
        exit /b 1
    )
    echo 前端依赖安装完成
    echo.
)

echo URL: http://localhost:5173
echo.
echo 提示: 关闭此窗口将停止前端服务器
echo.

REM 启动前端开发服务器
npm run dev

pause

