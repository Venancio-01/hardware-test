package screen

import (
	"fmt"
	"net"
	"time"

	"github.com/tarm/serial"
)

// ConnectionType 连接类型
type ConnectionType string

const (
	TypeSerial ConnectionType = "serial"
	TypeSocket ConnectionType = "socket"
)

// Controller 屏幕控制器
type Controller struct {
	connType    ConnectionType
	path        string
	baudRate    int
	port        int
	socketConn  net.Conn
	serialConn  *serial.Port
	isConnected bool
}

// NewController 创建屏幕控制器实例
func NewController(connType ConnectionType, path string, baudRate, port int) *Controller {
	return &Controller{
		connType: connType,
		path:     path,
		baudRate: baudRate,
		port:     port,
	}
}

// Connect 连接屏幕
func (c *Controller) Connect() error {
	if c.connType == TypeSerial {
		return c.connectSerial()
	}
	return c.connectSocket()
}

// connectSerial 串口连接
func (c *Controller) connectSerial() error {
	config := &serial.Config{
		Name:        c.path,
		Baud:        c.baudRate,
		ReadTimeout: 5 * time.Second,
	}

	conn, err := serial.OpenPort(config)
	if err != nil {
		return fmt.Errorf("屏幕串口连接失败: %w", err)
	}
	c.serialConn = conn
	c.isConnected = true
	return nil
}

// connectSocket Socket 连接
func (c *Controller) connectSocket() error {
	addr := fmt.Sprintf("%s:%d", c.path, c.port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("屏幕 Socket 连接失败: %w", err)
	}
	c.socketConn = conn
	c.isConnected = true
	return nil
}

// Disconnect 断开连接
func (c *Controller) Disconnect() error {
	if !c.isConnected {
		return nil
	}

	var err error
	if c.serialConn != nil {
		err = c.serialConn.Close()
		c.serialConn = nil
	}
	if c.socketConn != nil {
		err2 := c.socketConn.Close()
		if err == nil {
			err = err2
		}
		c.socketConn = nil
	}
	c.isConnected = false
	return err
}

// Write 写入数据
func (c *Controller) Write(data []byte) (int, error) {
	if !c.isConnected {
		return 0, fmt.Errorf("未连接")
	}

	if c.serialConn != nil {
		return c.serialConn.Write(data)
	}
	if c.socketConn != nil {
		return c.socketConn.Write(data)
	}
	return 0, fmt.Errorf("未连接")
}

// Read 读取数据
func (c *Controller) Read(data []byte) (int, error) {
	if !c.isConnected {
		return 0, fmt.Errorf("未连接")
	}

	if c.serialConn != nil {
		return c.serialConn.Read(data)
	}
	if c.socketConn != nil {
		return c.socketConn.Read(data)
	}
	return 0, fmt.Errorf("未连接")
}

// stringToGBKHex 将字符串转换为 GBK 编码的十六进制
func stringToGBKHex(s string) string {
	result := ""
	for _, c := range s {
		if c < 128 {
			result += fmt.Sprintf("%02X", c)
		} else {
			r := []byte(string(c))
			for _, b := range r {
				result += fmt.Sprintf("%02X", b)
			}
		}
	}
	return result
}

// generateCommand 生成屏幕命令
func generateCommand(cmdID, command string) []byte {
	dataHex := stringToGBKHex(command)
	frame := "EE" + cmdID + dataHex + "FF"

	length := (len(frame) / 2) - 1
	lengthHex := fmt.Sprintf("%02X", length)

	fullCmd := "EE" + lengthHex + frame + "FC"

	return hexToBytes(fullCmd)
}

// hexToBytes 十六进制字符串转字节
func hexToBytes(hexStr string) []byte {
	bytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		var b byte
		b1 := hexStr[i]
		b2 := hexStr[i+1]

		switch {
		case b1 >= '0' && b1 <= '9':
			b = (b1 - '0') << 4
		case b1 >= 'A' && b1 <= 'F':
			b = (b1 - 'A' + 10) << 4
		case b1 >= 'a' && b1 <= 'f':
			b = (b1 - 'a' + 10) << 4
		}

		switch {
		case b2 >= '0' && b2 <= '9':
			b |= b2 - '0'
		case b2 >= 'A' && b2 <= 'F':
			b |= b2 - 'A' + 10
		case b2 >= 'a' && b2 <= 'f':
			b |= b2 - 'a' + 10
		}
		bytes[i/2] = b
	}
	return bytes
}

// SendCommand 发送命令
func (c *Controller) SendCommand(cmdID, command string) error {
	if !c.isConnected {
		return fmt.Errorf("未连接")
	}
	cmd := generateCommand(cmdID, command)
	_, err := c.Write(cmd)
	return err
}

// TestConnection 测试连接
func (c *Controller) TestConnection() (bool, error) {
	if err := c.Connect(); err != nil {
		return false, err
	}
	defer c.Disconnect()

	if err := c.SendCommand("00", `t0.txt=""`); err != nil {
		return false, err
	}

	time.Sleep(100 * time.Millisecond)

	return true, nil
}
