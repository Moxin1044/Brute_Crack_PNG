@echo off
setlocal enabledelayedexpansion

rem 清理旧构建
if exist build (
    rmdir /s /q build
    echo 已清理旧构建目录
)

rem 创建输出目录
mkdir build

rem 架构列表
set arch_list=386 amd64 arm64

for %%a in (%arch_list%) do (
    echo 正在编译 %%a 架构...
    set GOARCH=%%a
    set GOOS=windows

    mkdir build\%%a
    go build -o build\%%a\Brute_Crack_PNG_Windows_%%a.exe .\cli\brute.go
@REM     go build -tags main -o Brute_Crack_PNG.exe
@REM     fyne package -os windows -icon icon.png

    if errorlevel 1 (
        echo 编译 %%a 架构失败!
        pause
        exit /b 1
    )
)

echo 全部架构编译完成
pause