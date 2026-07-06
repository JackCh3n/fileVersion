@echo off
setlocal
REM 编译 FileVersion 小工具（windowsgui 子系统，避免右键“发送到”时弹出黑窗口）
set CGO_ENABLED=0
set GOOS=windows

echo [1/2] Building windows/amd64 ...
go build -ldflags="-H=windowsgui" -o fileversion.exe
if errorlevel 1 goto :fail

echo [2/2] Building windows/386 (兼容旧版 32 位 Win7) ...
set GOARCH=386
go build -ldflags="-H=windowsgui" -o fileversion-386.exe
set GOARCH=amd64

echo.
echo 构建完成：
echo   fileversion.exe      (64 位，Win7/10/11 主流)
echo   fileversion-386.exe  (32 位，旧版 Win7)
echo.
echo 首次使用请以当前用户运行一次：
echo   fileversion.exe install
goto :eof

:fail
echo 构建失败，请检查 Go 环境（go version）。
exit /b 1
