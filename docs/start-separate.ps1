# go-stock 分别启动前后端脚本
# 功能：分别在两个终端窗口启动前端和后端

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   go-stock 分别启动前后端" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 设置环境变量
$env:PATH = "$env:PATH;C:\Program Files\Go\bin;$env:USERPROFILE\go\bin"

# 检查环境
Write-Host "检查环境..." -ForegroundColor Yellow
$goOk = $false
$nodeOk = $false
$wailsOk = $false

try {
    $goVersion = & go version 2>$null
    if ($LASTEXITCODE -eq 0) {
        $goOk = $true
        Write-Host "✓ Go: $goVersion" -ForegroundColor Green
    }
} catch {
    Write-Host "✗ Go未安装" -ForegroundColor Red
}

try {
    $nodeVersion = & node --version 2>$null
    if ($LASTEXITCODE -eq 0) {
        $nodeOk = $true
        Write-Host "✓ Node.js: $nodeVersion" -ForegroundColor Green
    }
} catch {
    Write-Host "✗ Node.js未安装" -ForegroundColor Red
}

try {
    $wailsVersion = & wails version 2>$null
    if ($LASTEXITCODE -eq 0) {
        $wailsOk = $true
        Write-Host "✓ Wails: $wailsVersion" -ForegroundColor Green
    }
} catch {
    Write-Host "✗ Wails未安装" -ForegroundColor Red
}

if (-not $goOk -or -not $nodeOk -or -not $wailsOk) {
    Write-Host ""
    Write-Host "环境检查失败，请先安装缺失的工具" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   启动服务..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 获取当前目录
$projectRoot = $PWD.Path

# 启动前端开发服务器（新窗口）
Write-Host "启动前端开发服务器（新窗口）..." -ForegroundColor Green
Start-Process powershell -ArgumentList @(
    "-NoExit",
    "-Command",
    "cd '$projectRoot\frontend'; Write-Host '前端开发服务器' -ForegroundColor Cyan; Write-Host 'URL: http://localhost:5173' -ForegroundColor Yellow; Write-Host ''; npm run dev"
) -WindowStyle Normal

# 等待前端启动
Write-Host "等待前端服务器启动..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# 启动后端（当前窗口）
Write-Host "启动后端服务器（当前窗口）..." -ForegroundColor Green
Write-Host ""
Write-Host "提示:" -ForegroundColor Yellow
Write-Host "  - 前端: http://localhost:5173 (另一个窗口)" -ForegroundColor Gray
Write-Host "  - 后端: 当前窗口" -ForegroundColor Gray
Write-Host "  - 按 Ctrl+C 停止后端服务器" -ForegroundColor Gray
Write-Host ""

# 启动Wails（连接到前端服务器）
wails dev

Write-Host ""
Write-Host "后端服务器已停止" -ForegroundColor Yellow
Write-Host "提示: 前端服务器仍在运行，如需停止请关闭前端窗口" -ForegroundColor Yellow

