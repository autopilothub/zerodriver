# 소프트웨어 아키텍처

## 레이어 구조

```
┌─────────────────────────────────────────┐
│           cmd/zerodriver (main)         │  오케스트레이션
├─────────────────────────────────────────┤
│              control                    │  PID, 상태 머신, 모터
├─────────────────────────────────────────┤
│               fusion                    │  센서 융합
├─────────────────────────────────────────┤
│             perception                  │  라인 검출, LiDAR, IMU
├─────────────────────────────────────────┤
│                 hal                     │  GPIO, I2C, UART, Camera
├─────────────────────────────────────────┤
│               domain                    │  엔티티, 값 객체
└─────────────────────────────────────────┘
```

## 패키지 책임

### `internal/domain`
- 순수 도메인 타입, 비즈니스 규칙
- 외부 의존성 없음

### `internal/hal`
- 하드웨어 인터페이스 정의 및 구현
- `Mock*` / `Real*` 구현체 (mode에 따라 선택)

```go
type IMU interface {
    Read() (domain.Attitude, error)
    Close() error
}

type Lidar interface {
    Scan() (domain.ObstacleScan, error)
    Close() error
}

type Camera interface {
    Capture() ([]byte, error)  // raw frame
    Close() error
}

type Motor interface {
    Drive(steering, throttle float64) error  // steering: -1~+1, throttle: 0~1
    Stop() error
    Close() error
}
```

### `internal/perception`
- HAL raw 데이터 → 도메인 객체 변환
- `LineDetector`: 프레임 → `LinePosition`
- `IMUReader`: raw → `Attitude`
- `LidarParser`: scan → `ObstacleScan`

### `internal/fusion`
- 다중 센서 데이터 융합
- 상황별 가중치 적용

| 상황 | 카메라 | IMU | LiDAR |
|------|--------|-----|-------|
| TRACING (정상) | 1.0 | 0.3 (보정) | 0.0 |
| TRACING (라인 소실) | 0.0 | 0.8 | 0.2 |
| AVOIDING | 0.0 | 0.0 | 1.0 |

### `internal/control`
- `PIDController`: steering/throttle 계산
- `StateMachine`: RaceState 전이
- `MotorDriver`: ControlCommand → HAL Motor

## 동시성 모델

```
main goroutine
├── cameraLoop   (30Hz) → lineCh
├── imuLoop      (50Hz) → imuCh
├── lidarLoop    (5Hz)  → lidarCh
├── fusionLoop   (50Hz) → fusedCh
├── controlLoop  (50Hz) → motor
└── telemetryLoop (1Hz, optional)
```

채널 버퍼 크기 1 (최신 값만 유지, stale 데이터 방지).

## 제어 루프

```
1. fusionLoop: lineCh, imuCh, lidarCh에서 읽기
2. StateMachine.Update(obstacle) → state
3. Fusion.Fuse(line, imu, lidar, state) → FusedInput
4. PID.Compute(FusedInput) → ControlCommand
5. Motor.Drive(cmd.Steering, cmd.Throttle)
```

## PID 구조

**Steering PID** (입력: line offset -1~+1, 출력: steering -1~+1)
- Kp=0.8, Ki=0.01, Kd=0.05 (시작값, `configs/zerodriver.yaml`)

**Throttle**
- 기본 `base_speed` (0.0~1.0)
- 코너 감지 시 `|steering| > corner_threshold` → `corner_speed`로 감속

## 설정 파일

`configs/zerodriver.yaml` — 핀맵, PID 게인, 임계값, 모드.

## 빌드 타겟

```bash
# 개발 (macOS/Linux amd64, mock)
go build -o zerodriver ./cmd/zerodriver

# Pi Zero W (ARMv6)
GOOS=linux GOARCH=arm GOARM=6 go build -o zerodriver-armv6 ./cmd/zerodriver
```

## 테스트 전략

| 레벨 | 대상 | 방법 |
|------|------|------|
| Unit | domain, control/pid | `go test` |
| Integration | fusion, state machine | mock HAL |
| E2E | 전체 루프 | mock mode 10초 실행 |
| Hardware | Pi 실기 | systemd + 로그 |
