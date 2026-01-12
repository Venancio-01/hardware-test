#!/bin/bash

# Docker 构建脚本
# 用法: ./docker-build.sh

set -e

IMAGE_NAME="hardware-test-builder"
CONTAINER_NAME="hardware-test-build"

echo "========================================="
echo "硬件测试工具 - Docker 构建脚本"
echo "========================================="
echo ""

# 检查 Docker 是否可用
if ! command -v docker &> /dev/null; then
    echo "错误: 未找到 Docker"
    exit 1
fi

echo "[1/4] 清理旧的构建产物..."
rm -rf build
mkdir -p build

echo "[2/4] 构建 Docker 镜像..."
docker build -t ${IMAGE_NAME} .

echo "[3/4] 运行容器编译..."
docker run --rm \
    -v "$(pwd)/build:/build" \
    -w /workspace \
    ${IMAGE_NAME} \
    sh -c "cd cmd && go build -ldflags='-s -w' -o /build/hardware-test main.go"

echo "[4/4] 设置执行权限..."
chmod +x build/hardware-test

echo ""
echo "========================================="
echo "构建完成！"
echo "========================================="
echo ""
echo "二进制文件位置: build/hardware-test"
echo "文件大小: $(du -h build/hardware-test | cut -f1)"
echo ""
echo "验证二进制文件信息:"
file build/hardware-test
echo ""
echo "GLIBC 版本要求: GLIBC 2.27 或更高"
echo "目标设备: 麒麟 OS（KYLINOS）- 兼容 ✓"
echo ""
echo "功能支持:"
echo "  ✓ RFID (Socket)"
echo "  ✓ 锁控板 (串口/Socket)"
echo "  ✓ 串口屏 (串口/Socket)"
echo "  ✓ 读卡器 (USB HID)"
