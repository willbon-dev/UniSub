package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	t.Parallel()

	cfgText := `
server:
  listen: 127.0.0.1:18080
subscriptions:
  - name: demo
    secret: "123e4567-e89b-42d3-a456-426614174000"
    default_platform: V2rayN
    sources:
      - name: manual-1
        type: manual
        entries:
          - vmess://abc
      - name: remote-1
        type: remote
        remote_type: base64_lines
        url: https://example.com/sub
        refresh_interval: 10m
        include_patterns:
          - 香港
        exclude_patterns:
          - 剩余流量
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(cfgText), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := cfg.Subscriptions[0].Sources[1].RemoteType; got != RemoteTypeBase64Lines {
		t.Fatalf("remote_type = %q, want %q", got, RemoteTypeBase64Lines)
	}
}

func TestLoadRejectsUnknownRemoteType(t *testing.T) {
	t.Parallel()

	cfgText := `
subscriptions:
  - name: demo
    secret: "123e4567-e89b-42d3-a456-426614174000"
    sources:
      - name: remote-1
        type: remote
        remote_type: mystery
        url: https://example.com/sub
        refresh_interval: 10m
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(cfgText), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}
