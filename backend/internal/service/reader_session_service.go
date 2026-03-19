package service

import (
	"comic-viewer-claude/internal/crawler"
	"comic-viewer-claude/internal/repository"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	readerSessionTTL         = 30 * time.Minute
	readerSessionCleanupTick = 5 * time.Minute
)

type ReaderSessionImage struct {
	Sort      int    `json:"sort"`
	Filename  string `json:"filename"`
	Extension string `json:"extension"`
	Status    string `json:"status"`
	Ready     bool   `json:"ready"`
	Source    string `json:"source"`
	ViewURL   string `json:"view_url"`
	Error     string `json:"error,omitempty"`

	RemoteURL string `json:"-"`
	LocalPath string `json:"-"`
	TempPath  string `json:"-"`
}

type ReaderSessionSnapshot struct {
	SessionID  string               `json:"session_id"`
	ComicID    int                  `json:"comic_id"`
	ReadyCount int                  `json:"ready_count"`
	Total      int                  `json:"total"`
	CreatedAt  string               `json:"created_at"`
	UpdatedAt  string               `json:"updated_at"`
	Images     []ReaderSessionImage `json:"images"`
}

type readerSession struct {
	id        string
	comicID   int
	images    []ReaderSessionImage
	focusSort int
	createdAt time.Time
	updatedAt time.Time
	mu        sync.RWMutex
}

type ReaderSessionService struct {
	repo        *repository.Repository
	crawler     *crawler.Client
	downloadDir string
	tempRoot    string

	mu           sync.RWMutex
	sessions     map[string]*readerSession
	comicSession map[int]string
}

func NewReaderSessionService(repo *repository.Repository, crawler *crawler.Client, downloadDir, tempRoot string) *ReaderSessionService {
	if absTempRoot, err := filepath.Abs(tempRoot); err == nil {
		tempRoot = absTempRoot
	}
	if absDownloadDir, err := filepath.Abs(downloadDir); err == nil {
		downloadDir = absDownloadDir
	}

	_ = os.MkdirAll(tempRoot, 0755)

	s := &ReaderSessionService{
		repo:         repo,
		crawler:      crawler,
		downloadDir:  downloadDir,
		tempRoot:     tempRoot,
		sessions:     make(map[string]*readerSession),
		comicSession: make(map[int]string),
	}

	go s.cleanupLoop()
	return s
}

func (s *ReaderSessionService) CreateSession(comicID int) (*ReaderSessionSnapshot, error) {
	if existing := s.getReusableSession(comicID); existing != nil {
		s.touchSession(existing)
		s.UpdateFocus(existing.id, 1)
		return s.snapshot(existing), nil
	}

	dbImages, err := s.repo.GetComicImages(comicID)
	if err != nil {
		return nil, fmt.Errorf("获取阅读图片失败: %w", err)
	}
	if len(dbImages) == 0 {
		return nil, errors.New("阅读图片尚未准备好")
	}

	now := time.Now()
	sessionID := uuid.NewString()
	tempDir := filepath.Join(s.tempRoot, strconv.Itoa(comicID), sessionID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建阅读缓存目录失败: %w", err)
	}

	images := make([]ReaderSessionImage, 0, len(dbImages))
	for _, img := range dbImages {
		tempPath := filepath.Join(tempDir, fmt.Sprintf("%03d%s", img.Sort, img.Extension))
		state := ReaderSessionImage{
			Sort:      img.Sort,
			Filename:  img.Filename,
			Extension: img.Extension,
			Status:    "pending",
			Ready:     false,
			Source:    "remote",
			ViewURL:   fmt.Sprintf("/api/v2/reader/sessions/%s/images/%d", sessionID, img.Sort),
			RemoteURL: img.URL,
			LocalPath: s.normalizePath(img.LocalPath),
			TempPath:  tempPath,
		}

		if path, source := firstExistingPath(img.LocalPath, tempPath); path != "" {
			state.Ready = true
			state.Status = "ready"
			state.Source = source
		}

		images = append(images, state)
	}

	session := &readerSession{
		id:        sessionID,
		comicID:   comicID,
		images:    images,
		focusSort: 1,
		createdAt: now,
		updatedAt: now,
	}

	s.mu.Lock()
	s.sessions[sessionID] = session
	s.comicSession[comicID] = sessionID
	s.mu.Unlock()

	go s.prefetch(session)
	return s.snapshot(session), nil
}

func (s *ReaderSessionService) GetSession(sessionID string) (*ReaderSessionSnapshot, error) {
	session := s.getSession(sessionID)
	if session == nil {
		return nil, errors.New("阅读会话不存在")
	}
	s.touchSession(session)
	return s.snapshot(session), nil
}

