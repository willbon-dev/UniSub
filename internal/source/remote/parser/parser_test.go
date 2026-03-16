package parser

import "testing"

func TestBase64LinesDecoder(t *testing.T) {
	t.Parallel()

	decoder := Base64LinesDecoder{}
	lines, err := decoder.DecodeResponse([]byte("dm1lc3M6Ly9hYmMKdm1lc3M6Ly9kZWY="))
	if err != nil {
		t.Fatalf("DecodeResponse() error = %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("len(lines) = %d, want 2", len(lines))
	}
	if lines[0] != "vmess://abc" {
		t.Fatalf("unexpected line %q", lines[0])
	}
}
