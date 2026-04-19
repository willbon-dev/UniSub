package service

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/willbon-dev/UniSub/internal/config"
	nodeparser "github.com/willbon-dev/UniSub/internal/parser"
	remoteparser "github.com/willbon-dev/UniSub/internal/source/remote/parser"
	"gopkg.in/yaml.v3"
)

type Service struct {
	cfg           *config.Config
	httpClient    *http.Client
	maxBodyBytes  int64
	subscriptions map[string]*subscriptionRuntime
}

type subscriptionRuntime struct {
	cfg     config.SubscriptionConfig
	sources []sourceRuntime
}

type sourceRuntime struct {
	cfg            config.SourceConfig
	includeRegexps []*regexp.Regexp
	excludeRegexps []*regexp.Regexp
	cache          remoteCache
}

type remoteCache struct {
	mu        sync.Mutex
	entries    resolvedSourceEntries
	fetchedAt  time.Time
	lastErr    error
	lastErrAt  time.Time
	hasEntries bool
}

type resolvedSourceEntries struct {
	LinkEntries  []string
	ClashProxies []map[string]any
}

type Result struct {
	Platform    string
	ContentType string
	Body        []byte
}

func New(cfg *config.Config) (*Service, error) {
	subscriptions := make(map[string]*subscriptionRuntime, len(cfg.Subscriptions))
	for _, sub := range cfg.Subscriptions {
		rt := &subscriptionRuntime{cfg: sub, sources: make([]sourceRuntime, 0, len(sub.Sources))}
		for _, src := range sub.Sources {
			sourceRT := sourceRuntime{cfg: src}
			for _, expr := range src.IncludePatterns {
				re, err := regexp.Compile(expr)
				if err != nil {
					return nil, fmt.Errorf("compile include regexp %q: %w", expr, err)
				}
				sourceRT.includeRegexps = append(sourceRT.includeRegexps, re)
			}
			for _, expr := range src.ExcludePatterns {
				re, err := regexp.Compile(expr)
				if err != nil {
					return nil, fmt.Errorf("compile exclude regexp %q: %w", expr, err)
				}
				sourceRT.excludeRegexps = append(sourceRT.excludeRegexps, re)
			}
			rt.sources = append(rt.sources, sourceRT)
		}
		subscriptions[sub.Secret] = rt
	}

	return &Service{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Server.FetchTimeout,
		},
		maxBodyBytes:  cfg.Server.MaxResponseBytes,
		subscriptions: subscriptions,
	}, nil
}

func (s *Service) RenderSubscription(ctx context.Context, secret, platform string, forceRefresh bool) (Result, error) {
	sub := s.subscriptions[secret]
	if sub == nil {
		return Result{}, ErrNotFound
	}

	platform = config.CanonicalPlatform(platform)
	if platform == "" {
		platform = config.CanonicalPlatform(sub.cfg.DefaultPlatform)
	}
	if platform == "" {
		return Result{}, fmt.Errorf("subscription %q has unsupported default platform", sub.cfg.Name)
	}

	linkLines := make([]string, 0)
	seenLinks := make(map[string]struct{})
	clashProxies := make([]map[string]any, 0)
	seenClash := make(map[string]struct{})

	for i := range sub.sources {
		src := &sub.sources[i]
		if !sourceSupportsPlatform(src.cfg, platform) {
			continue
		}

		entries, err := s.resolveSource(ctx, src, forceRefresh)
		if err != nil {
			return Result{}, err
		}

		for _, line := range entries.LinkEntries {
			if _, ok := seenLinks[line]; ok {
				continue
			}
			seenLinks[line] = struct{}{}
			linkLines = append(linkLines, line)
		}

		for _, proxy := range entries.ClashProxies {
			key := nodeparser.ClashProxyDedupKey(proxy)
			if _, ok := seenClash[key]; ok {
				continue
			}
			seenClash[key] = struct{}{}
			clashProxies = append(clashProxies, proxy)
		}
	}

	switch platform {
	case config.PlatformHapp:
		if happLines := sub.cfg.PlatformOptions.Happ.RenderSubscriptionLines(); len(happLines) > 0 {
			linkLines = append(happLines, linkLines...)
		}
		return Result{
			Platform:    platform,
			ContentType: "text/plain; charset=utf-8",
			Body:        []byte(strings.Join(linkLines, "\n")),
		}, nil
	case config.PlatformV2RayN:
		return Result{
			Platform:    platform,
			ContentType: "text/plain; charset=utf-8",
			Body:        []byte(strings.Join(linkLines, "\n")),
		}, nil
	case config.PlatformClash:
		if strings.TrimSpace(sub.cfg.PlatformOptions.Clash.Template) == "" {
			return Result{}, fmt.Errorf("subscription %q requires platform_options.clash.template for Clash output", sub.cfg.Name)
		}
		body, err := s.renderClash(ctx, sub.cfg, clashProxies)
		if err != nil {
			return Result{}, err
		}
		return Result{
			Platform:    platform,
			ContentType: "application/yaml; charset=utf-8",
			Body:        body,
		}, nil
	default:
		return Result{}, fmt.Errorf("platform %q is not supported", platform)
	}
}

