package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/autopilothub/zerodriver/internal/config"
	"github.com/autopilothub/zerodriver/internal/domain"
)

func TestCalibrationAPI(t *testing.T) {
	store := NewStore()
	store.SetCalibrationHooks(CalibrationHooks{
		Drive: func(steering, throttle float64) error { return nil },
		ReadIMU: func() (domain.Attitude, error) {
			return domain.Attitude{GyroZ: -5, HasMag: true, Heading: 90, HasAccel: true, AccelX: 0.5}, nil
		},
		Stop: func() error { return nil },
	})
	srv := NewServer(&config.WebConfig{Addr: ":0"}, store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/calibrate", bytes.NewReader([]byte(`{"confirm":true,"type":"imu"}`)))
	srv.handleCalibrate(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("start: %d %s", rec.Code, rec.Body.String())
	}

	deadline := time.Now().Add(12 * time.Second)
	for time.Now().Before(deadline) {
		state, _, _, _, _ := store.calibrationSnapshot()
		if state == CalibrationDone || state == CalibrationError {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("calibration did not finish")
}

func TestDeadzoneCalibrationStart(t *testing.T) {
	store := NewStore()
	store.SetCalibrationHooks(CalibrationHooks{
		SetThrottleUS: func(us int) error { return nil },
		ReadIMU: func() (domain.Attitude, error) {
			return domain.Attitude{HasAccel: true, AccelX: 0.5}, nil
		},
		Stop:      func() error { return nil },
		NeutralUs: 1500, MaxUs: 2000, ReverseUs: 1000,
	})

	if err := store.StartCalibration(CalibrationDeadzone); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Millisecond)
	if !store.IsCalibrating() {
		t.Fatal("expected running")
	}
}

func TestCalibrationBlocksManualDrive(t *testing.T) {
	store := NewStore()
	store.calState = CalibrationRunning
	srv := NewServer(&config.WebConfig{Addr: ":0", ManualTimeoutMS: 500}, store)
	store.SetDriveMode(DriveManual)

	rec := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]float64{"steering": 0.5, "throttle": 0.4})
	req := httptest.NewRequest(http.MethodPost, "/api/drive", bytes.NewReader(body))
	srv.handleDrive(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}
