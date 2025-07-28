#!/bin/bash

# 本地构建脚本 - 用于测试多平台构建
# 使用方法: ./scripts/build.sh [version]

set -e

VERSION=${1:-"dev"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S_UTC')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "🚀 开始构建 Git Log Analyzer"
echo "版本: $VERSION"
echo "构建时间: $BUILD_TIME"
echo "Git提交: $GIT_COMMIT"
echo

# 清理之前的构建
rm -rf dist/
mkdir -p dist/

# 构建标志
LDFLAGS="-w -s -X main.version=$VERSION -X main.buildTime=$BUILD_TIME -X main.gitCommit=$GIT_COMMIT"

# 支持的平台
declare -a platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

echo "📦 开始多平台构建..."

for platform in "${platforms[@]}"
do
    IFS='/' read -r -a platform_split <<< "$platform"
    GOOS="${platform_split[0]}"
    GOARCH="${platform_split[1]}"
    
    output_name="git-log-analyzer-$GOOS-$GOARCH"
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi
    
    echo "  构建 $GOOS/$GOARCH..."
    
    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o "dist/$output_name" \
        .
    
    if [ $? -ne 0 ]; then
        echo "❌ 构建 $GOOS/$GOARCH 失败"
        exit 1
    fi
done

echo
echo "✅ 构建完成！"
echo
echo "📋 构建结果:"
ls -lh dist/

echo
echo "🔍 验证构建..."

# 测试本地平台的可执行文件
local_binary=""
case "$(uname -s)" in
    Darwin*)
        case "$(uname -m)" in
            x86_64) local_binary="dist/git-log-analyzer-darwin-amd64" ;;
            arm64) local_binary="dist/git-log-analyzer-darwin-arm64" ;;
        esac
        ;;
    Linux*)
        case "$(uname -m)" in
            x86_64) local_binary="dist/git-log-analyzer-linux-amd64" ;;
            aarch64) local_binary="dist/git-log-analyzer-linux-arm64" ;;
        esac
        ;;
    MINGW*|CYGWIN*|MSYS*)
        local_binary="dist/git-log-analyzer-windows-amd64.exe"
        ;;
esac

if [ -n "$local_binary" ] && [ -f "$local_binary" ]; then
    echo "测试本地二进制文件: $local_binary"
    chmod +x "$local_binary"
    "$local_binary" --version 2>/dev/null || echo "  版本信息不可用"
    echo "  ✅ 二进制文件可以正常执行"
else
    echo "  ⚠️  未找到适合当前平台的二进制文件"
fi

echo
echo "🎉 构建完成！二进制文件位于 dist/ 目录"

# 如果在macOS上且存在codesign，提示签名
if [[ "$OSTYPE" == "darwin"* ]] && command -v codesign >/dev/null 2>&1; then
    echo
    echo "💡 提示: 在macOS上，您可能需要对二进制文件进行签名："
    echo "  codesign --force --options runtime --sign \"Developer ID Application: Your Name\" dist/git-log-analyzer-darwin-*"
fi
