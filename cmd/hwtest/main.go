// hwtest runs all hardware diagnostics in sequence.
// Usage: ./hwtest -config configs/zerodriver-hardware.yaml -motor -confirm
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/autopilothub/zerodriver/internal/config"
	"github.com/autopilothub/zerodriver/internal/hal"
	"github.com/autopilothub/zerodriver/internal/perception"
)

func main() {
	configPath := flag.String("config", "configs/zerodriver-hardware.yaml", "config file")
	mode := flag.String("mode", "hardware", "mode: mock or hardware")
	runMotor := flag.Bool("motor", false, "include motor test")
	confirm := flag.Bool("confirm", false, "confirm motor spin")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	cfg.Mode = *mode

	log.Printf("=== ZeroDriver Hardware Test (mode=%s) ===\n", cfg.Mode)

	passed := 0
	total := 0

	total++
	if testIMU(cfg) {
		passed++
	}

	total++
	if testLidar(cfg) {
		passed++
	}

	total++
	if testCamera(cfg) {
		passed++
	}

	if *runMotor {
		total++
		if testMotor(cfg, *confirm) {
			passed++
		}
	}

	fmt.Printf("\n=== Result: %d/%d passed ===\n", passed, total)
	if passed < total {
		os.Exit(1)
	}
}

func testIMU(cfg *config.Config) bool {
	fmt.Println("\n--- IMU (MPU-6500/9250) ---")
	imu, err := hal.NewIMU(cfg)
	if err != nil {
		fmt.Printf("FAIL init: %v\n", err)
		return false
	}
	defer imu.Close()

	for i := 0; i < 5; i++ {
		att, err := imu.Read()
		if err != nil {
			fmt.Printf("FAIL read: %v\n", err)
			return false
		}
		fmt.Printf("  yaw=%.1f° pitch=%.1f° roll=%.1f° gyroZ=%.1f°/s\n",
			att.Yaw, att.Pitch, att.Roll, att.GyroZ)
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Println("PASS")
	return true
}

func testLidar(cfg *config.Config) bool {
	fmt.Println("\n--- LiDAR (RPLidar A1) ---")
	lidar, err := hal.NewLidar(cfg)
	if err != nil {
		fmt.Printf("FAIL init: %v\n", err)
		return false
	}
	defer lidar.Close()

	time.Sleep(1 * time.Second) // wait for scan buffer

	for i := 0; i < 3; i++ {
		scan, err := lidar.Scan()
		if err != nil {
			fmt.Printf("FAIL scan: %v\n", err)
			return false
		}
		fmt.Printf("  front=%.0fcm points=%d\n", scan.FrontDistance, len(scan.Distances))
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println("PASS")
	return true
}

func testCamera(cfg *config.Config) bool {
	fmt.Println("\n--- Camera + Line Detection ---")
	camera, err := hal.NewCamera(cfg)
	if err != nil {
		fmt.Printf("FAIL init: %v\n", err)
		return false
	}
	defer camera.Close()

	detector := perception.NewLineDetector(cfg.Camera.Width, cfg.Camera.Height, cfg.Camera.ROIY, cfg.Camera.LineThreshold)

	for i := 0; i < 3; i++ {
		frame, err := camera.Capture()
		if err != nil {
			fmt.Printf("FAIL capture: %v\n", err)
			return false
		}
		pos := detector.Detect(frame)
		if pos.Detected {
			fmt.Printf("  line offset=%.3f (detected)\n", pos.Offset)
		} else {
			fmt.Printf("  line not detected\n")
		}
		time.Sleep(300 * time.Millisecond)
	}
	fmt.Println("PASS")
	return true
}

func testMotor(cfg *config.Config, confirm bool) bool {
	fmt.Println("\n--- Motor ---")
	if cfg.Mode == "hardware" && !confirm {
		fmt.Println("SKIP (use -motor -confirm)")
		return false
	}

	motor, err := hal.NewMotor(cfg)
	if err != nil {
		fmt.Printf("FAIL init: %v\n", err)
		return false
	}
	defer motor.Close()

	fmt.Println("  forward 1s...")
	motor.Drive(0, 0.3)
	time.Sleep(1 * time.Second)
	motor.Stop()
	time.Sleep(300 * time.Millisecond)

	fmt.Println("  steer left 1s...")
	motor.Drive(-0.5, 0)
	time.Sleep(1 * time.Second)
	motor.Stop()

	fmt.Println("PASS")
	return true
}
