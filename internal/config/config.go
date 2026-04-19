package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
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
	PlatformClash  = "Clash"

	StyleLinkLine        = "link_line"
	StyleLinkLinesBase64 = "link_lines_base64"
	StyleClashProxy      = "clash_proxy"
	StyleClashProxiesYML = "clash_proxies_yaml"
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
	Happ  HappOptions  `yaml:"happ"`
	Clash ClashOptions `yaml:"clash"`
}

type ClashOptions struct {
	Template string `yaml:"template"`
}

type HappOptions struct {
	Routing string `yaml:"routing"`

	ProfileUpdateInterval            *int    `yaml:"profile_update_interval"`
	ProfileTitle                     *string `yaml:"profile_title"`
	SubscriptionUserinfo             *string `yaml:"subscription_userinfo"`
	SupportURL                       *string `yaml:"support_url"`
	ProfileWebPageURL                *string `yaml:"profile_web_page_url"`
	Announce                         *string `yaml:"announce"`
	RoutingEnable                    *bool   `yaml:"routing_enable"`
	CustomTunnelConfig               *string `yaml:"custom_tunnel_config"`
	ProviderID                       *string `yaml:"provider_id"`
	NewURL                           *string `yaml:"new_url"`
	NewDomain                        *string `yaml:"new_domain"`
	FallbackURL                      *string `yaml:"fallback_url"`
	NoLimitEnabled                   *bool   `yaml:"no_limit_enabled"`
	NoLimitXHTTPEnabled              *bool   `yaml:"no_limit_xhttp_enabled"`
	SubscriptionAlwaysHWIDEnable     *bool   `yaml:"subscription_always_hwid_enable"`
	NotificationSubsExpire           *bool   `yaml:"notification_subs_expire"`
	HideSettings                     *bool   `yaml:"hide_settings"`
	ServerAddressResolveEnable       *bool   `yaml:"server_address_resolve_enable"`
	ServerAddressResolveDNSDomain    *string `yaml:"server_address_resolve_dns_domain"`
	ServerAddressResolveDNSIP        *string `yaml:"server_address_resolve_dns_ip"`
	SubscriptionAutoconnect          *bool   `yaml:"subscription_autoconnect"`
	SubscriptionAutoconnectType      *string `yaml:"subscription_autoconnect_type"`
	SubscriptionPingOnOpenEnabled    *bool   `yaml:"subscription_ping_onopen_enabled"`
	SubscriptionAutoUpdateEnable     *bool   `yaml:"subscription_auto_update_enable"`
	FragmentationEnable              *bool   `yaml:"fragmentation_enable"`
	FragmentationPackets             *string `yaml:"fragmentation_packets"`
	FragmentationLength              *string `yaml:"fragmentation_length"`
	FragmentationInterval            *string `yaml:"fragmentation_interval"`
	FragmentationMaxSplit            *string `yaml:"fragmentation_maxsplit"`
	NoisesEnable                     *bool   `yaml:"noises_enable"`
	NoisesType                       *string `yaml:"noises_type"`
	NoisesPacket                     *string `yaml:"noises_packet"`
	NoisesDelay                      *string `yaml:"noises_delay"`
	NoisesApplyTo                    *string `yaml:"noises_applyto"`
	PingType                         *string `yaml:"ping_type"`
	CheckURLViaProxy                 *string `yaml:"check_url_via_proxy"`
	ChangeUserAgent                  *string `yaml:"change_user_agent"`
	AppAutoStart                     *bool   `yaml:"app_auto_start"`
	SubscriptionAutoUpdateOpenEnable *bool   `yaml:"subscription_auto_update_open_enable"`
	PerAppProxyMode                  *string `yaml:"per_app_proxy_mode"`
	PerAppProxyList                  *string `yaml:"per_app_proxy_list"`
	SniffingEnable                   *bool   `yaml:"sniffing_enable"`
	SubscriptionsCollapse            *bool   `yaml:"subscriptions_collapse"`
	SubscriptionsExpandNow           *bool   `yaml:"subscriptions_expand_now"`
	PingResult                       *string `yaml:"ping_result"`
	MuxEnable                        *bool   `yaml:"mux_enable"`
	MuxTCPConnections                *string `yaml:"mux_tcp_connections"`
	MuxXUDPConnections               *string `yaml:"mux_xudp_connections"`
	MuxQUIC                          *string `yaml:"mux_quic"`
	ProxyEnable                      *bool   `yaml:"proxy_enable"`
	TunEnable                        *bool   `yaml:"tun_enable"`
	TunMode                          *string `yaml:"tun_mode"`
	TunType                          *string `yaml:"tun_type"`
	ExcludeRoutes                    *string `yaml:"exclude_routes"`
	ColorProfile                     *string `yaml:"color_profile"`
	SubInfoColor                     *string `yaml:"sub_info_color"`
	SubInfoText                      *string `yaml:"sub_info_text"`
	SubInfoButtonText                *string `yaml:"sub_info_button_text"`
	SubInfoButtonLink                *string `yaml:"sub_info_button_link"`
	SubExpire                        *bool   `yaml:"sub_expire"`
	SubExpireButtonLink              *string `yaml:"sub_expire_button_link"`
}

