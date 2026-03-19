package source

import (
	"comic-viewer-claude/internal/model"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	xproxy "golang.org/x/net/proxy"
)

const (
	SourceStatusActive   = "active"
	SourceStatusProxy    = "proxy"
	SourceStatusInactive = "inactive"
	SourceStatusUnknown  = "unknown"

	AccessStatusAvailable   = "available"
	AccessStatusUnavailable = "unavailable"
	AccessStatusUnknown     = "unknown"

	ProxyModeOff      = "off"
	ProxyModeFallback = "fallback"
	ProxyModeAlways   = "always"
)

const proxySettingsCacheTTL = 5 * time.Second
const proxyHealthCacheTTL = 30 * time.Second

var proxyProbeTargets = []struct {
	Name string
	URL  string
}{
	{Name: "Google", URL: "https://www.google.com/generate_204"},
	{Name: "YouTube", URL: "https://www.youtube.com/generate_204"},
}

type ProxySettings struct {
	Mode       string
	Type       string
	Host       string
	Port       int
	Username   string
	Password   string
	Timeout    time.Duration
	Configured bool
}

type AccessCandidate struct {
	Source   *model.SourceSite
	UseProxy bool
}

type AccessResult struct {
	Source   *model.SourceSite
	UseProxy bool
	Latency  int
}

func (s ProxySettings) Enabled() bool {
	return s.Configured && s.Host != "" && s.Port > 0 && (s.Type == "http" || s.Type == "socks5")
}

func (s ProxySettings) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func BuildHTTPClient(useProxy bool, settings ProxySettings, timeout time.Duration) (*http.Client, error) {
	transport, err := buildHTTPTransport(useProxy, settings, timeout)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}, nil
}

func (m *Manager) GetProxySettings() ProxySettings {
	m.mu.RLock()
	if time.Since(m.proxyLoadedAt) < proxySettingsCacheTTL {
		cached := m.proxySettings
		m.mu.RUnlock()
		return cached
	}
	m.mu.RUnlock()

	settingsMap := make(map[string]string)
	settings, err := m.repo.GetAllUserSettings(1)
	if err == nil {
		for _, setting := range settings {
			settingsMap[setting.Key] = setting.Value
		}
	}

	timeout := 6 * time.Second
	if raw := strings.TrimSpace(settingsMap["network_proxy_timeout_ms"]); raw != "" {
		if timeoutMS, parseErr := strconv.Atoi(raw); parseErr == nil && timeoutMS > 0 {
			timeout = time.Duration(timeoutMS) * time.Millisecond
		}
	}

	cfg := ProxySettings{
		Mode:       normalizeProxyMode(settingsMap["network_proxy_mode"]),
		Type:       normalizeProxyType(settingsMap["network_proxy_type"]),
		Host:       strings.TrimSpace(settingsMap["network_proxy_host"]),
		Username:   strings.TrimSpace(settingsMap["network_proxy_username"]),
		Password:   settingsMap["network_proxy_password"],
		Timeout:    timeout,
		Configured: false,
	}
	if port, parseErr := strconv.Atoi(strings.TrimSpace(settingsMap["network_proxy_port"])); parseErr == nil && port > 0 {
		cfg.Port = port
	}
	cfg.Configured = cfg.Host != "" && cfg.Port > 0 && cfg.Type != ""

	m.mu.Lock()
	m.proxySettings = cfg
	m.proxyLoadedAt = time.Now()
	m.mu.Unlock()

	return cfg
}

func (m *Manager) InvalidateProxySettingsCache() {
	m.mu.Lock()
	m.proxyLoadedAt = time.Time{}
	m.proxyHealth = nil
	m.proxyHealthAt = time.Time{}
	m.mu.Unlock()
}

