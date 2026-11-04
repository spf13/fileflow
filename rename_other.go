//go:build !windows && !linux && !darwin

package fileflow

// renameExclusive reports done=false: this platform has no atomic
// no-replace rename, so renameNoReplace uses link+remove.
func renameExclusive(src, dst string) (done bool, err error) {
	return false, nil
}
