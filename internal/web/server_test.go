package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/autopilothub/zerodriver/internal/config"
	"github.com/autopilothub/zerodriver/internal/domain"
)

func TestStoreUpdate(t *testing.T) {
	store := NewStore()
	started := time.Now().Add(-10 * time.Second)
	store.Update(UpdateInput{
		Mode:  "hardware",
		State: domain.StateTracing,
		Fused: domain.FusedInput{
			LineOffset:      0.1,
			LookaheadOffset: 0.2,
			LineDetected:    true,
			HeadingError:    5,
			YawRate:         1.5,
			FrontDistance:   80,
		},
		Command:   domain.ControlCommand{Steering: 0.3, Throttle: 0.6},
		Attitude:  domain.Attitude{Roll: 1, Pitch: 2, Yaw: 90, Heading: 270, HasMag: true},
		StartedAt: started,
	})

	s := store.Status()
	if s.State != "TRACING" {
		t.Fatalf("state: %s", s.State)
	}
	if s.LookaheadOffset != 0.2 {
		t.Fatalf("lookahead: %f", s.LookaheadOffset)
	}
	if s.Heading != 270 {
		t.Fatalf("heading: %f", s.Heading)
	}
	if s.UptimeSec < 9 {
		t.Fatalf("uptime: %f", s.UptimeSec)
	}
}

func TestRGBToJPEG(t *testing.T) {
	rgb := make([]byte, 4*4*3)
	for i := 0; i < len(rgb); i += 3 {
		rgb[i], rgb[i+1], rgb[i+2] = 255, 0, 0
	}
	jpeg, err := RGBToJPEG(rgb, 4, 4, 80)
	if err != nil {
		t.Fatal(err)
	}
	if len(jpeg) < 10 {
		t.Fatal("expected jpeg data")
	}
}

func TestAPIStatus(t *testing.T) {
	store := NewStore()
	store.Update(UpdateInput{Mode: "mock", State: domain.StateIdle})
	srv := NewServer(&config.WebConfig{Addr: ":0", RefreshMS: 100}, store)

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	srv.handleStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code %d", rec.Code)
	}
	var out Status
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Mode != "mock" {
		t.Fatalf("mode: %s", out.Mode)
	}
}

func TestDashboardIndex(t *testing.T) {
	store := NewStore()
	srv := NewServer(&config.WebConfig{Addr: ":0"}, store)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code %d", rec.Code)
	}
	if len(rec.Body.Bytes()) < 100 {
		t.Fatal("expected html body")
	}
}
