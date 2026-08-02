# ZeroDriver — 라인 트레이서 RC카 개요

## 프로젝트 목적

Raspberry Pi Zero W와 MPU-9250, 카메라, LiDAR를 활용하여 자율 라인 트레이싱 RC카를 구축한다.
소프트웨어는 Go로 작성하며, AWS AI-DLC(AI-Driven Development Life Cycle) 방법론으로 개발한다.

## 시스템 구성

```
┌─────────────────────────────────────────────────────────┐
│                   Raspberry Pi Zero W                   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────┐  │
│  │ Camera   │ │ MPU-9250 │ │  LiDAR   │ │  Motor    │  │
│  │ (CSI)    │ │ (I2C)    │ │ (UART)   │ │ (GPIO/PWM)│  │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └─────┬─────┘  │
│       │            │            │              │        │
│  ┌────▼────────────▼────────────▼──────────────▼─────┐  │
│  │              Sensor Fusion Layer (Go)             │  │
│  └──────────────────────┬────────────────────────────┘  │
│                         │                               │
│  ┌──────────────────────▼────────────────────────────┐  │
│  │         Control Loop (PID / State Machine)        │  │
│  └──────────────────────┬────────────────────────────┘  │
│                         │                               │
│  ┌──────────────────────▼────────────────────────────┐  │
│  │    AWS IoT Core / CloudWatch (선택, Operations)   │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## 센서 역할

| 센서 | 역할 | 우선순위 |
|------|------|----------|
| 카메라 (CSI) | 라인 위치 검출 (주 센서) | 1 |
| MPU-9250 | 자세/각속도, 코너 보정, dead reckoning | 2 |
| LiDAR (UART) | 장애물 회피, 트랙 경계 보조 | 3 |

## 기술 스택

- **언어**: Go 1.22+
- **HAL**: periph.io/x/host/v3 (GPIO, I2C, UART)
- **비전**: gocv (OpenCV 바인딩) — Pi 배포 시
- **설정**: YAML
- **클라우드**: AWS IoT Core (선택)
- **방법론**: AWS AI-DLC

## 디렉토리 구조

```
zerodriver/
├── cmd/zerodriver/       # 진입점
├── internal/
│   ├── domain/           # DDD 엔티티, 값 객체
│   ├── hal/              # 하드웨어 추상화
│   ├── perception/       # 센서 데이터 처리
│   ├── fusion/           # 센서 융합
│   ├── control/          # PID, 상태 머신, 모터
│   └── telemetry/        # AWS IoT (선택)
├── configs/              # 런타임 설정
├── deploy/               # systemd, 빌드 스크립트
└── docs/                 # 개발 문서
```

## 개발 모드

| 모드 | 설명 | 사용 시점 |
|------|------|-----------|
| `mock` | 센서/모터 시뮬레이션 | Mac/PC 개발, 단위 테스트 |
| `hardware` | 실제 Pi 하드웨어 | Pi Zero W 배포 |

```bash
# Mock 모드 (개발 PC)
go run ./cmd/zerodriver -config configs/zerodriver.yaml -mode mock

# 하드웨어 모드 (Pi)
./zerodriver -config /etc/zerodriver/zerodriver.yaml -mode hardware
```
