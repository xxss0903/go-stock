# go-stock 开发环境启动脚本
# 功能：同时启动前端和后端开发服务器

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   go-stock 开发环境启动脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 检查Go环境
Write-Host "[1/4] 检查Go环境..." -ForegroundColor Yellow
$goPath = "C:\Program Files\Go\bin\go.exe"
if (-not (Test-Path $goPath)) {
    Write-Host "错误: 未找到Go，请确保Go已安装在 C:\Program Files\Go\bin\" -ForegroundColor Red
    Write-Host "或者修改脚本中的Go路径" -ForegroundColor Red
    exit 1
}
$env:PATH = "$env:PATH;C:\Program Files\Go\bin"

# 检查Wails CLI
Write-Host "[2/4] 检查Wails CLI..." -ForegroundColor Yellow
$wailsPath = "$env:USERPROFILE\go\bin\wails.exe"
if (-not (Test-Path $wailsPath)) {
    Write-Host "警告: 未找到Wails CLI，尝试安装..." -ForegroundColor Yellow
    $env:PATH = "$env:PATH;$env:USERPROFILE\go\bin"
    & go install github.com/wailsapp/wails/v2/cmd/wails@latest
    if ($LASTEXITCODE -ne 0) {
        Write-Host "错误: Wails CLI安装失败" -ForegroundColor Red
        exit 1
    }
    Write-Host "Wails CLI安装成功" -ForegroundColor Green
} else {
    $env:PATH = "$env:PATH;$env:USERPROFILE\go\bin"
    Write-Host "Wails CLI已找到" -ForegroundColor Green
}

# 检查Node.js环境
Write-Host "[3/4] 检查Node.js环境..." -ForegroundColor Yellow
$nodeVersion = node --version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "错误: 未找到Node.js，请先安装Node.js" -ForegroundColor Red
    exit 1
}
Write-Host "Node.js版本: $nodeVersion" -ForegroundColor Green

# 检查前端依赖
Write-Host "[4/4] 检查前端依赖..." -ForegroundColor Yellow
if (-not (Test-Path "frontend\node_modules")) {
    Write-Host "前端依赖未安装，正在安装..." -ForegroundColor Yellow
    Set-Location frontend
    npm install
    if ($LASTEXITCODE -ne 0) {
        Write-Host "错误: 前端依赖安装失败" -ForegroundColor Red
        exit 1
    }
    Set-Location ..
    Write-Host "前端依赖安装完成" -ForegroundColor Green
} else {
    Write-Host "前端依赖已安装" -ForegroundColor Green
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   启动开发服务器..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "提示:" -ForegroundColor Yellow
Write-Host "  - 前端开发服务器: http://localhost:5173" -ForegroundColor Gray
Write-Host "  - 应用窗口将自动打开" -ForegroundColor Gray
Write-Host "  - 按 Ctrl+C 停止服务器" -ForegroundColor Gray
Write-Host ""

# 启动Wails开发服务器（会自动启动前后端）
Write-Host "正在启动Wails开发服务器..." -ForegroundColor Green
wails dev

# 如果wails dev失败，尝试分别启动
if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "Wails dev启动失败，尝试分别启动前后端..." -ForegroundColor Yellow
    Write-Host ""
    
    # 启动前端开发服务器（后台）
    Write-Host "启动前端开发服务器..." -ForegroundColor Green
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PWD\frontend'; npm run dev" -WindowStyle Normal
    
    # 等待前端服务器启动
    Start-Sleep -Seconds 5
    
    # 启动后端（使用wails dev，但只启动后端部分）
    Write-Host "启动后端服务器..." -ForegroundColor Green
    Write-Host "注意: 如果前端服务器已启动，Wails会连接到它" -ForegroundColor Yellow
    wails dev
}

Write-Host ""
Write-Host "开发服务器已停止" -ForegroundColor Yellow

