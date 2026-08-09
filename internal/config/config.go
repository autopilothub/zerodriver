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
	Web       WebConfig       `yaml:"web"`
}

type ControlConfig struct {
	LoopHz          int              `yaml:"loop_hz"`
	BaseSpeed       float64          `yaml:"base_speed"`
	CornerSpeed     float64          `yaml:"corner_speed"`
	CornerThreshold float64          `yaml:"corner_threshold"`
	LineLostSpeed   float64          `yaml:"line_lost_speed"`
	ESCArmDelaySec  int              `yaml:"esc_arm_delay_sec"`
	PurePursuit     PurePursuitConfig `yaml:"pure_pursuit"`
	IMUFusion       IMUFusionConfig   `yaml:"imu_fusion"`
}

type PurePursuitConfig struct {
	Lookahead float64 `yaml:"lookahead"` // normalized forward distance (0–1)
	Gain      float64 `yaml:"gain"`
}

type IMUFusionConfig struct {
	HeadingKp         float64 `yaml:"heading_kp"`           // heading error → steering (line follow)
	YawDamping        float64 `yaml:"yaw_damping"`          // gyro Z damping
	LineLostHeadingKp float64 `yaml:"line_lost_heading_kp"` // heading hold when line lost
	HeadingLockOffset float64 `yaml:"heading_lock_offset"`  // |line offset| below this locks heading ref
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
	Width          int `yaml:"width"`
	Height         int `yaml:"height"`
	ROIY           int `yaml:"roi_y"`
	LookaheadRows  int `yaml:"lookahead_rows"` // rows above roi_y for Pure Pursuit target
	LineThreshold  int `yaml:"line_threshold"`
}

type HardwareConfig struct {
	I2CBus            string `yaml:"i2c_bus"`
	I2CAddr           int    `yaml:"i2c_addr"`
	IMUModel          string `yaml:"imu_model"` // bno085, mpu9250, mpu6500, auto
	LidarEnabled      *bool  `yaml:"lidar_enabled"` // default true; false skips serial open
	LidarModel        string `yaml:"lidar_model"`
	LidarPort         string `yaml:"lidar_port"`
	LidarBaud         int    `yaml:"lidar_baud"`
	CameraDevice      string `yaml:"camera_device"`
	CameraBackend     string `yaml:"camera_backend"` // rpicam (default), v4l2, auto
	PCA9685Addr       int `yaml:"pca9685_addr"`
	PCA9685FreqHz     int `yaml:"pca9685_freq_hz"`
	SteeringChannel   int `yaml:"steering_channel"`
	ThrottleChannel   int `yaml:"throttle_channel"`
	ServoMinUs        int `yaml:"servo_min_us"`
	ServoCenterUs     int `yaml:"servo_center_us"`
	ServoMaxUs        int `yaml:"servo_max_us"`
	ThrottleMinUs     int `yaml:"throttle_min_us"`
	ThrottleMaxUs     int `yaml:"throttle_max_us"`
}

type TelemetryConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Endpoint    string `yaml:"endpoint"`
	Topic       string `yaml:"topic"`
	IntervalSec int    `yaml:"interval_sec"`
	CertDir     string `yaml:"cert_dir"`
}

