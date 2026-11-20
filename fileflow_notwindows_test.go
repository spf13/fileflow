//go:build !windows

package fileflow

import "testing"

// TestRenameLink covers the link+remove fallback directly, since platforms
// with an atomic no-replace rename never reach it in normal operation.
func TestRenameLink(t *testing.T) {
	testRenameNoReplace(t, renameLink)
}