type SourceConfig struct {
	Name            string        `yaml:"name"`
	Type            string        `yaml:"type"`
	Platforms       []string      `yaml:"platforms"`
	Style           string        `yaml:"style"`
	Prefix          string        `yaml:"prefix"`
	Entries         []string      `yaml:"entries"`
	RemoteType      string        `yaml:"remote_type"`
	URL             string        `yaml:"url"`
	RequestHeaders  map[string]string `yaml:"request_headers"`
	RefreshInterval time.Duration `yaml:"refresh_interval"`
	Timeout         time.Duration `yaml:"timeout"`
	IncludePatterns []string      `yaml:"include_patterns"`
	ExcludePatterns []string      `yaml:"exclude_patterns"`
}

func (s *SourceConfig) UnmarshalYAML(value *yaml.Node) error {
	type sourceConfigAlias struct {
		Name            string        `yaml:"name"`
		Type            string        `yaml:"type"`
		Platforms       []string      `yaml:"platforms"`
		Style           string        `yaml:"style"`
		Prefix          string        `yaml:"prefix"`
		Entries         []yaml.Node   `yaml:"entries"`
		RemoteType      string        `yaml:"remote_type"`
		URL             string        `yaml:"url"`
		RequestHeaders  map[string]string `yaml:"request_headers"`
		RefreshInterval time.Duration `yaml:"refresh_interval"`
		Timeout         time.Duration `yaml:"timeout"`
		IncludePatterns []string      `yaml:"include_patterns"`
		ExcludePatterns []string      `yaml:"exclude_patterns"`
	}

	var aux sourceConfigAlias
	if err := value.Decode(&aux); err != nil {
		return err
	}

	entries := make([]string, 0, len(aux.Entries))
	for i := range aux.Entries {
		text, err := renderEntryNode(&aux.Entries[i])
		if err != nil {
			return fmt.Errorf("render source entry: %w", err)
		}
		entries = append(entries, text)
	}

	*s = SourceConfig{
		Name:            aux.Name,
		Type:            aux.Type,
		Platforms:       aux.Platforms,
		Style:           aux.Style,
		Prefix:          aux.Prefix,
		Entries:         entries,
		RemoteType:      aux.RemoteType,
		URL:             aux.URL,
		RequestHeaders:  aux.RequestHeaders,
		RefreshInterval: aux.RefreshInterval,
		Timeout:         aux.Timeout,
		IncludePatterns: aux.IncludePatterns,
		ExcludePatterns: aux.ExcludePatterns,
	}
	return nil
}

