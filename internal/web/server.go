package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/autopilothub/zerodriver/internal/config"
)

//go:embed static/*
var staticFS embed.FS

// Server serves the live status dashboard and JSON API.
type Server struct {
	store *Store
	cfg   *config.WebConfig
	srv   *http.Server
}

func NewServer(cfg *config.WebConfig, store *Store) *Server {
	s := &Server{store: store, cfg: cfg}
	mux := http.NewServeMux()

	static, _ := fs.Sub(staticFS, "static")
	mux.Handle("/", http.FileServer(http.FS(static)))
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/events", s.handleEvents)
	mux.HandleFunc("/api/camera.jpg", s.handleCamera)
	mux.HandleFunc("/api/mode", s.handleMode)
	mux.HandleFunc("/api/drive", s.handleDrive)
	mux.HandleFunc("/api/stop", s.handleStop)

	s.srv = &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

func (s *Server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, s.store.Status())
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	interval := time.Duration(s.cfg.RefreshMS) * time.Millisecond
	if interval <= 0 {
		interval = 200 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			data, err := json.Marshal(s.store.Status())
			if err != nil {
				return
			}
			fmt.Fprintf(w, "event: status\ndata: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (s *Server) handleCamera(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rgb, width, height := s.store.Frame()
	if len(rgb) == 0 {
		http.Error(w, "no frame", http.StatusNotFound)
		return
	}

	jpeg, err := RGBToJPEG(rgb, width, height, s.cfg.CameraQuality)
	if err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(jpeg)
}

func (s *Server) handleMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		DriveMode string `json:"drive_mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	s.store.SetDriveMode(DriveMode(body.DriveMode))
	writeJSON(w, map[string]string{"drive_mode": string(s.store.DriveMode())})
}

func (s *Server) handleDrive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Steering float64 `json:"steering"`
		Throttle float64 `json:"throttle"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := s.store.SetManualDrive(body.Steering, body.Throttle); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	writeJSON(w, map[string]any{
		"steering": body.Steering,
		"throttle": body.Throttle,
	})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.store.StopManual()
	writeJSON(w, map[string]string{"status": "stopped"})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
