# 하드웨어 명세

## BOM (Bill of Materials)

| 부품 | 모델 (권장) | 수량 | 비고 |
|------|------------|------|------|
| SBC | Raspberry Pi Zero W | 1 | WiFi 내장 |
| IMU | MPU-9250 / MPU-6500 breakout | 1 | I2C, 6~9-DOF |
| 카메라 | Pi Camera Module v2 / HQ | 1 | CSI 커넥터 |
| LiDAR | Slamtec RPLidar A1 | 1 | USB serial, 5V |
| 모터 드라이버 | DRV8833 / L298N | 1 | 2채널 DC |
| PWM/서보 | PCA9685 breakout | 1 | I2C 16채널 PWM |
| DC 모터 | TT 기어드 모터 | 2 | 차동 구동 |
| 배터리 | LiPo 2S 7.4V | 1 | BEC 5V for Pi |
| 섀시 | 3D 프린트 / 키트 | 1 | 카메라 전방 하향 |

## 핀맵 (Raspberry Pi Zero W)

### MPU-9250 / MPU-6500 (I2C)

| IMU | Pi Zero W | GPIO (BCM) |
|-----|-----------|------------|
| VCC | 3.3V | — |
| GND | GND | — |
| SDA | Pin 3 | GPIO2 |
| SCL | Pin 5 | GPIO3 |

I2C 주소: `0x68` (AD0=LOW) 또는 `0x69` (AD0=HIGH)

지원 칩 (WHO_AM_I 자동 인식):
- `0x70` — MPU-6500 (6축: 가속도+자이로)
- `0x71` — MPU-9250 (9축)
- `0x73` — MPU-9255

### RPLidar A1 (USB Serial)

| RPLidar A1 | 연결 | 비고 |
|------------|------|------|
| VCC | 5V (USB 전원) | USB 어댑터 보드 사용 |
| GND | GND | |
| USB | Pi Zero W USB OTG | `/dev/ttyUSB0` |

- **모델**: Slamtec RPLidar A1
- **포트**: `/dev/ttyUSB0` (USB-UART 어댑터)
- **보드레이트**: `115200`
- **드라이버**: `internal/hal/rplidar_linux_arm.go`

```bash
# USB 권한 (Pi에서 1회)
sudo usermod -aG dialout pi
# 재로그인 후
ls -l /dev/ttyUSB0

# LiDAR 단독 테스트
go build -o lidartest ./cmd/lidartest
./lidartest -port /dev/ttyUSB0 -duration 10s
```

> Pi Zero W는 USB 포트가 1개(micro-USB OTG)이므로, LiDAR는 USB 허브 또는 OTG 어댑터로 연결한다.
> GPIO UART(`/dev/serial0`) 대신 USB serial을 사용하면 WiFi/Bluetooth와 핀 충돌이 없다.

### PCA9685 (서보 + ESC PWM)

MPU-9250과 동일 I2C 버스(`/dev/i2c-1`)에 연결.

| PCA9685 | Pi Zero W | 비고 |
|---------|-----------|------|
| VCC | 3.3V 또는 5V | 보드 사양 확인 |
| GND | GND | |
| SDA | Pin 3 | GPIO2 |
| SCL | Pin 5 | GPIO3 |

I2C 주소: `0x40` (기본)

| 채널 | 용도 | 연결 |
|------|------|------|
| CH0 | 조향 서보 | 서보 신호선 |
| CH1 | 스로틀 PWM | ESC 신호선 |

- PWM 주파수: 50Hz (서보/ESC 표준)
- 서보: 1000µs(좌) ~ 1500µs(중앙) ~ 2000µs(우)
- ESC: 1000µs(정지) ~ 2000µs(최대 속도)

```bash
# I2C 확인
sudo i2cdetect -y 1
# 0x40 (PCA9685), 0x68 (MPU-9250) 표시되어야 함
```

### 카메라 (CSI)

Pi Zero W 전용 CSI 리본 케이블 사용. 소프트웨어에서 `libcamera` 또는 V4L2로 접근.

## 배선 주의사항

1. **전원**: Pi Zero W는 5V 2A 이상 권장. 모터 전원과 Pi 전원 분리 (공통 GND).
2. **I2C 풀업**: MPU-9250 보드에 내장 풀업 저항 확인.
3. **UART 레벨**: LiDAR가 5V TTL이면 레벨 시프터 사용.
4. **카메라 각도**: 전방 30~45° 하향, 라인이 이미지 하단 1/3에 위치하도록 장착.

## Pi Zero W 제약

| 항목 | 사양 | 대응 |
|------|------|------|
| CPU | 1GHz 단일코어 ARM1176 | ROI 축소, 경량 알고리즘 |
| RAM | 512MB | 이미지 버퍼 재사용 |
| 카메라 | 30fps @ 640x480 한계 | 320x240 권장 |
| USB | micro-USB 1개 | WiFi 내장으로 절약 |

## 물리 치수 (권장)

- 바퀴 간격 (트랙 width): 12~15cm
- 카메라 높이: 8~12cm (바닥 기준)
- LiDAR 장착: 전방 중앙, 높이 10cm
