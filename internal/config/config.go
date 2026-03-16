package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	SourceTypeManual = "manual"
	SourceTypeRemote = "remote"

	RemoteTypeBase64Lines = "base64_lines"

	PlatformV2RayN = "V2rayN"
	PlatformHapp   = "Happ"
)

type Config struct {
	Server        ServerConfig         `yaml:"server"`
	Subscriptions []SubscriptionConfig `yaml:"subscriptions"`
}

type ServerConfig struct {
	Listen           string        `yaml:"listen"`
	ReadTimeout      time.Duration `yaml:"read_timeout"`
	WriteTimeout     time.Duration `yaml:"write_timeout"`
	ShutdownTimeout  time.Duration `yaml:"shutdown_timeout"`
	FetchTimeout     time.Duration `yaml:"fetch_timeout"`
	MaxResponseBytes int64         `yaml:"max_response_bytes"`
}

type SubscriptionConfig struct {
	Name            string          `yaml:"name"`
	Secret          string          `yaml:"secret"`
	DefaultPlatform string          `yaml:"default_platform"`
	PlatformOptions PlatformOptions `yaml:"platform_options"`
	Sources         []SourceConfig  `yaml:"sources"`
}

type PlatformOptions struct {
	Happ HappOptions `yaml:"happ"`
}

type HappOptions struct {
	Routing string `yaml:"routing"`
}

type SourceConfig struct {
	Name            string        `yaml:"name"`
	Type            string        `yaml:"type"`
	Prefix          string        `yaml:"prefix"`
	Entries         []string      `yaml:"entries"`
	RemoteType      string        `yaml:"remote_type"`
	URL             string        `yaml:"url"`
	RefreshInterval time.Duration `yaml:"refresh_interval"`
	Timeout         time.Duration `yaml:"timeout"`
	IncludePatterns []string      `yaml:"include_patterns"`
	ExcludePatterns []string      `yaml:"exclude_patterns"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server.Listen == "" {
		c.Server.Listen = "127.0.0.1:8080"
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 10 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 30 * time.Second
	}
	if c.Server.ShutdownTimeout == 0 {
		c.Server.ShutdownTimeout = 10 * time.Second
	}
	if c.Server.FetchTimeout == 0 {
		c.Server.FetchTimeout = 20 * time.Second
	}
	if c.Server.MaxResponseBytes == 0 {
		c.Server.MaxResponseBytes = 8 << 20
	}

	for i := range c.Subscriptions {
		if c.Subscriptions[i].DefaultPlatform == "" {
			c.Subscriptions[i].DefaultPlatform = PlatformV2RayN
		}
	}
}

func (c *Config) Validate() error {
	if len(c.Subscriptions) == 0 {
		return errors.New("subscriptions must not be empty")
	}

	seenSecrets := map[string]struct{}{}
	for i, sub := range c.Subscriptions {
		if strings.TrimSpace(sub.Name) == "" {
			return fmt.Errorf("subscriptions[%d].name must not be empty", i)
		}
		if !uuidPattern.MatchString(sub.Secret) {
			return fmt.Errorf("subscriptions[%d].secret must be a UUID", i)
		}
		if _, ok := seenSecrets[sub.Secret]; ok {
			return fmt.Errorf("subscriptions[%d].secret must be unique", i)
		}
		seenSecrets[sub.Secret] = struct{}{}

		if !isSupportedPlatform(sub.DefaultPlatform) {
			return fmt.Errorf("subscriptions[%d].default_platform %q is not supported", i, sub.DefaultPlatform)
		}

		if len(sub.Sources) == 0 {
			return fmt.Errorf("subscriptions[%d].sources must not be empty", i)
		}
		for j, src := range sub.Sources {
			if strings.TrimSpace(src.Name) == "" {
				return fmt.Errorf("subscriptions[%d].sources[%d].name must not be empty", i, j)
			}
			switch src.Type {
			case SourceTypeManual:
				if len(src.Entries) == 0 {
					return fmt.Errorf("subscriptions[%d].sources[%d].entries must not be empty for manual source", i, j)
				}
			case SourceTypeRemote:
				if src.RemoteType == "" {
					return fmt.Errorf("subscriptions[%d].sources[%d].remote_type must not be empty", i, j)
				}
				if !isSupportedRemoteType(src.RemoteType) {
					return fmt.Errorf("subscriptions[%d].sources[%d].remote_type %q is not supported", i, j, src.RemoteType)
				}
				if strings.TrimSpace(src.URL) == "" {
					return fmt.Errorf("subscriptions[%d].sources[%d].url must not be empty", i, j)
				}
				if src.RefreshInterval <= 0 {
					return fmt.Errorf("subscriptions[%d].sources[%d].refresh_interval must be greater than 0", i, j)
				}
				for _, expr := range append(src.IncludePatterns, src.ExcludePatterns...) {
					if _, err := regexp.Compile(expr); err != nil {
						return fmt.Errorf("subscriptions[%d].sources[%d] invalid regexp %q: %w", i, j, expr, err)
					}
				}
			default:
				return fmt.Errorf("subscriptions[%d].sources[%d].type %q is not supported", i, j, src.Type)
			}
		}
	}
	return nil
}

func CanonicalPlatform(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "v2rayn":
		return PlatformV2RayN
	case "happ":
		return PlatformHapp
	default:
		return ""
	}
}

func isSupportedPlatform(name string) bool {
	switch strings.TrimSpace(name) {
	case PlatformV2RayN, PlatformHapp:
		return true
	default:
		return false
	}
}

func isSupportedRemoteType(name string) bool {
	switch strings.TrimSpace(name) {
	case RemoteTypeBase64Lines:
		return true
	default:
		return false
	}
}

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