func (m *Manager) GetProxyHealth(force bool) *model.ProxyHealthResponse {
	settings := m.GetProxySettings()
	if !settings.Enabled() {
		return &model.ProxyHealthResponse{
			Configured: false,
			Available:  false,
			Status:     "unconfigured",
			Message:    "代理未配置完整",
			Targets:    []model.ProxyTargetProbeResponse{},
		}
	}

	m.mu.RLock()
	if !force && m.proxyHealth != nil && time.Since(m.proxyHealthAt) < proxyHealthCacheTTL {
		cached := *m.proxyHealth
		cached.Targets = append([]model.ProxyTargetProbeResponse(nil), m.proxyHealth.Targets...)
		m.mu.RUnlock()
		return &cached
	}
	m.mu.RUnlock()

	timeout := settings.Timeout
	if timeout <= 0 {
		timeout = 6 * time.Second
	}

	client, err := BuildHTTPClient(true, settings, timeout)
	if err != nil {
		response := &model.ProxyHealthResponse{
			Configured: true,
			Available:  false,
			Status:     AccessStatusUnavailable,
			Message:    err.Error(),
			Targets:    []model.ProxyTargetProbeResponse{},
		}
		m.cacheProxyHealth(response)
		return response
	}

	results := make([]model.ProxyTargetProbeResponse, len(proxyProbeTargets))
	var wg sync.WaitGroup
	for i := range proxyProbeTargets {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			target := proxyProbeTargets[index]
			results[index] = probeProxyTarget(client, target.Name, target.URL)
		}(i)
	}
	wg.Wait()

	response := &model.ProxyHealthResponse{
		Configured: true,
		Available:  false,
		Status:     AccessStatusUnavailable,
		Targets:    results,
		CheckedAt:  time.Now().Format(time.RFC3339),
	}

	bestLatency := 0
	for _, result := range results {
		if result.Status != AccessStatusAvailable {
			continue
		}
		response.Available = true
		if bestLatency == 0 || (result.Latency > 0 && result.Latency < bestLatency) {
			bestLatency = result.Latency
		}
	}
	response.Latency = bestLatency

	if response.Available {
		response.Status = AccessStatusAvailable
		response.Message = "代理联通正常"
	} else {
		failures := make([]string, 0, len(results))
		for _, result := range results {
			if result.Error != "" {
				failures = append(failures, fmt.Sprintf("%s: %s", result.Name, result.Error))
			}
		}
		if len(failures) == 0 {
			response.Message = "代理检测失败"
		} else {
			response.Message = strings.Join(failures, "；")
		}
	}

	m.cacheProxyHealth(response)
	return response
}

func (m *Manager) IsProxyAvailable(force bool) bool {
	return m.GetProxyHealth(force).Available
}

func (m *Manager) GetActive() *model.SourceSite {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.active != nil {
		return m.active
	}
	for _, source := range m.sources {
		if source.DirectStatus == AccessStatusAvailable {
			return source
		}
	}
	return nil
}

func (m *Manager) PreferredSourceBaseURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, source := range m.sources {
		if source.DirectStatus == AccessStatusAvailable {
			return source.URL
		}
	}
	for _, source := range m.sources {
		if source.ProxyStatus == AccessStatusAvailable {
			return source.URL
		}
	}
	for _, source := range m.sources {
		return source.URL
	}
	return ""
}

func (m *Manager) ExecuteRequest(
	build func(baseURL string) (targetURL string, referer string),
	do func(targetURL, referer string, useProxy bool) ([]byte, int, error),
) ([]byte, *AccessResult, error) {
	proxySettings := m.GetProxySettings()
	candidates, err := m.buildAccessCandidates(proxySettings)
	if err != nil {
		return nil, nil, err
	}

	errors := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		targetURL, referer := build(candidate.Source.URL)
		body, latency, requestErr := do(targetURL, referer, candidate.UseProxy)
		if requestErr != nil {
			m.recordSourceAccess(candidate.Source.ID, candidate.UseProxy, AccessStatusUnavailable, latency, requestErr.Error())
			errors = append(errors, formatAttemptError(candidate.Source.URL, candidate.UseProxy, requestErr))
			continue
		}

		m.recordSourceAccess(candidate.Source.ID, candidate.UseProxy, AccessStatusAvailable, latency, "")
		return body, &AccessResult{
			Source:   candidate.Source,
			UseProxy: candidate.UseProxy,
			Latency:  latency,
		}, nil
	}

	if len(errors) == 0 {
		return nil, nil, fmt.Errorf("没有可用的源站，请先在设置中添加源站")
	}
	return nil, nil, fmt.Errorf("所有可用源站都访问失败: %s", strings.Join(errors, "; "))
}

