package source

import (
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/parser"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/pkg/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

const (
	maxFailCount             = 3
	healthCheckInterval      = 60 * time.Minute
	sourceFingerprintMetaKey = "source_validation_fingerprints"
	sourceFingerprintLimit   = 3
	releasePageFetchTimeout  = 6 * time.Second
	sourceValidateTimeout    = 3 * time.Second
	sourceImportWorkerCount  = 4
)

// Manager 源站管理器
type Manager struct {
	repo          *repository.Repository
	sources       []*model.SourceSite
	active        *model.SourceSite
	mu            sync.RWMutex
	proxySettings ProxySettings
	proxyLoadedAt time.Time
	proxyHealth   *model.ProxyHealthResponse
	proxyHealthAt time.Time
}

// NewManager 创建源站管理器
func NewManager(repo *repository.Repository) *Manager {
	return &Manager{
		repo: repo,
	}
}

// Load 从数据库加载源站列表
func (m *Manager) Load() error {
	sources, err := m.repo.GetAllSources()
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sources = make([]*model.SourceSite, len(sources))
	for i := range sources {
		sources[i].DirectStatus = normalizedAccessStatus(sources[i].DirectStatus)
		sources[i].ProxyStatus = normalizedAccessStatus(sources[i].ProxyStatus)
		sources[i].Status = deriveSourceOverallStatus(&sources[i])
		sources[i].Latency = choosePrimaryLatency(&sources[i])
		m.sources[i] = &sources[i]
	}
	m.refreshActiveLocked()
	return nil
}

// GetAll 获取所有源站
func (m *Manager) GetAll() []*model.SourceSite {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*model.SourceSite, len(m.sources))
	copy(result, m.sources)
	return result
}

// Add 添加源站
func (m *Manager) Add(url, name, imageCDN string) (*model.SourceSite, error) {
	normalizedURL, err := normalizeSourceURL(url)
	if err != nil {
		return nil, err
	}

	m.mu.RLock()
	for _, existing := range m.sources {
		if normalizeSourceURLOrOriginal(existing.URL) == normalizedURL {
			m.mu.RUnlock()
			return nil, fmt.Errorf("源站已存在: %s", normalizedURL)
		}
	}
	priority := len(m.sources) + 1
	m.mu.RUnlock()

	source := &model.SourceSite{
		URL:          normalizedURL,
		Name:         name,
		ImageCDN:     imageCDN,
		Status:       SourceStatusUnknown,
		Priority:     priority,
		DirectStatus: AccessStatusUnknown,
		ProxyStatus:  AccessStatusUnknown,
	}
	if err := m.repo.AddSource(source); err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.sources = append(m.sources, source)
	m.refreshActiveLocked()
	m.mu.Unlock()

	return source, nil
}

