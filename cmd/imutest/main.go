// IMU diagnostic tool for MPU-9250 / MPU-6500.
// Usage: go run ./cmd/imutest -config configs/zerodriver-hardware.yaml -duration 5s
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/autopilothub/zerodriver/internal/config"
	"github.com/autopilothub/zerodriver/internal/hal"
)

func main() {
	configPath := flag.String("config", "configs/zerodriver.yaml", "config file")
	mode := flag.String("mode", "hardware", "mode: mock or hardware")
	duration := flag.Duration("duration", 0, "run duration (0 = until Ctrl+C)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	cfg.Mode = *mode

	imu, err := hal.NewIMU(cfg)
	if err != nil {
		log.Fatalf("init IMU: %v", err)
	}
	defer imu.Close()

	chip := "IMU"
	if named, ok := imu.(interface{ ChipName() string }); ok {
		chip = named.ChipName()
	}
	log.Printf("%s on %s addr=0x%02X (mode=%s)", chip, cfg.Hardware.I2CBus, cfg.Hardware.I2CAddr, cfg.Mode)
	if who, ok := imu.(interface{ WhoAmI() byte }); ok {
		log.Printf("WHO_AM_I=0x%02X", who.WhoAmI())
	}
	if mag, ok := imu.(interface{ HasMagnetometer() bool }); ok {
		if mag.HasMagnetometer() {
			log.Printf("compass: AK8963 active")
		} else {
			log.Printf("compass: not available (6-axis IMU or AK8963 not detected)")
		}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	deadline := time.Now().Add(*duration)
	if *duration == 0 {
		deadline = time.Time{}
	}

	for {
		select {
		case <-sigCh:
			log.Println("stopped")
			return
		case <-ticker.C:
			if !deadline.IsZero() && time.Now().After(deadline) {
				log.Println("done")
				return
			}
			att, err := imu.Read()
			if err != nil {
				log.Printf("read error: %v", err)
				continue
			}
			line := fmt.Sprintf("yaw=%.1f° pitch=%.1f° roll=%.1f° gyroZ=%.1f°/s",
				att.Yaw, att.Pitch, att.Roll, att.GyroZ)
			if att.HasMag {
				line += fmt.Sprintf(" heading=%.0f° mag=(%.1f,%.1f,%.1f)µT",
					att.Heading, att.MagX, att.MagY, att.MagZ)
			}
			fmt.Println(line)
		}
	}
}