func (m *Manager) CheckHealth(source *model.SourceSite) (*model.SourceHealthResponse, error) {
	if source == nil {
		return nil, fmt.Errorf("源站不存在")
	}

	proxySettings := m.GetProxySettings()
	proxyUsable := proxySettings.Enabled() && m.IsProxyAvailable(false)
	timeout := proxySettings.Timeout
	if timeout <= 0 {
		timeout = 6 * time.Second
	}

	var directLatency int
	var directErr error
	var proxyLatency int
	var proxyErr error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		directLatency, directErr = m.probeSource(source.URL, false, timeout, proxySettings)
		if directErr != nil {
			m.recordSourceAccess(source.ID, false, AccessStatusUnavailable, directLatency, directErr.Error())
			return
		}
		m.recordSourceAccess(source.ID, false, AccessStatusAvailable, directLatency, "")
	}()

	if proxyUsable {
		wg.Add(1)
		go func() {
			defer wg.Done()
			proxyLatency, proxyErr = m.probeSource(source.URL, true, timeout, proxySettings)
			if proxyErr != nil {
				m.recordSourceAccess(source.ID, true, AccessStatusUnavailable, proxyLatency, proxyErr.Error())
				return
			}
			m.recordSourceAccess(source.ID, true, AccessStatusAvailable, proxyLatency, "")
		}()
	} else if proxySettings.Enabled() {
		m.recordSourceAccess(source.ID, true, AccessStatusUnknown, 0, "代理当前不可用，已跳过代理测速")
	}

	wg.Wait()

	updated := m.findSourceByID(source.ID)
	response := &model.SourceHealthResponse{
		ID:            source.ID,
		URL:           source.URL,
		Status:        deriveSourceOverallStatus(updated),
		Latency:       choosePrimaryLatency(updated),
		DirectStatus:  updated.DirectStatus,
		DirectLatency: updated.DirectLatency,
		ProxyStatus:   updated.ProxyStatus,
		ProxyLatency:  updated.ProxyLatency,
	}
	if updated.DirectLastError != "" {
		response.DirectError = updated.DirectLastError
	}
	if updated.ProxyLastError != "" {
		response.ProxyError = updated.ProxyLastError
	}

	if directErr == nil || proxyErr == nil {
		return response, nil
	}
	if proxyUsable {
		return response, fmt.Errorf("直连失败: %v；代理失败: %v", directErr, proxyErr)
	}
	return response, directErr
}

func (m *Manager) buildAccessCandidates(proxySettings ProxySettings) ([]AccessCandidate, error) {
	sources := m.GetAll()
	if len(sources) == 0 {
		return nil, fmt.Errorf("暂无源站，请先在设置中添加源站")
	}

	mode := proxySettings.Mode
	if mode == ProxyModeAlways && !proxySettings.Enabled() {
		return nil, fmt.Errorf("代理模式已开启，但代理未配置完整")
	}
	proxyUsable := proxySettings.Enabled() && m.IsProxyAvailable(false)

	candidates := make([]AccessCandidate, 0, len(sources)*2)
	switch mode {
	case ProxyModeAlways:
		if proxyUsable {
			candidates = append(candidates, collectCandidates(sources, true)...)
		}
	case ProxyModeFallback:
		candidates = append(candidates, collectCandidates(sources, false)...)
		if proxyUsable {
			candidates = append(candidates, collectCandidates(sources, true)...)
		}
	default:
		candidates = append(candidates, collectCandidates(sources, false)...)
	}

	if len(candidates) == 0 {
		switch mode {
		case ProxyModeAlways:
			if proxySettings.Enabled() && !proxyUsable {
				return nil, fmt.Errorf("代理当前不可用，请先修复代理连接")
			}
			return nil, fmt.Errorf("当前没有可用的代理源站，请先手动检测或等待心跳恢复")
		case ProxyModeFallback:
			if proxySettings.Enabled() && proxyUsable {
				return nil, fmt.Errorf("当前直连源站和代理源站都不可用，请先手动检测或等待心跳恢复")
			}
			if proxySettings.Enabled() && !proxyUsable {
				return nil, fmt.Errorf("当前没有可用的直连源站，且代理当前不可用")
			}
			return nil, fmt.Errorf("当前没有可用的直连源站，请先手动检测或等待心跳恢复")
		default:
			return nil, fmt.Errorf("当前没有可用的直连源站，请先手动检测或等待心跳恢复")
		}
	}

	return candidates, nil
}

