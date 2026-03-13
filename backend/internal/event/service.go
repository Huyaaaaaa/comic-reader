package event

import (
	"comic-viewer-claude/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

// 事件类型常量
const (
	EventDownloadProgress = "download_progress"
	EventCacheProgress    = "cache_progress"
	EventSourceStatus     = "source_status"
	EventLog              = "log"
)

// Event SSE 事件
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Service SSE 事件服务
type Service struct {
	clients map[string]chan Event
	mu      sync.RWMutex
}

// NewService 创建事件服务
func NewService() *Service {
	return &Service{
		clients: make(map[string]chan Event),
	}
}

// Register 注册客户端，返回事件通道
func (s *Service) Register(clientID string) chan Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan Event, 64)
	s.clients[clientID] = ch
	logger.Info("SSE 客户端连接", zap.String("client_id", clientID), zap.Int("total", len(s.clients)))
	return ch
}

// Unregister 注销客户端
func (s *Service) Unregister(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.clients[clientID]; ok {
		close(ch)
		delete(s.clients, clientID)
		logger.Info("SSE 客户端断开", zap.String("client_id", clientID), zap.Int("total", len(s.clients)))
	}
}

// Broadcast 广播事件到所有客户端
func (s *Service) Broadcast(evt Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id, ch := range s.clients {
		select {
		case ch <- evt:
		default:
			logger.Warn("SSE 客户端通道已满，丢弃事件", zap.String("client_id", id), zap.String("type", evt.Type))
		}
	}
}

// ClientCount 返回当前连接的客户端数量
func (s *Service) ClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}