// ImportFromReleasePage 从发布页导入可用源站
func (m *Manager) ImportFromReleasePage(releasePageURL string) (*model.ImportReleasePageSourcesResponse, error) {
	normalizedReleaseURL, err := normalizeSourceURL(releasePageURL)
	if err != nil {
		return nil, fmt.Errorf("发布页地址无效: %w", err)
	}

	fingerprints, err := m.ensureValidationFingerprints()
	if err != nil {
		return nil, err
	}
	if len(fingerprints) == 0 {
		return nil, fmt.Errorf("暂无可用的本地漫画校验指纹，请先浏览并保存几部漫画详情后再导入")
	}

	html, err := m.fetchHTMLWithTimeout(normalizedReleaseURL, releasePageFetchTimeout)
	if err != nil {
		return nil, fmt.Errorf("获取发布页失败: %w", err)
	}

	candidateURLs, err := extractReleasePageLinks(html)
	if err != nil {
		return nil, fmt.Errorf("解析发布页失败: %w", err)
	}

	result := &model.ImportReleasePageSourcesResponse{
		ReleasePageURL: normalizedReleaseURL,
		CandidateCount: len(candidateURLs),
		Fingerprints:   fingerprints,
		Added:          make([]model.SourceSite, 0),
		Skipped:        make([]model.SourceImportSkipped, 0),
	}

	type pendingCandidate struct {
		index int
		url   string
	}
	type candidateOutcome struct {
		index   int
		added   *model.SourceSite
		skipped *model.SourceImportSkipped
	}

	pending := make([]pendingCandidate, 0, len(candidateURLs))
	for _, candidateURL := range candidateURLs {
		normalizedCandidateURL, normalizeErr := normalizeSourceURL(candidateURL)
		if normalizeErr != nil {
			result.Skipped = append(result.Skipped, model.SourceImportSkipped{
				URL:    candidateURL,
				Reason: "地址格式无效",
			})
			continue
		}

		if normalizeSourceURLOrOriginal(normalizedReleaseURL) == normalizedCandidateURL {
			result.Skipped = append(result.Skipped, model.SourceImportSkipped{
				URL:    normalizedCandidateURL,
				Reason: "发布页地址本身不是源站",
			})
			continue
		}

		if m.hasSource(normalizedCandidateURL) {
			result.Skipped = append(result.Skipped, model.SourceImportSkipped{
				URL:    normalizedCandidateURL,
				Reason: "多源站列表中已存在",
			})
			continue
		}

		pending = append(pending, pendingCandidate{
			index: len(pending),
			url:   normalizedCandidateURL,
		})
	}

	if len(pending) > 0 {
		workerCount := sourceImportWorkerCount
		if len(pending) < workerCount {
			workerCount = len(pending)
		}

		jobs := make(chan pendingCandidate)
		outcomes := make(chan candidateOutcome, len(pending))
		var wg sync.WaitGroup

		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for candidate := range jobs {
					outcome := candidateOutcome{index: candidate.index}

					validationErr := m.validateCandidateSource(candidate.url, fingerprints)
					if validationErr != nil {
						outcome.skipped = &model.SourceImportSkipped{
							URL:    candidate.url,
							Reason: validationErr.Error(),
						}
						outcomes <- outcome
						continue
					}

					sourceName := inferSourceName(candidate.url)
					addedSource, addErr := m.Add(candidate.url, sourceName, "")
					if addErr != nil {
						outcome.skipped = &model.SourceImportSkipped{
							URL:    candidate.url,
							Reason: "添加源站失败: " + addErr.Error(),
						}
						outcomes <- outcome
						continue
					}

					outcome.added = addedSource
					outcomes <- outcome
				}
			}()
		}

		go func() {
			for _, candidate := range pending {
				jobs <- candidate
			}
			close(jobs)
			wg.Wait()
			close(outcomes)
		}()

		orderedOutcomes := make([]candidateOutcome, 0, len(pending))
		for outcome := range outcomes {
			orderedOutcomes = append(orderedOutcomes, outcome)
		}
		sort.Slice(orderedOutcomes, func(i, j int) bool {
			return orderedOutcomes[i].index < orderedOutcomes[j].index
		})

		for _, outcome := range orderedOutcomes {
			if outcome.added != nil {
				result.Added = append(result.Added, *outcome.added)
			}
			if outcome.skipped != nil {
				result.Skipped = append(result.Skipped, *outcome.skipped)
			}
		}
	}

	result.AddedCount = len(result.Added)
	result.SkippedCount = len(result.Skipped)
	return result, nil
}

// Remove 删除源站
func (m *Manager) Remove(id int) error {
	if err := m.repo.DeleteSource(id); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for i, s := range m.sources {
		if s.ID == id {
			m.sources = append(m.sources[:i], m.sources[i+1:]...)
			m.refreshActiveLocked()
			break
		}
	}
	return nil
}

// MarkFailure 记录源站失败
func (m *Manager) MarkFailure(sourceID int) error {
	m.recordSourceAccess(sourceID, false, AccessStatusUnavailable, 0, "请求失败")
	return nil
}