type WebConfig struct {
	Enabled         bool   `yaml:"enabled"`
	Addr            string `yaml:"addr"`
	RefreshMS       int    `yaml:"refresh_ms"`
	CameraQuality   int    `yaml:"camera_quality"`
	ManualTimeoutMS int    `yaml:"manual_timeout_ms"`
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

// LidarEnabled reports whether LiDAR hardware should be used.
func (c *Config) LidarEnabled() bool {
	if c.Hardware.LidarEnabled == nil {
		return true
	}
	return *c.Hardware.LidarEnabled
}

// ApplyLidarOverride sets lidar_enabled from a CLI value ("on"/"off", etc.).
func (c *Config) ApplyLidarOverride(v string) {
	switch v {
	case "off", "false", "0", "no":
		enabled := false
		c.Hardware.LidarEnabled = &enabled
	case "on", "true", "1", "yes":
		enabled := true
		c.Hardware.LidarEnabled = &enabled
	}
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
	if c.Control.LineLostSpeed == 0 {
		c.Control.LineLostSpeed = 0.2
	}
	if c.Control.ESCArmDelaySec == 0 {
		c.Control.ESCArmDelaySec = 2
	}
	if c.Control.PurePursuit.Lookahead == 0 {
		c.Control.PurePursuit.Lookahead = 0.35
	}
	if c.Control.PurePursuit.Gain == 0 {
		c.Control.PurePursuit.Gain = 1.0
	}
	if c.Control.IMUFusion.HeadingKp == 0 {
		c.Control.IMUFusion.HeadingKp = 0.03
	}
	if c.Control.IMUFusion.YawDamping == 0 {
		c.Control.IMUFusion.YawDamping = 0.015
	}
	if c.Control.IMUFusion.LineLostHeadingKp == 0 {
		c.Control.IMUFusion.LineLostHeadingKp = 0.06
	}
	if c.Control.IMUFusion.HeadingLockOffset == 0 {
		c.Control.IMUFusion.HeadingLockOffset = 0.15
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
	if c.Camera.LookaheadRows == 0 {
		c.Camera.LookaheadRows = 40
	}
	if c.Camera.LineThreshold == 0 {
		c.Camera.LineThreshold = 100
	}
	if c.Hardware.IMUModel == "" {
		c.Hardware.IMUModel = "bno085"
	}
	if c.Hardware.I2CAddr == 0 {
		switch c.Hardware.IMUModel {
		case "bno085", "bno08x", "bno080":
			c.Hardware.I2CAddr = 0x4B
		default:
			c.Hardware.I2CAddr = 0x68
		}
	}
	if c.Hardware.I2CBus == "" {
		c.Hardware.I2CBus = "/dev/i2c-1"
	}
	if c.Hardware.LidarEnabled == nil {
		enabled := true
		c.Hardware.LidarEnabled = &enabled
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
	if c.Hardware.CameraBackend == "" {
		c.Hardware.CameraBackend = "rpicam"
	}
	if c.Hardware.CameraDevice == "" {
		c.Hardware.CameraDevice = "/dev/video0"
	}
	if c.Hardware.PCA9685Addr == 0 {
		c.Hardware.PCA9685Addr = 0x40
	}
	if c.Hardware.PCA9685FreqHz == 0 {
		c.Hardware.PCA9685FreqHz = 50
	}
	if c.Hardware.ThrottleChannel == 0 {
		c.Hardware.ThrottleChannel = 1
	}
	if c.Hardware.ServoMinUs == 0 {
		c.Hardware.ServoMinUs = 1000
	}
	if c.Hardware.ServoCenterUs == 0 {
		c.Hardware.ServoCenterUs = 1500
	}
	if c.Hardware.ServoMaxUs == 0 {
		c.Hardware.ServoMaxUs = 2000
	}
	if c.Hardware.ThrottleMinUs == 0 {
		c.Hardware.ThrottleMinUs = 1000
	}
	if c.Hardware.ThrottleMaxUs == 0 {
		c.Hardware.ThrottleMaxUs = 2000
	}
	if c.Telemetry.IntervalSec == 0 {
		c.Telemetry.IntervalSec = 1
	}
	if c.Telemetry.Topic == "" {
		c.Telemetry.Topic = "zerodriver/telemetry"
	}
	if c.Web.Addr == "" {
		c.Web.Addr = ":8080"
	}
	if c.Web.RefreshMS == 0 {
		c.Web.RefreshMS = 200
	}
	if c.Web.CameraQuality == 0 {
		c.Web.CameraQuality = 70
	}
	if c.Web.ManualTimeoutMS == 0 {
		c.Web.ManualTimeoutMS = 500
	}
}
