//go:build linux

package fileflow

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

const (
	unprivSrcEnv = "FILEFLOW_UNPRIV_SRC"
	unprivDstEnv = "FILEFLOW_UNPRIV_DST"
)

// TestRenameNoReplaceUnprivileged runs renameNoReplace in a helper process
// as nobody, where Linux link restrictions apply, and checks that the rename
// is all-or-nothing: it never overwrites dst and never leaves dst as a hard
// link to a source it failed to remove.
func TestRenameNoReplaceUnprivileged(t *testing.T) {
	if src := os.Getenv(unprivSrcEnv); src != "" {
		err := renameNoReplace(src, os.Getenv(unprivDstEnv))
		fmt.Printf("exist=%v err=%v\n", errors.Is(err, fs.ErrExist), err)
		return
	}
	if os.Geteuid() != 0 {
		t.Skip("needs root to run a helper process as an unprivileged user")
	}

	t.Run("link not permitted", func(t *testing.T) {
		// With fs.protected_hardlinks=1, nobody may not link a file it can
		// only read, so link fails with EPERM.
		if b, err := os.ReadFile("/proc/sys/fs/protected_hardlinks"); err != nil || strings.TrimSpace(string(b)) != "1" {
			t.Skip("fs.protected_hardlinks is not enabled")
		}
		dir := unprivilegedDir(t)
		src := writeOwned(t, filepath.Join(dir, "src"), "new", 0644)
		dst := writeOwned(t, filepath.Join(dir, "dst"), "precious", 0644)
		out := runUnprivileged(t, dir, src, dst)
		assertContent(t, dst, "precious")
		assertContent(t, src, "new")
		if !strings.Contains(out, "exist=true") {
			t.Fatalf("want fs.ErrExist; helper output:\n%s", out)
		}
	})

	t.Run("source not removable", func(t *testing.T) {
		// nobody may link a world-writable file but, in a sticky directory,
		// may not remove it. A link+remove rename would leave dst as a hard
		// link to someone else's file.
		dir := unprivilegedDir(t)
		sticky := filepath.Join(dir, "sticky")
		if err := os.Mkdir(sticky, 0777); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(sticky, 0777|os.ModeSticky); err != nil {
			t.Fatal(err)
		}
		src := writeOwned(t, filepath.Join(sticky, "src"), "theirs", 0666)
		dst := filepath.Join(dir, "dst")
		out := runUnprivileged(t, dir, src, dst)
		if _, err := os.Lstat(dst); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("failed rename left %s behind (%v); helper output:\n%s", dst, err, out)
		}
		assertContent(t, src, "theirs")
		if strings.Contains(out, "err=<nil>") {
			t.Fatalf("want an error; helper output:\n%s", out)
		}
	})
}

// unprivilegedDir returns a temporary directory that nobody can write to.
func unprivilegedDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, d := range []string{filepath.Dir(dir), dir} {
		if err := os.Chmod(d, 0777); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// writeOwned writes a file owned by a uid that is neither root nor nobody.
func writeOwned(t *testing.T, path, data string, mode fs.FileMode) string {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chown(path, 12345, 12345); err != nil {
		t.Fatal(err)
	}
	return path
}

// runUnprivileged runs renameNoReplace(src, dst) as nobody using a copy of
// the test binary placed in dir, and returns the helper's output.
func runUnprivileged(t *testing.T, dir, src, dst string) string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	bin, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	helper := filepath.Join(dir, "helper.test")
	if err := os.WriteFile(helper, bin, 0755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(helper, "-test.run=^TestRenameNoReplaceUnprivileged$")
	cmd.Env = append(os.Environ(), unprivSrcEnv+"="+src, unprivDstEnv+"="+dst)
	cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: 65534, Gid: 65534}}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helper failed: %v\n%s", err, out)
	}
	return string(out)
}

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	if got, err := os.ReadFile(path); err != nil || string(got) != want {
		t.Fatalf("%s = %q, %v; want %q", path, got, err, want)
	}
}
