package system

import "syscall"

type DiskMetrics struct {
	Path       string  `json:"path"`
	TotalBytes uint64  `json:"totalBytes"`
	FreeBytes  uint64  `json:"freeBytes"`
	UsedBytes  uint64  `json:"usedBytes"`
	UsagePct   float64 `json:"usagePct"`
}

func CollectDiskMetrics(path string) (DiskMetrics, error) {
	var stat syscall.Statfs_t

	if err := syscall.Statfs(path, &stat); err != nil {
		return DiskMetrics{}, err
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	var pct float64
	if total > 0 {
		pct = (float64(used) / float64(total)) * 100.0
	}

	return DiskMetrics{
		Path:       path,
		TotalBytes: total,
		FreeBytes:  free,
		UsedBytes:  used,
		UsagePct:   pct,
	}, nil
}
