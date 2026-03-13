package crawler
import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"comic-viewer-claude/pkg/config"
	"comic-viewer-claude/pkg/utils"

	"github.com/go-resty/resty/v2"
)

// Client 爬虫客户端
type Client struct {
	client           *resty.Client
	config       *config.CrawlerConfig
	lastRequestTime  time.Time
	requestLock      sync.Mutex
	banBackoffUntil  time.Time
}

// NewClient 创建新的爬虫客户端
func NewClient(cfg *config.CrawlerConfig) *Client {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(5 * time.Second)

	return &Client{
		client: client,
		config: cfg,
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

// Get 发送GET请求
func (c *Client) Get(url string, referer string) (string, error) {
	c.requestLock.Lock()
	defer c.requestLock.Unlock()

	// 检查是否在冷却期
	now := time.Now()
	if now.Before(c.banBackoffUntil) {
		cooldown := c.banBackoffUntil.Sub(now)
		time.Sleep(cooldown)
	}

	// 请求延迟
	c.wait()

	// 发送请求
	resp, err := c.client.R().
		SetHeaders(c.getHeaders(referer)).
		Get(url)

	c.lastRequestTime = time.Now()

	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}

	// 检测风控状态码
	if resp.StatusCode() == 403 || resp.StatusCode() == 429 || resp.StatusCode() == 503 {
		// 进入冷却期
		cooldown := utils.RandomDelay(120, 300) // 2-5分钟
		c.banBackoffUntil = time.Now().Add(cooldown)
		return "", fmt.Errorf("触发风控，状态码: %d", resp.StatusCode())
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("HTTP错误: %d", resp.StatusCode())
	}

	return resp.String(), nil
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
	url := fmt.Sprintf("%s/dnew.php?t=%d&page=%d", c.config.ActiveBaseURL, timestamp, page)

	if tagID != nil {
		url += fmt.Sprintf("&tag_id=%d", *tagID)
	}
	if categoryID != nil {
		url += fmt.Sprintf("&category_id=%d", *categoryID)
	}
	if author != "" {
		url += fmt.Sprintf("&author=%s", author)
	}
	if authorID != nil {
		url += fmt.Sprintf("&author_id=%d", *authorID)
	}

	return c.Get(url, "")
}

// GetDetailPage 获取详情页
func (c *Client) GetDetailPage(comicID int) (string, error) {
	url := fmt.Sprintf("%s/post.php?ID=%d", c.config.ActiveBaseURL, comicID)
	return c.Get(url, "")
}

// GetReaderPage 获取阅读页
func (c *Client) GetReaderPage(comicID int) (string, error) {
	url := fmt.Sprintf("%s/readOnline2.php?ID=%d&host_id=0", c.config.ActiveBaseURL, comicID)
	referer := fmt.Sprintf("%s/post.php?ID=%d", c.config.ActiveBaseURL, comicID)
	return c.Get(url, referer)
}

// SearchOnline 在线搜索
func (c *Client) SearchOnline(keyword string, page int) (string, error) {
	timestamp := c.GetCurrentMonthTimestamp()
	url := fmt.Sprintf("%s/dnew.php?t=%d&page=%d&keyword=%s",
		c.config.ActiveBaseURL, timestamp, page, keyword)
	return c.Get(url, "")
}

// GetCoverURL 获取封面URL
func (c *Client) GetCoverURL(comicID int) string {
	idRange := c.getIDRange(comicID)
	return fmt.Sprintf("%s/image/comic/cover/thumbnail/h300/%s/%d.jpg?v=1",
		c.config.ImageCDN, idRange, comicID)
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
func (c *Client) DownloadImage(url string) ([]byte, error) {
	resp, err := c.client.R().
		SetHeaders(map[string]string{
		"User-Agent": c.randomUserAgent(),
			"Referer":    c.config.ActiveBaseURL + "/",
		}).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("下载图片失败: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("HTTP错误: %d", resp.StatusCode())
	}

	return resp.Body(), nil
}
