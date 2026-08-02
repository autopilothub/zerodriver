package telemetry

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/autopilothub/zerodriver/internal/config"
)

// IoTPublisher publishes telemetry to AWS IoT Core via MQTT.
type IoTPublisher struct {
	client mqtt.Client
	topic  string
}

// NewPublisher creates a telemetry publisher based on config.
func NewPublisher(cfg *config.TelemetryConfig) (Publisher, error) {
	if !cfg.Enabled {
		return NewNoopPublisher(), nil
	}
	if cfg.Endpoint == "" {
		return NewLogPublisher(), nil
	}
	return newIoTPublisher(cfg)
}

func newIoTPublisher(cfg *config.TelemetryConfig) (*IoTPublisher, error) {
	certDir := cfg.CertDir
	if certDir == "" {
		certDir = "/etc/zerodriver/certs"
	}

	cert, err := tls.LoadX509KeyPair(
		filepath.Join(certDir, "device.pem.crt"),
		filepath.Join(certDir, "private.pem.key"),
	)
	if err != nil {
		return nil, fmt.Errorf("load device cert: %w", err)
	}

	caCert, err := os.ReadFile(filepath.Join(certDir, "AmazonRootCA1.pem"))
	if err != nil {
		return nil, fmt.Errorf("load CA cert: %w", err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}

	broker := fmt.Sprintf("ssl://%s:8883", cfg.Endpoint)
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID("zerodriver").
		SetTLSConfig(tlsCfg).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		return nil, fmt.Errorf("mqtt connect timeout")
	}
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("mqtt connect: %w", err)
	}

	return &IoTPublisher{client: client, topic: cfg.Topic}, nil
}

func (p *IoTPublisher) Publish(s Snapshot) error {
	data, err := s.JSON()
	if err != nil {
		return err
	}
	token := p.client.Publish(p.topic, 0, false, data)
	token.Wait()
	return token.Error()
}

func (p *IoTPublisher) Close() error {
	if p.client.IsConnected() {
		p.client.Disconnect(250)
	}
	return nil
}
