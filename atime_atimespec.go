//go:build darwin || freebsd || netbsd

package fileflow

import (
	"io/fs"
	"syscall"
	"time"
)

// atime returns fi's access time, or the zero time if unavailable.
func atime(fi fs.FileInfo) time.Time {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return time.Time{}
	}
	return time.Unix(int64(st.Atimespec.Sec), int64(st.Atimespec.Nsec))
}
