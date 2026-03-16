package parser

import (
	"encoding/json"
	"net/url"
	"strings"
	"unicode/utf8"

	remoteparser "github.com/willbon-dev/UniSub/internal/source/remote/parser"
)

type Node struct {
	Raw         string
	DisplayName string
}

type vmessNode struct {
	V    string `json:"v,omitempty"`
	PS   string `json:"ps,omitempty"`
	Add  string `json:"add,omitempty"`
	Port string `json:"port,omitempty"`
	ID   string `json:"id,omitempty"`
	Aid  string `json:"aid,omitempty"`
	Net  string `json:"net,omitempty"`
	Type string `json:"type,omitempty"`
	Host string `json:"host,omitempty"`
	Path string `json:"path,omitempty"`
	TLS  string `json:"tls,omitempty"`
}

func ParseNode(raw string) Node {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Node{}
	}

	if strings.HasPrefix(strings.ToLower(raw), "vmess://") {
		if name := parseVMessName(strings.TrimPrefix(raw, "vmess://")); name != "" {
			return Node{Raw: raw, DisplayName: name}
		}
	}

	if strings.HasPrefix(strings.ToLower(raw), "ssr://") {
		if name := parseSSRName(strings.TrimPrefix(raw, "ssr://")); name != "" {
			return Node{Raw: raw, DisplayName: name}
		}
	}

	if u, err := url.Parse(raw); err == nil {
		name := parseNameFromURL(u)
		if name == "" {
			name = strings.TrimSpace(hostFromURL(u))
		}
		if name == "" {
			name = raw
		}
		return Node{Raw: raw, DisplayName: name}
	}

	return Node{Raw: raw, DisplayName: raw}
}

func parseVMessName(payload string) string {
	decoded, err := remoteparser.DecodeBase64String(payload)
	if err != nil {
		return ""
	}

	var item vmessNode
	if err := json.Unmarshal(decoded, &item); err != nil {
		return ""
	}
	return strings.TrimSpace(item.PS)
}

func parseSSRName(payload string) string {
	decoded, err := remoteparser.DecodeBase64String(payload)
	if err != nil {
		return ""
	}

	parts := strings.SplitN(string(decoded), "/?", 2)
	if len(parts) != 2 {
		return ""
	}

	query, err := url.ParseQuery(parts[1])
	if err != nil {
		return ""
	}

	for _, key := range []string{"remarks", "remark", "ps"} {
		value := strings.TrimSpace(query.Get(key))
		if value == "" {
			continue
		}
		decodedValue, err := remoteparser.DecodeBase64String(value)
		if err == nil {
			return strings.TrimSpace(string(decodedValue))
		}
		return value
	}

	return ""
}

func parseNameFromURL(u *url.URL) string {
	if name := strings.TrimSpace(u.Fragment); name != "" {
		return name
	}

	query := u.Query()
	for _, key := range []string{"remarks", "remark", "ps", "name", "title", "display_name", "peer"} {
		if value := strings.TrimSpace(query.Get(key)); value != "" {
			return maybeDecodeBase64Value(value)
		}
	}

	switch strings.ToLower(u.Scheme) {
	case "ss":
		if name := parseSSName(u); name != "" {
			return name
		}
	}

	return ""
}

func parseSSName(u *url.URL) string {
	if name := strings.TrimSpace(u.Fragment); name != "" {
		return name
	}

	// Legacy SIP002 may place the base64-encoded userinfo in opaque form.
	if u.Host == "" && u.Opaque != "" {
		if decoded, err := remoteparser.DecodeBase64String(u.Opaque); err == nil {
			if at := strings.LastIndexByte(string(decoded), '@'); at >= 0 && at+1 < len(decoded) {
				return strings.TrimSpace(string(decoded[at+1:]))
			}
		}
	}

	return ""
}

func maybeDecodeBase64Value(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	decoded, err := remoteparser.DecodeBase64String(value)
	if err != nil {
		return value
	}
	decodedText := strings.TrimSpace(string(decoded))
	if decodedText == "" || !looksLikeReadableText(decodedText) {
		return value
	}
	return decodedText
}

func looksLikeReadableText(s string) bool {
	for _, r := range s {
		if r == utf8.RuneError {
			return false
		}
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}

func PrefixNode(raw, prefix string) string {
	if strings.TrimSpace(prefix) == "" {
		return raw
	}
	node := ParseNode(raw)
	if node.Raw == "" {
		return raw
	}
	newName := prefix + node.DisplayName

	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "vmess://"):
		if renamed, ok := renameVMess(raw, newName); ok {
			return renamed
		}
	case strings.HasPrefix(lower, "ssr://"):
		if renamed, ok := renameSSR(raw, newName); ok {
			return renamed
		}
	default:
		if renamed, ok := renameURLFragment(raw, newName); ok {
			return renamed
		}
	}

	return raw
}

func renameVMess(raw, newName string) (string, bool) {
	decoded, err := remoteparser.DecodeBase64String(strings.TrimPrefix(raw, "vmess://"))
	if err != nil {
		return "", false
	}

	var item map[string]any
	if err := json.Unmarshal(decoded, &item); err != nil {
		return "", false
	}
	item["ps"] = newName

	encoded, err := json.Marshal(item)
	if err != nil {
		return "", false
	}
	return "vmess://" + remoteparser.EncodeBase64String(encoded), true
}

func renameSSR(raw, newName string) (string, bool) {
	decoded, err := remoteparser.DecodeBase64String(strings.TrimPrefix(raw, "ssr://"))
	if err != nil {
		return "", false
	}

	parts := strings.SplitN(string(decoded), "/?", 2)
	if len(parts) != 2 {
		return "", false
	}
	values, err := url.ParseQuery(parts[1])
	if err != nil {
		return "", false
	}
	values.Set("remarks", remoteparser.EncodeBase64StringRaw([]byte(newName)))
	rebuilt := parts[0] + "/?" + values.Encode()
	return "ssr://" + remoteparser.EncodeBase64StringRaw([]byte(rebuilt)), true
}

func renameURLFragment(raw, newName string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	u.Fragment = newName
	return u.String(), true
}

func hostFromURL(u *url.URL) string {
	if u.Host != "" {
		return u.Host
	}
	if u.Opaque == "" {
		return ""
	}

	if strings.HasPrefix(u.Scheme, "ss") {
		decoded, err := remoteparser.DecodeBase64String(u.Opaque)
		if err == nil {
			if at := strings.LastIndexByte(string(decoded), '@'); at >= 0 && at+1 < len(decoded) {
				return string(decoded[at+1:])
			}
		}
	}

	if idx := strings.LastIndexByte(u.Opaque, '@'); idx >= 0 && idx+1 < len(u.Opaque) {
		return u.Opaque[idx+1:]
	}

	return u.Opaque
}