// SwitchToNext 切换到下一个可用源站
func (m *Manager) SwitchToNext() (*model.SourceSite, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.switchToNextLocked()
}

func (m *Manager) switchToNextLocked() (*model.SourceSite, error) {
	for _, s := range m.sources {
		if s.DirectStatus == AccessStatusAvailable && (m.active == nil || s.ID != m.active.ID) {
			m.active = s
			logger.Info("源站切换", zap.String("url", s.URL))
			return s, nil
		}
	}
	return nil, fmt.Errorf("没有可用的源站")
}

// StartHealthChecker 启动定时健康检查
func (m *Manager) StartHealthChecker(stopCh <-chan struct{}) {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.RLock()
			sources := make([]*model.SourceSite, len(m.sources))
			copy(sources, m.sources)
			m.mu.RUnlock()

			for _, s := range sources {
				if _, err := m.CheckHealth(s); err != nil {
					logger.Warn("源站健康检查失败", zap.String("url", s.URL), zap.Error(err))
				}
			}
		case <-stopCh:
			return
		}
	}
}

func (m *Manager) hasSource(rawURL string) bool {
	normalizedURL := normalizeSourceURLOrOriginal(rawURL)

	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, source := range m.sources {
		if normalizeSourceURLOrOriginal(source.URL) == normalizedURL {
			return true
		}
	}
	return false
}

func (m *Manager) ensureValidationFingerprints() ([]model.SourceValidationFingerprint, error) {
	var stored []model.SourceValidationFingerprint
	if raw, err := m.repo.GetSystemMetadata(sourceFingerprintMetaKey); err == nil && strings.TrimSpace(raw) != "" {
		if unmarshalErr := json.Unmarshal([]byte(raw), &stored); unmarshalErr != nil {
			logger.Warn("读取源站校验特征点失败，将回退到本地漫画重新生成", zap.Error(unmarshalErr))
			stored = nil
		}
	}

	fresh, err := m.repo.GetSourceValidationFingerprints(sourceFingerprintLimit)
	if err != nil {
		return nil, fmt.Errorf("读取本地漫画特征点失败: %w", err)
	}
	fingerprints := mergeValidationFingerprints(fresh, stored, sourceFingerprintLimit)
	if len(fingerprints) == 0 {
		return nil, nil
	}

	payload, err := json.Marshal(fingerprints)
	if err == nil {
		if saveErr := m.repo.SetSystemMetadata(sourceFingerprintMetaKey, string(payload)); saveErr != nil {
			logger.Warn("保存源站校验特征点失败", zap.Error(saveErr))
		}
	}

	return fingerprints, nil
}

func (m *Manager) validateCandidateSource(baseURL string, fingerprints []model.SourceValidationFingerprint) error {
	homeHTML, err := m.fetchHTMLWithTimeout(baseURL, sourceValidateTimeout)
	if err != nil {
		return fmt.Errorf("首页不可访问")
	}
	if !matchesSourceMarkers(homeHTML) {
		return fmt.Errorf("首页特征不匹配")
	}

	for _, fingerprint := range fingerprints {
		detailHTML, detailErr := m.fetchHTMLWithTimeout(buildDetailURL(baseURL, fingerprint.ComicID), sourceValidateTimeout)
		if detailErr != nil {
			continue
		}

		detail, parseErr := parser.ParseDetailPage(detailHTML, fingerprint.ComicID)
		if parseErr != nil {
			continue
		}

		if matchesFingerprint(detail, fingerprint) {
			return nil
		}
	}

	return fmt.Errorf("内容指纹校验失败")
}

func (m *Manager) fetchHTML(rawURL string) (string, error) {
	return m.fetchHTMLWithTimeout(rawURL, 0)
}

