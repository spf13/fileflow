package fileflow

import (
	"io/fs"
	"syscall"
	"time"
)

// atime returns fi's access time, or the zero time if unavailable.
func atime(fi fs.FileInfo) time.Time {
	d, ok := fi.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return time.Time{}
	}
	return time.Unix(0, d.LastAccessTime.Nanoseconds())
}
