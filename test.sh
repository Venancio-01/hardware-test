#!/bin/bash

# 硬件测试脚本
# 用法: ./test.sh [rfid|lock|screen|cardreader|all]

MODULE=${1:-all}

echo "========================================="
echo "硬件模块测试工具"
echo "========================================="
echo ""

# 检查是否已编译
if [ ! -f "./hardware-test" ]; then
    echo "未找到可执行文件，正在编译..."
    go build -o hardware-test cmd/main.go
    if [ $? -ne 0 ]; then
        echo "编译失败！"
        exit 1
    fi
    echo "编译成功！"
    echo ""
fi

# 根据参数执行测试
case $MODULE in
    rfid)
        echo "测试 RFID 读写器..."
        ./hardware-test -module rfid -host 192.168.1.100 -port 8086
        ;;
    lock)
        echo "测试锁控板..."
        ./hardware-test -module lock -host 192.168.1.101 -port 8080
        ;;
    screen)
        echo "测试串口屏..."
        ./hardware-test -module screen -host 192.168.1.102 -port 8081
        ;;
    cardreader)
        echo "测试读卡器..."
        ./hardware-test -module cardreader -vid 0x1234 -pid 0x5678
        ;;
    all)
        echo "测试所有模块..."
        ./hardware-test -module all -host 192.168.1.100 -port 8086 -vid 0x1234 -pid 0x5678
        ;;
    *)
        echo "用法: $0 [rfid|lock|screen|cardreader|all]"
        exit 1
        ;;
esac

echo ""
echo "========================================="
echo "测试完成"
echo "========================================="
