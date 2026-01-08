@echo off
chcp 65001 >nul
echo ========================================
echo    go-stock 开发环境启动脚本
echo ========================================
echo.

REM 设置Go环境变量
set "PATH=%PATH%;C:\Program Files\Go\bin"
set "PATH=%PATH%;%USERPROFILE%\go\bin"

REM 检查Go
echo [1/4] 检查Go环境...
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到Go，请确保Go已安装
    pause
    exit /b 1
)
for /f "tokens=*" %%i in ('go version') do set GO_VERSION=%%i
echo %GO_VERSION%
echo Go环境正常

REM 检查Node.js
echo [2/4] 检查Node.js环境...
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到Node.js，请先安装Node.js
    pause
    exit /b 1
)
for /f "tokens=*" %%i in ('node --version') do set NODE_VERSION=%%i
echo Node.js版本: %NODE_VERSION%
echo Node.js环境正常

REM 检查前端依赖
echo [3/4] 检查前端依赖...
if not exist "frontend\node_modules" (
    echo 前端依赖未安装，正在安装...
    cd frontend
    call npm install
    if %errorlevel% neq 0 (
        echo 错误: 前端依赖安装失败
        pause
        exit /b 1
    )
    cd ..
    echo 前端依赖安装完成
) else (
    echo 前端依赖已安装
)

REM 检查Wails
echo [4/4] 检查Wails CLI...
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
) else (
    for /f "tokens=*" %%i in ('wails version') do set WAILS_VERSION=%%i
    echo Wails版本: %WAILS_VERSION%
    echo Wails CLI已找到
)

echo.
echo ========================================
echo    启动开发服务器...
echo ========================================
echo.
echo 提示:
echo   - 前端开发服务器: http://localhost:5173
echo   - 应用窗口将自动打开
echo   - 关闭窗口即可停止对应服务
echo.

REM 获取当前脚本所在目录
set "PROJECT_ROOT=%~dp0"

REM 启动前端开发服务器（新窗口）
echo 启动前端开发服务器（新窗口）...
start "go-stock 前端开发服务器" cmd /k "cd /d %PROJECT_ROOT%frontend && echo ======================================== && echo   go-stock 前端开发服务器 && echo ======================================== && echo. && echo URL: http://localhost:5173 && echo. && echo 提示: 关闭此窗口将停止前端服务器 && echo. && npm run dev"

REM 等待前端服务器启动
echo 等待前端服务器启动（3秒）...
timeout /t 3 /nobreak >nul

REM 启动后端服务器（新窗口）
echo 启动后端服务器（新窗口）...
start "go-stock 后端开发服务器" cmd /k "cd /d %PROJECT_ROOT% && echo ======================================== && echo   go-stock 后端开发服务器 && echo ======================================== && echo. && echo 提示: 关闭此窗口将停止后端服务器 && echo. && wails dev"

echo.
echo ========================================
echo    启动完成！
echo ========================================
echo.
echo 已打开两个窗口:
echo   - 前端开发服务器窗口
echo   - 后端开发服务器窗口
echo.
echo 提示:
echo   - 关闭对应窗口即可停止对应服务
echo   - 前端: http://localhost:5173
echo   - 应用窗口将自动打开
echo.
echo 按任意键关闭此窗口（不会影响已启动的服务）...
pause >nul