func (s *Service) resolveSource(ctx context.Context, src *sourceRuntime, forceRefresh bool) (resolvedSourceEntries, error) {
	switch src.cfg.Type {
	case config.SourceTypeManual:
		return resolveManualSource(src.cfg, src.includeRegexps, src.excludeRegexps)
	case config.SourceTypeRemote:
		return s.fetchRemote(ctx, src, forceRefresh)
	default:
		return resolvedSourceEntries{}, fmt.Errorf("unsupported source type %q", src.cfg.Type)
	}
}

func resolveManualSource(cfg config.SourceConfig, includes, excludes []*regexp.Regexp) (resolvedSourceEntries, error) {
	switch cfg.Style {
	case config.StyleLinkLine:
		entries := applyPrefixToLinks(filterLinkEntries(cfg.Entries, includes, excludes), cfg.Prefix)
		return resolvedSourceEntries{LinkEntries: entries}, nil
	case config.StyleClashProxy:
		proxies := make([]map[string]any, 0, len(cfg.Entries))
		for _, entry := range cfg.Entries {
			proxy, err := nodeparser.ParseClashProxy(entry)
			if err != nil {
				return resolvedSourceEntries{}, fmt.Errorf("parse manual clash proxy for source %q: %w", cfg.Name, err)
			}
			proxies = append(proxies, proxy)
		}
		return resolvedSourceEntries{ClashProxies: applyClashFilters(proxies, includes, excludes, cfg.Prefix)}, nil
	default:
		return resolvedSourceEntries{}, fmt.Errorf("unsupported manual style %q", cfg.Style)
	}
}

func (s *Service) fetchRemote(ctx context.Context, src *sourceRuntime, forceRefresh bool) (resolvedSourceEntries, error) {
	src.cache.mu.Lock()
	if !forceRefresh && src.cache.hasEntries && time.Since(src.cache.fetchedAt) < src.cfg.RefreshInterval {
		entries := cloneResolvedEntries(src.cache.entries)
		src.cache.mu.Unlock()
		return entries, nil
	}
	src.cache.mu.Unlock()

	body, err := s.fetchURL(ctx, src.cfg.URL, src.cfg.Timeout, src.cfg.RequestHeaders)
	if err != nil {
		return resolvedSourceEntries{}, fmt.Errorf("fetch source %q: %w", src.cfg.Name, err)
	}

	var entries resolvedSourceEntries
	switch src.cfg.Style {
	case config.StyleLinkLinesBase64:
		decoder, err := remoteparser.New(config.RemoteTypeBase64Lines)
		if err != nil {
			return resolvedSourceEntries{}, err
		}
		lines, err := decoder.DecodeResponse(body)
		if err != nil {
			return resolvedSourceEntries{}, fmt.Errorf("decode source %q: %w", src.cfg.Name, err)
		}
		entries.LinkEntries = applyPrefixToLinks(filterLinkEntries(lines, src.includeRegexps, src.excludeRegexps), src.cfg.Prefix)
	case config.StyleClashProxiesYML:
		proxies, err := nodeparser.ParseClashProxyList(body)
		if err != nil {
			return resolvedSourceEntries{}, fmt.Errorf("decode source %q: %w", src.cfg.Name, err)
		}
		entries.ClashProxies = applyClashFilters(proxies, src.includeRegexps, src.excludeRegexps, src.cfg.Prefix)
	default:
		return resolvedSourceEntries{}, fmt.Errorf("unsupported remote style %q", src.cfg.Style)
	}

	src.cache.mu.Lock()
	src.cache.entries = cloneResolvedEntries(entries)
	src.cache.fetchedAt = time.Now()
	src.cache.lastErr = nil
	src.cache.lastErrAt = time.Time{}
	src.cache.hasEntries = true
	src.cache.mu.Unlock()

	return entries, nil
}

