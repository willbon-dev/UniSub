package parser

import (
	"encoding/base64"
	"testing"
)

func TestParseNodeVMessChineseName(t *testing.T) {
	t.Parallel()

	raw := "vmess://eyJ2IjoiMiIsInBzIjoi6aaZ5rivLea1i+ivlSIsImFkZCI6ImV4YW1wbGUuY29tIiwicG9ydCI6IjQ0MyIsImlkIjoiMTIzZTQ1NjctZTg5Yi00MmQzLWE0NTYtNDI2NjE0MTc0MDAwIiwiYWlkIjoiMCIsIm5ldCI6IndzIiwidHlwZSI6Im5vbmUiLCJob3N0IjoiZXhhbXBsZS5jb20iLCJwYXRoIjoiLyIsInRscyI6InRscyJ9"
	node := ParseNode(raw)
	if node.DisplayName != "香港-测试" {
		t.Fatalf("DisplayName = %q, want %q", node.DisplayName, "香港-测试")
	}
}

func TestParseNodeVLESSFragmentName(t *testing.T) {
	t.Parallel()

	raw := "vless://uuid@example.com:443?security=reality&type=xhttp#%E6%97%A5%E6%9C%AC%28Vless%2CXhttp%2CReality%29"
	node := ParseNode(raw)
	if node.DisplayName != "日本(Vless,Xhttp,Reality)" {
		t.Fatalf("DisplayName = %q", node.DisplayName)
	}
}

func TestParseNodeTrojanFragmentName(t *testing.T) {
	t.Parallel()

	raw := "trojan://password@example.com:443?sni=example.com#Singapore-Trojan"
	node := ParseNode(raw)
	if node.DisplayName != "Singapore-Trojan" {
		t.Fatalf("DisplayName = %q", node.DisplayName)
	}
}

func TestParseNodeSSFragmentName(t *testing.T) {
	t.Parallel()

	raw := "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo4NDQz#%E9%A6%99%E6%B8%AF-SS"
	node := ParseNode(raw)
	if node.DisplayName != "香港-SS" {
		t.Fatalf("DisplayName = %q", node.DisplayName)
	}
}

func TestParseNodeSSRRemarks(t *testing.T) {
	t.Parallel()

	remarks := base64.RawURLEncoding.EncodeToString([]byte("台湾-SSR"))
	payload := "example.com:443:origin:aes-256-cfb:plain:" + base64.RawURLEncoding.EncodeToString([]byte("password")) + "/?remarks=" + remarks
	raw := "ssr://" + base64.RawURLEncoding.EncodeToString([]byte(payload))
	node := ParseNode(raw)
	if node.DisplayName != "台湾-SSR" {
		t.Fatalf("DisplayName = %q", node.DisplayName)
	}
}

func TestParseNodeFallsBackToHost(t *testing.T) {
	t.Parallel()

	raw := "tuic://uuid:password@node.example.com:443?congestion_control=bbr"
	node := ParseNode(raw)
	if node.DisplayName != "node.example.com:443" {
		t.Fatalf("DisplayName = %q", node.DisplayName)
	}
}

func TestPrefixNodeVLESS(t *testing.T) {
	t.Parallel()

	raw := "vless://uuid@example.com:443?security=reality#%E6%97%A5%E6%9C%AC"
	renamed := PrefixNode(raw, "[JP] ")
	node := ParseNode(renamed)
	if node.DisplayName != "[JP] 日本" {
		t.Fatalf("DisplayName = %q", node.DisplayName)
	}
}

func TestPrefixNodeVMess(t *testing.T) {
	t.Parallel()

	raw := "vmess://eyJ2IjoiMiIsInBzIjoi6aaZ5rivLea1i+ivlSIsImFkZCI6ImV4YW1wbGUuY29tIiwicG9ydCI6IjQ0MyIsImlkIjoiMTIzZTQ1NjctZTg5Yi00MmQzLWE0NTYtNDI2NjE0MTc0MDAwIiwiYWlkIjoiMCIsIm5ldCI6IndzIiwidHlwZSI6Im5vbmUiLCJob3N0IjoiZXhhbXBsZS5jb20iLCJwYXRoIjoiLyIsInRscyI6InRscyJ9"
	renamed := PrefixNode(raw, "[HK] ")
	node := ParseNode(renamed)
	if node.DisplayName != "[HK] 香港-测试" {
		t.Fatalf("DisplayName = %q", node.DisplayName)
	}
}
