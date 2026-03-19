package crawler

import (
	"comic-viewer-claude/internal/source"
	"comic-viewer-claude/pkg/config"
	"comic-viewer-claude/pkg/utils"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const imageRequestTimeout = 30 * time.Second

// Client 爬虫客户端
type Client struct {
	config          *config.CrawlerConfig
	sourceManager   *source.Manager
	lastRequestTime time.Time
	requestLock     sync.Mutex
	banBackoffUntil time.Time
}

// NewClient 创建新的爬虫客户端
func NewClient(cfg *config.CrawlerConfig, sourceManager *source.Manager) *Client {
	return &Client{
		config:        cfg,
		sourceManager: sourceManager,
	}
}

// getHeaders 获取请求头
func (c *Client) getHeaders(referer string) map[string]string {
	headers := map[string]string{
		"User-Agent":      c.randomUserAgent(),
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "zh-TW,zh;q=0.9,en-US;q=0.8,en;q=0.7",
	}
	if referer != "" {
		headers["Referer"] = referer
	}
	return headers
}

// randomUserAgent 随机选择User-Agent
func (c *Client) randomUserAgent() string {
	if len(c.config.UserAgents) == 0 {
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
	}
	return c.config.UserAgents[rand.Intn(len(c.config.UserAgents))]
}

// wait 请求前等待
func (c *Client) wait() {
	elapsed := time.Since(c.lastRequestTime)
	delay := utils.RandomDelay(c.config.RequestDelayMin, c.config.RequestDelayMax)
	if elapsed < delay {
		time.Sleep(delay - elapsed)
	}
}

// GetCurrentMonthTimestamp 获取当前月份第一天的时间戳
func (c *Client) GetCurrentMonthTimestamp() int64 {
	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 8, 0, 0, 0, now.Location())
	return firstDay.Unix()
}

// GetListPage 获取列表页
func (c *Client) GetListPage(page int, tagID, categoryID, authorID *int, author string) (string, error) {
	timestamp := c.GetCurrentMonthTimestamp()

	return c.fetchSourceHTML(func(baseURL string) (string, string) {
		targetURL := fmt.Sprintf("%s/dnew.php?t=%d&page=%d", baseURL, timestamp, page)
		if tagID != nil {
			targetURL += fmt.Sprintf("&tag_id=%d", *tagID)
		}
		if categoryID != nil {
			targetURL += fmt.Sprintf("&category_id=%d", *categoryID)
		}
		if author != "" {
			targetURL += fmt.Sprintf("&author=%s", author)
		}
		if authorID != nil {
			targetURL += fmt.Sprintf("&author_id=%d", *authorID)
		}
		return targetURL, ""
	})
}

// GetDetailPage 获取详情页
func (c *Client) GetDetailPage(comicID int) (string, error) {
	return c.fetchSourceHTML(func(baseURL string) (string, string) {
		targetURL := fmt.Sprintf("%s/post.php?ID=%d", baseURL, comicID)
		return targetURL, ""
	})
}

// GetReaderPage 获取阅读页
func (c *Client) GetReaderPage(comicID int) (string, error) {
	return c.fetchSourceHTML(func(baseURL string) (string, string) {
		targetURL := fmt.Sprintf("%s/readOnline2.php?ID=%d&host_id=0", baseURL, comicID)
		referer := fmt.Sprintf("%s/post.php?ID=%d", baseURL, comicID)
		return targetURL, referer
	})
}

// SearchOnline 在线搜索
func (c *Client) SearchOnline(keyword string, page int) (string, error) {
	timestamp := c.GetCurrentMonthTimestamp()
	return c.fetchSourceHTML(func(baseURL string) (string, string) {
		targetURL := fmt.Sprintf("%s/dnew.php?t=%d&page=%d&keyword=%s", baseURL, timestamp, page, keyword)
		return targetURL, ""
	})
}

// GetCoverURL 获取封面URL
func (c *Client) GetCoverURL(comicID int) string {
	idRange := c.getIDRange(comicID)
	return fmt.Sprintf("%s/image/comic/cover/thumbnail/h300/%s/%d.jpg?v=1", c.config.ImageCDN, idRange, comicID)
}

