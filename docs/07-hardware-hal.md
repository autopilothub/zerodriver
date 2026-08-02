# Hardware HAL

Pi Zero W 실제 하드웨어 드라이버. `linux && arm` 빌드 태그로만 컴파일된다.

## 드라이버 목록

| 파일 | 장치 | 인터페이스 |
|------|------|-----------|
| `mpu9250_linux_arm.go` | MPU-9250 IMU | I2C (`/dev/i2c-1`, 0x68) |
| `rplidar_linux_arm.go` | RPLidar A1 | USB Serial (`/dev/ttyUSB0`, 115200) |
| `pca9685_linux_arm.go` | PCA9685 PWM | I2C (`0x40`, 50Hz) |
| `pca9685_motor_linux_arm.go` | 서보 + ESC | CH0=조향, CH1=스로틀 |
| `camera_linux_arm.go` | Pi Camera | V4L2 (`/dev/video0`) |
| `camera_rpicam_linux_arm.go` | Pi Camera | rpicam-vid (libcamera) |

## Mock vs Hardware

```bash
# 개발 PC (mock)
go run ./cmd/zerodriver -mode mock

# Pi Zero W (hardware) — Pi에서 직접 빌드
go build -o zerodriver ./cmd/zerodriver
./zerodriver -config zerodriver.yaml -mode hardware

# 개발 PC에서 cross-compile
./deploy/build-arm.sh
scp bin/zerodriver-armv6 pi@raspberrypi.local:/home/pi/zerodriver
```

개발 PC에서 `-mode hardware`를 사용하면 `linux/arm` 빌드가 아니므로 에러가 반환된다.

## Pi 사전 설정

```bash
# I2C
sudo raspi-config  # Interface Options → I2C → Enable
sudo i2cdetect -y 1  # 0x68 확인

# UART (LiDAR)
sudo raspi-config  # Serial → No login shell, Serial enabled
# /boot/config.txt: enable_uart=1

# 카메라
sudo raspi-config  # Interface Options → Camera → Enable
rpicam-hello --list-cameras  # OV5647 등 확인

# Bookworm+ 에서 /dev/video0 은 raw unicam 노드라 V4L2 직접 스트리밍이 실패할 수 있음.
# 기본 camera_backend: rpicam (rpicam-vid 파이프).
sudo apt install -y rpicam-apps
```

## MPU-9250

- WHO_AM_I 레지스터(0x75)에서 0x71 확인
- 가속도 ±2g, 자이로 ±250°/s 기본 스케일
- Yaw는 자이로 Z축 적분 (magnetometer 미사용)

## RPLidar A1

- **연결**: USB serial `/dev/ttyUSB0` (115200 baud)
- **모델 설정**: `hardware.lidar_model: rplidar_a1`
- 부팅 시 RESET → SCAN 명령으로 연속 스캔
- 종료 시 STOP 명령(0xA5 0x25) 전송
- 5바이트 노드 파싱 (`internal/hal/rplidar_protocol.go`)
- 전방 `scan_front_angle` 범위 내 최소 거리를 `FrontDistance`로 반환 (cm)

### Pi 설정

```bash
# USB serial 권한
sudo usermod -aG dialout pi

# 포트 확인
ls -l /dev/ttyUSB0

# 단독 테스트
./lidartest -port /dev/ttyUSB0 -duration 10s

# mock 모드 (개발 PC)
./lidartest -mode mock -duration 5s
```

## PCA9685 (RC카: 서보 + ESC)

- **I2C 주소**: `0x40` (기본)
- **주파수**: 50Hz
- **2채널만 사용**:

```yaml
steering_channel: 0    # 조향 서보
throttle_channel: 1    # ESC 스로틀
servo_min_us: 1000
servo_center_us: 1500
servo_max_us: 2000
throttle_min_us: 1000  # 정지
throttle_max_us: 2000  # 최대 속도
```

- **조향**: PID steering (-1~+1) → 서보 펄스
- **속도**: throttle (0~1) → ESC 펄스

## 카메라 (V4L2)

- MJPEG 우선, 미지원 시 YUYV 폴백
- MJPEG → `image/jpeg` 디코드 → RGB
- 권장 해상도: 320x240

## 의존성

```
periph.io/x/host/v3        GPIO, I2C
go.bug.st/serial           UART (LiDAR)
github.com/blackjack/webcam V4L2
```

이 패키지들은 `linux/arm` 빌드 태그 파일에서만 import되므로, mock 모드 빌드에는 링크되지 않는다.
