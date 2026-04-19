package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/willbon-dev/UniSub/internal/config"
	nodeparser "github.com/willbon-dev/UniSub/internal/parser"
)

func TestRenderSubscriptionHappAndRefresh(t *testing.T) {
	t.Parallel()

	var hits atomic.Int32
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		count := hits.Add(1)
		body := samplePayload(fmt.Sprintf("香港-%d", count), "剩余流量：10GB")
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})

	cfg := &config.Config{
		Server: config.ServerConfig{
			FetchTimeout:     5 * time.Second,
			MaxResponseBytes: 1 << 20,
		},
		Subscriptions: []config.SubscriptionConfig{
			{
				Name:            "demo",
				Secret:          "123e4567-e89b-42d3-a456-426614174000",
				DefaultPlatform: config.PlatformHapp,
				PlatformOptions: config.PlatformOptions{
					Happ: config.HappOptions{
						Routing:               "happ://routing/onadd/abc",
						ProfileUpdateInterval: intPtr(1),
						ProfileTitle:          strPtr("UniSub"),
						PingType:              strPtr("proxy"),
					},
				},
				Sources: []config.SourceConfig{
					{
						Name:            "remote-1",
						Type:            config.SourceTypeRemote,
						Platforms:       []string{config.PlatformV2RayN, config.PlatformHapp},
						Style:           config.StyleLinkLinesBase64,
						Prefix:          "[Remote] ",
						URL:             "https://example.com/sub",
						RefreshInterval: time.Hour,
						ExcludePatterns: []string{"剩余流量"},
					},
				},
			},
		},
	}

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	svc.httpClient.Transport = transport

	result, err := svc.RenderSubscription(context.Background(), "123e4567-e89b-42d3-a456-426614174000", "", false)
	if err != nil {
		t.Fatalf("RenderSubscription() error = %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(result.Body)), "\n")
	if len(lines) != 5 {
		t.Fatalf("len(lines) = %d, want 5", len(lines))
	}
	if result.ContentType != "text/plain; charset=utf-8" {
		t.Fatalf("ContentType = %q", result.ContentType)
	}
	if lines[0] != "happ://routing/onadd/abc" {
		t.Fatalf("routing line = %q", lines[0])
	}
	if got := nodeparser.ParseNode(lines[4]).DisplayName; got != "[Remote] 香港-1" {
		t.Fatalf("prefixed name = %q", got)
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("hits after first render = %d, want 1", got)
	}

	result, err = svc.RenderSubscription(context.Background(), "123e4567-e89b-42d3-a456-426614174000", "V2rayN", false)
	if err != nil {
		t.Fatalf("RenderSubscription() cached error = %v", err)
	}
	lines = strings.Split(strings.TrimSpace(string(result.Body)), "\n")
	if len(lines) != 1 {
		t.Fatalf("len(lines) = %d, want 1", len(lines))
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("hits after cached render = %d, want 1", got)
	}

	result, err = svc.RenderSubscription(context.Background(), "123e4567-e89b-42d3-a456-426614174000", "V2rayN", true)
	if err != nil {
		t.Fatalf("RenderSubscription() force refresh error = %v", err)
	}
	if got := hits.Load(); got != 2 {
		t.Fatalf("hits after force refresh = %d, want 2", got)
	}
	if strings.TrimSpace(string(result.Body)) == "" {
		t.Fatal("expected subscription body after force refresh")
	}
}

func TestRenderSubscriptionClash(t *testing.T) {
	t.Parallel()

	templateDir := t.TempDir()
	templatePath := filepath.Join(templateDir, "Self.ini")
	templateBody := strings.Join([]string{
		"[custom]",
		"custom_proxy_group=节点选择`select`[]DIRECT`[]REJECT`.*",
		"custom_proxy_group=自动选择`url-test`.*`https://www.gstatic.com/generate_204`300``50",
		"ruleset=节点选择,https://example.com/rules/list.txt",
		"ruleset=DIRECT,[]GEOIP,CN",
		"ruleset=节点选择,[]FINAL",
	}, "\n")
	if err := os.WriteFile(templatePath, []byte(templateBody), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	cfg := &config.Config{
		Server: config.ServerConfig{
			FetchTimeout:     5 * time.Second,
			MaxResponseBytes: 1 << 20,
		},
		Subscriptions: []config.SubscriptionConfig{
			{
				Name:            "clash-demo",
				Secret:          "123e4567-e89b-42d3-a456-426614174001",
				DefaultPlatform: config.PlatformClash,
				PlatformOptions: config.PlatformOptions{
					Clash: config.ClashOptions{Template: templatePath},
				},
				Sources: []config.SourceConfig{
					{
						Name:      "manual-clash",
						Type:      config.SourceTypeManual,
						Platforms: []string{config.PlatformClash},
						Style:     config.StyleClashProxy,
						Prefix:    "[Manual] ",
						Entries: []string{
							"{ name: hk, type: vmess, server: example.com, port: 443 }",
						},
					},
					{
						Name:            "remote-clash",
						Type:            config.SourceTypeRemote,
						Platforms:       []string{config.PlatformClash},
						Style:           config.StyleClashProxiesYML,
						URL:             "https://example.com/clash.yaml",
						Prefix:          "[Remote] ",
						RefreshInterval: time.Hour,
					},
				},
			},
		},
	}

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	svc.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := "proxies:\n  - { name: jp, type: trojan, server: example.net, port: 443 }\n"
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})

	result, err := svc.RenderSubscription(context.Background(), "123e4567-e89b-42d3-a456-426614174001", "Clash", false)
	if err != nil {
		t.Fatalf("RenderSubscription() error = %v", err)
	}
	body := string(result.Body)
	if result.ContentType != "application/yaml; charset=utf-8" {
		t.Fatalf("ContentType = %q", result.ContentType)
	}
	for _, want := range []string{"proxies:", "proxy-groups:", "rules:", "rule-providers:", "[Manual] hk", "[Remote] jp", "MATCH,节点选择", "RULE-SET,provider-1,节点选择", "format: text", "path: ./ruleset/provider-1.txt"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q:\n%s", want, body)
		}
	}
}

