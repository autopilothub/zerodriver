package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/autopilothub/zerodriver/internal/config"
	"github.com/autopilothub/zerodriver/internal/control"
	"github.com/autopilothub/zerodriver/internal/domain"
	"github.com/autopilothub/zerodriver/internal/fusion"
	"github.com/autopilothub/zerodriver/internal/hal"
	"github.com/autopilothub/zerodriver/internal/perception"
	"github.com/autopilothub/zerodriver/internal/telemetry"
	"github.com/autopilothub/zerodriver/internal/web"
)

func main() {
	configPath := flag.String("config", "configs/zerodriver.yaml", "path to config file")
	mode := flag.String("mode", "", "override mode: mock or hardware")
	lidar := flag.String("lidar", "", "override lidar: on or off")
	verbose := flag.Bool("v", false, "verbose logging")
	duration := flag.Duration("duration", 0, "run duration (0 = until signal)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if *mode != "" {
		cfg.Mode = *mode
	}
	cfg.ApplyLidarOverride(*lidar)

	devices, err := hal.NewDevices(cfg)
	if err != nil {
		log.Fatalf("init devices: %v", err)
	}
	defer closeAll(devices)

	pub, err := telemetry.NewPublisher(&cfg.Telemetry)
	if err != nil {
		log.Fatalf("init telemetry: %v", err)
	}
	defer pub.Close()

	lineDetector := perception.NewLineDetector(
		cfg.Camera.Width, cfg.Camera.Height, cfg.Camera.ROIY,
		cfg.Camera.LookaheadRows, cfg.Camera.LineThreshold,
	)
	imuReader := perception.NewIMUReader(devices.IMU)
	lidarParser := perception.NewLidarParser(devices.Lidar)
	fuser := fusion.NewWithHeadingLock(cfg.Control.IMUFusion.HeadingLockOffset)
	ctrl := control.NewController(cfg, devices.Motor)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *duration > 0 {
		go func() {
			time.Sleep(*duration)
			cancel()
		}()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	log.Printf("zerodriver starting (mode=%s, loop=%dHz, lidar=%v)", cfg.Mode, cfg.Control.LoopHz, cfg.LidarEnabled())

	var statusStore *web.Store
	if cfg.Web.Enabled {
		statusStore = web.NewStore()
		webSrv := web.NewServer(&cfg.Web, statusStore)
		go func() {
			log.Printf("web dashboard http://localhost%s", cfg.Web.Addr)
			if err := webSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("web server error: %v", err)
			}
		}()
		defer func() {
			shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancelShutdown()
			_ = webSrv.Shutdown(shutdownCtx)
		}()
	}

	startedAt := time.Now()

	if err := ctrl.Arm(ctx); err != nil {
		log.Fatalf("ESC arm: %v", err)
	}
	log.Printf("ESC armed (%ds)", cfg.Control.ESCArmDelaySec)
	log.Printf("throttle pulse range: %dµs (stop) → %dµs (max), CH%d",
		cfg.Hardware.ThrottleMinUs, cfg.Hardware.ThrottleMaxUs, cfg.Hardware.ThrottleChannel)
	ctrl.Start()

	interval := time.Second / time.Duration(cfg.Control.LoopHz)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	telemetryInterval := time.Duration(cfg.Telemetry.IntervalSec) * time.Second
	telemetryTicker := time.NewTicker(telemetryInterval)
	defer telemetryTicker.Stop()

	var lastLine domain.LinePosition
	var lastAtt domain.Attitude
	var lastObs = domain.ObstacleScan{FrontDistance: 999}
	var lastCmd domain.ControlCommand

	var camErrors, imuErrors, lidarErrors atomic.Uint64

	lineCh := make(chan domain.LinePosition, 1)
	imuCh := make(chan domain.Attitude, 1)
	lidarCh := make(chan domain.ObstacleScan, 1)

	go sensorLoop(ctx, time.Second/30, func() {
		frame, err := devices.Camera.Capture()
		if err != nil {
			n := camErrors.Add(1)
			if *verbose || n%30 == 1 {
				log.Printf("camera error (#%d): %v", n, err)
			}
			return
		}
		if statusStore != nil {
			statusStore.SetFrame(frame, cfg.Camera.Width, cfg.Camera.Height)
		}
		pos := lineDetector.Detect(frame)
		send(lineCh, pos)
	})

	go sensorLoop(ctx, time.Second/50, func() {
		att, err := imuReader.Read()
		if err != nil {
			n := imuErrors.Add(1)
			if *verbose || n%50 == 1 {
				log.Printf("imu error (#%d): %v", n, err)
			}
			return
		}
		send(imuCh, att)
	})

	if cfg.LidarEnabled() {
		go sensorLoop(ctx, time.Second/5, func() {
			obs, err := lidarParser.Scan()
			if err != nil {
				n := lidarErrors.Add(1)
				if *verbose || n%5 == 1 {
					log.Printf("lidar error (#%d): %v", n, err)
				}
				return
			}
			send(lidarCh, obs)
		})
	} else {
		log.Printf("lidar disabled (obstacle avoidance off)")
	}

	for {
		select {
		case <-ctx.Done():
			ctrl.Stop()
			log.Printf("zerodriver stopped (cam_err=%d imu_err=%d lidar_err=%d)",
				camErrors.Load(), imuErrors.Load(), lidarErrors.Load())
			return

		case lastLine = <-lineCh:
		case lastAtt = <-imuCh:
		case lastObs = <-lidarCh:

		case <-ticker.C:
			fused := fuser.Fuse(lastLine, lastAtt, lastObs, ctrl.State())

			manualTimeout := time.Duration(cfg.Web.ManualTimeoutMS) * time.Millisecond
			if statusStore != nil {
				if active, mcmd := statusStore.ManualCommand(manualTimeout); active {
					lastCmd = ctrl.DriveManual(mcmd)
				} else {
					lastCmd = ctrl.Tick(fused, interval.Seconds())
				}
			} else {
				lastCmd = ctrl.Tick(fused, interval.Seconds())
			}

			if statusStore != nil {
				statusStore.Update(web.UpdateInput{
					Mode:        cfg.Mode,
					State:       ctrl.State(),
					Fused:       fused,
					Command:     lastCmd,
					Attitude:    lastAtt,
					CamErrors:   camErrors.Load(),
					IMUErrors:   imuErrors.Load(),
					LidarErrors: lidarErrors.Load(),
					StartedAt:   startedAt,
				})
			}

			if *verbose {
				log.Printf("state=%s line=%.2f steering=%.2f throttle=%.2f obstacle=%.0fcm",
					ctrl.State(), fused.LineOffset, lastCmd.Steering, lastCmd.Throttle, fused.FrontDistance)
			}

		case <-telemetryTicker.C:
			if !cfg.Telemetry.Enabled && cfg.Telemetry.Endpoint == "" {
				continue
			}
			fused := fuser.Fuse(lastLine, lastAtt, lastObs, ctrl.State())
			snap := telemetry.NewSnapshot(ctrl.State(), fused, lastCmd, lastAtt)
			if err := pub.Publish(snap); err != nil {
				log.Printf("telemetry error: %v", err)
			}
		}
	}
}

func send[T any](ch chan T, val T) {
	select {
	case ch <- val:
	default:
		ch <- val
	}
}

func sensorLoop(ctx context.Context, interval time.Duration, fn func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn()
		}
	}
}

func closeAll(d *hal.Devices) {
	d.IMU.Close()
	d.Lidar.Close()
	d.Camera.Close()
	d.Motor.Close()
}
