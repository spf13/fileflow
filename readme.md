# FileFlow Package

The `fileflow` package provides a robust set of utilities to safely move, copy, and rename files even across different drives/filesystems. It includes various safety mechanisms to ensure files are moved efficiently, and destination files are not overwritten unless identical to the source.

## Features

- **Safe File Move**: Moves files between paths, with support for cross-filesystem transfers.
- **Unique Destination Naming**: If the destination file already exists, appends incrementing suffixes (`-1`, `-2`, etc.) to avoid overwriting non-identical files.
- **Identical File Check**: Compares files to determine if they are identical, preventing unnecessary overwrites.
- **Path Creation**: Automatically creates directories for destination paths if they don't exist.
- **Customizable Naming Strategy**: Flexible naming strategy for handling file conflicts through customizable functions.

## Safety guarantees

FileFlow never overwrites a non-identical file during normal use within a single process. Conflict detection is a check-then-act operation, however, so it is not atomic against concurrent external writers. Copies are written to a temporary file in the destination directory, synced, set to the source file's mode, and renamed into place, so a failed copy does not leave a partial or truncated destination. When the destination is identical to the source, FileFlow deduplicates the operation instead of writing another copy.

## Installation
Simply include the package in your Go project:

```go
import "github.com/spf13/fileflow"
```

## Usage
### Move
Moves a file from src to dst and returns the final destination path, which may differ from dst after conflict resolution. Move uses an atomic rename when possible and falls back to copy and delete across filesystems or drives, including on Windows. If src and dst refer to the same file through different paths, Move returns `ErrSameFile` rather than risking data loss.

```go
finalPath, err := fileflow.Move("source.txt", "destination.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println("File moved to:", finalPath)
```
### Rename
Renames src to dst, creates missing destination directories using `DirMode`, and returns the final destination path. Rename fails across filesystems; use Move if you want the copy fallback.

```go
destination, err := fileflow.Rename("source.txt", "destination.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println("File renamed to:", destination)
```

### Copy
Copies src to dst and returns the final destination path. If dst exists with different content, the returned path contains the conflict suffix. If dst is identical, Copy is a no-op and returns dst. Copy expects the destination directory to exist.

```go
finalPath, err := fileflow.Copy("source.txt", "backup/source.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println("File copied to:", finalPath)
```

### CopyWithPaths
Works like Copy but creates missing destination directories first using `DirMode`.

```go
finalPath, err := fileflow.CopyWithPaths("source.txt", "backup/source.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println("File copied to:", finalPath)
```

### Exists
Reports whether an accessible regular file exists at the specified path. It returns false for directories.

```go
exists := fileflow.Exists("path/to/file.txt")
if exists {
    fmt.Println("File exists.")
}
```

### Equal
Compares two files byte by byte to check if they are identical.

```go
identical, err := fileflow.Equal("file1.txt", "file2.txt")
if err != nil {
    log.Fatal(err)
}
if identical {
    fmt.Println("Files are identical.")
}
```

### FindAvailableName
The package provides flexible naming strategies for handling file conflicts through the `FindAvailableName` variable. This variable holds a function that determines how to generate alternative filenames when a conflict occurs. The package includes two built-in implementations:

#### FindAvailableNameInc (Default)
The default implementation appends `-1`, `-2`, and so on, up to `MaxIncrementAttempts` (100 by default), then returns `ErrMaxAttemptsReached`. Existing numeric endings are preserved:
- For "document.txt", generates "document-1.txt", "document-2.txt", etc.
- For "report-2024.txt", generates "report-2024-1.txt", not "report-1.txt".

#### FindAvailableNameTS
An alternative implementation that appends one nanosecond-precision timestamp using the format `20060102-150405.000000000`:
- For "document.txt", generates a name such as "document-20230615-143022.123456789.txt".

If that exact timestamped candidate already exists, it falls back to `FindAvailableNameInc` on the candidate.

You can customize the naming strategy by providing your own implementation:

```go
// Example of a custom naming strategy that adds a random suffix
func customNamingStrategy(baseName string) (string, error) {
    ext := filepath.Ext(baseName)
    nameWOExt := baseName[:len(baseName)-len(ext)]

    // Generate a random 6-character string
    const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
    suffix := make([]byte, 6)
    for i := range suffix {
        suffix[i] = charset[rand.Intn(len(charset))]
    }

    newName := fmt.Sprintf("%s-%s%s", nameWOExt, string(suffix), ext)
    if !fileflow.Exists(newName) {
        return newName, nil
    }
    return "", fileflow.ErrMaxAttemptsReached
}

// Set the custom strategy
fileflow.FindAvailableName = customNamingStrategy
```

### Package configuration

The package has three tunable variables:

- `FindAvailableName`: conflict naming function; defaults to `FindAvailableNameInc`.
- `MaxIncrementAttempts`: maximum incrementing suffix attempts; defaults to 100.
- `DirMode`: permission mode used when creating destination directories; defaults to `0755`.

## Error Handling
The package exposes sentinel errors for `errors.Is`:

* `ErrSameFile`: The source and destination refer to the same underlying file, including through different paths.
* `ErrMaxAttemptsReached`: The incrementing naming strategy exhausted `MaxIncrementAttempts`.

It also exposes structured errors for `errors.As`; both unwrap their underlying cause:

* `ErrFailedRemovingOriginal`: Indicates failure to remove the original file after copying or deduplicating.
* `ErrFailedMovingFile`: Indicates failure to rename a file from the source to the destination.

Each error type includes relevant file path information to help with debugging.

## Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/spf13/fileflow"
)

func main() {
    src := "example.txt"
    dst := "new_location/example.txt"

    movedFile, err := fileflow.Move(src, dst)
    if err != nil {
        log.Fatalf("Failed to move file: %v", err)
    }

    fmt.Printf("File successfully moved to %s\n", movedFile)
}
```

## License
This package is open-source and available under the Apache 2.0 License.
