package service

import (
	"archive/zip"
	"comic-viewer-claude/internal/model"
	"comic-viewer-claude/internal/repository"
	"comic-viewer-claude/pkg/config"
	"comic-viewer-claude/pkg/logger"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	maxExportSize       = 10 * 1024 * 1024 * 1024 // 10GB
	maxExportConcurrent = 3
)

// ImportExportService 导入导出服务
type ImportExportService struct {
	repo      *repository.Repository
	dlCfg     *config.DownloadConfig
	exportDir string
	importDir string
}

// NewImportExportService 创建导入导出服务
func NewImportExportService(repo *repository.Repository, dlCfg *config.DownloadConfig) *ImportExportService {
	exportDir := filepath.Join(dlCfg.Dir, "exports")
	os.MkdirAll(exportDir, 0755)
	importDir := filepath.Join(exportDir, "imports")
	os.MkdirAll(importDir, 0755)

	return &ImportExportService{
		repo:      repo,
		dlCfg:     dlCfg,
		exportDir: exportDir,
		importDir: importDir,
	}
}

// SaveImportUpload 保存前端上传的导入包，返回服务端文件路径
func (s *ImportExportService) SaveImportUpload(file multipart.File, filename string) (string, error) {
	safeName := filepath.Base(filename)
	if safeName == "." || safeName == string(filepath.Separator) || safeName == "" {
		safeName = "import.cpack"
	}

	targetName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName)
	targetPath := filepath.Join(s.importDir, targetName)

	dst, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("创建上传文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("保存上传文件失败: %w", err)
	}

	return targetPath, nil
}

