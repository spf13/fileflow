/*
Copyright © 2024 The FileFlow Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package fileflow safely moves, copies, and renames files, including
// across filesystems. Move, Copy, and Rename never overwrite a non-identical
// file: renames use no-replace primitives, copies land atomically via a
// temporary file, and naming conflicts are resolved by configurable strategies
// (see Flow and the FindAvailableName functions). Destinations identical to
// the source are deduplicated rather than rewritten. Replace is the explicit
// opt-in operation for atomically overwriting an exact destination path.
// Unlike the standard library, FileFlow creates missing destination directories
// by default; set Flow.NoCreateDirs to require existing directories instead.
package fileflow

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	// DefaultDirMode is the default permission mode for new directories.
	DefaultDirMode = 0755

	// copyChunkSize is the buffer size of the cancellable, progress-reporting
	// copy loop, and so the granularity of cancellation and Progress calls.
	copyChunkSize = 1 << 20

	// maxIncrementAttempts is the maximum number of candidate names tried
	// when resolving a naming conflict by incrementing a counter.
	maxIncrementAttempts = 100
)

// errDstIsDir is returned, wrapped once in ErrFailedMovingFile, when the
// destination is an existing directory.
var errDstIsDir = fmt.Errorf("destination is a directory: %w", fs.ErrExist)

var (
	ErrSameFile           = errors.New("source and destination are the same")
	ErrMaxAttemptsReached = errors.New("maximum increment attempts reached")

	// ErrDigestMismatch reports that a copy made with Flow.Verify did not
	// read back with the source's digest. Nothing is published and the
	// source is left intact.
	ErrDigestMismatch = errors.New("copy verification failed: digest mismatch")
)

// testHookBeforeVerify, when non-nil, is called with the temporary file's
// path after it is synced and before it is re-read for verification.
var testHookBeforeVerify func(tmpName string)

// Digest is a SHA-256 content digest.
type Digest [sha256.Size]byte

// String returns d as lowercase hex.
func (d Digest) String() string {
	return hex.EncodeToString(d[:])
}

// Result describes a completed Copy or Move.
type Result struct {
	Path   string // final destination path (may differ from dst after conflict resolution)
	Digest Digest // SHA-256 of the content at Path when Flow.Verify is set; zero otherwise
}

// Flow provides per-instance configuration for file operations.
// The zero value is ready to use. Configure a Flow before first use;
// do not mutate it concurrently with calls to its methods.
type Flow struct {
	// FindAvailableName resolves naming conflicts when a destination
	// already exists with different content.
	// Nil means FindAvailableNameInc.
	FindAvailableName func(string) (string, error)

	// DirMode is the permission mode used when creating destination
	// directories. Zero means 0755.
	DirMode fs.FileMode

	// NoCreateDirs, when true, disables automatic creation of missing
	// destination directories: operations fail with the underlying
	// ErrNotExist instead, matching os.Rename semantics. The zero value
	// (false) creates missing directories, which departs from the standard
	// library; see the package documentation.
	NoCreateDirs bool

	// AllowOverwriteFallback, when true, allows the Unix fallback from
	// os.Link to os.Rename on filesystems that do not support hard links
	// (e.g. FAT, SMB). When false (the default) the fallback is disabled
	// and the link error is returned instead, preventing silent overwrite
	// of a concurrently created destination.
	AllowOverwriteFallback bool

	// NoPreserveTimes, when true, leaves copied files with the times of the
	// copy. When false (the default), copies made by Copy, and by Move and
	// Replace across filesystems, set the destination's modification and
	// access times to the source's. Preservation is best-effort: errors
	// setting times are ignored, and precision is limited by the destination
	// filesystem (FAT: 2 s mtime, date-only atime; exFAT: 10 ms).
	// Same-filesystem renames always keep times because no copy happens.
	NoPreserveTimes bool

	// Progress, when non-nil, is called synchronously from the copying
	// goroutine after each chunk (up to 1 MiB) is written to the temporary
	// file. total is the source size from the pre-copy stat; the last call
	// has done equal to the bytes copied. It is not called when no bytes are
	// copied (same-filesystem rename, identical-destination dedup). It runs
	// on the copy path and must not block for long.
	Progress func(done, total int64)

	// Verify, when true, makes copies verified: Copy, and Move and Replace
	// across filesystems, hash the source with SHA-256 while copying, sync
	// the temporary file, re-read and hash it, and publish only if the
	// digests match. A mismatch returns an error wrapping ErrDigestMismatch,
	// removes the temporary file, and leaves the source intact. CopyResult
	// and MoveResult return the digest. Identical-destination dedup hashes
	// the source during the byte comparison, and a same-filesystem Move
	// hashes the file once after the rename. Rename ignores Verify.
	//
	// The re-read may be served from the page cache, so verification catches
	// short writes and corruption in the copy path, not media faults below
	// the cache. Verify disables the io.Copy kernel fast path.
	Verify bool
}

// findAvailableName resolves the configured naming strategy, defaulting to
// FindAvailableNameInc when none is set.
func (f *Flow) findAvailableName(baseName string) (string, error) {
	if f.FindAvailableName != nil {
		return f.FindAvailableName(baseName)
	}
	return FindAvailableNameInc(baseName)
}

// dirMode resolves the directory permission mode, defaulting to DefaultDirMode
// when unset.
func (f *Flow) dirMode() fs.FileMode {
	if f.DirMode != 0 {
		return f.DirMode
	}
	return DefaultDirMode
}

// splitExt splits name into everything before the extension and the
// extension itself. A leading dot marks a hidden file, not an extension, so
// ".bashrc" has none.
func splitExt(name string) (stem, ext string) {
	ext = filepath.Ext(name)
	if ext == filepath.Base(name) {
		ext = ""
	}
	return name[:len(name)-len(ext)], ext
}

// findAvailable generates candidate names by calling next with incrementing
// counters starting at 1 and returns the first path that nothing occupies.
func findAvailable(baseName string, next func(stem, ext string, i int) string) (string, error) {
	stem, ext := splitExt(baseName)
	for i := 1; i <= maxIncrementAttempts; i++ {
		cand := next(stem, ext, i)
		if !taken(cand) {
			return cand, nil
		}
	}
	return "", ErrMaxAttemptsReached
}

// ensureDir creates the parent directory of dst when NoCreateDirs is false.
func (f *Flow) ensureDir(dst string) error {
	if f.NoCreateDirs {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), f.dirMode()); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}
	return nil
}

// validateRegularFile checks that src is a regular file via Lstat.
// Symlinks, directories, and other non-regular files return fs.ErrInvalid
// wrapped in *ErrFailedMovingFile.
func (f *Flow) validateRegularFile(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return &ErrFailedMovingFile{Err: err, Src: src, Dst: dst}
	}
	if !info.Mode().IsRegular() {
		return &ErrFailedMovingFile{Err: fs.ErrInvalid, Src: src, Dst: dst}
	}
	return nil
}

// parseContSuffix parses a trailing "-N" suffix from s.
// It returns base (s without suffix), the numeric value, and whether a
// suffix was found. Used by FindAvailableNameCont/Auto without regexp.
func parseContSuffix(s string) (base string, n int, ok bool) {
	dash := -1
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '-' {
			dash = i
			break
		}
		if s[i] < '0' || s[i] > '9' {
			return s, 0, false
		}
	}
	if dash == -1 || dash == len(s)-1 {
		return s, 0, false
	}
	n, err := strconv.Atoi(s[dash+1:])
	if err != nil {
		return s, 0, false
	}
	return s[:dash], n, true
}

// ErrFailedRemovingOriginal occurs when the original file cannot be removed
// after its content has safely reached the destination. When returned alongside
// a non-empty destination path, the destination file is already committed.
type ErrFailedRemovingOriginal struct {
	// Err is the underlying error from the remove operation.
	Err error
	// File is the path of the original file that could not be removed.
	File string
}

func (e *ErrFailedRemovingOriginal) Error() string {
	return fmt.Sprintf("failed removing original file %q: %v", e.File, e.Err)
}

func (e *ErrFailedRemovingOriginal) Unwrap() error {
	return e.Err
}

// ErrFailedMovingFile occurs when a file move operation fails
type ErrFailedMovingFile struct {
	// Err is the underlying error from the rename operation.
	Err error
	// Src and Dst are the source and destination paths of the failed move.
	Src, Dst string
}

func (e *ErrFailedMovingFile) Error() string {
	return fmt.Sprintf("failed moving file %q to %q: %v", e.Src, e.Dst, e.Err)
}

func (e *ErrFailedMovingFile) Unwrap() error {
	return e.Err
}

// sameFile reports whether src and dst refer to the same underlying file,
// even when the paths are spelled differently (e.g. "./a" vs "a", symlinks,
// or case differences on case-insensitive filesystems).
func sameFile(src, dst string) bool {
	if src == dst {
		return true
	}
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false
	}
	dstInfo, err := os.Stat(dst)
	if err != nil {
		return false
	}
	return os.SameFile(srcInfo, dstInfo)
}

// Move tries to move a file atomically using rename if possible,
// falling back to copy+delete if files are on different filesystems.
func Move(src, dst string) (string, error) {
	var f Flow
	return f.Move(src, dst)
}

// MoveContext is like [Move] but honours ctx; see [Flow.MoveContext].
func MoveContext(ctx context.Context, src, dst string) (string, error) {
	var f Flow
	return f.MoveContext(ctx, src, dst)
}

// Move tries to move a file atomically using rename if possible,
// falling back to copy+delete if files are on different filesystems.
// Conflicts are handled as in Rename; an existing directory at dst is an
// error, not a place to move into.
func (f *Flow) Move(src, dst string) (string, error) {
	return f.MoveContext(context.Background(), src, dst)
}

// MoveContext is like Move but honours ctx. It returns ctx's error, wrapped
// in *ErrFailedMovingFile, if ctx is done before the move starts or while
// comparing against an existing destination; src is left intact. The rename
// itself is not interrupted. Across filesystems the copy is
// cancelled between chunks and its temporary file removed, leaving src
// intact; once the destination is published the move completes, including
// removal of src.
func (f *Flow) MoveContext(ctx context.Context, src, dst string) (string, error) {
	r, err := f.MoveResult(ctx, src, dst)
	return r.Path, err
}

// MoveResult is like MoveContext but returns a Result, which carries the
// moved content's digest when f.Verify is set. A same-filesystem move is a
// rename, so with Verify the file is hashed once at its final path after the
// rename; if that read fails, MoveResult returns the final Path together
// with the error. As with *ErrFailedRemovingOriginal, a non-empty Path
// alongside an error means the file is already at Path.
func (f *Flow) MoveResult(ctx context.Context, src, dst string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, &ErrFailedMovingFile{Err: err, Src: src, Dst: dst}
	}
	if err := f.validateRegularFile(src, dst); err != nil {
		return Result{}, err
	}
	if sameFile(src, dst) {
		return Result{}, ErrSameFile
	}

	r, renamed, err := f.rename(ctx, src, dst, f.Verify)
	if err != nil {
		if isCrossDeviceError(err) {
			// If the file is on a different drive, copy it instead
			return f.moveByCopy(ctx, src, dst)
		}
		return Result{}, err
	}
	if f.Verify && renamed {
		// The move is committed, so hash regardless of ctx.
		d, herr := hashFile(context.Background(), r.Path)
		if herr != nil {
			return r, fmt.Errorf("hashing moved file %q: %w", r.Path, herr)
		}
		r.Digest = d
	}
	return r, nil
}

// resolveDst decides where an operation on src may write. It returns the
// destination path to use and whether dst already holds content identical
// to src (in which case no write is needed). When verify is set and dst is
// identical, sum is src's digest, computed during the comparison.
func (f *Flow) resolveDst(ctx context.Context, src, dst string, verify bool) (path string, identical bool, sum Digest, err error) {
	if sameFile(src, dst) {
		return "", false, sum, ErrSameFile
	}
	if !taken(dst) {
		return dst, false, sum, nil
	}
	if info, err := os.Stat(dst); err == nil {
		if info.IsDir() {
			return "", false, sum, errDstIsDir
		}
		if info.Mode().IsRegular() {
			var h hash.Hash
			if verify {
				h = sha256.New()
			}
			same, err := compare(ctx, src, dst, h)
			if err != nil {
				return "", false, sum, fmt.Errorf("checking file identity: %w", err)
			}
			if same {
				if h != nil {
					h.Sum(sum[:0])
				}
				return dst, true, sum, nil
			}
		}
	}
	// dst holds different content, a dangling symlink, or a special file.
	path, err = f.findAvailableName(dst)
	if err != nil {
		return "", false, sum, fmt.Errorf("finding available name: %w", err)
	}
	return path, false, sum, nil
}

// Rename attempts to rename a file from src to dst, handling naming conflicts.
// It returns the final destination path.
func Rename(src, dst string) (string, error) {
	var f Flow
	return f.Rename(src, dst)
}

// publishWithRetry resolves naming conflicts, stages the source if needed, and
// publishes to the resolved destination with no-replace retry on fs.ErrExist.
// When verify is set, the Result's digest comes from the identical-destination
// comparison or from stage.
func (f *Flow) publishWithRetry(
	ctx context.Context,
	src, dst string,
	verify bool,
	onIdentical func() error,
	stage func() (stagedPath string, sum Digest, cleanup func(), err error),
) (Result, error) {
	var staged string
	var sum Digest
	var cleanup func()
	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()

	for range maxIncrementAttempts {
		path, identical, dupSum, err := f.resolveDst(ctx, src, dst, verify)
		if err != nil {
			if errors.Is(err, ErrSameFile) {
				return Result{}, err
			}
			return Result{}, &ErrFailedMovingFile{Err: err, Src: src, Dst: dst}
		}
		if identical {
			r := Result{Path: dst, Digest: dupSum}
			if err := onIdentical(); err != nil {
				return r, err
			}
			return r, nil
		}
		if err := f.ensureDir(path); err != nil {
			return Result{}, &ErrFailedMovingFile{Err: err, Src: src, Dst: path}
		}
		if staged == "" {
			staged, sum, cleanup, err = stage()
			if err != nil {
				return Result{}, &ErrFailedMovingFile{Err: err, Src: src, Dst: path}
			}
		}

		err = renameNoReplace(staged, path, f.AllowOverwriteFallback)
		if err == nil {
			_ = syncDir(filepath.Dir(path))
			return Result{Path: path, Digest: sum}, nil
		}
		var rmErr *ErrFailedRemovingOriginal
		if errors.As(err, &rmErr) {
			// The file landed at path but src could not be removed.
			return Result{Path: path, Digest: sum}, err
		}
		if errors.Is(err, fs.ErrExist) {
			// A concurrent writer took the name; resolve again.
			continue
		}
		return Result{}, &ErrFailedMovingFile{Err: err, Src: src, Dst: path}
	}

	return Result{}, &ErrFailedMovingFile{Err: ErrMaxAttemptsReached, Src: src, Dst: dst}
}

// Rename attempts to rename a file from src to dst, handling naming conflicts.
// It returns the final destination path. The rename never replaces an
// existing file: if a concurrent writer claims the chosen name between
// conflict resolution and the rename, Rename re-resolves and retries.
// If dst is an existing directory, Rename fails with an *ErrFailedMovingFile
// matching fs.ErrExist; it never renames into the directory. Move and Copy
// behave the same way.
func (f *Flow) Rename(src, dst string) (string, error) {
	r, _, err := f.rename(context.Background(), src, dst, false)
	return r.Path, err
}

// rename is Rename, honouring ctx while comparing against an existing
// destination. renamed reports that src was renamed rather than deduplicated
// against an identical destination; when verify is set the dedup Result
// carries src's digest.
func (f *Flow) rename(ctx context.Context, src, dst string, verify bool) (r Result, renamed bool, err error) {
	if err := f.validateRegularFile(src, dst); err != nil {
		return Result{}, false, err
	}
	renamed = true
	r, err = f.publishWithRetry(ctx, src, dst, verify, func() error {
		renamed = false
		if err := os.Remove(src); err != nil {
			return &ErrFailedRemovingOriginal{Err: err, File: src}
		}
		return nil
	}, func() (string, Digest, func(), error) {
		return src, Digest{}, nil, nil
	})
	return r, renamed, err
}

// fileMove moves a file across filesystems by copy + remove; see moveByCopy.
func (f *Flow) fileMove(ctx context.Context, src, dst string) (string, error) {
	r, err := f.moveByCopy(ctx, src, dst)
	return r.Path, err
}

// moveByCopy moves a file across filesystems by copy + remove. src is removed
// only after the copy, verified when f.Verify is set, is published.
func (f *Flow) moveByCopy(ctx context.Context, src, dst string) (Result, error) {
	r, err := f.CopyResult(ctx, src, dst)
	if err != nil {
		return Result{}, err
	}
	if err := os.Remove(src); err != nil {
		return r, &ErrFailedRemovingOriginal{Err: err, File: src}
	}
	return r, nil
}

// taken reports whether anything, including a directory or a dangling
// symlink, occupies path. Naming strategies use it so they never pick a
// candidate that the no-replace rename would refuse.
func taken(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// Exists reports whether an accessible regular file exists at the given path. It returns false for directories and inaccessible paths.
func Exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// FindAvailableNameInc returns an available filename by incrementing a counter
func FindAvailableNameInc(baseName string) (string, error) {
	return findAvailable(baseName, func(stem, ext string, i int) string {
		return fmt.Sprintf("%s-%d%s", stem, i, ext)
	})
}

// FindAvailableNameCont returns an available filename by treating a
// trailing -N suffix as a counter and continuing the sequence from N+1.
// A name without a trailing -N behaves like FindAvailableNameInc.
// Do not use this strategy when filenames end in meaningful digits
// (e.g. years): "report-2024.txt" would continue as "report-2025.txt".
func FindAvailableNameCont(baseName string) (string, error) {
	stem, ext := splitExt(baseName)
	base := stem
	start := 1
	if b, n, ok := parseContSuffix(stem); ok {
		base = b
		start = n + 1
	}
	return findAvailable(baseName, func(_, _ string, i int) string {
		return fmt.Sprintf("%s-%d%s", base, start+i-1, ext)
	})
}

// FindAvailableNameAuto chooses between the Inc and Cont strategies based
// on the trailing suffix. A trailing -N counter of one or two digits
// (e.g. "file-1", "file-42") is treated as a counter and continued from
// N+1, like FindAvailableNameCont. Three or more trailing digits
// (e.g. "report-2024", "img-001") are assumed to be meaningful and are
// preserved, appending a new counter like FindAvailableNameInc.
func FindAvailableNameAuto(baseName string) (string, error) {
	stem, _ := splitExt(baseName)
	if b, _, ok := parseContSuffix(stem); ok {
		digits := stem[len(b)+1:]
		if len(digits) <= 2 {
			return FindAvailableNameCont(baseName)
		}
	}
	return FindAvailableNameInc(baseName)
}

// FindAvailableNameTS returns an available filename by appending a timestamp.
func FindAvailableNameTS(baseName string) (string, error) {
	stem, ext := splitExt(baseName)
	candidate := fmt.Sprintf("%s-%s%s", stem, time.Now().Format("20060102-150405.000000000"), ext)
	if !taken(candidate) {
		return candidate, nil
	}
	return FindAvailableNameInc(candidate)
}

// Equal reports whether two files have identical content by comparing them byte for byte.
func Equal(file1, file2 string) (bool, error) {
	return equal(context.Background(), file1, file2)
}

// equal is Equal but stops reading when ctx is done.
func equal(ctx context.Context, file1, file2 string) (bool, error) {
	return compare(ctx, file1, file2, nil)
}

// compare is equal, additionally writing the bytes read from file1 to h when
// h is non-nil. When the files are identical, h has seen all of file1.
func compare(ctx context.Context, file1, file2 string, h hash.Hash) (bool, error) {
	f1Info, err := os.Stat(file1)
	if err != nil {
		return false, fmt.Errorf("stat file1: %w", err)
	}
	f2Info, err := os.Stat(file2)
	if err != nil {
		return false, fmt.Errorf("stat file2: %w", err)
	}

	// Quick check: if sizes differ, files are not identical
	if f1Info.Size() != f2Info.Size() {
		return false, nil
	}

	f1, err := os.Open(file1)
	if err != nil {
		return false, fmt.Errorf("opening first file: %w", err)
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return false, fmt.Errorf("opening second file: %w", err)
	}
	defer f2.Close()

	var r1, r2 io.Reader = f1, f2
	if ctx.Done() != nil {
		r1, r2 = ctxReader{ctx, f1}, ctxReader{ctx, f2}
	}
	if h != nil {
		r1 = io.TeeReader(r1, h)
	}
	return readersEqual(r1, r2)
}

// ctxReader fails reads with ctx's error once ctx is done.
type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func (r ctxReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

func readersEqual(r1, r2 io.Reader) (bool, error) {
	b1 := bufio.NewReader(r1)
	b2 := bufio.NewReader(r2)

	for {
		byte1, err1 := b1.ReadByte()
		byte2, err2 := b2.ReadByte()

		if err1 != nil && err1 != io.EOF {
			return false, fmt.Errorf("reading first file: %w", err1)
		}
		if err2 != nil && err2 != io.EOF {
			return false, fmt.Errorf("reading second file: %w", err2)
		}
		if err1 == io.EOF || err2 == io.EOF {
			return err1 == io.EOF && err2 == io.EOF, nil
		}
		if byte1 != byte2 {
			return false, nil
		}
	}
}

// Copy performs an efficient copy of a file from src to dst.
// If the destination file exists and is identical, it returns early.
// If the destination exists and is different, it finds an available name.
// It returns the final destination path written, which may differ from dst
// when a naming conflict was resolved. On error it returns an empty path.
func Copy(src, dst string) (string, error) {
	var f Flow
	return f.Copy(src, dst)
}

// CopyContext is like [Copy] but honours ctx; see [Flow.CopyContext].
func CopyContext(ctx context.Context, src, dst string) (string, error) {
	var f Flow
	return f.CopyContext(ctx, src, dst)
}

// Copy performs an efficient copy of a file from src to dst.
// It creates missing destination directories by default, using DirMode;
// set Flow.NoCreateDirs to require an existing directory instead.
// If the destination file exists and is identical, it returns early.
// If the destination exists and is different, it finds an available name.
// It returns the final destination path written, which may differ from dst
// when a naming conflict was resolved. On error it returns an empty path.
// The final rename never replaces an existing file: if a concurrent writer
// claims the chosen name, Copy re-resolves and retries. If dst is an
// existing directory, Copy fails with an *ErrFailedMovingFile matching
// fs.ErrExist.
func (f *Flow) Copy(src, dst string) (string, error) {
	return f.CopyContext(context.Background(), src, dst)
}

// CopyContext is like Copy but honours ctx. It checks ctx before starting,
// while comparing against an existing destination,
// between 1 MiB chunks of the copy, and once more before publishing; on
// cancellation it removes the temporary file and returns ctx's error wrapped
// in *ErrFailedMovingFile, leaving no destination. Once the destination is
// published the copy completes.
func (f *Flow) CopyContext(ctx context.Context, src, dst string) (string, error) {
	r, err := f.CopyResult(ctx, src, dst)
	return r.Path, err
}

// CopyResult is like CopyContext but returns a Result, which carries the
// copied content's digest when f.Verify is set.
func (f *Flow) CopyResult(ctx context.Context, src, dst string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, &ErrFailedMovingFile{Err: err, Src: src, Dst: dst}
	}
	if err := f.validateRegularFile(src, dst); err != nil {
		return Result{}, err
	}
	return f.publishWithRetry(ctx, src, dst, f.Verify, func() error {
		return nil
	}, func() (string, Digest, func(), error) {
		tmpName, sum, err := f.copyToTemp(ctx, src, filepath.Dir(dst))
		if err != nil {
			return "", sum, nil, err
		}
		if err := ctx.Err(); err != nil {
			os.Remove(tmpName)
			return "", sum, nil, err
		}
		return tmpName, sum, func() { os.Remove(tmpName) }, nil
	})
}

// copyToTemp copies src to a new temporary file in dir, preserving src's
// permission mode and, unless f.NoPreserveTimes is set, its modification and
// access times (best-effort). It reports progress to f.Progress and stops
// between chunks when ctx is done. When f.Verify is set it hashes src while
// copying, re-reads the synced temporary file, and fails with
// ErrDigestMismatch if the digests differ. It returns the temporary file's
// path and, with Verify, its digest.
func (f *Flow) copyToTemp(ctx context.Context, src, dir string) (string, Digest, error) {
	var sum Digest
	sourceFile, err := os.Open(src)
	if err != nil {
		return "", sum, fmt.Errorf("opening source file: %w", err)
	}
	defer sourceFile.Close()

	// Stat the open source before io.Copy so the captured access time is the
	// pre-copy value, not one bumped by reading the file.
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return "", sum, fmt.Errorf("getting source file info: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".fileflow-*")
	if err != nil {
		return "", sum, fmt.Errorf("creating temporary destination file: %w", err)
	}
	tmpName := tmp.Name()
	success := false
	defer func() {
		if !success {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	var r io.Reader = sourceFile
	var h hash.Hash
	if f.Verify {
		h = sha256.New()
		r = io.TeeReader(sourceFile, h)
	}
	if _, err := copyData(ctx, tmp, r, sourceInfo.Size(), f.Progress); err != nil {
		return "", sum, fmt.Errorf("copying file content: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		return "", sum, fmt.Errorf("syncing file: %w", err)
	}

	if h != nil {
		h.Sum(sum[:0])
		if testHookBeforeVerify != nil {
			testHookBeforeVerify(tmpName)
		}
		got, err := hashFile(ctx, tmpName)
		if err != nil {
			return "", sum, fmt.Errorf("verifying copy: %w", err)
		}
		if got != sum {
			return "", sum, ErrDigestMismatch
		}
	}

	if err := os.Chmod(tmpName, sourceInfo.Mode()); err != nil {
		return "", sum, fmt.Errorf("setting destination file mode: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return "", sum, fmt.Errorf("closing temporary destination file: %w", err)
	}

	// Set times after Close, which can update mtime on some platforms.
	// Best-effort: some filesystems (SMB, FUSE, FAT drivers) reject or cannot
	// represent the times, and that must not fail an otherwise good copy.
	if !f.NoPreserveTimes {
		_ = os.Chtimes(tmpName, atime(sourceInfo), sourceInfo.ModTime())
	}

	success = true
	return tmpName, sum, nil
}

// copyData copies src to dst and returns the number of bytes copied. When ctx
// can never be cancelled and progress is nil it uses io.Copy, keeping kernel
// fast paths such as copy_file_range. Otherwise it copies in chunks of
// copyChunkSize, checking ctx before each read and calling progress after
// each write.
func copyData(ctx context.Context, dst io.Writer, src io.Reader, total int64, progress func(done, total int64)) (int64, error) {
	if ctx.Done() == nil && progress == nil {
		return io.Copy(dst, src)
	}
	buf := make([]byte, copyChunkSize)
	var done int64
	for {
		if err := ctx.Err(); err != nil {
			return done, err
		}
		n, rerr := src.Read(buf)
		if n > 0 {
			w, werr := dst.Write(buf[:n])
			done += int64(w)
			if werr != nil {
				return done, werr
			}
			if w != n {
				return done, io.ErrShortWrite
			}
			if progress != nil {
				progress(done, total)
			}
		}
		if rerr == io.EOF {
			return done, nil
		}
		if rerr != nil {
			return done, rerr
		}
	}
}

// hashFile returns the SHA-256 digest of the file at path, stopping between
// reads when ctx is done.
func hashFile(ctx context.Context, path string) (Digest, error) {
	var sum Digest
	file, err := os.Open(path)
	if err != nil {
		return sum, err
	}
	defer file.Close()
	var r io.Reader = file
	if ctx.Done() != nil {
		r = ctxReader{ctx, file}
	}
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return sum, err
	}
	h.Sum(sum[:0])
	return sum, nil
}
