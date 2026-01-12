# 硬件测试工具

用于测试智能柜硬件模块连接的 Go 命令行工具。

## 参考代码

本项目的协议实现基于以下 Node.js 代码：

| 模块 | Go 代码 | Node.js 参考代码 | 说明 |
|------|---------|------------------|------|
| **RFID 读写器** | `pkg/rfid/rfid.go` | `/packages/rfid/src/rfid-reader.ts` | TCP Socket 连接，支持查询功率、读取 EPC 等命令 |
| **锁控板** | `pkg/lock/lock.go` | `/packages/lock-control/src/lock-controller.ts` | 支持 Socket/串口连接，异或校验协议 |
| **串口屏** | `pkg/screen/screen.go` | `/packages/screen/src/screen-controller.ts` | 支持 Socket/串口连接，EE/FC 帧协议 |
| **读卡器** | `pkg/cardreader/cardreader.go` | `/packages/card-reader/src/index.ts` | HID USB 设备，控制传输初始化 |

### 协议文档参考

- **锁控板协议**: `/packages/lock-control/docs/两路锁控板通讯协议说明书.md`
- **锁控板通信说明**: `/packages/lock-control/docs/锁控板通信说明.md`
- **串口屏协议**: `/packages/screen/docs/串口屏命令文档.md`

## 支持的模块

- **RFID 读写器** - TCP Socket 连接
- **锁控板** - 串口或 Socket 连接
- **串口屏** - 串口或 Socket 连接
- **读卡器** - HID USB 连接

## 编译

### Linux

```bash
cd hardware-test
go mod tidy
go build -o hardware-test cmd/main.go
```

### Docker 构建（推荐用于老版本系统）

如果目标系统 GLIBC 版本较低（如麒麟操作系统），建议使用 Docker 在 Ubuntu 18.04 环境中编译：

```bash
# 使用 Docker 构建脚本
./docker-build.sh

# 或手动运行
docker build -t hardware-test-builder .
docker run --rm -v "$(pwd)/build:/build" hardware-test-builder \
  sh -c "cd cmd && go build -ldflags='-s -w' -o /build/hardware-test main.go"
```

Docker 构建产生的二进制文件兼容 GLIBC 2.27 及更高版本，适合在麒麟操作系统等老版本系统上运行。

### Windows

```bash
cd hardware-test
go mod tidy
go build -o hardware-test.exe cmd/main.go
```

## 使用方法

### RFID 读写器测试

```bash
# TCP Socket 连接
./hardware-test -module rfid -host 192.168.1.100 -port 8086 -antennas "1,2,3,4"
```

参数说明:
- `-host`: RFID 读写器 IP 地址
- `-port`: RFID 读写器端口号
- `-antennas`: 要启用的天线列表 (默认: 1,2,3,4)

### 锁控板测试

```bash
# 串口连接
./hardware-test -module lock -serial /dev/ttyUSB0 -baud 9600

# Socket 连接
./hardware-test -module lock -host 192.168.1.101 -port 8080
```

参数说明:
- `-serial`: 串口设备路径 (如 /dev/ttyUSB0 或 COM1)
- `-baud`: 波特率 (默认: 115200)
- `-host`: Socket 连接的 IP 地址
- `-port`: Socket 连接的端口号

### 串口屏测试

```bash
# 串口连接
./hardware-test -module screen -serial /dev/ttyUSB1 -baud 115200

# Socket 连接
./hardware-test -module screen -host 192.168.1.102 -port 8081
```

### 读卡器测试

```bash
# USB HID 连接
./hardware-test -module cardreader -vid 0x1234 -pid 0x5678
```

参数说明:
- `-vid`: USB 厂商 ID (十六进制)
- `-pid`: USB 产品 ID (十六进制)

### 测试所有模块

```bash
# 测试所有网络连接的模块
./hardware-test -module all -host 192.168.1.100 -port 8086

# 测试指定模块组合
./hardware-test -module "rfid,lock,screen" -host 192.168.1.100 -port 8086
```

## 测试成功标准

程序通过发送简单的通信命令并验证设备响应来判断连接是否成功:

- **RFID**: 发送查询功率命令，检查是否有有效响应
- **锁控板**: 发送查询状态命令，检查是否有有效响应
- **串口屏**: 发送清屏命令，检查连接是否建立
- **读卡器**: 打开 HID 设备，验证设备可访问

## 项目结构

```
hardware-test/
├── cmd/
│   └── main.go          # 命令行入口
├── pkg/
│   ├── rfid/            # RFID 模块
│   │   └── rfid.go
│   ├── lock/            # 锁控模块
│   │   └── lock.go
│   ├── screen/          # 屏幕模块
│   │   └── screen.go
│   └── cardreader/      # 读卡器模块
│       └── cardreader.go
├── go.mod
└── README.md
```

## 依赖

- Go 1.23+ (仅使用标准库)

### 串口和 HID 支持说明

当前版本使用纯 Go 标准库实现，**仅支持 Socket 网络连接**。如需支持：

- **串口连接**: 需要添加 CGO 依赖 `github.com/tarm/serial`
- **HID USB**: 需要添加 HID 库依赖（在 Linux/Windows 原生环境）

完整功能需要在有串口和 USB 支持的环境中编译。

## 注意事项

1. Linux 下串口设备可能需要 sudo 权限
2. USB HID 设备可能需要配置 udev 规则
3. Windows 下串口设备名为 COM1, COM2 等
4. WSL 下可能无法直接访问串口和 USB 设备

## 故障排除

### 找不到串口设备

```bash
# Linux 查看可用串口
ls -l /dev/ttyUSB*

# Windows 查看可用串口
# 在设备管理器中查看 "端口 (COM 和 LPT)"
```

### 找不到 USB HID 设备

```bash
# Linux 查看 USB 设备
lsusb

# 查看设备详细信息
sudo usbhid-dump
```

### 权限问题

```bash
# Linux 添加用户到 dialout 组 (串口访问)
sudo usermod -a -G dialout $USER

# 配置 HID 设备 udev 规则
sudo nano /etc/udev/rules.d/99-hid.rules
# 添加: KERNEL=="hidraw*", SUBSYSTEM=="hidraw", MODE="0666"
sudo udevadm control --reload-rules
```

## 协议说明

### RFID 协议

- 帧头: 0x5A
- 协议控制字: 4 字节
- 长度: 2 字节
- 数据: N 字节
- 校验: CRC-16

### 锁控板协议

- 命令格式: 所有字节异或校验
- 查询命令: 80010033
- 开锁命令: 8A + 板地址 + 锁地址 + 11

### 串口屏协议

- 帧头: EE
- 长度: 1 字节
- 前缀: FF
- 数据: GBK 编码
- 帧尾: FC

### 读卡器

- HID USB 设备
- 使用控制传输进行初始化
- 中断传输读取卡片数据
