// Command smoke exercises fileflow's golden path and known edge cases and
// prints one deterministic line per scenario. Compare against expected.txt.
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/fileflow"
)

var root string

func write(rel, data string) string {
	p := filepath.Join(root, rel)
	must(os.MkdirAll(filepath.Dir(p), 0o755))
	must(os.WriteFile(p, []byte(data), 0o644))
	return p
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func rel(p string) string {
	if p == "" {
		return `""`
	}
	r, err := filepath.Rel(root, p)
	if err != nil || strings.HasPrefix(r, "..") {
		return "<outside>/" + filepath.Base(p)
	}
	return r
}

func exists(rel string) bool {
	_, err := os.Lstat(filepath.Join(root, rel))
	return err == nil
}

func content(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return "<missing>"
	}
	return string(b)
}

func errClass(err error) string {
	var mv *fileflow.ErrFailedMovingFile
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, fileflow.ErrMaxAttemptsReached):
		return "ErrMaxAttemptsReached"
	case errors.As(err, &mv) && errors.Is(err, fs.ErrExist):
		return "ErrFailedMovingFile+ErrExist"
	case errors.Is(err, fs.ErrNotExist):
		return "ErrNotExist"
	default:
		return "other"
	}
}

func main() {
	var err error
	root, err = os.MkdirTemp("", "fileflow-smoke-")
	must(err)
	defer os.RemoveAll(root)

	src := write("a.txt", "A")
	p, err := fileflow.Move(src, filepath.Join(root, "out/a.txt"))
	fmt.Printf("move-plain: %s %s src=%v data=%s\n", rel(p), errClass(err), exists("a.txt"), content(p))

	src = write("b.txt", "B")
	p, err = fileflow.Move(src, filepath.Join(root, "out/a.txt"))
	fmt.Printf("move-conflict: %s %s data=%s\n", rel(p), errClass(err), content(p))

	src = write("c.txt", "A")
	p, err = fileflow.Move(src, filepath.Join(root, "out/a.txt"))
	fmt.Printf("move-identical: %s %s src=%v\n", rel(p), errClass(err), exists("c.txt"))

	src = write("d.txt", "D")
	p, err = fileflow.Copy(src, filepath.Join(root, "out/a.txt"))
	fmt.Printf("copy-conflict: %s %s src=%v data=%s\n", rel(p), errClass(err), exists("d.txt"), content(p))

	p, err = fileflow.Copy(src, filepath.Join(root, "out/a-2.txt"))
	fmt.Printf("copy-identical: %s %s\n", rel(p), errClass(err))

	if shm, err := os.MkdirTemp("/dev/shm", "fileflow-smoke-"); err == nil {
		x := filepath.Join(shm, "x.txt")
		must(os.WriteFile(x, []byte("X"), 0o644))
		p, err = fileflow.Move(x, filepath.Join(root, "out/x.txt"))
		_, serr := os.Stat(x)
		fmt.Printf("move-crossfs: %s %s src=%v data=%s\n", rel(p), errClass(err), serr == nil, content(p))
		os.RemoveAll(shm)
	} else {
		fmt.Println("move-crossfs: skipped")
	}

	src = write("new.toml", "v2")
	write("cfg/app.toml", "v1")
	p, err = fileflow.Replace(src, filepath.Join(root, "cfg/app.toml"))
	fmt.Printf("replace: %s %s data=%s\n", rel(p), errClass(err), content(p))

	must(os.MkdirAll(filepath.Join(root, "dir"), 0o755))
	for _, op := range []struct {
		name string
		fn   func(string, string) (string, error)
	}{{"move", fileflow.Move}, {"copy", fileflow.Copy}, {"rename", fileflow.Rename}} {
		src = write(op.name+"-e.txt", "E")
		p, err = op.fn(src, filepath.Join(root, "dir"))
		fmt.Printf("%s-onto-dir: %s %s src=%v\n", op.name, rel(p), errClass(err), exists(op.name+"-e.txt"))
	}

	write("home/.bashrc", "old")
	src = write("rc", "new")
	p, err = fileflow.Copy(src, filepath.Join(root, "home/.bashrc"))
	fmt.Printf("dotfile-conflict: %s %s\n", rel(p), errClass(err))

	write("t/x.tar.gz", "1")
	src = write("t2", "2")
	p, err = fileflow.Copy(src, filepath.Join(root, "t/x.tar.gz"))
	fmt.Printf("targz-conflict: %s %s\n", rel(p), errClass(err))

	must(os.Symlink(filepath.Join(root, "nowhere"), filepath.Join(root, "dangling")))
	src = write("f.txt", "F")
	p, err = fileflow.Move(src, filepath.Join(root, "dangling"))
	fmt.Printf("dangling-dst: %s %s\n", rel(p), errClass(err))

	_, err = fileflow.Move(filepath.Join(root, "missing"), filepath.Join(root, "new/deep/x"))
	fmt.Printf("missing-src: %s dircreated=%v\n", errClass(err), exists("new/deep"))
}
