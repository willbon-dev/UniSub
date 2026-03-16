package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/willbon-dev/UniSub/internal/config"
	nodeparser "github.com/willbon-dev/UniSub/internal/parser"
	remoteparser "github.com/willbon-dev/UniSub/internal/source/remote/parser"
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
	mu         sync.Mutex
	entries    []string
	fetchedAt  time.Time
	lastErr    error
	lastErrAt  time.Time
	refreshing bool
}

type Result struct {
	Platform string
	Lines    []string
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

	lines := make([]string, 0)
	seen := make(map[string]struct{})
	for i := range sub.sources {
		entries, err := s.resolveSource(ctx, &sub.sources[i], forceRefresh)
		if err != nil {
			return Result{}, err
		}
		for _, line := range entries {
			if _, ok := seen[line]; ok {
				continue
			}
			seen[line] = struct{}{}
			lines = append(lines, line)
		}
	}

	if platform == config.PlatformHapp {
		if routing := strings.TrimSpace(sub.cfg.PlatformOptions.Happ.Routing); routing != "" {
			lines = append([]string{routing}, lines...)
		}
	}

	return Result{Platform: platform, Lines: lines}, nil
}

func (s *Service) resolveSource(ctx context.Context, src *sourceRuntime, forceRefresh bool) ([]string, error) {
	switch src.cfg.Type {
	case config.SourceTypeManual:
		return applyPrefix(filterEntries(src.cfg.Entries, src.includeRegexps, src.excludeRegexps), src.cfg.Prefix), nil
	case config.SourceTypeRemote:
		return s.fetchRemote(ctx, src, forceRefresh)
	default:
		return nil, fmt.Errorf("unsupported source type %q", src.cfg.Type)
	}
}

func (s *Service) fetchRemote(ctx context.Context, src *sourceRuntime, forceRefresh bool) ([]string, error) {
	src.cache.mu.Lock()
	if !forceRefresh && len(src.cache.entries) > 0 && time.Since(src.cache.fetchedAt) < src.cfg.RefreshInterval {
		entries := append([]string(nil), src.cache.entries...)
		src.cache.mu.Unlock()
		return entries, nil
	}
	src.cache.mu.Unlock()

	timeout := src.cfg.Timeout
	if timeout <= 0 {
		timeout = s.cfg.Server.FetchTimeout
	}
	fetchCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(fetchCtx, http.MethodGet, src.cfg.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request for source %q: %w", src.cfg.Name, err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch source %q: %w", src.cfg.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch source %q: unexpected status %s", src.cfg.Name, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, s.maxBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read source %q response: %w", src.cfg.Name, err)
	}
	if int64(len(body)) > s.maxBodyBytes {
		return nil, fmt.Errorf("source %q response exceeded max_response_bytes", src.cfg.Name)
	}

	decoder, err := remoteparser.New(src.cfg.RemoteType)
	if err != nil {
		return nil, err
	}
	entries, err := decoder.DecodeResponse(body)
	if err != nil {
		return nil, fmt.Errorf("decode source %q: %w", src.cfg.Name, err)
	}

	filtered := filterEntries(entries, src.includeRegexps, src.excludeRegexps)
	filtered = applyPrefix(filtered, src.cfg.Prefix)

	src.cache.mu.Lock()
	src.cache.entries = append([]string(nil), filtered...)
	src.cache.fetchedAt = time.Now()
	src.cache.lastErr = nil
	src.cache.lastErrAt = time.Time{}
	src.cache.mu.Unlock()

	return filtered, nil
}

func filterEntries(entries []string, includes, excludes []*regexp.Regexp) []string {
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

func applyPrefix(entries []string, prefix string) []string {
	if strings.TrimSpace(prefix) == "" {
		return entries
	}
	renamed := make([]string, 0, len(entries))
	for _, entry := range entries {
		renamed = append(renamed, nodeparser.PrefixNode(entry, prefix))
	}
	return renamed
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

var ErrNotFound = errors.New("subscription not found")
