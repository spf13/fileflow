package fileflow

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// renameExclusive renames src to dst with renamex_np(RENAME_EXCL), which
// fails with EEXIST if dst exists. It reports done=false, without changing
// anything, when the system or filesystem does not support the flag.
func renameExclusive(src, dst string) (done bool, err error) {
	err = unix.RenamexNp(src, dst, unix.RENAME_EXCL)
	if errors.Is(err, unix.ENOTSUP) || errors.Is(err, unix.ENOSYS) || errors.Is(err, unix.EINVAL) {
		return false, nil
	}
	if err != nil {
		return true, &os.LinkError{Op: "rename", Old: src, New: dst, Err: err}
	}
	return true, nil
}
