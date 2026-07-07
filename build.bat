@echo off
setlocal
REM FileVersion GUI 构建脚本
REM 首次使用需先：go install github.com/wailsapp/wails/v2/cmd/wails@latest
REM 并确保已安装：WebKit2 (Win10/11 自带)、UPX(可选)
set CGO_ENABLED=1

if "%1"=="dev" (
  echo [dev] 启动 Wails 开发模式（热重载）...
  wails dev
  goto :eof
)

echo [1/2] Building windows/amd64 (GUI) ...
wails build -platform windows/amd64
if errorlevel 1 goto :fail

echo [2/2] Building windows/386 (兼容旧版 32 位 Win7) ...
wails build -platform windows/386 -o fileversion-386.exe
if errorlevel 1 goto :fail

echo.
echo 构建完成：fileversion.exe (64 位) / fileversion-386.exe (32 位)
echo 首次使用请以当前用户运行一次：fileversion.exe install
goto :eof

:fail
echo 构建失败，请确认已安装 Wails (go install github.com/wailsapp/wails/v2/cmd/wails@latest) 与 Go 1.22+
exit /b 1