func (m *Manager) fetchHTMLWithTimeout(rawURL string, timeout time.Duration) (string, error) {
	proxySettings := m.GetProxySettings()
	if timeout <= 0 {
		timeout = proxySettings.Timeout
	}
	if timeout <= 0 {
		timeout = 6 * time.Second
	}

	tryFetch := func(useProxy bool) (string, error) {
		client, err := BuildHTTPClient(useProxy, proxySettings, timeout)
		if err != nil {
			return "", err
		}

		req, err := http.NewRequest(http.MethodGet, rawURL, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusBadRequest {
			return "", fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}

	html, err := tryFetch(false)
	if err == nil {
		return html, nil
	}
	if !proxySettings.Enabled() {
		return "", err
	}
	return tryFetch(true)
}

func extractReleasePageLinks(html string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	links := make([]string, 0)
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}

		normalized, normalizeErr := normalizeSourceURL(href)
		if normalizeErr != nil {
			return
		}
		if _, exists := seen[normalized]; exists {
			return
		}

		seen[normalized] = struct{}{}
		links = append(links, normalized)
	})

	return links, nil
}

func matchesSourceMarkers(html string) bool {
	markers := []string{
		"catalog/view/theme/default/css/bootstrap.min.css",
		"catalog/view/theme/default/css/smartadmin-production.min.css",
		"catalog/view/core/crypto/aes.js",
		"var aei =",
	}

	matched := 0
	for _, marker := range markers {
		if strings.Contains(html, marker) {
			matched++
		}
	}
	return matched >= 2
}

func matchesFingerprint(detail *model.ComicDetail, fingerprint model.SourceValidationFingerprint) bool {
	if detail == nil {
		return false
	}

	titleMatch := normalizeText(detail.Title) == normalizeText(fingerprint.Title)
	authorMatch := normalizeText(detail.Author) == normalizeText(fingerprint.Author) ||
		strings.Contains(normalizeText(detail.Author), normalizeText(fingerprint.Author)) ||
		strings.Contains(normalizeText(fingerprint.Author), normalizeText(detail.Author))

	return titleMatch && authorMatch
}

func buildDetailURL(baseURL string, comicID int) string {
	return fmt.Sprintf("%s/post.php?ID=%d", strings.TrimRight(baseURL, "/"), comicID)
}

func inferSourceName(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "自动导入源站"
	}
	return parsed.Host
}

func normalizeSourceURL(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("缺少协议或域名")
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawQuery = ""
	parsed.Fragment = ""

	if parsed.Path == "" {
		return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host), nil
	}
	return fmt.Sprintf("%s://%s%s", parsed.Scheme, parsed.Host, parsed.Path), nil
}

func normalizeSourceURLOrOriginal(rawURL string) string {
	normalized, err := normalizeSourceURL(rawURL)
	if err != nil {
		return strings.TrimSpace(rawURL)
	}
	return normalized
}

func normalizeText(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(strings.ToLower(text))), "")
}

func mergeValidationFingerprints(primary []model.SourceValidationFingerprint, secondary []model.SourceValidationFingerprint, limit int) []model.SourceValidationFingerprint {
	if limit <= 0 {
		return nil
	}

	merged := make([]model.SourceValidationFingerprint, 0, limit)
	seen := make(map[string]struct{}, limit)

	appendUnique := func(items []model.SourceValidationFingerprint) {
		for _, item := range items {
			if len(merged) >= limit {
				return
			}
			if item.ComicID <= 0 || strings.TrimSpace(item.Title) == "" || strings.TrimSpace(item.Author) == "" {
				continue
			}

			key := fmt.Sprintf("%d|%s|%s", item.ComicID, normalizeText(item.Title), normalizeText(item.Author))
			if _, exists := seen[key]; exists {
				continue
			}

			seen[key] = struct{}{}
			merged = append(merged, item)
		}
	}

	appendUnique(primary)
	appendUnique(secondary)
	return merged
}
