#!/bin/bash

set -e

BUILD_DIR="build"
BINARY_NAME="hardware-test"

echo "========================================="
echo "硬件测试工具 - Linux x64 构建脚本"
echo "========================================="
echo ""

echo "[1/4] 清理旧的构建产物..."
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

echo "[2/4] 检查 Go 环境..."
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go 环境"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "Go 版本: ${GO_VERSION}"

echo "[3/4] 下载依赖..."
go mod download

echo "[4/4] 编译 Linux x64 二进制文件..."
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

echo "编译参数: GOOS=${GOOS}, GOARCH=${GOARCH}, CGO_ENABLED=${CGO_ENABLED}"

cd cmd
go build -ldflags="-s -w" -tags netgo -o "../${BUILD_DIR}/${BINARY_NAME}" main.go
cd ..

if [ $? -ne 0 ]; then
    echo "编译失败！"
    exit 1
fi

chmod +x "${BUILD_DIR}/${BINARY_NAME}"

echo ""
echo "========================================="
echo "构建完成！"
echo "========================================="
echo ""
echo "二进制文件位置: ${BUILD_DIR}/${BINARY_NAME}"
echo "文件大小: $(du -h ${BUILD_DIR}/${BINARY_NAME} | cut -f1)"
echo ""
echo "运行方式:"
echo "  ${BUILD_DIR}/${BINARY_NAME} -help"
echo ""
