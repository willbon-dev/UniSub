package config

import (
	"os"
	"path/filepath"
	"reflect"
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
    default_platform: Clash
    platform_options:
      happ:
        routing: happ://routing/onadd/abc
        profile_update_interval: 1
        profile_title: UniSub
        ping_type: proxy
      clash:
        template: ./Self.ini
    sources:
      - name: manual-1
        type: manual
        style: clash_proxy
        platforms: [Clash]
        entries:
          - "{ name: hk, type: vmess, server: example.com, port: 443 }"
      - name: remote-1
        type: remote
        remote_type: base64_lines
        platforms: [V2rayN, Happ]
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
	if got := cfg.Subscriptions[0].Sources[1].Style; got != StyleLinkLinesBase64 {
		t.Fatalf("style = %q, want %q", got, StyleLinkLinesBase64)
	}
	if got := cfg.Subscriptions[0].PlatformOptions.Happ.ProfileUpdateInterval; got == nil || *got != 1 {
		t.Fatalf("profile_update_interval = %#v", got)
	}
	if got := cfg.Subscriptions[0].PlatformOptions.Happ.ProfileTitle; got == nil || *got != "UniSub" {
		t.Fatalf("profile_title = %#v", got)
	}
	if got := cfg.Subscriptions[0].PlatformOptions.Happ.PingType; got == nil || *got != "proxy" {
		t.Fatalf("ping_type = %#v", got)
	}
	if got := cfg.Subscriptions[0].PlatformOptions.Clash.Template; got != "./Self.ini" {
		t.Fatalf("clash template = %q", got)
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
        style: mystery
        platforms: [V2rayN]
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

func TestLoadRejectsMissingPlatforms(t *testing.T) {
	t.Parallel()

	cfgText := `
subscriptions:
  - name: demo
    secret: "123e4567-e89b-42d3-a456-426614174000"
    sources:
      - name: manual-1
        type: manual
        style: link_line
        entries:
          - vmess://abc
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(cfgText), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadSupportsClashProxyObjectEntries(t *testing.T) {
	t.Parallel()

	cfgText := `
subscriptions:
  - name: demo
    secret: "123e4567-e89b-42d3-a456-426614174000"
    default_platform: Clash
    platform_options:
      clash:
        template: ./Self.ini
    sources:
      - name: manual-clash
        type: manual
        style: clash_proxy
        platforms: [Clash]
        entries:
          - { name: 'Example Clash Node', type: vmess, server: demo.example.com, port: 443, uuid: 123e4567-e89b-42d3-a456-426614174001, alterId: 0, cipher: auto, udp: true, network: ws, ws-opts: { path: /websocket, headers: { Host: demo-edge.example.com } } }
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(cfgText), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Subscriptions[0].Sources[0].Entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(cfg.Subscriptions[0].Sources[0].Entries))
	}
	if got := cfg.Subscriptions[0].Sources[0].Entries[0]; got == "" || got == "{}" {
		t.Fatalf("entry = %q, want non-empty rendered yaml", got)
	}
}

func TestHappOptionsRenderSubscriptionLines(t *testing.T) {
	t.Parallel()

	got := HappOptions{
		Routing:               "happ://routing/onadd/abc",
		ProfileUpdateInterval: intPtr(1),
		ProfileTitle:          strPtr("UniSub"),
		ProviderID:            strPtr("provider-id-demo"),
		PingType:              strPtr("proxy"),
		SubInfoText:           strPtr(""),
	}.RenderSubscriptionLines()

	want := []string{
		"happ://routing/onadd/abc",
		"#profile-update-interval: 1",
		"#profile-title: UniSub",
		"#providerid provider-id-demo",
		"#ping-type proxy",
		"#sub-info-text:",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RenderSubscriptionLines() = %#v, want %#v", got, want)
	}
}

func strPtr(v string) *string {
	return &v
}

func intPtr(v int) *int {
	return &v
}
