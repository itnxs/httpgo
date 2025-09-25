@echo off
REM HttpGo Windows 构建脚本
REM 支持构建 Windows 平台的多个架构版本

setlocal enabledelayedexpansion

REM 设置颜色代码
set "RED=[91m"
set "GREEN=[92m"
set "YELLOW=[93m"
set "BLUE=[94m"
set "NC=[0m"

REM 打印带颜色的消息
:print_info
echo %BLUE%[INFO]%NC% %~1
goto :eof

:print_success
echo %GREEN%[SUCCESS]%NC% %~1
goto :eof

:print_warning
echo %YELLOW%[WARNING]%NC% %~1
goto :eof

:print_error
echo %RED%[ERROR]%NC% %~1
goto :eof

REM 检查 Go 环境
:check_go
where go >nul 2>nul
if errorlevel 1 (
    call :print_error "Go 未安装，请先安装 Go 语言环境"
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
set GO_VERSION=!GO_VERSION:go=!
call :print_info "检测到 Go 版本: !GO_VERSION!"
goto :eof

REM 获取版本号
:get_version
if not "%~1"=="" (
    set VERSION=%~1
) else (
    REM 尝试从 git tag 获取版本号
    git describe --tags --exact-match HEAD >nul 2>nul
    if errorlevel 1 (
        set VERSION=v1.0.0-dev
        call :print_warning "未找到 git tag，使用默认版本: !VERSION!"
    ) else (
        for /f %%i in ('git describe --tags --exact-match HEAD 2^>nul') do set VERSION=%%i
    )
)
call :print_info "构建版本: !VERSION!"
goto :eof

REM 清理构建目录
:clean_build
if exist build (
    call :print_info "清理构建目录..."
    rmdir /s /q build
)
mkdir build
goto :eof

REM 构建单个平台
:build_platform
set GOOS=%~1
set GOARCH=%~2
set SUFFIX=%~3
set NAME=httpgo-!GOOS!-!GOARCH!

call :print_info "构建 !GOOS!/!GOARCH!..."

REM 设置环境变量
set CGO_ENABLED=0

REM 构建
go build -ldflags="-s -w -X 'github.com/itnxs/httpgo/internal/pkg.Version=!VERSION!'" -o "build\!NAME!!SUFFIX!" .
if errorlevel 1 (
    call :print_error "构建失败: !GOOS!/!GOARCH!"
    goto :eof
)

call :print_success "构建完成: !NAME!!SUFFIX!"

REM 创建压缩包
cd build
if "!GOOS!"=="windows" (
    powershell -command "Compress-Archive -Path '!NAME!!SUFFIX!' -DestinationPath '!NAME!.zip' -Force"
    call :print_success "创建压缩包: !NAME!.zip"
) else (
    tar -czf "!NAME!.tar.gz" "!NAME!!SUFFIX!"
    call :print_success "创建压缩包: !NAME!.tar.gz"
)
cd ..
goto :eof

REM 主构建函数
:build_all
call :print_info "开始构建 Windows 平台版本..."

set SUCCESS_COUNT=0
set TOTAL_COUNT=0

REM Windows 构建目标
set TARGETS[0]=windows:amd64:.exe
set TARGETS[1]=windows:386:.exe
set TARGETS[2]=windows:arm64:.exe

for /L %%i in (0,1,2) do (
    set /a TOTAL_COUNT+=1
    for /f "tokens=1,2,3 delims=:" %%a in ("!TARGETS[%%i]!") do (
        call :build_platform "%%a" "%%b" "%%c"
        if not errorlevel 1 set /a SUCCESS_COUNT+=1
    )
)

call :print_info "构建完成: !SUCCESS_COUNT!/!TOTAL_COUNT! 成功"
goto :eof

REM 生成校验和
:generate_checksums
call :print_info "生成校验和文件..."
cd build

REM 生成 SHA256 校验和
where certutil >nul 2>nul
if not errorlevel 1 (
    for %%f in (*.zip *.exe) do (
        certutil -hashfile "%%f" SHA256 | find /v "CertUtil" | find /v "SHA256" >> SHA256SUMS.tmp
        echo %%f >> SHA256SUMS.tmp
    )
    if exist SHA256SUMS.tmp (
        move SHA256SUMS.tmp SHA256SUMS
        call :print_success "校验和文件已生成: SHA256SUMS"
    )
) else (
    call :print_warning "未找到 certutil 命令，跳过校验和生成"
)

cd ..
goto :eof

REM 显示构建结果
:show_results
call :print_info "构建结果:"
echo.
if exist build (
    dir build
    echo.
    
    REM 计算文件数量
    for /f %%i in ('dir /b build ^| find /c /v ""') do set FILE_COUNT=%%i
    call :print_info "构建文件总数: !FILE_COUNT!"
) else (
    call :print_error "构建目录不存在"
)
goto :eof

REM 使用说明
:show_usage
echo HttpGo Windows 构建脚本
echo.
echo 用法:
echo     %~nx0 [版本号]
echo.
echo 参数:
echo     版本号          指定构建版本号 (例如: v1.0.0)
echo                    如果不指定，将尝试从 git tag 获取
echo.
echo 示例:
echo     %~nx0                      # 使用默认版本构建
echo     %~nx0 v1.2.0              # 构建指定版本
echo.
echo 支持的 Windows 平台:
echo     - Windows amd64 (64位)
echo     - Windows 386 (32位)  
echo     - Windows arm64 (ARM64)
echo.
goto :eof

REM 主函数
:main
if "%~1"=="/?" goto show_usage
if "%~1"=="-h" goto show_usage
if "%~1"=="--help" goto show_usage

call :print_info "HttpGo Windows 构建脚本"
echo ================================

REM 检查环境
call :check_go
if errorlevel 1 exit /b 1

REM 获取版本号
call :get_version "%~1"

REM 清理构建目录
call :clean_build

REM 执行构建
call :build_all

REM 生成校验和
call :generate_checksums

REM 显示结果
call :show_results

call :print_success "Windows 平台构建完成！"
echo.
call :print_info "构建文件位于 build\ 目录中"

goto :eof

REM 运行主函数
call :main %*