func TestRemoteSourceRequestHeaders(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Server: config.ServerConfig{
			FetchTimeout:     5 * time.Second,
			MaxResponseBytes: 1 << 20,
		},
		Subscriptions: []config.SubscriptionConfig{
			{
				Name:            "headers-demo",
				Secret:          "123e4567-e89b-42d3-a456-426614174099",
				DefaultPlatform: config.PlatformClash,
				PlatformOptions: config.PlatformOptions{
					Clash: config.ClashOptions{Template: "unused"},
				},
				Sources: []config.SourceConfig{
					{
						Name:            "remote-clash",
						Type:            config.SourceTypeRemote,
						Platforms:       []string{config.PlatformClash},
						Style:           config.StyleClashProxiesYML,
						URL:             "https://example.com/clash.yaml",
						RequestHeaders:  map[string]string{"User-Agent": "clash-verge"},
						RefreshInterval: time.Hour,
					},
				},
			},
		},
	}

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	svc.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.Header.Get("User-Agent"); got != "clash-verge" {
			t.Fatalf("User-Agent = %q, want %q", got, "clash-verge")
		}
		body := "proxies:\n  - { name: jp, type: trojan, server: example.net, port: 443 }\n"
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})

	entries, err := svc.fetchRemote(context.Background(), &svc.subscriptions["123e4567-e89b-42d3-a456-426614174099"].sources[0], false)
	if err != nil {
		t.Fatalf("fetchRemote() error = %v", err)
	}
	if len(entries.ClashProxies) != 1 {
		t.Fatalf("len(entries.ClashProxies) = %d, want 1", len(entries.ClashProxies))
	}
}

func samplePayload(names ...string) string {
	lines := make([]string, 0, len(names))
	for _, name := range names {
		lines = append(lines, vmessLine(name))
	}
	return base64.StdEncoding.EncodeToString([]byte(stringsJoin(lines, "\n")))
}

func vmessLine(name string) string {
	payload := fmt.Sprintf(`{"v":"2","ps":"%s","add":"example.com","port":"443","id":"123e4567-e89b-42d3-a456-426614174000","aid":"0","net":"ws","type":"none","host":"example.com","path":"/","tls":"tls"}`, name)
	return "vmess://" + base64.StdEncoding.EncodeToString([]byte(payload))
}

func stringsJoin(items []string, sep string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	}
	result := items[0]
	for _, item := range items[1:] {
		result += sep + item
	}
	return result
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func strPtr(v string) *string {
	return &v
}

func intPtr(v int) *int {
	return &v
}
