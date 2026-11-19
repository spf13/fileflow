//go:build !(linux || openbsd || dragonfly || solaris || aix || darwin || freebsd || netbsd || windows)

package fileflow

import (
	"io/fs"
	"time"
)

// atime returns the zero time, which os.Chtimes leaves unchanged, on
// platforms where the access time is not extracted.
func atime(fs.FileInfo) time.Time { return time.Time{} }
