package agent

import (
	"os"
	"syscall"
)

func hasStickyRootOwnedDirectory(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}
	return info.Mode()&os.ModeSticky != 0 && stat.Uid == 0
}
