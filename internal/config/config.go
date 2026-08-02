package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Mode      string          `yaml:"mode"`
	Control   ControlConfig   `yaml:"control"`
	PID       PIDConfig       `yaml:"pid"`
	Obstacle  ObstacleConfig  `yaml:"obstacle"`
	Camera    CameraConfig    `yaml:"camera"`
	Hardware  HardwareConfig  `yaml:"hardware"`
	Telemetry TelemetryConfig `yaml:"telemetry"`
}

type ControlConfig struct {
	LoopHz           int     `yaml:"loop_hz"`
	BaseSpeed        float64 `yaml:"base_speed"`
	CornerSpeed      float64 `yaml:"corner_speed"`
	CornerThreshold  float64 `yaml:"corner_threshold"`
}

type PIDConfig struct {
	Steering PIDGains `yaml:"steering"`
}

type PIDGains struct {
	Kp float64 `yaml:"kp"`
	Ki float64 `yaml:"ki"`
	Kd float64 `yaml:"kd"`
}

type ObstacleConfig struct {
	StopDistanceCM  float64 `yaml:"stop_distance_cm"`
	ScanFrontAngle  float64 `yaml:"scan_front_angle"`
}

type CameraConfig struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
	ROIY   int `yaml:"roi_y"` // start row of region of interest
}

type HardwareConfig struct {
	I2CBus       string `yaml:"i2c_bus"`
	I2CAddr      int    `yaml:"i2c_addr"`
	LidarModel   string `yaml:"lidar_model"`
	LidarPort    string `yaml:"lidar_port"`
	LidarBaud    int    `yaml:"lidar_baud"`
	CameraDevice string `yaml:"camera_device"`
	MotorLPWM    int    `yaml:"motor_l_pwm"`
	MotorLDir    int    `yaml:"motor_l_dir"`
	MotorRPWM    int    `yaml:"motor_r_pwm"`
	MotorRDir    int    `yaml:"motor_r_dir"`
}

type TelemetryConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Endpoint    string `yaml:"endpoint"`
	Topic       string `yaml:"topic"`
	IntervalSec int    `yaml:"interval_sec"`
	CertDir     string `yaml:"cert_dir"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyDefaults()
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Control.LoopHz == 0 {
		c.Control.LoopHz = 50
	}
	if c.Control.BaseSpeed == 0 {
		c.Control.BaseSpeed = 0.6
	}
	if c.Control.CornerSpeed == 0 {
		c.Control.CornerSpeed = 0.3
	}
	if c.Control.CornerThreshold == 0 {
		c.Control.CornerThreshold = 0.5
	}
	if c.PID.Steering.Kp == 0 {
		c.PID.Steering.Kp = 0.8
	}
	if c.Obstacle.StopDistanceCM == 0 {
		c.Obstacle.StopDistanceCM = 20
	}
	if c.Obstacle.ScanFrontAngle == 0 {
		c.Obstacle.ScanFrontAngle = 60
	}
	if c.Camera.Width == 0 {
		c.Camera.Width = 320
	}
	if c.Camera.Height == 0 {
		c.Camera.Height = 240
	}
	if c.Camera.ROIY == 0 {
		c.Camera.ROIY = 160
	}
	if c.Hardware.I2CAddr == 0 {
		c.Hardware.I2CAddr = 0x68
	}
	if c.Hardware.I2CBus == "" {
		c.Hardware.I2CBus = "/dev/i2c-1"
	}
	if c.Hardware.LidarPort == "" {
		c.Hardware.LidarPort = "/dev/ttyUSB0"
	}
	if c.Hardware.LidarModel == "" {
		c.Hardware.LidarModel = "rplidar_a1"
	}
	if c.Hardware.LidarBaud == 0 {
		c.Hardware.LidarBaud = 115200
	}
	if c.Hardware.CameraDevice == "" {
		c.Hardware.CameraDevice = "/dev/video0"
	}
	if c.Telemetry.IntervalSec == 0 {
		c.Telemetry.IntervalSec = 1
	}
	if c.Telemetry.Topic == "" {
		c.Telemetry.Topic = "zerodriver/telemetry"
	}
}