// getIDRange 获取ID范围
func (c *Client) getIDRange(id int) string {
	if id < 1000 {
		return "0-999"
	}
	start := (id / 1000) * 1000
	end := start + 999
	return fmt.Sprintf("%d-%d", start, end)
}

// DownloadImage 下载图片
func (c *Client) DownloadImage(rawURL string) ([]byte, error) {
	proxySettings := c.currentProxySettings()
	refererBase := ""
	if c.sourceManager != nil {
		refererBase = c.sourceManager.PreferredSourceBaseURL()
	}

	attempts, err := c.buildProxyAttempts(proxySettings)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, useProxy := range attempts {
		data, _, requestErr := c.doBinaryRequest(rawURL, refererBase, useProxy, imageRequestTimeout)
		if requestErr == nil {
			return data, nil
		}
		lastErr = requestErr
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("下载图片失败")
	}
	return nil, lastErr
}

func (c *Client) fetchSourceHTML(build func(baseURL string) (targetURL string, referer string)) (string, error) {
	if c.sourceManager == nil {
		return "", fmt.Errorf("源站管理器未初始化")
	}

	c.requestLock.Lock()
	defer c.requestLock.Unlock()

	now := time.Now()
	if now.Before(c.banBackoffUntil) {
		time.Sleep(c.banBackoffUntil.Sub(now))
	}
	c.wait()

	body, _, err := c.sourceManager.ExecuteRequest(build, func(targetURL, referer string, useProxy bool) ([]byte, int, error) {
		return c.doTextRequest(targetURL, referer, useProxy)
	})
	c.lastRequestTime = time.Now()
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *Client) doTextRequest(targetURL, referer string, useProxy bool) ([]byte, int, error) {
	timeout := c.currentSourceRequestTimeout()
	client, err := source.BuildHTTPClient(useProxy, c.currentProxySettings(), timeout)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, 0, err
	}
	for key, value := range c.getHeaders(referer) {
		req.Header.Set(key, value)
	}

	start := time.Now()
	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())
	if err != nil {
		return nil, latency, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
		cooldown := utils.RandomDelay(120, 300)
		c.banBackoffUntil = time.Now().Add(cooldown)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, latency, fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, latency, err
	}
	return body, latency, nil
}

func (c *Client) doBinaryRequest(rawURL, refererBase string, useProxy bool, timeout time.Duration) ([]byte, int, error) {
	client, err := source.BuildHTTPClient(useProxy, c.currentProxySettings(), timeout)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", c.randomUserAgent())
	if refererBase != "" {
		req.Header.Set("Referer", refererBase+"/")
	}

	start := time.Now()
	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())
	if err != nil {
		return nil, latency, fmt.Errorf("下载图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, latency, fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, latency, err
	}
	return body, latency, nil
}

func (c *Client) currentSourceRequestTimeout() time.Duration {
	proxySettings := c.currentProxySettings()
	if proxySettings.Timeout > 0 {
		return proxySettings.Timeout
	}
	return 6 * time.Second
}

func (c *Client) currentProxySettings() source.ProxySettings {
	if c.sourceManager == nil {
		return source.ProxySettings{Mode: source.ProxyModeOff, Timeout: 6 * time.Second}
	}
	return c.sourceManager.GetProxySettings()
}

func (c *Client) buildProxyAttempts(proxySettings source.ProxySettings) ([]bool, error) {
	proxyUsable := false
	if proxySettings.Enabled() && c.sourceManager != nil {
		proxyUsable = c.sourceManager.IsProxyAvailable(false)
	}
	switch proxySettings.Mode {
	case source.ProxyModeAlways:
		if !proxySettings.Enabled() {
			return nil, fmt.Errorf("代理模式已开启，但代理未配置完整")
		}
		if !proxyUsable {
			return nil, fmt.Errorf("代理当前不可用")
		}
		return []bool{true}, nil
	case source.ProxyModeFallback:
		if proxySettings.Enabled() && proxyUsable {
			return []bool{false, true}, nil
		}
		return []bool{false}, nil
	default:
		return []bool{false}, nil
	}
}
