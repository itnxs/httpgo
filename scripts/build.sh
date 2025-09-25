#!/bin/bash

# HttpGo 多平台构建脚本
# 支持构建 Windows、Linux、macOS、FreeBSD 版本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Go 环境
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go 未安装，请先安装 Go 语言环境"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "检测到 Go 版本: $GO_VERSION"
}

# 获取版本号
get_version() {
    if [ -n "$1" ]; then
        VERSION="$1"
    else
        # 尝试从 git tag 获取版本号
        if git describe --tags --exact-match HEAD 2>/dev/null; then
            VERSION=$(git describe --tags --exact-match HEAD 2>/dev/null)
        else
            VERSION="v1.0.0-dev"
            print_warning "未找到 git tag，使用默认版本: $VERSION"
        fi
    fi
    print_info "构建版本: $VERSION"
}

# 清理构建目录
clean_build() {
    if [ -d "build" ]; then
        print_info "清理构建目录..."
        rm -rf build
    fi
    mkdir -p build
}

# 构建单个平台
build_platform() {
    local goos=$1
    local goarch=$2
    local suffix=$3
    local name="httpgo-${goos}-${goarch}"
    
    print_info "构建 ${goos}/${goarch}..."
    
    # 设置环境变量
    export GOOS=$goos
    export GOARCH=$goarch
    export CGO_ENABLED=0
    
    # 构建
    if go build -ldflags="-s -w -X 'github.com/itnxs/httpgo/internal/pkg.Version=${VERSION}'" \
        -o "build/${name}${suffix}" .; then
        print_success "构建完成: ${name}${suffix}"
        
        # 创建压缩包
        cd build
        if [ "$goos" = "windows" ]; then
            zip "${name}.zip" "${name}${suffix}"
            print_success "创建压缩包: ${name}.zip"
        else
            tar -czf "${name}.tar.gz" "${name}${suffix}"
            print_success "创建压缩包: ${name}.tar.gz"
        fi
        cd ..
    else
        print_error "构建失败: ${goos}/${goarch}"
        return 1
    fi
}

# 主构建函数
build_all() {
    print_info "开始构建所有平台版本..."
    
    # 构建计数器
    local success_count=0
    local total_count=0
    
    # 定义构建目标
    # 格式: "操作系统:架构:后缀"
    local targets=(
        "windows:amd64:.exe"
        "windows:386:.exe"
        "windows:arm64:.exe"
        "linux:amd64:"
        "linux:386:"
        "linux:arm64:"
        "linux:arm:"
        "darwin:amd64:"
        "darwin:arm64:"
        "freebsd:amd64:"
    )
    
    for target in "${targets[@]}"; do
        IFS=':' read -r goos goarch suffix <<< "$target"
        total_count=$((total_count + 1))
        
        if build_platform "$goos" "$goarch" "$suffix"; then
            success_count=$((success_count + 1))
        fi
    done
    
    print_info "构建完成: $success_count/$total_count 成功"
}

# 生成校验和
generate_checksums() {
    print_info "生成校验和文件..."
    cd build
    
    # 生成 SHA256 校验和
    if command -v sha256sum &> /dev/null; then
        sha256sum *.zip *.tar.gz 2>/dev/null > SHA256SUMS || true
    elif command -v shasum &> /dev/null; then
        shasum -a 256 *.zip *.tar.gz 2>/dev/null > SHA256SUMS || true
    else
        print_warning "未找到 sha256sum 或 shasum 命令，跳过校验和生成"
        cd ..
        return
    fi
    
    if [ -f "SHA256SUMS" ]; then
        print_success "校验和文件已生成: SHA256SUMS"
        echo "文件列表:"
        cat SHA256SUMS
    fi
    cd ..
}

# 显示构建结果
show_results() {
    print_info "构建结果:"
    echo
    if [ -d "build" ]; then
        ls -lh build/
        echo
        print_info "构建文件总数: $(ls -1 build/ | wc -l)"
        
        # 计算总大小
        if command -v du &> /dev/null; then
            total_size=$(du -sh build/ | cut -f1)
            print_info "构建文件总大小: $total_size"
        fi
    else
        print_error "构建目录不存在"
    fi
}

# 使用说明
show_usage() {
    cat << EOF
HttpGo 多平台构建脚本

用法:
    $0 [选项] [版本号]

选项:
    -h, --help      显示此帮助信息
    -c, --clean     仅清理构建目录
    -v, --verbose   详细输出模式

参数:
    版本号          指定构建版本号 (例如: v1.0.0)
                   如果不指定，将尝试从 git tag 获取

示例:
    $0                      # 使用默认版本构建
    $0 v1.2.0              # 构建指定版本
    $0 --clean             # 清理构建目录
    
支持的平台:
    - Windows (amd64, 386, arm64)
    - Linux (amd64, 386, arm64, arm)
    - macOS (amd64, arm64)
    - FreeBSD (amd64)

EOF
}

# 主函数
main() {
    local clean_only=false
    local version=""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -c|--clean)
                clean_only=true
                shift
                ;;
            -v|--verbose)
                set -x
                shift
                ;;
            -*)
                print_error "未知选项: $1"
                show_usage
                exit 1
                ;;
            *)
                version="$1"
                shift
                ;;
        esac
    done
    
    # 仅清理模式
    if [ "$clean_only" = true ]; then
        clean_build
        print_success "构建目录已清理"
        exit 0
    fi
    
    print_info "HttpGo 多平台构建脚本"
    echo "================================"
    
    # 检查环境
    check_go
    
    # 获取版本号
    get_version "$version"
    
    # 清理构建目录
    clean_build
    
    # 执行构建
    if build_all; then
        # 生成校验和
        generate_checksums
        
        # 显示结果
        show_results
        
        print_success "所有平台构建完成！"
        echo
        print_info "构建文件位于 build/ 目录中"
        print_info "可以将这些文件上传到 GitHub Releases 或其他分发平台"
    else
        print_error "构建过程中出现错误"
        exit 1
    fi
}

# 运行主函数
main "$@"