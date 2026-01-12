package rfid

import (
	"fmt"
	"net"
	"time"
)

// Reader RFID 读写器
type Reader struct {
	host     string
	port     int
	conn     net.Conn
	antennas []int
}

// NewReader 创建 RFID 读写器实例
func NewReader(host string, port int, antennas []int) *Reader {
	return &Reader{
		host:     host,
		port:     port,
		antennas: antennas,
	}
}

// Connect 连接 RFID 读写器
func (r *Reader) Connect() error {
	addr := fmt.Sprintf("%s:%d", r.host, r.port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("RFID 连接失败: %w", err)
	}
	r.conn = conn
	return nil
}

// Disconnect 断开连接
func (r *Reader) Disconnect() error {
	if r.conn != nil {
		// 发送停止命令
		r.Stop()
		return r.conn.Close()
	}
	return nil
}

// buildRFIDCommand 构建 RFID 命令
func buildRFIDCommand(cmdType uint16, dataParams string) []byte {
	// 协议: 帧头(1) + 协议控制字(4) + 长度(2) + 数据(N) + 校验(2)
	// 帧头: 0x5A
	// 协议控制字: 设备类型(2字节) + 命令类型(1字节) + 命令方向(1字节)

	// 简化版本：直接使用已知格式的命令
	// 5A + 00000100 + 1000 + length + data + CRC
	command := "5A00000100"

	// 长度字段 (数据长度，2字节)
	dataLen := len(dataParams) / 2
	lengthHex := fmt.Sprintf("%04X", dataLen)
	command += lengthHex

	// 数据参数
	command += dataParams

	// CRC 校验 (使用简单的累加异或)
	crc := calculateCRC(command)
	command += crc

	// 转换为字节
	return hexToBytes(command)
}

// calculateCRC 计算 CRC 校验码
func calculateCRC(hexStr string) string {
	var crc uint16 = 0xFFFF
	for i := 0; i < len(hexStr); i += 2 {
		b := parseHexByte(hexStr, i)
		crc ^= uint16(b) << 8
		for j := 0; j < 8; j++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc = crc << 1
			}
		}
	}
	return fmt.Sprintf("%04X", crc)
}

// parseHexByte 从十六进制字符串解析一个字节
func parseHexByte(s string, i int) byte {
	b1 := s[i]
	b2 := s[i+1]
	var v byte
	if b1 >= '0' && b1 <= '9' {
		v = (b1 - '0') << 4
	} else {
		v = (b1-'A'+10) << 4
	}
	if b2 >= '0' && b2 <= '9' {
		v |= b2 - '0'
	} else {
		v |= b2 - 'A' + 10
	}
	return v
}

// hexToBytes 十六进制字符串转字节
func hexToBytes(hexStr string) []byte {
	bytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		bytes[i/2] = parseHexByte(hexStr, i)
	}
	return bytes
}

// generateStopCommand 生成停止命令
func generateStopCommand() []byte {
	// MID = 0xFF, 停止操作命令
	return buildRFIDCommand(0xFF, "")
}

// generateReadEPCCommand 生成读取 EPC 命令
func generateReadEPCCommand(antennas []int) []byte {
	// 构建天线端口位掩码
	var antennaMask uint32
	for _, ant := range antennas {
		if ant >= 1 && ant <= 32 {
			antennaMask |= 1 << (ant - 1)
		}
	}

	dataParams := fmt.Sprintf("%08X", antennaMask) // 天线端口
	dataParams += "01"                              // 连续读取
	dataParams += "02"                              // 读取参数 (TID)
	dataParams += "0006"                            // TID 读取参数

	return buildRFIDCommand(0x10, dataParams)
}

// generateQueryPowerCommand 生成查询功率命令
func generateQueryPowerCommand() []byte {
	return buildRFIDCommand(0x02, "")
}

// Stop 停止读取
func (r *Reader) Stop() error {
	if r.conn == nil {
		return fmt.Errorf("未连接")
	}
	cmd := generateStopCommand()
	_, err := r.conn.Write(cmd)
	return err
}

// StartReading 开始读取 RFID 标签
func (r *Reader) StartReading() error {
	if r.conn == nil {
		return fmt.Errorf("未连接")
	}

	// 先发送停止命令
	r.Stop()
	time.Sleep(100 * time.Millisecond)

	// 发送读取命令
	cmd := generateReadEPCCommand(r.antennas)
	_, err := r.conn.Write(cmd)
	return err
}

// QueryPower 查询功率
func (r *Reader) QueryPower() error {
	if r.conn == nil {
		return fmt.Errorf("未连接")
	}
	cmd := generateQueryPowerCommand()
	_, err := r.conn.Write(cmd)
	return err
}

// TestConnection 测试连接
func (r *Reader) TestConnection() (bool, error) {
	// 连接设备
	if err := r.Connect(); err != nil {
		return false, err
	}
	defer r.Disconnect()

	// 发送查询功率命令作为测试
	if err := r.QueryPower(); err != nil {
		return false, err
	}

	// 设置读取超时
	r.conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	// 尝试读取响应
	buf := make([]byte, 256)
	n, err := r.conn.Read(buf)
	if err != nil {
		return false, fmt.Errorf("读取响应失败: %w", err)
	}

	if n > 0 {
		fmt.Printf("RFID 响应: %X\n", buf[:n])
		return true, nil
	}

	return false, fmt.Errorf("无响应")
}
