@echo off
chcp 65001 >nul
title go-stock 后端开发服务器

echo ========================================
echo   go-stock 后端开发服务器
echo ========================================
echo.

REM 设置Go环境变量
set "PATH=%PATH%;C:\Program Files\Go\bin"
set "PATH=%PATH%;%USERPROFILE%\go\bin"

REM 检查Go
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到Go，请确保Go已安装
    pause
    exit /b 1
)

REM 检查Wails
where wails >nul 2>&1
if %errorlevel% neq 0 (
    echo 警告: 未找到Wails CLI，尝试安装...
    call go install github.com/wailsapp/wails/v2/cmd/wails@latest
    if %errorlevel% neq 0 (
        echo 错误: Wails CLI安装失败
        pause
        exit /b 1
    )
    echo Wails CLI安装成功
    echo.
)

echo 提示: 关闭此窗口将停止后端服务器
echo.

REM 启动后端开发服务器
wails dev

pause

