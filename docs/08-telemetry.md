# Telemetry

주행 데이터를 외부로 전송하는 텔레메트리 패키지.

## Publisher 종류

| Publisher | 조건 | 설명 |
|-----------|------|------|
| `NoopPublisher` | `telemetry.enabled: false` | 데이터 무시 |
| `LogPublisher` | `enabled: true`, endpoint 없음 | JSON 로그 출력 |
| `IoTPublisher` | `enabled: true`, endpoint 설정 | AWS IoT Core MQTT |

## 설정

```yaml
telemetry:
  enabled: true
  endpoint: xxxxx-ats.iot.ap-northeast-2.amazonaws.com
  topic: zerodriver/telemetry
  interval_sec: 1
  cert_dir: /etc/zerodriver/certs
```

## 인증서 (AWS IoT)

`/etc/zerodriver/certs/` 디렉토리에 배치:

```
device.pem.crt      # 디바이스 인증서
private.pem.key     # 개인키
AmazonRootCA1.pem   # AWS Root CA
```

AWS IoT Console → Things → Create → Create certificate

## JSON 페이로드

```json
{
  "state": "TRACING",
  "line_offset": 0.12,
  "line_detected": true,
  "steering": 0.10,
  "throttle": 0.60,
  "front_distance_cm": 999,
  "yaw": 2.5,
  "timestamp": "2026-08-02T20:30:01Z"
}
```

## CloudWatch 연동

AWS IoT Core Rule로 MQTT 메시지를 CloudWatch Logs 또는 Timestream에 저장:

```sql
SELECT * FROM 'zerodriver/telemetry'
```

## 로컬 테스트

```bash
# 로그 출력 모드
# configs/zerodriver.yaml: enabled: true, endpoint: ""
go run ./cmd/zerodriver -mode mock -v -duration 5s
```
