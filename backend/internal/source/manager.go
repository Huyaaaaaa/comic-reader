package source

import (
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/pkg/logger"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	maxFailCount        = 3
	healthCheckInterval = 60 * time.Minute
)

// Manager 源站管理器
type Manager struct {
	repo    *repository.Repository
	sources []*model.SourceSite
	active  *model.SourceSite
	mu      sync.RWMutex
	client  *http.Client
}

// NewManager 创建源站管理器
func NewManager(repo *repository.Repository) *Manager {
	return &Manager{
		repo: repo,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
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
		m.sources[i] = &sources[i]
	}

	for _, s := range m.sources {
		if s.Status == "active" {
			m.active = s
			break
		}
	}
	return nil
}

// GetActive 获取当前活跃源站
func (m *Manager) GetActive() *model.SourceSite {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.active
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
	source := &model.SourceSite{
		URL:      url,
		Name:     name,
		ImageCDN: imageCDN,
		Status:   "active",
		Priority: len(m.sources) + 1,
	}
	if err := m.repo.AddSource(source); err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.sources = append(m.sources, source)
	if m.active == nil {
		m.active = source
	}
	m.mu.Unlock()

	return source, nil
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
			if m.active != nil && m.active.ID == id {
				m.active = nil
				for _, s2 := range m.sources {
					if s2.Status == "active" {
						m.active = s2
						break
					}
				}
			}
			break
		}
	}
	return nil
}

// CheckHealth 检查单个源站健康状态
func (m *Manager) CheckHealth(s *model.SourceSite) (int, error) {
	start := time.Now()
	resp, err := m.client.Get(s.URL)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		m.repo.UpdateSourceStatus(s.ID, "failed", latency)
		return latency, fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		m.repo.UpdateSourceStatus(s.ID, "failed", latency)
		return latency, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	m.repo.UpdateSourceStatus(s.ID, "active", latency)
	m.repo.ResetSourceFailCount(s.ID)

	m.mu.Lock()
	s.Status = "active"
	s.Latency = latency
	now := time.Now()
	s.LastCheck = &now
	m.mu.Unlock()

	return latency, nil
}

// MarkFailure 记录源站失败
func (m *Manager) MarkFailure(sourceID int) error {
	m.repo.IncrementSourceFailCount(sourceID)

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, s := range m.sources {
		if s.ID == sourceID {
			s.FailCount++
			if s.FailCount >= maxFailCount {
				s.Status = "failed"
				m.repo.UpdateSourceStatus(s.ID, "failed", s.Latency)
				if m.active != nil && m.active.ID == sourceID {
					m.switchToNextLocked()
				}
			}
			break
		}
	}
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
		if s.Status == "active" && (m.active == nil || s.ID != m.active.ID) {
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