func (s *ReaderSessionService) UpdateFocus(sessionID string, sortValue int) error {
	session := s.getSession(sessionID)
	if session == nil {
		return errors.New("阅读会话不存在")
	}

	session.mu.Lock()
	if sortValue < 1 {
		sortValue = 1
	}
	if len(session.images) > 0 && sortValue > len(session.images) {
		sortValue = len(session.images)
	}
	session.focusSort = sortValue
	session.updatedAt = time.Now()
	session.mu.Unlock()
	return nil
}

func (s *ReaderSessionService) ServeImagePath(sessionID string, sortValue int, waitFor time.Duration) (string, error) {
	session := s.getSession(sessionID)
	if session == nil {
		return "", errors.New("阅读会话不存在")
	}

	if err := s.UpdateFocus(sessionID, sortValue); err != nil {
		return "", err
	}

	deadline := time.Now().Add(waitFor)
	for {
		if path, err := s.resolveReadyPath(session, sortValue); err != nil || path != "" {
			return path, err
		}

		if time.Now().After(deadline) {
			return "", errors.New("图片尚未就绪")
		}
		time.Sleep(120 * time.Millisecond)
	}
}

func (s *ReaderSessionService) WaitForReusablePath(comicID, sortValue int, waitFor time.Duration) string {
	session := s.getReusableSession(comicID)
	if session == nil {
		return ""
	}

	deadline := time.Now().Add(waitFor)
	for {
		if path, err := s.resolveReadyPath(session, sortValue); err == nil && path != "" {
			return path
		}
		if time.Now().After(deadline) {
			return ""
		}
		time.Sleep(120 * time.Millisecond)
	}
}

func (s *ReaderSessionService) MarkPermanentReady(comicID, sortValue int, localPath string) {
	session := s.getReusableSession(comicID)
	if session == nil {
		return
	}

	localPath = s.normalizePath(localPath)

	session.mu.Lock()
	defer session.mu.Unlock()

	index := s.findImageIndexLocked(session, sortValue)
	if index < 0 {
		return
	}

	image := &session.images[index]
	image.LocalPath = localPath
	if readerFileExists(localPath) {
		image.Status = "ready"
		image.Ready = true
		image.Source = "local"
		image.Error = ""
	}
	session.updatedAt = time.Now()
}

func (s *ReaderSessionService) prefetch(session *readerSession) {
	for {
		index := s.nextPendingIndex(session)
		if index < 0 {
			return
		}
		s.downloadOne(session, index)
	}
}

func (s *ReaderSessionService) downloadOne(session *readerSession, index int) {
	session.mu.Lock()
	if index < 0 || index >= len(session.images) {
		session.mu.Unlock()
		return
	}
	image := session.images[index]
	if image.Ready {
		session.mu.Unlock()
		return
	}
	session.images[index].Status = "downloading"
	session.images[index].Error = ""
	session.updatedAt = time.Now()
	session.mu.Unlock()

	if path, source := firstExistingPath(image.LocalPath, image.TempPath); path != "" {
		session.mu.Lock()
		session.images[index].Status = "ready"
		session.images[index].Ready = true
		session.images[index].Source = source
		session.updatedAt = time.Now()
		session.mu.Unlock()
		return
	}

	data, err := s.crawler.DownloadImage(image.RemoteURL)
	if err != nil {
		session.mu.Lock()
		session.images[index].Status = "failed"
		session.images[index].Ready = false
		session.images[index].Error = err.Error()
		session.updatedAt = time.Now()
		session.mu.Unlock()
		return
	}

	if err := os.WriteFile(image.TempPath, data, 0644); err != nil {
		session.mu.Lock()
		session.images[index].Status = "failed"
		session.images[index].Ready = false
		session.images[index].Error = err.Error()
		session.updatedAt = time.Now()
		session.mu.Unlock()
		return
	}

	session.mu.Lock()
	session.images[index].Status = "ready"
	session.images[index].Ready = true
	session.images[index].Source = "temp"
	session.images[index].Error = ""
	session.updatedAt = time.Now()
	session.mu.Unlock()
}

func (s *ReaderSessionService) nextPendingIndex(session *readerSession) int {
	session.mu.RLock()
	defer session.mu.RUnlock()

	pending := make(map[int]int)
	sorts := make([]int, 0)
	for i, image := range session.images {
		if image.Ready || image.Status == "failed" {
			continue
		}
		pending[image.Sort] = i
		sorts = append(sorts, image.Sort)
	}
	if len(pending) == 0 {
		return -1
	}

	sort.Ints(sorts)
	seen := make(map[int]struct{})
	for _, candidate := range s.priorityOrder(session, sorts) {
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		if index, ok := pending[candidate]; ok {
			return index
		}
	}

	return pending[sorts[0]]
}

func (s *ReaderSessionService) priorityOrder(session *readerSession, sorts []int) []int {
	order := make([]int, 0, len(sorts)+8)
	appendSort := func(value int) {
		if value > 0 {
			order = append(order, value)
		}
	}

	appendSort(1)
	appendSort(2)
	appendSort(3)
	appendSort(4)

	focus := session.focusSort
	appendSort(focus)
	appendSort(focus + 1)
	appendSort(focus + 2)
	appendSort(focus - 1)

	order = append(order, sorts...)
	return order
}

