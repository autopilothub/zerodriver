# 요구사항 명세

## 기능 요구사항 (User Stories)

### US-01: 라인 추종
- **As a** 라인 트레이서
- **I want** 검은 라인 위에서 자율 주행
- **So that** 대회/시연에서 탈선 없이 완주할 수 있다

**인수 조건**
- [ ] 직선 구간에서 라인 중심 오차 ±3cm 이내
- [ ] 기본 속도 30cm/s 이상 유지
- [ ] 라인 검출 주기 30Hz 이상

### US-02: 코너 주행
- **As a** 라인 트레이서
- **I want** 90° 코너에서 IMU 보정으로 안정적으로 통과
- **So that** 급커브에서 탈선하지 않는다

**인수 조건**
- [ ] 90° 코너 탈선률 5% 미만 (10회 시험)
- [ ] 코너 진입 시 자동 감속
- [ ] IMU yaw 각도로 steering 보정

### US-03: 장애물 회피
- **As a** 라인 트레이서
- **I want** 전방 장애물 감지 시 정지 또는 우회
- **So that** 충돌 없이 안전하게 주행한다

**인수 조건**
- [ ] 전방 20cm 이내 장애물 감지 시 500ms 이내 정지
- [ ] 장애물 제거 후 자동 재개 (TRACING 상태 복귀)
- [ ] LiDAR 스캔 주기 5Hz 이상

### US-04: 텔레메트리 (선택)
- **As a** 운영자
- **I want** AWS IoT로 주행 데이터 전송
- **So that** 원격 모니터링 및 튜닝이 가능하다

**인수 조건**
- [ ] 1Hz 이상 텔레메트리 전송
- [ ] line offset, speed, state, obstacle distance 포함

## 비기능 요구사항 (NFR)

| ID | 항목 | 목표 |
|----|------|------|
| NFR-01 | 제어 루프 지연 | < 50ms (센서 읽기 → 모터 출력) |
| NFR-02 | CPU 사용률 | Pi Zero W 기준 < 80% |
| NFR-03 | 메모리 | < 256MB RSS |
| NFR-04 | 부팅 후 자동 시작 | systemd, 30초 이내 주행 준비 |
| NFR-05 | 설정 변경 | YAML 재시작 없이 PID 게인 핫리로드 (v2) |

## 도메인 모델

```
LineTracer (Aggregate Root)
├── PerceptionSnapshot
│   ├── LinePosition      // -1.0 ~ +1.0 (라인 중심 오프셋)
│   ├── ObstacleScan      // 전방 거리 배열 (cm)
│   └── VehicleAttitude   // yaw, pitch, roll (degrees)
├── ControlCommand
│   ├── Steering          // -1.0 ~ +1.0 (좌 ~ 우)
│   └── Throttle          // 0.0 ~ 1.0
└── RaceState             // IDLE | TRACING | AVOIDING | STOPPED
```

## 상태 전이

```
        ┌──────┐
        │ IDLE │
        └──┬───┘
           │ start
           ▼
      ┌─────────┐  obstacle detected   ┌──────────┐
      │ TRACING │ ──────────────────► │ AVOIDING │
      └────┬────┘                      └────┬─────┘
           │                                │ cleared
           │ stop                           │
           ▼                                │
      ┌─────────┐ ◄────────────────────────┘
      │ STOPPED │
      └─────────┘
```
