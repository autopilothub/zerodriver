# 개발 가이드

## 사전 요구사항

### 개발 PC (macOS/Linux)
- Go 1.22+
- Git

### Raspberry Pi Zero W
- Raspberry Pi OS Lite (32-bit)
- I2C, UART, SPI 활성화 (`raspi-config`)
- Go 1.22+ (또는 cross-compile)

## 로컬 개발

```bash
# 저장소 클론
git clone <repo-url> zerodriver
cd zerodriver

# 의존성 설치
go mod download

# Mock 모드 실행 (하드웨어 없이)
go run ./cmd/zerodriver -config configs/zerodriver.yaml -mode mock

# 테스트
go test ./...

# 빌드
go build -o bin/zerodriver ./cmd/zerodriver
```

## Pi Zero W 배포

### 1. Cross-compile (개발 PC)

```bash
./deploy/build-arm.sh
# → bin/zerodriver-armv6
```

### 2. Pi에 복사

```bash
scp bin/zerodriver-armv6 pi@raspberrypi.local:/home/pi/zerodriver
scp configs/zerodriver.yaml pi@raspberrypi.local:/home/pi/
```

### 3. Pi 설정

```bash
# I2C 활성화
sudo raspi-config  # Interface Options → I2C → Enable

# UART 활성화 (LiDAR)
sudo raspi-config  # Interface Options → Serial → No login shell, Serial enabled

# 카메라 활성화
sudo raspi-config  # Interface Options → Camera → Enable
```

### 4. systemd 서비스

```bash
sudo cp deploy/zerodriver.service /etc/systemd/system/
sudo systemctl enable zerodriver
sudo systemctl start zerodriver
sudo journalctl -u zerodriver -f
```

## 설정 파일

`configs/zerodriver.yaml` 주요 항목:

```yaml
mode: mock          # mock | hardware
control:
  loop_hz: 50
  base_speed: 0.6
  corner_speed: 0.3
  corner_threshold: 0.5
pid:
  steering:
    kp: 0.8
    ki: 0.01
    kd: 0.05
obstacle:
  stop_distance_cm: 20
  scan_front_angle: 60  # degrees
```

## PID 튜닝

1. **Kp만**으로 시작 — 진동 없을 때까지 증가
2. **Kd** 추가 — 오버슈트 감소
3. **Ki** 추가 — 정상 상태 오차 제거 (작은 값)

로그에서 `line_offset`, `steering`, `throttle` 확인:

```bash
# mock mode verbose
go run ./cmd/zerodriver -config configs/zerodriver.yaml -mode mock -v
```

## 트러블슈팅

| 증상 | 원인 | 해결 |
|------|------|------|
| I2C device not found | 배선/활성화 | `i2cdetect -y 1`로 0x68 확인 |
| LiDAR no data | UART 설정 | `/boot/config.txt`에서 `enable_uart=1` |
| 카메라 black frame | CSI 케이블 | 리본 방향, `libcamera-hello` 테스트 |
| 높은 CPU | 해상도 | config에서 320x240으로 축소 |
| 탈선 | PID 미튜닝 | Kp 감소, corner_speed 감소 |

## 로그 형식

```
2026/08/02 20:30:01 state=TRACING line=0.12 steering=0.10 throttle=0.60 obstacle=999cm
```

## AWS IoT (선택)

`configs/zerodriver.yaml`에서 telemetry 활성화:

```yaml
telemetry:
  enabled: true
  endpoint: xxxxxx-ats.iot.ap-northeast-2.amazonaws.com
  topic: zerodriver/telemetry
  interval_sec: 1
```

인증서는 `/etc/zerodriver/certs/`에 배치.