func (s *ReaderSessionService) resolveReadyPath(session *readerSession, sortValue int) (string, error) {
	session.mu.Lock()
	defer session.mu.Unlock()

	index := s.findImageIndexLocked(session, sortValue)
	if index < 0 {
		return "", errors.New("图片不存在")
	}

	image := &session.images[index]
	if path, source := firstExistingPath(image.LocalPath, image.TempPath); path != "" {
		image.Status = "ready"
		image.Ready = true
		image.Source = source
		image.Error = ""
		session.updatedAt = time.Now()
		return path, nil
	}

	if image.Status == "failed" {
		return "", errors.New(image.Error)
	}

	return "", nil
}

func (s *ReaderSessionService) snapshot(session *readerSession) *ReaderSessionSnapshot {
	session.mu.RLock()
	defer session.mu.RUnlock()

	images := make([]ReaderSessionImage, len(session.images))
	copy(images, session.images)

	readyCount := 0
	for i := range images {
		if path, source := firstExistingPath(images[i].LocalPath, images[i].TempPath); path != "" {
			images[i].Ready = true
			images[i].Status = "ready"
			images[i].Source = source
		}
		if images[i].Ready {
			readyCount++
		}
	}

	return &ReaderSessionSnapshot{
		SessionID:  session.id,
		ComicID:    session.comicID,
		ReadyCount: readyCount,
		Total:      len(images),
		CreatedAt:  session.createdAt.Format(time.RFC3339),
		UpdatedAt:  session.updatedAt.Format(time.RFC3339),
		Images:     images,
	}
}

func (s *ReaderSessionService) cleanupLoop() {
	ticker := time.NewTicker(readerSessionCleanupTick)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupExpiredSessions()
	}
}

func (s *ReaderSessionService) cleanupExpiredSessions() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for sessionID, session := range s.sessions {
		session.mu.RLock()
		expired := now.Sub(session.updatedAt) > readerSessionTTL
		comicID := session.comicID
		tempDir := filepath.Dir(session.images[0].TempPath)
		session.mu.RUnlock()

		if !expired {
			continue
		}

		delete(s.sessions, sessionID)
		if currentID, ok := s.comicSession[comicID]; ok && currentID == sessionID {
			delete(s.comicSession, comicID)
		}
		_ = os.RemoveAll(tempDir)
	}
}

func (s *ReaderSessionService) getReusableSession(comicID int) *readerSession {
	s.mu.RLock()
	sessionID, ok := s.comicSession[comicID]
	if !ok {
		s.mu.RUnlock()
		return nil
	}
	session := s.sessions[sessionID]
	s.mu.RUnlock()

	if session == nil || s.isExpired(session) {
		s.removeSession(comicID, sessionID)
		return nil
	}
	return session
}

func (s *ReaderSessionService) getSession(sessionID string) *readerSession {
	s.mu.RLock()
	session := s.sessions[sessionID]
	s.mu.RUnlock()

	if session == nil || s.isExpired(session) {
		if session != nil {
			s.removeSession(session.comicID, sessionID)
		}
		return nil
	}
	return session
}

func (s *ReaderSessionService) removeSession(comicID int, sessionID string) {
	s.mu.Lock()
	session := s.sessions[sessionID]
	delete(s.sessions, sessionID)
	if currentID, ok := s.comicSession[comicID]; ok && currentID == sessionID {
		delete(s.comicSession, comicID)
	}
	s.mu.Unlock()

	if session == nil || len(session.images) == 0 {
		return
	}
	_ = os.RemoveAll(filepath.Dir(session.images[0].TempPath))
}

func (s *ReaderSessionService) isExpired(session *readerSession) bool {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return time.Since(session.updatedAt) > readerSessionTTL
}

func (s *ReaderSessionService) touchSession(session *readerSession) {
	session.mu.Lock()
	session.updatedAt = time.Now()
	session.mu.Unlock()
}

func (s *ReaderSessionService) findImageIndexLocked(session *readerSession, sortValue int) int {
	for i, image := range session.images {
		if image.Sort == sortValue {
			return i
		}
	}
	return -1
}

func firstExistingPath(localPath, tempPath string) (string, string) {
	if readerFileExists(localPath) {
		return localPath, "local"
	}
	if readerFileExists(tempPath) {
		return tempPath, "temp"
	}
	return "", ""
}

func readerFileExists(path string) bool {
	if path == "" {
		return false
	}
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func (s *ReaderSessionService) normalizePath(path string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	if absPath, err := filepath.Abs(path); err == nil {
		return absPath
	}
	return filepath.Join(s.downloadDir, path)
}
