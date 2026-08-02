package hal

import (
	"testing"

	"github.com/autopilothub/zerodriver/internal/config"
)

func TestNewDevices_Mock(t *testing.T) {
	cfg := &config.Config{Mode: "mock"}
	devices, err := NewDevices(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer devices.IMU.Close()
	defer devices.Lidar.Close()
	defer devices.Camera.Close()
	defer devices.Motor.Close()

	att, err := devices.IMU.Read()
	if err != nil {
		t.Fatal(err)
	}
	if att.GyroZ == 0 && att.Yaw == 0 {
		t.Fatal("mock IMU should return non-zero values")
	}
}

func TestNewDevices_HardwareOnNonARM(t *testing.T) {
	cfg := &config.Config{Mode: "hardware"}
	_, err := NewDevices(cfg)
	if err == nil {
		t.Fatal("expected error for hardware mode on non-arm build")
	}
}

func TestNewDevices_UnknownMode(t *testing.T) {
	cfg := &config.Config{Mode: "invalid"}
	_, err := NewDevices(cfg)
	if err == nil {
		t.Fatal("expected error for unknown mode")
	}
}
