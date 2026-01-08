# 启动脚本使用说明

## 📋 脚本说明

项目提供了多个启动脚本，方便不同场景使用：

### 1. `start-dev.ps1` - 一键启动（推荐）

**功能**: 自动检查环境并启动完整的开发服务器

**使用方法**:
```powershell
.\start-dev.ps1
```

**特点**:
- ✅ 自动检查Go、Node.js、Wails环境
- ✅ 自动检查并安装前端依赖
- ✅ 自动安装Wails CLI（如果缺失）
- ✅ 使用 `wails dev` 同时启动前后端
- ✅ 如果失败，自动尝试分别启动

**适用场景**: 
- 首次运行项目
- 日常开发
- 快速启动

---

### 2. `start-dev.bat` - Windows批处理版本

**功能**: 与PowerShell版本相同，但使用批处理格式

**使用方法**:
```cmd
start-dev.bat
```

**特点**:
- ✅ 兼容性更好（不需要PowerShell执行策略）
- ✅ 功能与PowerShell版本相同

**适用场景**:
- PowerShell执行策略受限时
- 喜欢使用CMD的用户

---

### 3. `start-separate.ps1` - 分别启动前后端

**功能**: 在前端和后端分别打开两个终端窗口

**使用方法**:
```powershell
.\start-separate.ps1
```

**特点**:
- ✅ 前端在独立窗口运行（便于查看日志）
- ✅ 后端在当前窗口运行
- ✅ 可以分别查看前后端日志
- ✅ 可以分别停止前后端

**适用场景**:
- 需要分别查看前后端日志
- 调试前端或后端问题
- 需要分别控制前后端

---

## 🚀 快速开始

### 方法1：使用一键启动脚本（最简单）

```powershell
# PowerShell
.\start-dev.ps1

# 或使用批处理
start-dev.bat
```

### 方法2：手动启动

```powershell
# 1. 设置环境变量
$env:PATH = "$env:PATH;C:\Program Files\Go\bin;$env:USERPROFILE\go\bin"

# 2. 启动开发服务器
wails dev
```

### 方法3：分别启动（调试时推荐）

```powershell
# 终端1：启动前端
cd frontend
npm run dev

# 终端2：启动后端
wails dev
```

---

## ⚙️ 环境要求

脚本会自动检查以下环境：

- ✅ **Go 1.25.0+** - 安装在 `C:\Program Files\Go\bin\`
- ✅ **Node.js 20.19.0+ 或 22.12.0+** - 需要全局安装
- ✅ **Wails CLI v2** - 会自动安装到 `%USERPROFILE%\go\bin\`
- ✅ **前端依赖** - 会自动检查并安装

---

## 🔧 自定义配置

### 修改Go路径

如果Go安装在其他位置，修改脚本中的路径：

**PowerShell脚本** (`start-dev.ps1`):
```powershell
$goPath = "你的Go路径\go.exe"
```

**批处理脚本** (`start-dev.bat`):
```batch
set "PATH=%PATH%;你的Go路径"
```

### 修改Wails路径

如果Wails安装在自定义位置：

```powershell
$wailsPath = "你的Wails路径\wails.exe"
```

---

## 📝 常见问题

### 1. PowerShell执行策略错误

**错误**: `无法加载文件，因为在此系统上禁止运行脚本`

**解决方法**:
```powershell
# 以管理员身份运行PowerShell，执行：
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

或者直接使用 `start-dev.bat` 批处理文件。

### 2. 找不到Wails命令

**解决方法**: 脚本会自动尝试安装，如果失败，手动安装：

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

然后确保 `%USERPROFILE%\go\bin` 在PATH中。

### 3. 前端依赖安装失败

**解决方法**:
```powershell
cd frontend
npm install
```

### 4. 端口被占用

**错误**: `Port 5173 is already in use`

**解决方法**:
- 关闭占用端口的程序
- 或修改 `frontend/vite.config.js` 中的端口配置

---

## 🎯 推荐工作流程

### 日常开发

```powershell
# 每天第一次启动
.\start-dev.ps1

# 修改代码后，保存即可自动刷新
```

### 调试问题

```powershell
# 需要查看详细日志时
.\start-separate.ps1

# 可以分别查看前后端日志
```

### 首次运行

```powershell
# 1. 确保环境已安装
# 2. 运行启动脚本
.\start-dev.ps1

# 脚本会自动：
# - 检查环境
# - 安装依赖
# - 启动服务器
```

---

## 📊 脚本执行流程

```
启动脚本
    ↓
检查Go环境
    ↓
检查Node.js环境
    ↓
检查Wails CLI
    ↓
检查前端依赖
    ↓
启动开发服务器
    ├─ 成功 → 运行中
    └─ 失败 → 尝试分别启动
```

---

## 💡 提示

1. **首次运行**: 建议使用 `start-dev.ps1`，会自动处理所有依赖
2. **日常使用**: 直接运行 `wails dev` 即可
3. **调试时**: 使用 `start-separate.ps1` 分别查看日志
4. **遇到问题**: 查看脚本输出的错误信息，按提示解决

---

## 🔗 相关文件

- `wails.json` - Wails配置文件
- `frontend/package.json` - 前端依赖配置
- `go.mod` - Go依赖配置

---

**提示**: 如果遇到任何问题，请查看脚本输出的错误信息，或查看项目日志文件 `logs/info.log` 和 `logs/error.log`。

