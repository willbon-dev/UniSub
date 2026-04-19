package parser

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseClashProxy(raw string) (map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty clash proxy")
	}

	var list []map[string]any
	if err := yaml.Unmarshal([]byte(raw), &list); err == nil && len(list) > 0 {
		return normalizeClashProxy(list[0]), nil
	}

	var proxy map[string]any
	if err := yaml.Unmarshal([]byte(raw), &proxy); err != nil {
		return nil, fmt.Errorf("parse clash proxy: %w", err)
	}
	if len(proxy) == 0 {
		return nil, fmt.Errorf("parse clash proxy: empty object")
	}
	return normalizeClashProxy(proxy), nil
}

func ParseClashProxyList(raw []byte) ([]map[string]any, error) {
	var doc struct {
		Proxies []map[string]any `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse clash proxies yaml: %w", err)
	}
	if len(doc.Proxies) == 0 {
		return nil, fmt.Errorf("parse clash proxies yaml: proxies is empty")
	}

	proxies := make([]map[string]any, 0, len(doc.Proxies))
	for _, proxy := range doc.Proxies {
		proxies = append(proxies, normalizeClashProxy(proxy))
	}
	return proxies, nil
}

func ClashProxyName(proxy map[string]any) string {
	if proxy == nil {
		return ""
	}
	if name, ok := proxy["name"].(string); ok {
		return strings.TrimSpace(name)
	}
	return ""
}

func PrefixClashProxy(proxy map[string]any, prefix string) map[string]any {
	if strings.TrimSpace(prefix) == "" || proxy == nil {
		return proxy
	}

	cloned := cloneMap(proxy)
	name := ClashProxyName(cloned)
	if name == "" {
		name = ClashProxyStableName(cloned)
	}
	cloned["name"] = prefix + name
	return cloned
}

func ClashProxyStableName(proxy map[string]any) string {
	if name := ClashProxyName(proxy); name != "" {
		return name
	}
	sum := sha1.Sum([]byte(ClashProxyDedupKey(proxy)))
	return "proxy-" + hex.EncodeToString(sum[:4])
}

func ClashProxyDedupKey(proxy map[string]any) string {
	if proxy == nil {
		return ""
	}
	name := ClashProxyName(proxy)
	typ, _ := proxy["type"].(string)
	server, _ := proxy["server"].(string)
	return name + "\x00" + strings.TrimSpace(typ) + "\x00" + strings.TrimSpace(server) + "\x00" + stringifyValue(proxy["port"])
}

func stringifyValue(v any) string {
	switch value := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(value)
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func normalizeClashProxy(proxy map[string]any) map[string]any {
	if proxy == nil {
		return nil
	}
	cloned := cloneMap(proxy)
	if _, ok := cloned["name"]; !ok {
		cloned["name"] = ClashProxyStableName(cloned)
	}
	return cloned
}

func cloneMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = cloneValue(v)
	}
	return dst
}

func cloneValue(v any) any {
	switch value := v.(type) {
	case map[string]any:
		return cloneMap(value)
	case []any:
		items := make([]any, len(value))
		for i, item := range value {
			items[i] = cloneValue(item)
		}
		return items
	default:
		return value
	}
}
