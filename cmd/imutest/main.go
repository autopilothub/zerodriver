// IMU diagnostic tool for MPU-9250.
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

	log.Printf("IMU on %s addr=0x%02X (mode=%s)", cfg.Hardware.I2CBus, cfg.Hardware.I2CAddr, cfg.Mode)

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
			fmt.Printf("yaw=%.1f° pitch=%.1f° roll=%.1f° gyroZ=%.1f°/s\n",
				att.Yaw, att.Pitch, att.Roll, att.GyroZ)
		}
	}
}
