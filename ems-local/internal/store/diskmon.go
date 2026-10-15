package store

import (
	"fmt"
	"os"
	"syscall"
)

// DiskUsage returns disk usage info for the database file.
type DiskUsage struct {
	DatabaseBytes  int64   `json:"database_bytes"`
	DatabaseMB     float64 `json:"database_mb"`
	DiskTotalBytes int64   `json:"disk_total_bytes,omitempty"`
	DiskFreeBytes  int64   `json:"disk_free_bytes,omitempty"`
	DiskUsagePct   float64 `json:"disk_usage_pct,omitempty"`
}

// DiskUsage calculates and returns the disk usage statistics.
func (s *Store) DiskUsage() (DiskUsage, error) {
	s.mu.Lock()
	path := s.path
	s.mu.Unlock()

	var usage DiskUsage

	if path != ":memory:" {
		info, err := os.Stat(path)
		if err == nil {
			usage.DatabaseBytes = info.Size()
			usage.DatabaseMB = float64(usage.DatabaseBytes) / (1024 * 1024)
		}

		var stat syscall.Statfs_t
		if err := syscall.Statfs(path, &stat); err == nil {
			bsize := uint64(stat.Bsize)
			blocks := uint64(stat.Blocks)
			bavail := uint64(stat.Bavail)
			bfree := uint64(stat.Bfree)

			usage.DiskTotalBytes = int64(blocks * bsize)
			usage.DiskFreeBytes = int64(bavail * bsize)

			if blocks > 0 {
				used := blocks - bfree
				usage.DiskUsagePct = (float64(used) / float64(blocks)) * 100.0
			}
		}
	}
	return usage, nil
}

// DiskAlert checks if the disk usage exceeds 80%.
func (s *Store) DiskAlert() (string, bool) {
	usage, err := s.DiskUsage()
	if err != nil {
		return fmt.Sprintf("error checking disk usage: %v", err), false
	}
	if usage.DiskUsagePct > 80.0 {
		return fmt.Sprintf("disk usage is high: %.1f%%", usage.DiskUsagePct), true
	}
	return "", false
}
