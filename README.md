# ZeroDriver

Raspberry Pi Zero W 기반 라인 트레이서 RC카. Go + AWS AI-DLC.

## Diagnostic Tools

| Tool | Purpose |
|------|---------|
| `hwtest` | All sensors sequential check |
| `imutest` | MPU-9250 I2C |
| `lidartest` | RPLidar A1 `/dev/ttyUSB0` |
| `camtest` | Camera + line detection |
| `motortest` | Motor spin test (`-confirm` required) |

```bash
# Mock (dev PC)
go run ./cmd/hwtest -mode mock

# Build all for Pi
./deploy/build-arm.sh
```

## Documentation

- [Overview](docs/01-overview.md)
- [Requirements](docs/02-requirements.md)
- [Hardware](docs/03-hardware.md)
- [Architecture](docs/04-architecture.md)
- [AI-DLC Process](docs/05-aidlc-process.md)
- [Development Guide](docs/06-development-guide.md)
- [Hardware HAL](docs/07-hardware-hal.md)
- [Telemetry](docs/08-telemetry.md)
- [Field Test](docs/09-field-test.md)

## Project Structure

```
cmd/zerodriver/     Main entry point
cmd/lidartest/      RPLidar A1 diagnostic tool
internal/domain/    Domain types
internal/hal/       Hardware abstraction (mock + hardware)
internal/perception/ Line detection, IMU, LiDAR
internal/fusion/    Sensor fusion
internal/control/   PID, state machine, motor
configs/            Runtime configuration
deploy/             Build & systemd scripts
docs/               Development documentation
```

## Pi Deployment

```bash
./deploy/build-arm.sh
scp bin/zerodriver-armv6 pi@raspberrypi.local:/home/pi/zerodriver
scp bin/lidartest-armv6 pi@raspberrypi.local:/home/pi/lidartest
scp configs/zerodriver-hardware.yaml pi@raspberrypi.local:/home/pi/zerodriver.yaml

# LiDAR 단독 테스트 (RPLidar A1, /dev/ttyUSB0)
./lidartest -port /dev/ttyUSB0 -duration 10s

# 전체 주행
./zerodriver -config zerodriver.yaml -mode hardware -v
```
