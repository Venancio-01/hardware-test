package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"hardware-test/pkg/cardreader"
	"hardware-test/pkg/lock"
	"hardware-test/pkg/rfid"
	"hardware-test/pkg/screen"
)

func main() {
	// 定义命令行参数
	module := flag.String("module", "", "要测试的模块: rfid, lock, screen, cardreader, all")
	host := flag.String("host", "", "设备地址 (用于 rfid, lock, screen 的 socket 连接)")
	port := flag.Int("port", 0, "端口号 (用于 rfid, lock, screen 的 socket 连接)")
	serialPort := flag.String("serial", "/dev/ttyS0", "串口路径 (用于 lock, screen 的串口连接，默认: /dev/ttyS0)")
	baudRate := flag.Int("baud", 115200, "波特率 (用于 lock, screen 的串口连接)")
	vid := flag.Int("vid", 0x1A86, "读卡器 VID (十六进制, 如 0x1234, 默认: 0x1A86)")
	pid := flag.Int("pid", 0xE000, "读卡器 PID (十六进制, 如 0x5678, 默认: 0xE000)")
	antennas := flag.String("antennas", "1,2,3,4", "RFID 天线列表 (逗号分隔)")
	flag.Parse()

	if *module == "" {
		printUsage()
		os.Exit(1)
	}

	// 解析天线列表
	antennaList := parseAntennas(*antennas)

	// 根据模块执行测试
	modules := strings.Split(*module, ",")
	successCount := 0
	failCount := 0

	for _, m := range modules {
		m = strings.TrimSpace(m)
		fmt.Printf("\n========== 测试 %s 模块 ==========\n", strings.ToUpper(m))
		success, err := testModule(m, *host, *port, *serialPort, *baudRate, *vid, *pid, antennaList)
		if success {
			fmt.Printf("✓ %s 模块测试通过\n", strings.ToUpper(m))
			successCount++
		} else {
			fmt.Printf("✗ %s 模块测试失败: %v\n", strings.ToUpper(m), err)
			failCount++
		}
	}

	fmt.Printf("\n========== 测试结果 ==========\n")
	fmt.Printf("成功: %d, 失败: %d, 总计: %d\n", successCount, failCount, successCount+failCount)

	if failCount > 0 {
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("硬件测试工具")
	fmt.Println("\n用法:")
	fmt.Println("  hardware-test [选项]")
	fmt.Println("\n选项:")
	fmt.Println("  -module string")
	fmt.Println("        要测试的模块: rfid, lock, screen, cardreader, all")
	fmt.Println("  -host string")
	fmt.Println("        设备地址 (用于 socket 连接)")
	fmt.Println("  -port int")
	fmt.Println("        端口号 (用于 socket 连接)")
	fmt.Println("  -serial string")
	fmt.Println("        串口路径 (用于串口连接，默认: /dev/ttyS0)")
	fmt.Println("  -baud int")
	fmt.Println("        波特率 (默认: 115200)")
	fmt.Println("  -vid int")
	fmt.Println("        读卡器 VID (十六进制)")
	fmt.Println("  -pid int")
	fmt.Println("        读卡器 PID (十六进制)")
	fmt.Println("  -antennas string")
	fmt.Println("        RFID 天线列表 (默认: 1,2,3,4)")
	fmt.Println("\n示例:")
	fmt.Println("  # 测试 RFID (socket)")
	fmt.Println("  hardware-test -module rfid -host 192.168.1.100 -port 8086")
	fmt.Println("\n  # 测试锁控板 (串口，默认 /dev/ttyS0)")
	fmt.Println("  hardware-test -module lock")
	fmt.Println("  # 或指定串口")
	fmt.Println("  hardware-test -module lock -serial /dev/ttyUSB0 -baud 9600")
	fmt.Println("\n  # 测试屏幕 (socket)")
	fmt.Println("  hardware-test -module screen -host 192.168.1.101 -port 8080")
	fmt.Println("\n  # 测试读卡器 (USB HID，使用默认 VID/PID)")
	fmt.Println("  hardware-test -module cardreader")
	fmt.Println("  # 或指定 VID/PID")
	fmt.Println("  hardware-test -module cardreader -vid 0x1234 -pid 0x5678")
	fmt.Println("\n  # 测试所有模块")
	fmt.Println("  hardware-test -module all")
}

func parseAntennas(s string) []int {
	parts := strings.Split(s, ",")
	antennas := make([]int, 0, len(parts))
	for _, p := range parts {
		var ant int
		fmt.Sscanf(strings.TrimSpace(p), "%d", &ant)
		if ant > 0 {
			antennas = append(antennas, ant)
		}
	}
	if len(antennas) == 0 {
		return []int{1, 2, 3, 4}
	}
	return antennas
}

func testModule(module, host string, port int, serialPort string, baudRate int, vid, pid int, antennas []int) (bool, error) {
	switch module {
	case "rfid":
		return testRFID(host, port, antennas)
	case "lock":
		return testLock(host, port, serialPort, baudRate)
	case "screen":
		return testScreen(host, port, serialPort, baudRate)
	case "cardreader":
		return testCardReader(vid, pid)
	case "all":
		// 测试所有模块
		allSuccess := true
		var lastErr error

		if host != "" && port > 0 {
			// 测试基于 socket 的模块
			if success, err := testRFID(host, port, antennas); !success {
				allSuccess = false
				lastErr = err
			}
			if success, err := testLock(host, port, "", 0); !success {
				allSuccess = false
				lastErr = err
			}
			if success, err := testScreen(host, port, "", 0); !success {
				allSuccess = false
				lastErr = err
			}
		}

		if vid != 0 && pid != 0 {
			if success, err := testCardReader(vid, pid); !success {
				allSuccess = false
				lastErr = err
			}
		}

		return allSuccess, lastErr
	default:
		return false, fmt.Errorf("未知模块: %s", module)
	}
}

func testRFID(host string, port int, antennas []int) (bool, error) {
	if host == "" || port == 0 {
		return false, fmt.Errorf("RFID 测试需要 -host 和 -port 参数")
	}

	reader := rfid.NewReader(host, port, antennas)
	fmt.Printf("连接 RFID 读写器: %s:%d (天线: %v)\n", host, port, antennas)
	return reader.TestConnection()
}

func testLock(host string, port int, serialPort string, baudRate int) (bool, error) {
	var controller *lock.Controller

	if host != "" && port > 0 {
		controller = lock.NewController(lock.TypeSocket, host, 0, port)
		fmt.Printf("连接锁控板 (Socket): %s:%d\n", host, port)
	} else {
		controller = lock.NewController(lock.TypeSerial, serialPort, baudRate, 0)
		fmt.Printf("连接锁控板 (串口): %s (波特率: %d)\n", serialPort, baudRate)
	}

	success, err := controller.TestConnection()
	if !success {
		return false, err
	}

	if err := controller.Connect(); err != nil {
		return false, err
	}
	defer controller.Disconnect()

	allStatus, err := controller.QueryAll()
	if err != nil {
		return false, err
	}

	fmt.Printf("\n========== 锁状态报告 ==========\n")
	for _, status := range allStatus {
		fmt.Printf("板地址: 0x%02X\n", status.BoardAddr)
		fmt.Printf("  命令长度: %d 字节\n", status.Length)
		fmt.Printf("  原始数据: %X\n", status.Data)
		fmt.Printf("  十六进制: ")
		for _, b := range status.Data {
			fmt.Printf("%02X ", b)
		}
		fmt.Println()
	}

	return true, nil
}

func testScreen(host string, port int, serialPort string, baudRate int) (bool, error) {
	var controller *screen.Controller

	if host != "" && port > 0 {
		controller = screen.NewController(screen.TypeSocket, host, 0, port)
		fmt.Printf("连接屏幕 (Socket): %s:%d\n", host, port)
	} else {
		controller = screen.NewController(screen.TypeSerial, serialPort, baudRate, 0)
		fmt.Printf("连接屏幕 (串口): %s (波特率: %d)\n", serialPort, baudRate)
	}

	return controller.TestConnection()
}

func testCardReader(vid, pid int) (bool, error) {
	if vid == 0 || pid == 0 {
		return false, fmt.Errorf("读卡器测试需要 -vid 和 -pid 参数")
	}

	reader := cardreader.NewReader(vid, pid)
	fmt.Printf("连接读卡器: VID=0x%04X, PID=0x%04X\n", vid, pid)
	return reader.TestConnection()
}