func (s *Service) fetchURL(ctx context.Context, rawURL string, timeout time.Duration, headers map[string]string) ([]byte, error) {
	if timeout <= 0 {
		timeout = s.cfg.Server.FetchTimeout
	}
	fetchCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(fetchCtx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	for key, value := range headers {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		req.Header.Set(key, value)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, s.maxBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if int64(len(body)) > s.maxBodyBytes {
		return nil, fmt.Errorf("response exceeded max_response_bytes")
	}
	return body, nil
}

func (s *Service) renderClash(ctx context.Context, sub config.SubscriptionConfig, proxies []map[string]any) ([]byte, error) {
	templateRef := strings.TrimSpace(sub.PlatformOptions.Clash.Template)
	templateContent, err := s.readTemplate(ctx, templateRef)
	if err != nil {
		return nil, fmt.Errorf("read clash template for subscription %q from %q: %w", sub.Name, templateRef, err)
	}

	rendered, err := buildClashYAML(ctx, s, templateContent, templateRef, proxies)
	if err != nil {
		return nil, fmt.Errorf("render clash template for subscription %q: %w", sub.Name, err)
	}
	return rendered, nil
}

func (s *Service) readTemplate(ctx context.Context, ref string) ([]byte, error) {
	if isHTTPURL(ref) {
		return s.fetchURL(ctx, ref, s.cfg.Server.FetchTimeout, nil)
	}
	path := filepath.Clean(ref)
	data, err := osReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func filterLinkEntries(entries []string, includes, excludes []*regexp.Regexp) []string {
	filtered := make([]string, 0, len(entries))
	for _, entry := range entries {
		node := nodeparser.ParseNode(entry)
		if node.Raw == "" {
			continue
		}
		if !matchesIncludes(node.DisplayName, includes) {
			continue
		}
		if matchesAny(node.DisplayName, excludes) {
			continue
		}
		filtered = append(filtered, node.Raw)
	}
	return filtered
}

func applyPrefixToLinks(entries []string, prefix string) []string {
	if strings.TrimSpace(prefix) == "" {
		return entries
	}
	renamed := make([]string, 0, len(entries))
	for _, entry := range entries {
		renamed = append(renamed, nodeparser.PrefixNode(entry, prefix))
	}
	return renamed
}

func applyClashFilters(proxies []map[string]any, includes, excludes []*regexp.Regexp, prefix string) []map[string]any {
	filtered := make([]map[string]any, 0, len(proxies))
	for _, proxy := range proxies {
		name := nodeparser.ClashProxyStableName(proxy)
		if !matchesIncludes(name, includes) {
			continue
		}
		if matchesAny(name, excludes) {
			continue
		}
		filtered = append(filtered, nodeparser.PrefixClashProxy(proxy, prefix))
	}
	return filtered
}

func cloneResolvedEntries(entries resolvedSourceEntries) resolvedSourceEntries {
	cloned := resolvedSourceEntries{
		LinkEntries: append([]string(nil), entries.LinkEntries...),
	}
	if len(entries.ClashProxies) > 0 {
		cloned.ClashProxies = make([]map[string]any, 0, len(entries.ClashProxies))
		for _, proxy := range entries.ClashProxies {
			cloned.ClashProxies = append(cloned.ClashProxies, nodeparser.PrefixClashProxy(proxy, ""))
		}
	}
	return cloned
}

func matchesIncludes(name string, regexps []*regexp.Regexp) bool {
	if len(regexps) == 0 {
		return true
	}
	return matchesAny(name, regexps)
}

func matchesAny(name string, regexps []*regexp.Regexp) bool {
	for _, re := range regexps {
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

func sourceSupportsPlatform(src config.SourceConfig, platform string) bool {
	for _, candidate := range src.Platforms {
		if candidate == platform {
			return true
		}
	}
	return false
}

func isHTTPURL(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func buildClashYAML(ctx context.Context, svc *Service, templateContent []byte, templateRef string, proxies []map[string]any) ([]byte, error) {
	baseDoc := map[string]any{}
	template := parseClashTemplateINI(templateContent)

	if strings.TrimSpace(template.ClashRuleBase) != "" {
		baseContent, err := readTemplateReference(ctx, svc, template.ClashRuleBase, templateRef)
		if err != nil {
			return nil, fmt.Errorf("read clash_rule_base %q: %w", template.ClashRuleBase, err)
		}
		if err := yaml.Unmarshal(baseContent, &baseDoc); err != nil {
			return nil, fmt.Errorf("parse clash_rule_base %q: %w", template.ClashRuleBase, err)
		}
	}

	if len(proxies) == 0 {
		proxies = []map[string]any{}
	}
	baseDoc["proxies"] = proxies

	proxyNames := make([]string, 0, len(proxies))
	for _, proxy := range proxies {
		proxyNames = append(proxyNames, nodeparser.ClashProxyStableName(proxy))
	}

	if len(template.ProxyGroups) > 0 {
		baseDoc["proxy-groups"] = renderProxyGroups(template.ProxyGroups, proxyNames)
	}

	rules, providers := renderRulesAndProviders(template.RuleSets)
	if len(rules) > 0 {
		baseDoc["rules"] = rules
	}
	if len(providers) > 0 {
		baseDoc["rule-providers"] = providers
	}

	ensureClashDefaults(baseDoc)

	rendered, err := yaml.Marshal(baseDoc)
	if err != nil {
		return nil, fmt.Errorf("marshal clash yaml: %w", err)
	}
	return rendered, nil
}

type clashTemplate struct {
	RuleSets      []clashRuleSet
	ProxyGroups   []clashProxyGroup
	ClashRuleBase string
}

type clashRuleSet struct {
	Policy string
	Target string
}

type clashProxyGroup struct {
	Name      string
	Type      string
	Selectors []string
	URL       string
	Interval  int
	Tolerance int
}

func parseClashTemplateINI(content []byte) clashTemplate {
	template := clashTemplate{}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "ruleset="):
			payload := strings.TrimSpace(strings.TrimPrefix(line, "ruleset="))
			parts := strings.SplitN(payload, ",", 2)
			if len(parts) != 2 {
				continue
			}
			template.RuleSets = append(template.RuleSets, clashRuleSet{
				Policy: strings.TrimSpace(parts[0]),
				Target: strings.TrimSpace(parts[1]),
			})
		case strings.HasPrefix(line, "custom_proxy_group="):
			if group, ok := parseProxyGroup(strings.TrimSpace(strings.TrimPrefix(line, "custom_proxy_group="))); ok {
				template.ProxyGroups = append(template.ProxyGroups, group)
			}
		case strings.HasPrefix(line, "clash_rule_base="):
			template.ClashRuleBase = strings.TrimSpace(strings.TrimPrefix(line, "clash_rule_base="))
		}
	}
	return template
}

func parseProxyGroup(payload string) (clashProxyGroup, bool) {
	parts := strings.Split(payload, "`")
	if len(parts) < 2 {
		return clashProxyGroup{}, false
	}
	group := clashProxyGroup{
		Name:      strings.TrimSpace(parts[0]),
		Type:      strings.TrimSpace(parts[1]),
		Selectors: make([]string, 0),
		Interval:  300,
	}
	for idx, part := range parts[2:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		switch {
		case strings.HasPrefix(part, "[]"):
			group.Selectors = append(group.Selectors, strings.TrimPrefix(part, "[]"))
		case idx == len(parts[2:])-1 && group.Type == "url-test":
			if value, err := strconv.Atoi(part); err == nil {
				group.Tolerance = value
			}
		case group.Type == "url-test" && group.URL == "" && (strings.HasPrefix(part, "http://") || strings.HasPrefix(part, "https://")):
			group.URL = part
		case group.Type == "url-test" && group.URL != "" && group.Interval == 300:
			if value, err := strconv.Atoi(part); err == nil {
				group.Interval = value
			}
		default:
			group.Selectors = append(group.Selectors, part)
		}
	}
	return group, group.Name != "" && group.Type != ""
}

func renderProxyGroups(groups []clashProxyGroup, proxyNames []string) []map[string]any {
	rendered := make([]map[string]any, 0, len(groups))
	for _, group := range groups {
		groupMap := map[string]any{
			"name": group.Name,
			"type": group.Type,
		}

		proxies := expandSelectors(group.Selectors, proxyNames)
		if len(proxies) > 0 {
			groupMap["proxies"] = proxies
		}
		if group.Type == "url-test" {
			if group.URL != "" {
				groupMap["url"] = group.URL
			}
			if group.Interval > 0 {
				groupMap["interval"] = group.Interval
			}
			if group.Tolerance > 0 {
				groupMap["tolerance"] = group.Tolerance
			}
		}

		rendered = append(rendered, groupMap)
	}
	return rendered
}

func expandSelectors(selectors []string, proxyNames []string) []string {
	expanded := make([]string, 0)
	seen := make(map[string]struct{})
	appendUnique := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		expanded = append(expanded, value)
	}

	for _, selector := range selectors {
		switch selector {
		case ".*", "*":
			for _, name := range proxyNames {
				appendUnique(name)
			}
		default:
			appendUnique(selector)
		}
	}
	return expanded
}

func renderRulesAndProviders(ruleSets []clashRuleSet) ([]string, map[string]map[string]any) {
	rules := make([]string, 0, len(ruleSets))
	providers := make(map[string]map[string]any)

	for i, ruleSet := range ruleSets {
		target := strings.TrimSpace(ruleSet.Target)
		policy := strings.TrimSpace(ruleSet.Policy)
		switch {
		case strings.HasPrefix(target, "[]FINAL"):
			rules = append(rules, "MATCH,"+policy)
		case strings.HasPrefix(target, "[]GEOIP,"):
			country := strings.TrimSpace(strings.TrimPrefix(target, "[]GEOIP,"))
			rules = append(rules, "GEOIP,"+country+","+policy)
		case isHTTPURL(target):
			providerName := fmt.Sprintf("provider-%d", i+1)
			providers[providerName] = map[string]any{
				"type":     "http",
				"behavior": "classical",
				"url":      target,
				"path":     "./ruleset/" + providerName + ".yaml",
				"interval": 86400,
			}
			rules = append(rules, "RULE-SET,"+providerName+","+policy)
		}
	}

	return rules, providers
}

func ensureClashDefaults(doc map[string]any) {
	if _, ok := doc["port"]; !ok {
		doc["port"] = 7890
	}
	if _, ok := doc["socks-port"]; !ok {
		doc["socks-port"] = 7891
	}
	if _, ok := doc["allow-lan"]; !ok {
		doc["allow-lan"] = true
	}
	if _, ok := doc["mode"]; !ok {
		doc["mode"] = "rule"
	}
	if _, ok := doc["log-level"]; !ok {
		doc["log-level"] = "info"
	}
}

func readTemplateReference(ctx context.Context, svc *Service, ref, parentRef string) ([]byte, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("empty template reference")
	}
	if isHTTPURL(ref) {
		return svc.fetchURL(ctx, ref, svc.cfg.Server.FetchTimeout, nil)
	}
	if isHTTPURL(parentRef) {
		return nil, fmt.Errorf("relative local path %q is not supported for remote template %q", ref, parentRef)
	}
	baseDir := filepath.Dir(parentRef)
	return osReadFile(filepath.Join(baseDir, ref))
}

var osReadFile = func(path string) ([]byte, error) {
	return os.ReadFile(path)
}

var ErrNotFound = errors.New("subscription not found")
