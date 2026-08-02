# Field Test 가이드 (Bolt 4.x)

Pi Zero W 실기 테스트 절차.

## 사전 준비

```bash
# Pi 설정
sudo raspi-config
# - I2C Enable
# - Camera Enable
# - Serial: login shell No, serial port Yes (LiDAR는 USB 사용 시 불필요)

# USB LiDAR 권한
sudo usermod -aG dialout pi
# 재로그인

# 바이너리 배포
./deploy/build-arm.sh
scp bin/*-armv6 configs/zerodriver-hardware.yaml pi@raspberrypi.local:/home/pi/
```

## 테스트 순서

### 1단계: 전체 하드웨어 점검

```bash
git pull
./deploy/build-arm.sh

# 순서대로 자동 점검 (IMU → Camera → LiDAR → hwtest)
chmod +x deploy/verify-pi.sh
./deploy/verify-pi.sh configs/zerodriver-hardware.yaml

# 또는 수동
./bin/hwtest-armv6 -config configs/zerodriver-hardware.yaml
```

예상 출력:
```
--- IMU (MPU-6500/9250) ---
  yaw=0.1° pitch=2.0° roll=-1.0° gyroZ=0.5°/s
PASS

--- LiDAR (RPLidar A1) ---
  front=999cm points=12
PASS

--- Camera + Line Detection ---
  line offset=0.050 (detected)
PASS

=== Result: 3/3 passed ===
```

### 2단계: 개별 센서 테스트

```bash
# PCA9685 (서보 + ESC) — 차량 들어 올린 상태!
./pcatest-armv6 -config zerodriver-hardware.yaml -confirm

# IMU
./imutest-armv6 -config zerodriver-hardware.yaml -duration 5s

# LiDAR (RPLidar A1, /dev/ttyUSB0)
./lidartest-armv6 -port /dev/ttyUSB0 -duration 10s

# 카메라 + 라인 검출 (프레임 저장)
./camtest-armv6 -config zerodriver-hardware.yaml -save frame.ppm -duration 3s
# frame.ppm을 PC로 복사하여 확인

# 모터 (차량을 들어 올린 상태에서!)
./motortest-armv6 -config zerodriver-hardware.yaml -confirm -duration 2s
```

### 3단계: 모터 포함 전체 점검

```bash
./hwtest-armv6 -config zerodriver-hardware.yaml -motor -confirm
```

### 4단계: 주행 테스트

```bash
# 바닥에 라인 테이프 부착 후
./zerodriver-armv6 -config zerodriver-hardware.yaml -mode hardware -v
```

## PID 튜닝

`zerodriver-hardware.yaml`에서 PID 게인 조정:

```yaml
pid:
  steering:
    kp: 0.8   # 반응성 (높으면 진동)
    ki: 0.01  # 정상상태 오차
    kd: 0.05  # 오버슈트 억제
```

튜닝 순서:
1. Kp만 증가 → 진동 전까지
2. Kd 추가 → 오버슈트 감소
3. Ki 소량 추가 → 잔여 오차 제거

## 트러블슈팅

| 증상 | 확인 | 해결 |
|------|------|------|
| IMU FAIL | `i2cdetect -y 1` | `0x68` 확인 (0x70은 PCA9685 All Call) |
| LiDAR FAIL | `ls /dev/ttyUSB0` | USB 연결, dialout 그룹 |
| Camera FAIL | `rpicam-hello --list-cameras` | CSI 케이블, `sudo apt install rpicam-apps` |
| `start streaming: invalid argument` | V4L2 unicam 노드 | `camera_backend: rpicam` (기본값) |
| `rpicam first frame: EOF` | rpicam-vid 즉시 종료 | `git pull` 후 재빌드; 수동 테스트: `rpicam-vid -n -t 5000 --width 640 --height 480 --codec yuv420 --inline -o /tmp/t.data` |
| line not detected | `camtest -save frame.ppm` | 조명, 라인 대비, ROI 조정 |
| 탈선 | verbose 로그 확인 | Kp 감소, corner_speed 감소 |
| AVOIDING 고착 | lidartest | 전방 오탐지, scan_front_angle 조정 |

## systemd 자동 시작

```bash
sudo cp deploy/zerodriver.service /etc/systemd/system/
# ExecStart 경로를 /home/pi/zerodriver-armv6 로 수정
sudo systemctl enable zerodriver
sudo systemctl start zerodriver
sudo journalctl -u zerodriver -f
```

## Human Validation Gate

Field Test 완료 체크리스트:

- [ ] hwtest 3/3 PASS (모터 4/4)
- [ ] 직선 구간 탈선 없음
- [ ] 90° 코너 통과
- [ ] 장애물 감지 후 정지
- [ ] 5분 이상 연속 주행 안정성
