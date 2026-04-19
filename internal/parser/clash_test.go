package parser

import "testing"

func TestParseClashProxy(t *testing.T) {
	t.Parallel()

	proxy, err := ParseClashProxy(`- { name: 'Example Clash Node', type: vmess, server: demo.example.com, port: 443 }`)
	if err != nil {
		t.Fatalf("ParseClashProxy() error = %v", err)
	}
	if got := ClashProxyName(proxy); got != "Example Clash Node" {
		t.Fatalf("ClashProxyName() = %q", got)
	}
}

func TestParseClashProxyList(t *testing.T) {
	t.Parallel()

	raw := []byte("proxies:\n  - { name: hk, type: vmess, server: example.com, port: 443 }\n  - { name: jp, type: trojan, server: example.net, port: 443 }\n")
	proxies, err := ParseClashProxyList(raw)
	if err != nil {
		t.Fatalf("ParseClashProxyList() error = %v", err)
	}
	if len(proxies) != 2 {
		t.Fatalf("len(proxies) = %d, want 2", len(proxies))
	}
}

func TestPrefixClashProxy(t *testing.T) {
	t.Parallel()

	proxy, err := ParseClashProxy(`{ name: hk, type: vmess, server: example.com, port: 443 }`)
	if err != nil {
		t.Fatalf("ParseClashProxy() error = %v", err)
	}
	renamed := PrefixClashProxy(proxy, "[Manual] ")
	if got := ClashProxyName(renamed); got != "[Manual] hk" {
		t.Fatalf("ClashProxyName() = %q", got)
	}
}