// CreateExport 创建导出任务
func (s *ImportExportService) CreateExport(scope string, comicIDs []int) (*model.ImportExportJob, error) {
	// 并发检查
	running, _ := s.repo.GetRunningExportCount()
	if running >= maxExportConcurrent {
		return nil, fmt.Errorf("已有 %d 个导出任务在运行，最多 %d 个", running, maxExportConcurrent)
	}

	// 解析 scope 获取漫画ID
	resolvedIDs, err := s.resolveComicIDs(scope, comicIDs)
	if err != nil {
		return nil, err
	}
	if len(resolvedIDs) == 0 {
		return nil, fmt.Errorf("没有可导出的漫画")
	}

	optionsBytes, _ := json.Marshal(map[string]interface{}{
		"comic_ids": resolvedIDs,
	})

	now := time.Now()
	job := &model.ImportExportJob{
		Direction:   "export",
		Scope:       scope,
		Status:      "queued",
		OptionsJSON: string(optionsBytes),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateImportExportJob(job); err != nil {
		return nil, err
	}

	// 异步执行导出
	go s.runExport(job.ID, resolvedIDs)

	return job, nil
}

// GetJob 获取任务状态
func (s *ImportExportService) GetJob(id int) (*model.ImportExportJob, error) {
	return s.repo.GetImportExportJob(id)
}

// GetExportFilePath 获取导出文件路径
func (s *ImportExportService) GetExportFilePath(id int) (string, error) {
	job, err := s.repo.GetImportExportJob(id)
	if err != nil {
		return "", fmt.Errorf("任务不存在")
	}
	if job.Direction != "export" {
		return "", fmt.Errorf("不是导出任务")
	}
	if job.Status != "completed" {
		return "", fmt.Errorf("任务未完成 (status=%s)", job.Status)
	}
	if job.FilePath == "" || !fileExists(job.FilePath) {
		return "", fmt.Errorf("导出文件不存在")
	}
	return job.FilePath, nil
}

// ScanImport 预扫描导入包
func (s *ImportExportService) ScanImport(filePath string) (*model.ImportScanResponse, error) {
	if !fileExists(filePath) {
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	manifest, err := s.readManifest(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取导入包失败: %w", err)
	}

	resp := &model.ImportScanResponse{}
	for _, entry := range manifest.Comics {
		detail := model.ImportConflictDetail{
			ComicID:         entry.ID,
			Title:           entry.Title,
			ImportUpdatedAt: manifest.ExportedAt,
		}

		// 检查本地是否存在
		comic, err := s.repo.GetComic(entry.ID)
		if err != nil || comic.Title == "" {
			detail.ConflictType = "add"
			resp.Conflicts.WillAdd++
		} else {
			detail.ConflictType = "overwrite"
			detail.LocalUpdatedAt = comic.UpdatedAt.Format("2006-01-02 15:04:05")
			resp.Conflicts.WillOverwrite++
		}

		resp.Details = append(resp.Details, detail)
	}

	return resp, nil
}

// ExecuteImport 执行导入
func (s *ImportExportService) ExecuteImport(filePath, strategy string) (*model.ImportExportJob, error) {
	if strategy == "" {
		strategy = "skip_existing"
	}
	if strategy == "cancel" {
		return nil, fmt.Errorf("导入已取消")
	}

	if !fileExists(filePath) {
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	optionsBytes, _ := json.Marshal(map[string]interface{}{
		"strategy": strategy,
	})

	now := time.Now()
	job := &model.ImportExportJob{
		Direction:   "import",
		Scope:       "file",
		Status:      "queued",
		FilePath:    filePath,
		OptionsJSON: string(optionsBytes),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateImportExportJob(job); err != nil {
		return nil, err
	}

	// 异步执行导入
	go s.runImport(job.ID, filePath, strategy)

	return job, nil
}

// resolveComicIDs 根据 scope 解析漫画ID
func (s *ImportExportService) resolveComicIDs(scope string, comicIDs []int) ([]int, error) {
	switch scope {
	case "single_comic":
		if len(comicIDs) == 0 {
			return nil, fmt.Errorf("single_comic 需要指定 comic_ids")
		}
		return comicIDs, nil
	case "selected_offline_ready":
		return s.repo.GetOfflineReadyComicIDs()
	case "all_cached_comics":
		return s.repo.GetAllCachedComicIDs()
	case "all_covers":
		return s.repo.GetAllCachedComicIDs()
	case "all_images":
		return s.repo.GetAllCachedComicIDs()
	case "all_cached":
		return s.repo.GetAllCachedComicIDs()
	case "covers":
		return s.repo.GetAllCachedComicIDs()
	case "images":
		return s.repo.GetAllCachedComicIDs()
	default:
		return nil, fmt.Errorf("未知 scope: %s", scope)
	}
}

// runExport 异步执行导出
func (s *ImportExportService) runExport(jobID int, comicIDs []int) {
	s.repo.UpdateImportExportJobStatus(jobID, "running", nil)

	filename := fmt.Sprintf("export_%d_%s.cpack", jobID, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(s.exportDir, filename)

	// 创建 ZIP 文件
	zipFile, err := os.Create(filePath)
	if err != nil {
		s.failJob(jobID, fmt.Sprintf("创建导出文件失败: %v", err))
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	manifest := model.ExportManifest{
		Version:    "1.0",
		ExportedAt: time.Now().Format(time.RFC3339),
		Comics:     make([]model.ExportComicEntry, 0),
	}

	var totalSize int64
	dlDir := s.dlCfg.Dir

	for _, comicID := range comicIDs {
		comic, err := s.repo.GetComic(comicID)
		if err != nil {
			continue
		}

		entry := model.ExportComicEntry{
			ID:    comicID,
			Title: comic.Title,
			Files: make([]model.ExportFileEntry, 0),
		}

		// 打包图片文件
		comicDir := filepath.Join(dlDir, fmt.Sprintf("%d", comicID))
		if dirExists(comicDir) {
			files, _ := os.ReadDir(comicDir)
			for _, f := range files {
				if f.IsDir() {
					continue
				}

				srcPath := filepath.Join(comicDir, f.Name())
				info, err := f.Info()
				if err != nil {
					continue
				}

				arcPath := fmt.Sprintf("%d/%s", comicID, f.Name())
				w, err := zipWriter.Create(arcPath)
				if err != nil {
					continue
				}

				src, err := os.Open(srcPath)
				if err != nil {
					continue
				}

				written, _ := io.Copy(w, src)
				src.Close()

				entry.Files = append(entry.Files, model.ExportFileEntry{
					Path: arcPath,
					Size: info.Size(),
				})
				totalSize += written
			}
		}

		// 导出封面 base64（如有）
		if comic.CoverBase64 != "" {
			coverPath := fmt.Sprintf("%d/cover.b64", comicID)
			w, _ := zipWriter.Create(coverPath)
			w.Write([]byte(comic.CoverBase64))
			entry.Files = append(entry.Files, model.ExportFileEntry{
				Path: coverPath,
				Size: int64(len(comic.CoverBase64)),
			})
		}

		// 导出漫画元数据
		metaBytes, _ := json.Marshal(comic)
		metaPath := fmt.Sprintf("%d/metadata.json", comicID)
		w, _ := zipWriter.Create(metaPath)
		w.Write(metaBytes)
		entry.Files = append(entry.Files, model.ExportFileEntry{
			Path: metaPath,
			Size: int64(len(metaBytes)),
		})

		manifest.Comics = append(manifest.Comics, entry)
	}

	manifest.TotalSize = totalSize

	// 写入 manifest.json
	manifestBytes, _ := json.MarshalIndent(manifest, "", "  ")
	w, _ := zipWriter.Create("manifest.json")
	w.Write(manifestBytes)

	zipWriter.Close()
	zipFile.Close()

	// 获取文件大小
	info, _ := os.Stat(filePath)
	summaryBytes, _ := json.Marshal(map[string]interface{}{
		"comics_count": len(manifest.Comics),
		"total_size":   totalSize,
		"file_size":    info.Size(),
	})

	now := time.Now()
	s.repo.UpdateImportExportJobStatus(jobID, "completed", map[string]interface{}{
		"file_path":    filePath,
		"summary_json": string(summaryBytes),
		"finished_at":  &now,
	})

	logger.Info("导出完成", zap.Int("job_id", jobID), zap.Int("comics", len(manifest.Comics)))
}

// runImport 异步执行导入
func (s *ImportExportService) runImport(jobID int, filePath, strategy string) {
	s.repo.UpdateImportExportJobStatus(jobID, "running", nil)

	manifest, err := s.readManifest(filePath)
	if err != nil {
		s.failJob(jobID, fmt.Sprintf("读取导入包失败: %v", err))
		return
	}

	reader, err := zip.OpenReader(filePath)
	if err != nil {
		s.failJob(jobID, fmt.Sprintf("打开 ZIP 失败: %v", err))
		return
	}
	defer reader.Close()

	// 建立文件索引
	fileMap := make(map[string]*zip.File)
	for _, f := range reader.File {
		fileMap[f.Name] = f
	}

	dlDir := s.dlCfg.Dir
	imported := 0
	skipped := 0
	overwritten := 0
	conflictCount := 0

	// 使用事务
	tx := s.repo.GetDB().Begin()

	for _, entry := range manifest.Comics {
		// 检查本地是否存在
		existing, err := s.repo.GetComic(entry.ID)
		localExists := err == nil && existing.Title != ""

		if localExists && strategy == "skip_existing" {
			skipped++
			continue
		}

		if localExists {
			conflictCount++
			overwritten++
		}

		// 导入元数据
		metaPath := fmt.Sprintf("%d/metadata.json", entry.ID)
		if metaFile, ok := fileMap[metaPath]; ok {
			rc, err := metaFile.Open()
			if err == nil {
				data, _ := io.ReadAll(rc)
				rc.Close()

				var comic model.Comic
				if json.Unmarshal(data, &comic) == nil {
					if err := tx.Save(&comic).Error; err != nil {
						tx.Rollback()
						s.failJob(jobID, fmt.Sprintf("导入漫画 %d 元数据失败: %v", entry.ID, err))
						return
					}
				}
			}
		}

		// 导入图片文件
		comicDir := filepath.Join(dlDir, fmt.Sprintf("%d", entry.ID))
		os.MkdirAll(comicDir, 0755)

		for _, fe := range entry.Files {
			if strings.HasSuffix(fe.Path, "metadata.json") || strings.HasSuffix(fe.Path, "cover.b64") {
				continue
			}

			if zipFile, ok := fileMap[fe.Path]; ok {
				rc, err := zipFile.Open()
				if err != nil {
					continue
				}

				destName := filepath.Base(fe.Path)
				destPath := filepath.Join(comicDir, destName)
				out, err := os.Create(destPath)
				if err != nil {
					rc.Close()
					continue
				}

				io.Copy(out, rc)
				out.Close()
				rc.Close()
			}
		}

		// 导入封面
		coverPath := fmt.Sprintf("%d/cover.b64", entry.ID)
		if coverFile, ok := fileMap[coverPath]; ok {
			rc, err := coverFile.Open()
			if err == nil {
				data, _ := io.ReadAll(rc)
				rc.Close()
				tx.Model(&model.Comic{}).Where("id = ?", entry.ID).
					Update("cover_base64", string(data))
			}
		}

		imported++
	}

	if err := tx.Commit().Error; err != nil {
		s.failJob(jobID, fmt.Sprintf("事务提交失败: %v", err))
		return
	}

	summaryBytes, _ := json.Marshal(map[string]interface{}{
		"imported":    imported,
		"skipped":     skipped,
		"overwritten": overwritten,
	})

	now := time.Now()
	s.repo.UpdateImportExportJobStatus(jobID, "completed", map[string]interface{}{
		"summary_json":   string(summaryBytes),
		"conflict_count": conflictCount,
		"finished_at":    &now,
	})

	logger.Info("导入完成",
		zap.Int("job_id", jobID),
		zap.Int("imported", imported),
		zap.Int("skipped", skipped),
		zap.Int("overwritten", overwritten))
}

// readManifest 从 .cpack 读取 manifest.json
func (s *ImportExportService) readManifest(filePath string) (*model.ExportManifest, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	for _, f := range reader.File {
		if f.Name == "manifest.json" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, err
			}

			var manifest model.ExportManifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				return nil, err
			}
			return &manifest, nil
		}
	}

	return nil, fmt.Errorf("manifest.json not found in package")
}

// failJob 标记任务失败
func (s *ImportExportService) failJob(jobID int, errMsg string) {
	logger.Error("导入导出任务失败", zap.Int("job_id", jobID), zap.String("error", errMsg))
	s.repo.UpdateImportExportJobStatus(jobID, "failed", map[string]interface{}{
		"last_error": errMsg,
	})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
