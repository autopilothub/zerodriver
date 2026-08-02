//go:build linux && arm

package hal

import (
	"fmt"

	"github.com/autopilothub/zerodriver/internal/config"
)

func newHardwareDevices(cfg *config.Config) (*Devices, error) {
	imu, err := NewIMU(cfg)
	if err != nil {
		return nil, fmt.Errorf("init IMU: %w", err)
	}

	lidar, err := NewLidar(cfg)
	if err != nil {
		imu.Close()
		return nil, fmt.Errorf("init LiDAR: %w", err)
	}

	camera, err := NewCamera(cfg)
	if err != nil {
		imu.Close()
		lidar.Close()
		return nil, fmt.Errorf("init camera: %w", err)
	}

	motor, err := NewMotor(cfg)
	if err != nil {
		imu.Close()
		lidar.Close()
		camera.Close()
		return nil, fmt.Errorf("init motor: %w", err)
	}

	return &Devices{
		IMU:    imu,
		Lidar:  lidar,
		Camera: camera,
		Motor:  motor,
	}, nil
}