func renderEntryNode(node *yaml.Node) (string, error) {
	if node == nil {
		return "", nil
	}
	if node.Kind == yaml.ScalarNode {
		return strings.TrimSpace(node.Value), nil
	}

	var value any
	if err := node.Decode(&value); err != nil {
		return "", err
	}

	rendered, err := yaml.Marshal(value)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(rendered)), nil
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
		c.Subscriptions[i].DefaultPlatform = CanonicalPlatform(c.Subscriptions[i].DefaultPlatform)

		for j := range c.Subscriptions[i].Sources {
			src := &c.Subscriptions[i].Sources[j]
			if src.Style == "" {
				src.Style = legacyRemoteTypeToStyle(src.RemoteType)
			}
			src.Style = canonicalStyle(src.Style)

			canonicalPlatforms := make([]string, 0, len(src.Platforms))
			for _, platform := range src.Platforms {
				canonicalPlatforms = append(canonicalPlatforms, CanonicalPlatform(platform))
			}
			src.Platforms = canonicalPlatforms
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
			if len(src.Platforms) == 0 {
				return fmt.Errorf("subscriptions[%d].sources[%d].platforms must not be empty", i, j)
			}
			for k, platform := range src.Platforms {
				if !isSupportedPlatform(platform) {
					return fmt.Errorf("subscriptions[%d].sources[%d].platforms[%d] %q is not supported", i, j, k, platform)
				}
			}
			if !isSupportedStyle(src.Style) {
				return fmt.Errorf("subscriptions[%d].sources[%d].style %q is not supported", i, j, src.Style)
			}
			switch src.Type {
			case SourceTypeManual:
				if len(src.Entries) == 0 {
					return fmt.Errorf("subscriptions[%d].sources[%d].entries must not be empty for manual source", i, j)
				}
				switch src.Style {
				case StyleLinkLine:
					if containsPlatform(src.Platforms, PlatformClash) {
						return fmt.Errorf("subscriptions[%d].sources[%d].style %q cannot be used with Clash platform", i, j, src.Style)
					}
				case StyleClashProxy:
					if !containsPlatform(src.Platforms, PlatformClash) {
						return fmt.Errorf("subscriptions[%d].sources[%d].style %q requires Clash platform", i, j, src.Style)
					}
				default:
					return fmt.Errorf("subscriptions[%d].sources[%d].style %q is not valid for manual source", i, j, src.Style)
				}
			case SourceTypeRemote:
				if src.Style == "" && src.RemoteType == "" {
					return fmt.Errorf("subscriptions[%d].sources[%d].style must not be empty", i, j)
				}
				if strings.TrimSpace(src.RemoteType) != "" && !isSupportedRemoteType(src.RemoteType) {
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
				switch src.Style {
				case StyleLinkLinesBase64:
					if containsPlatform(src.Platforms, PlatformClash) {
						return fmt.Errorf("subscriptions[%d].sources[%d].style %q cannot be used with Clash platform", i, j, src.Style)
					}
				case StyleClashProxiesYML:
					if !containsPlatform(src.Platforms, PlatformClash) {
						return fmt.Errorf("subscriptions[%d].sources[%d].style %q requires Clash platform", i, j, src.Style)
					}
				default:
					return fmt.Errorf("subscriptions[%d].sources[%d].style %q is not valid for remote source", i, j, src.Style)
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
	case "clash":
		return PlatformClash
	default:
		return ""
	}
}

func isSupportedPlatform(name string) bool {
	switch strings.TrimSpace(name) {
	case PlatformV2RayN, PlatformHapp, PlatformClash:
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

func canonicalStyle(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case StyleLinkLine:
		return StyleLinkLine
	case StyleLinkLinesBase64:
		return StyleLinkLinesBase64
	case StyleClashProxy:
		return StyleClashProxy
	case StyleClashProxiesYML:
		return StyleClashProxiesYML
	default:
		return ""
	}
}

func isSupportedStyle(name string) bool {
	switch name {
	case StyleLinkLine, StyleLinkLinesBase64, StyleClashProxy, StyleClashProxiesYML:
		return true
	default:
		return false
	}
}

func legacyRemoteTypeToStyle(name string) string {
	switch strings.TrimSpace(name) {
	case RemoteTypeBase64Lines:
		return StyleLinkLinesBase64
	default:
		return ""
	}
}

func containsPlatform(platforms []string, want string) bool {
	for _, platform := range platforms {
		if platform == want {
			return true
		}
	}
	return false
}

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func (o HappOptions) RenderSubscriptionLines() []string {
	lines := make([]string, 0, 1)
	seen := make(map[string]struct{}, 1)

	appendUnique := func(line string) {
		line = strings.TrimSpace(line)
		if line == "" {
			return
		}
		if _, ok := seen[line]; ok {
			return
		}
		seen[line] = struct{}{}
		lines = append(lines, line)
	}

	appendUnique(o.Routing)

	appendComment := func(key, value, separator string) {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			return
		}
		if separator == "" {
			separator = ": "
		}
		appendUnique("#" + key + separator + value)
	}
	appendCommentAllowEmpty := func(key string, value *string, separator string) {
		if value == nil {
			return
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return
		}
		if separator == "" {
			separator = ": "
		}
		appendUnique("#" + key + separator + strings.TrimSpace(*value))
	}
	appendBool := func(key string, value *bool) {
		if value == nil {
			return
		}
		if *value {
			appendComment(key, "1", ": ")
			return
		}
		appendComment(key, "0", ": ")
	}
	appendInt := func(key string, value *int) {
		if value == nil {
			return
		}
		appendComment(key, strconv.Itoa(*value), ": ")
	}

	appendInt("profile-update-interval", o.ProfileUpdateInterval)
	if o.ProfileTitle != nil {
		appendComment("profile-title", *o.ProfileTitle, ": ")
	}
	if o.SubscriptionUserinfo != nil {
		appendComment("subscription-userinfo", *o.SubscriptionUserinfo, ": ")
	}
	if o.SupportURL != nil {
		appendComment("support-url", *o.SupportURL, ": ")
	}
	if o.ProfileWebPageURL != nil {
		appendComment("profile-web-page-url", *o.ProfileWebPageURL, ": ")
	}
	if o.Announce != nil {
		appendComment("announce", *o.Announce, ": ")
	}
	appendBool("routing-enable", o.RoutingEnable)
	if o.CustomTunnelConfig != nil {
		appendComment("custom-tunnel-config", *o.CustomTunnelConfig, ": ")
	}
	if o.ProviderID != nil {
		appendComment("providerid", *o.ProviderID, " ")
	}
	if o.NewURL != nil {
		appendComment("new-url", *o.NewURL, " ")
	}
	if o.NewDomain != nil {
		appendComment("new-domain", *o.NewDomain, " ")
	}
	if o.FallbackURL != nil {
		appendComment("fallback-url", *o.FallbackURL, " ")
	}
	appendBool("no-limit-enabled", o.NoLimitEnabled)
	appendBool("no-limit-xhttp-enabled", o.NoLimitXHTTPEnabled)
	appendBool("subscription-always-hwid-enable", o.SubscriptionAlwaysHWIDEnable)
	appendBool("notification-subs-expire", o.NotificationSubsExpire)
	appendBool("hide-settings", o.HideSettings)
	appendBool("server-address-resolve-enable", o.ServerAddressResolveEnable)
	if o.ServerAddressResolveDNSDomain != nil {
		appendComment("server-address-resolve-dns-domain", *o.ServerAddressResolveDNSDomain, ": ")
	}
	if o.ServerAddressResolveDNSIP != nil {
		appendComment("server-address-resolve-dns-ip", *o.ServerAddressResolveDNSIP, ": ")
	}
	appendBool("subscription-autoconnect", o.SubscriptionAutoconnect)
	if o.SubscriptionAutoconnectType != nil {
		appendComment("subscription-autoconnect-type", *o.SubscriptionAutoconnectType, ": ")
	}
	appendBool("subscription-ping-onopen-enabled", o.SubscriptionPingOnOpenEnabled)
	appendBool("subscription-auto-update-enable", o.SubscriptionAutoUpdateEnable)
	appendBool("fragmentation-enable", o.FragmentationEnable)
	if o.FragmentationPackets != nil {
		appendComment("fragmentation-packets", *o.FragmentationPackets, ": ")
	}
	if o.FragmentationLength != nil {
		appendComment("fragmentation-length", *o.FragmentationLength, ": ")
	}
	if o.FragmentationInterval != nil {
		appendComment("fragmentation-interval", *o.FragmentationInterval, ": ")
	}
	if o.FragmentationMaxSplit != nil {
		appendComment("fragmentation-maxsplit", *o.FragmentationMaxSplit, ": ")
	}
	appendBool("noises-enable", o.NoisesEnable)
	if o.NoisesType != nil {
		appendComment("noises-type", *o.NoisesType, ": ")
	}
	if o.NoisesPacket != nil {
		appendComment("noises-packet", *o.NoisesPacket, ": ")
	}
	if o.NoisesDelay != nil {
		appendComment("noises-delay", *o.NoisesDelay, ": ")
	}
	if o.NoisesApplyTo != nil {
		appendComment("noises-applyto", *o.NoisesApplyTo, ": ")
	}
	if o.PingType != nil {
		appendComment("ping-type", *o.PingType, " ")
	}
	if o.CheckURLViaProxy != nil {
		appendComment("check-url-via-proxy", *o.CheckURLViaProxy, ": ")
	}
	if o.ChangeUserAgent != nil {
		appendComment("change-user-agent", *o.ChangeUserAgent, ": ")
	}
	appendBool("app-auto-start", o.AppAutoStart)
	appendBool("subscription-auto-update-open-enable", o.SubscriptionAutoUpdateOpenEnable)
	if o.PerAppProxyMode != nil {
		appendComment("per-app-proxy-mode", *o.PerAppProxyMode, ": ")
	}
	if o.PerAppProxyList != nil {
		appendComment("per-app-proxy-list", *o.PerAppProxyList, ": ")
	}
	appendBool("sniffing-enable", o.SniffingEnable)
	appendBool("subscriptions-collapse", o.SubscriptionsCollapse)
	appendBool("subscriptions-expand-now", o.SubscriptionsExpandNow)
	if o.PingResult != nil {
		appendComment("ping-result", *o.PingResult, ": ")
	}
	appendBool("mux-enable", o.MuxEnable)
	if o.MuxTCPConnections != nil {
		appendComment("mux-tcp-connections", *o.MuxTCPConnections, ": ")
	}
	if o.MuxXUDPConnections != nil {
		appendComment("mux-xudp-connections", *o.MuxXUDPConnections, ": ")
	}
	if o.MuxQUIC != nil {
		appendComment("mux-quic", *o.MuxQUIC, ": ")
	}
	appendBool("proxy-enable", o.ProxyEnable)
	appendBool("tun-enable", o.TunEnable)
	if o.TunMode != nil {
		appendComment("tun-mode", *o.TunMode, ": ")
	}
	if o.TunType != nil {
		appendComment("tun-type", *o.TunType, ": ")
	}
	if o.ExcludeRoutes != nil {
		appendComment("exclude-routes", *o.ExcludeRoutes, ": ")
	}
	if o.ColorProfile != nil {
		appendComment("color-profile", *o.ColorProfile, ": ")
	}
	if o.SubInfoColor != nil {
		appendComment("sub-info-color", *o.SubInfoColor, ": ")
	}
	appendCommentAllowEmpty("sub-info-text", o.SubInfoText, ": ")
	if o.SubInfoButtonText != nil {
		appendComment("sub-info-button-text", *o.SubInfoButtonText, ": ")
	}
	if o.SubInfoButtonLink != nil {
		appendComment("sub-info-button-link", *o.SubInfoButtonLink, ": ")
	}
	appendBool("sub-expire", o.SubExpire)
	if o.SubExpireButtonLink != nil {
		appendComment("sub-expire-button-link", *o.SubExpireButtonLink, ": ")
	}

	return lines
}