func collectCandidates(sources []*model.SourceSite, useProxy bool) []AccessCandidate {
	result := make([]AccessCandidate, 0, len(sources))

	appendMatching := func(expected string) {
		for _, source := range sources {
			status := source.DirectStatus
			if useProxy {
				status = source.ProxyStatus
			}
			status = normalizedAccessStatus(status)
			if status == expected {
				result = append(result, AccessCandidate{Source: source, UseProxy: useProxy})
			}
		}
	}

	appendMatching(AccessStatusAvailable)
	appendMatching(AccessStatusUnknown)
	return result
}

func (m *Manager) recordSourceAccess(sourceID int, useProxy bool, status string, latency int, lastError string) {
	var fields map[string]interface{}

	m.mu.Lock()
	var target *model.SourceSite
	for _, source := range m.sources {
		if source.ID == sourceID {
			target = source
			break
		}
	}

	if target == nil {
		m.mu.Unlock()
		return
	}

	now := time.Now()
	fields = map[string]interface{}{
		"last_check": &now,
	}
	if useProxy {
		target.ProxyStatus = status
		target.ProxyLatency = latency
		target.ProxyLastError = lastError
		fields["proxy_status"] = status
		fields["proxy_latency"] = latency
		fields["proxy_last_error"] = lastError
	} else {
		target.DirectStatus = status
		target.DirectLatency = latency
		target.DirectLastError = lastError
		fields["direct_status"] = status
		fields["direct_latency"] = latency
		fields["direct_last_error"] = lastError
	}

	target.LastCheck = &now
	target.Status = deriveSourceOverallStatus(target)
	target.Latency = choosePrimaryLatency(target)
	fields["status"] = target.Status
	fields["latency"] = target.Latency

	m.refreshActiveLocked()
	m.mu.Unlock()

	if err := m.repo.UpdateSourceFields(sourceID, fields); err != nil {
		// The in-memory state is still updated; health check will retry on next probe.
	}
}

func (m *Manager) refreshActiveLocked() {
	m.active = nil
	for _, source := range m.sources {
		if source.DirectStatus == AccessStatusAvailable {
			m.active = source
			return
		}
	}
}

func (m *Manager) findSourceByID(id int) *model.SourceSite {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, source := range m.sources {
		if source.ID == id {
			copySource := *source
			return &copySource
		}
	}
	return &model.SourceSite{ID: id}
}

func (m *Manager) probeSource(rawURL string, useProxy bool, timeout time.Duration, proxySettings ProxySettings) (int, error) {
	client, err := BuildHTTPClient(useProxy, proxySettings, timeout)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	start := time.Now()
	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())
	if err != nil {
		return latency, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return latency, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return latency, nil
}

