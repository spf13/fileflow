package fileflow

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// renameExclusive renames src to dst with renameat2(RENAME_NOREPLACE), which
// fails with EEXIST if dst exists. It reports done=false, without changing
// anything, when the kernel or filesystem does not support the flag.
func renameExclusive(src, dst string) (done bool, err error) {
	err = unix.Renameat2(unix.AT_FDCWD, src, unix.AT_FDCWD, dst, unix.RENAME_NOREPLACE)
	if errors.Is(err, unix.ENOSYS) || errors.Is(err, unix.EINVAL) {
		return false, nil
	}
	if err != nil {
		return true, &os.LinkError{Op: "rename", Old: src, New: dst, Err: err}
	}
	return true, nil
}
