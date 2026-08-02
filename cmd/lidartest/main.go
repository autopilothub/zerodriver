// Lidartest is a diagnostic tool for RPLidar A1 on Pi.
// Build on Pi: go build -o lidartest ./cmd/lidartest
// Usage: ./lidartest -port /dev/ttyUSB0 -duration 10s
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
	port := flag.String("port", "/dev/ttyUSB0", "serial port")
	baud := flag.Int("baud", 115200, "baud rate")
	angle := flag.Float64("angle", 60, "front scan angle (degrees)")
	mode := flag.String("mode", "hardware", "mode: hardware or mock")
	duration := flag.Duration("duration", 0, "run duration (0 = until Ctrl+C)")
	flag.Parse()

	cfg := &config.Config{
		Mode: *mode,
		Hardware: config.HardwareConfig{
			LidarPort:  *port,
			LidarBaud:  *baud,
			LidarModel: "rplidar_a1",
		},
		Obstacle: config.ObstacleConfig{
			ScanFrontAngle: *angle,
		},
	}

	lidar, err := hal.NewLidar(cfg)
	if err != nil {
		log.Fatalf("init lidar: %v", err)
	}
	defer lidar.Close()

	log.Printf("RPLidar A1 on %s @ %d baud (front ±%.0f°, mode=%s)", *port, *baud, *angle/2, *mode)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(500 * time.Millisecond)
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
			scan, err := lidar.Scan()
			if err != nil {
				log.Printf("scan error: %v", err)
				continue
			}
			fmt.Printf("front=%.0fcm  points=%d  time=%s\n",
				scan.FrontDistance, len(scan.Distances),
				scan.Timestamp.Format("15:04:05.000"))
		}
	}
}
