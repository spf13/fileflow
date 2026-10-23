//go:build windows

package fileflow

import (
	"os"

	"golang.org/x/sys/windows"
)

// renameReplace intentionally replaces dst, but never enables cross-volume copy:
// the caller stages cross-volume content before publishing it atomically.
func renameReplace(src, dst string) error {
	linkErr := func(err error) error {
		return &os.LinkError{Op: "rename", Old: src, New: dst, Err: err}
	}
	from, err := windows.UTF16PtrFromString(src)
	if err != nil {
		return linkErr(err)
	}
	to, err := windows.UTF16PtrFromString(dst)
	if err != nil {
		return linkErr(err)
	}
	if err := windows.MoveFileEx(from, to, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH); err != nil {
		return linkErr(err)
	}
	return nil
}
