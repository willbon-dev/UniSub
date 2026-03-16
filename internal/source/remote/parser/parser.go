package parser

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type Decoder interface {
	DecodeResponse(raw []byte) ([]string, error)
}

func New(name string) (Decoder, error) {
	switch name {
	case "base64_lines":
		return Base64LinesDecoder{}, nil
	default:
		return nil, fmt.Errorf("unsupported remote_type %q", name)
	}
}

type Base64LinesDecoder struct{}

func (Base64LinesDecoder) DecodeResponse(raw []byte) ([]string, error) {
	decoded, err := decodeBase64String(strings.TrimSpace(string(raw)))
	if err != nil {
		return nil, fmt.Errorf("decode base64_lines response: %w", err)
	}

	lines := make([]string, 0)
	for _, line := range strings.Split(strings.ReplaceAll(string(decoded), "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("decoded response contained no entries")
	}
	return lines, nil
}

func decodeBase64String(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty base64 payload")
	}

	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	for _, enc := range encodings {
		decoded, err := enc.DecodeString(s)
		if err == nil {
			return decoded, nil
		}
	}
	return nil, fmt.Errorf("unsupported base64 payload")
}

func DecodeBase64String(s string) ([]byte, error) {
	return decodeBase64String(s)
}

func EncodeBase64String(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func EncodeBase64StringRaw(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
