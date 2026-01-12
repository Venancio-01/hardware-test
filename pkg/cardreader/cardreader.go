package cardreader

import (
	"fmt"
	"time"

	"github.com/karalabe/hid"
)

// Reader 读卡器
type Reader struct {
	vid         int
	pid         int
	device      *hid.Device
	isConnected bool
}

// NewReader 创建读卡器实例
func NewReader(vid, pid int) *Reader {
	return &Reader{
		vid: vid,
		pid: pid,
	}
}

// Connect 连接读卡器
func (r *Reader) Connect() error {
	if r.vid == 0 || r.pid == 0 {
		return fmt.Errorf("无效的 VID/PID")
	}

	devices := hid.Enumerate(uint16(r.vid), uint16(r.pid))
	if len(devices) == 0 {
		return fmt.Errorf("未找到 HID 设备 (VID: 0x%04X, PID: 0x%04X)", r.vid, r.pid)
	}

	device, err := devices[0].Open()
	if err != nil {
		return fmt.Errorf("打开 HID 设备失败: %w", err)
	}

	r.device = device
	r.isConnected = true
	return nil
}

// Disconnect 断开连接
func (r *Reader) Disconnect() error {
	if r.device != nil {
		err := r.device.Close()
		r.device = nil
		r.isConnected = false
		return err
	}
	r.isConnected = false
	return nil
}

// Read 读取卡片数据
func (r *Reader) Read() (string, error) {
	if !r.isConnected {
		return "", fmt.Errorf("设备未连接")
	}

	data := make([]byte, 64)
	n, err := r.device.Read(data)
	if err != nil {
		return "", fmt.Errorf("读取数据失败: %w", err)
	}

	if n > 0 {
		return fmt.Sprintf("%X", data[:n]), nil
	}

	return "", fmt.Errorf("无数据")
}

// ReadWithTimeout 读取卡片数据（带超时）
func (r *Reader) ReadWithTimeout(timeout time.Duration) (string, error) {
	if !r.isConnected {
		return "", fmt.Errorf("设备未连接")
	}

	data := make([]byte, 64)
	done := make(chan struct{})
	var result string
	var err error

	go func() {
		n, readErr := r.device.Read(data)
		if readErr == nil && n > 0 {
			result = fmt.Sprintf("%X", data[:n])
		}
		err = readErr
		close(done)
	}()

	select {
	case <-done:
		if err != nil {
			return "", fmt.Errorf("读取数据失败: %w", err)
		}
		if result != "" {
			return result, nil
		}
		return "", fmt.Errorf("无数据")
	case <-time.After(timeout):
		return "", fmt.Errorf("读取超时")
	}
}

// TestConnection 测试连接
func (r *Reader) TestConnection() (bool, error) {
	if err := r.Connect(); err != nil {
		return false, err
	}

	fmt.Printf("读卡器已连接 (VID: 0x%04X, PID: 0x%04X)\n", r.vid, r.pid)
	fmt.Printf("设备信息: %s - %s\n", r.device.Product, r.device.Manufacturer)

	r.Disconnect()
	return true, nil
}