func buildHTTPTransport(useProxy bool, settings ProxySettings, timeout time.Duration) (*http.Transport, error) {
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           (&net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   timeout,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: timeout,
	}

	if !useProxy {
		return transport, nil
	}
	if !settings.Enabled() {
		return nil, fmt.Errorf("代理未配置完整")
	}

	switch settings.Type {
	case "http":
		proxyURL := &url.URL{
			Scheme: "http",
			Host:   settings.Address(),
		}
		if settings.Username != "" {
			proxyURL.User = url.UserPassword(settings.Username, settings.Password)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
		return transport, nil
	case "socks5":
		var auth *xproxy.Auth
		if settings.Username != "" {
			auth = &xproxy.Auth{
				User:     settings.Username,
				Password: settings.Password,
			}
		}
		dialer, err := xproxy.SOCKS5("tcp", settings.Address(), auth, &net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second})
		if err != nil {
			return nil, err
		}
		if contextDialer, ok := dialer.(xproxy.ContextDialer); ok {
			transport.DialContext = contextDialer.DialContext
		} else {
			transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
				return dialer.Dial(network, address)
			}
		}
		return transport, nil
	default:
		return nil, fmt.Errorf("不支持的代理类型: %s", settings.Type)
	}
}

func deriveSourceOverallStatus(source *model.SourceSite) string {
	if source == nil {
		return SourceStatusUnknown
	}
	directStatus := normalizedAccessStatus(source.DirectStatus)
	proxyStatus := normalizedAccessStatus(source.ProxyStatus)
	if directStatus == AccessStatusAvailable {
		return SourceStatusActive
	}
	if proxyStatus == AccessStatusAvailable {
		return SourceStatusProxy
	}
	if directStatus == AccessStatusUnavailable || proxyStatus == AccessStatusUnavailable {
		return SourceStatusInactive
	}
	return SourceStatusUnknown
}

func choosePrimaryLatency(source *model.SourceSite) int {
	if source == nil {
		return 0
	}
	if normalizedAccessStatus(source.DirectStatus) == AccessStatusAvailable {
		return source.DirectLatency
	}
	if normalizedAccessStatus(source.ProxyStatus) == AccessStatusAvailable {
		return source.ProxyLatency
	}
	return 0
}

func normalizeProxyMode(mode string) string {
	switch strings.TrimSpace(strings.ToLower(mode)) {
	case ProxyModeAlways:
		return ProxyModeAlways
	case ProxyModeFallback:
		return ProxyModeFallback
	default:
		return ProxyModeOff
	}
}

func normalizeProxyType(proxyType string) string {
	switch strings.TrimSpace(strings.ToLower(proxyType)) {
	case "socks5":
		return "socks5"
	default:
		return "http"
	}
}

func formatAttemptError(sourceURL string, useProxy bool, err error) string {
	mode := "直连"
	if useProxy {
		mode = "代理"
	}
	return fmt.Sprintf("%s(%s): %v", sourceURL, mode, err)
}

func (m *Manager) cacheProxyHealth(response *model.ProxyHealthResponse) {
	if response == nil {
		return
	}
	copied := *response
	copied.Targets = append([]model.ProxyTargetProbeResponse(nil), response.Targets...)
	m.mu.Lock()
	m.proxyHealth = &copied
	m.proxyHealthAt = time.Now()
	m.mu.Unlock()
}

func probeProxyTarget(client *http.Client, name, rawURL string) model.ProxyTargetProbeResponse {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
	if err != nil {
		return model.ProxyTargetProbeResponse{Name: name, URL: rawURL, Status: AccessStatusUnavailable, Error: err.Error()}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")

	start := time.Now()
	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())
	if err != nil {
		return model.ProxyTargetProbeResponse{
			Name:    name,
			URL:     rawURL,
			Status:  AccessStatusUnavailable,
			Latency: latency,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return model.ProxyTargetProbeResponse{
			Name:    name,
			URL:     rawURL,
			Status:  AccessStatusUnavailable,
			Latency: latency,
			Error:   fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	return model.ProxyTargetProbeResponse{
		Name:    name,
		URL:     rawURL,
		Status:  AccessStatusAvailable,
		Latency: latency,
	}
}

func normalizedAccessStatus(status string) string {
	switch status {
	case AccessStatusAvailable:
		return AccessStatusAvailable
	case AccessStatusUnavailable:
		return AccessStatusUnavailable
	default:
		return AccessStatusUnknown
	}
}
