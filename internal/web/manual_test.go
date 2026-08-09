package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/autopilothub/zerodriver/internal/config"
)

func TestManualModeAPI(t *testing.T) {
	store := NewStore()
	srv := NewServer(&config.WebConfig{Addr: ":0", ManualTimeoutMS: 500}, store)

	// switch to manual
	rec := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"drive_mode": "manual"})
	req := httptest.NewRequest(http.MethodPost, "/api/mode", bytes.NewReader(body))
	srv.handleMode(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("mode: %d %s", rec.Code, rec.Body.String())
	}
	if store.DriveMode() != DriveManual {
		t.Fatalf("expected manual, got %s", store.DriveMode())
	}

	// drive command
	rec = httptest.NewRecorder()
	body, _ = json.Marshal(map[string]float64{"steering": 0.5, "throttle": 0.4})
	req = httptest.NewRequest(http.MethodPost, "/api/drive", bytes.NewReader(body))
	srv.handleDrive(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("drive: %d %s", rec.Code, rec.Body.String())
	}

	active, cmd := store.ManualCommand(500 * time.Millisecond)
	if !active || cmd.Steering != 0.5 || cmd.Throttle != 0.4 {
		t.Fatalf("cmd: active=%v %+v", active, cmd)
	}

	// auto mode rejects drive
	store.SetDriveMode(DriveAuto)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/drive", bytes.NewReader(body))
	srv.handleDrive(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestManualWatchdog(t *testing.T) {
	store := NewStore()
	store.SetDriveMode(DriveManual)
	store.manualLastCmd = time.Now().Add(-2 * time.Second)
	store.manualSteer = 0.8
	store.manualThrottle = 0.6

	_, cmd := store.ManualCommand(500 * time.Millisecond)
	if cmd.Throttle != 0 || cmd.Steering != 0 {
		t.Fatalf("watchdog should zero: %+v", cmd)
	}
}
