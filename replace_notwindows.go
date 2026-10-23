//go:build !windows

package fileflow

import "os"

// renameReplace publishes src at dst without an intermediate missing-path window.
// Cross-device errors are handled by the caller using a destination-side copy.
func renameReplace(src, dst string) error {
	return os.Rename(src, dst)
}
