package lock

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

// LockStatus 锁状态
type LockStatus struct {
	BoardAddr int
	Data      []byte
	Length    int
}

// Controller 锁控板控制器
type Controller struct {
	connType    ConnectionType
	path        string
	baudRate    int
	port        int
	socketConn  net.Conn
	serialConn  *serial.Port
	isConnected bool
}

// NewController 创建锁控板控制器实例
func NewController(connType ConnectionType, path string, baudRate, port int) *Controller {
	return &Controller{
		connType: connType,
		path:     path,
		baudRate: baudRate,
		port:     port,
	}
}

// Connect 连接锁控板
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
		return fmt.Errorf("锁控板串口连接失败: %w", err)
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
		return fmt.Errorf("锁控板 Socket 连接失败: %w", err)
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

// generateCommand 生成锁控板命令
func generateCommand(hexCmd string) []byte {
	// 命令格式: 所有字节异或校验
	bytes := make([]byte, len(hexCmd)/2)
	for i := 0; i < len(hexCmd); i += 2 {
		b := parseHexByte(hexCmd, i)
		bytes[i/2] = b
	}

	// 计算校验码
	checksum := bytes[0]
	for i := 1; i < len(bytes); i++ {
		checksum ^= bytes[i]
	}

	// 添加校验码
	result := make([]byte, len(bytes)+1)
	copy(result, bytes)
	result[len(bytes)] = checksum

	return result
}

// parseHexByte 从十六进制字符串解析一个字节
func parseHexByte(s string, i int) byte {
	b1 := s[i]
	b2 := s[i+1]
	var v byte
	switch {
	case b1 >= '0' && b1 <= '9':
		v = (b1 - '0') << 4
	case b1 >= 'A' && b1 <= 'F':
		v = (b1 - 'A' + 10) << 4
	case b1 >= 'a' && b1 <= 'f':
		v = (b1 - 'a' + 10) << 4
	}

	switch {
	case b2 >= '0' && b2 <= '9':
		v |= b2 - '0'
	case b2 >= 'A' && b2 <= 'F':
		v |= b2 - 'A' + 10
	case b2 >= 'a' && b2 <= 'f':
		v |= b2 - 'a' + 10
	}
	return v
}

// generateQueryCommand 生成查询命令
func generateQueryCommand() []byte {
	return generateCommand("80010033")
}

// generateQueryAllCommand 生成查询所有锁状态命令
func generateQueryAllCommand(boardAddr int) []byte {
	boardHex := fmt.Sprintf("%02X", boardAddr)
	hexCmd := "80" + boardHex + "01"
	return generateCommand(hexCmd)
}

// generateOpenCommand 生成开锁命令
func generateOpenCommand(boardAddr, lockAddr int) []byte {
	boardHex := fmt.Sprintf("%02X", boardAddr)
	lockHex := fmt.Sprintf("%02X", lockAddr)
	hexCmd := "8A" + boardHex + lockHex + "11"
	return generateCommand(hexCmd)
}

// Query 查询锁状态
func (c *Controller) Query() error {
	if !c.isConnected {
		return fmt.Errorf("未连接")
	}
	cmd := generateQueryCommand()
	_, err := c.Write(cmd)
	return err
}

// QueryAll 查询所有锁的状态（每个板地址查询一次）
func (c *Controller) QueryAll() ([]LockStatus, error) {
	if !c.isConnected {
		return nil, fmt.Errorf("未连接")
	}

	var allStatus []LockStatus

	// 查询 1-8 号板（假设最大8个板）
	for boardAddr := 1; boardAddr <= 8; boardAddr++ {
		cmd := generateQueryAllCommand(boardAddr)
		_, err := c.Write(cmd)
		if err != nil {
			return nil, fmt.Errorf("查询板地址 %d 失败: %w", boardAddr, err)
		}

		time.Sleep(50 * time.Millisecond)

		// 读取响应
		if c.serialConn != nil {
			c.serialConn.Flush()
		} else if c.socketConn != nil {
			c.socketConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		}

		buf := make([]byte, 256)
		n, err := c.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("读取板地址 %d 响应失败: %w", boardAddr, err)
		}

		if n > 0 {
			status := LockStatus{
				BoardAddr: boardAddr,
				Data:      buf[:n],
				Length:    n,
			}
			allStatus = append(allStatus, status)
		}
	}

	return allStatus, nil
}

// Open 打开指定的锁
func (c *Controller) Open(boardAddr, lockAddr int) error {
	if !c.isConnected {
		return fmt.Errorf("未连接")
	}
	cmd := generateOpenCommand(boardAddr, lockAddr)
	_, err := c.Write(cmd)
	return err
}

// TestConnection 测试连接
func (c *Controller) TestConnection() (bool, error) {
	if err := c.Connect(); err != nil {
		return false, err
	}
	defer c.Disconnect()

	// 发送查询命令
	if err := c.Query(); err != nil {
		return false, err
	}

	// 设置读取超时
	if c.serialConn != nil {
		c.serialConn.Flush()
	}

	// 读取响应
	buf := make([]byte, 256)
	if c.serialConn != nil {
		// 串口超时已在 OpenPort 时设置
	} else if c.socketConn != nil {
		c.socketConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	}

	n, err := c.Read(buf)
	if err != nil {
		return false, fmt.Errorf("读取响应失败: %w", err)
	}
	if n > 0 {
		fmt.Printf("锁控板响应: %X\n", buf[:n])
		return true, nil
	}

	return false, fmt.Errorf("无响应")
}
