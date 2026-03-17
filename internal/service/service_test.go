package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
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
						Prefix:          "[Remote] ",
						RemoteType:      config.RemoteTypeBase64Lines,
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
	if len(result.Lines) != 5 {
		t.Fatalf("len(result.Lines) = %d, want 5", len(result.Lines))
	}
	if result.Lines[0] != "happ://routing/onadd/abc" {
		t.Fatalf("routing line = %q", result.Lines[0])
	}
	if result.Lines[1] != "#profile-update-interval: 1" {
		t.Fatalf("profile-update-interval line = %q", result.Lines[1])
	}
	if result.Lines[2] != "#profile-title: UniSub" {
		t.Fatalf("profile-title line = %q", result.Lines[2])
	}
	if result.Lines[3] != "#ping-type proxy" {
		t.Fatalf("ping-type line = %q", result.Lines[3])
	}
	if got := nodeparser.ParseNode(result.Lines[4]).DisplayName; got != "[Remote] 香港-1" {
		t.Fatalf("prefixed name = %q", got)
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("hits after first render = %d, want 1", got)
	}

	result, err = svc.RenderSubscription(context.Background(), "123e4567-e89b-42d3-a456-426614174000", "V2rayN", false)
	if err != nil {
		t.Fatalf("RenderSubscription() cached error = %v", err)
	}
	if len(result.Lines) != 1 {
		t.Fatalf("len(result.Lines) = %d, want 1", len(result.Lines))
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
	if result.Lines[0] == "" {
		t.Fatal("expected subscription line after force refresh")
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
