package download

import (
	"fmt"
	"syscall"
)

// ReservedSpace 保留磁盘空间 1GB
const ReservedSpace int64 = 1 * 1024 * 1024 * 1024

// CheckDiskSpace 检查磁盘空间
func CheckDiskSpace(path string, estimatedSize int64) (availableSpace int64, err error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("获取磁盘信息失败: %w", err)
	}

	availableSpace = int64(stat.Bavail) * int64(stat.Bsize)

	if availableSpace < estimatedSize+ReservedSpace {
		return availableSpace, fmt.Errorf(
			"磁盘空间不足: 需要 %dMB, 可用 %dMB (保留 %dMB)",
			estimatedSize/1024/1024,
			availableSpace/1024/1024,
			ReservedSpace/1024/1024,
		)
	}

	return availableSpace, nil
}

// GetAvailableSpace 获取可用空间
func GetAvailableSpace(path string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("获取磁盘信息失败: %w", err)
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}
