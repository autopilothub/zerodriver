package telemetry

import (
	"log"
)

// LogPublisher writes telemetry as JSON log lines.
type LogPublisher struct{}

func NewLogPublisher() *LogPublisher {
	return &LogPublisher{}
}

func (p *LogPublisher) Publish(s Snapshot) error {
	data, err := s.JSON()
	if err != nil {
		return err
	}
	log.Printf("telemetry: %s", data)
	return nil
}

func (p *LogPublisher) Close() error { return nil }

// NoopPublisher discards all telemetry.
type NoopPublisher struct{}

func NewNoopPublisher() *NoopPublisher {
	return &NoopPublisher{}
}

func (p *NoopPublisher) Publish(Snapshot) error { return nil }
func (p *NoopPublisher) Close() error           { return nil }
