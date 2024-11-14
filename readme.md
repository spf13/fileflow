# FileFlow Package

The `fileflow` package provides a robust set of utilities to safely move, copy, and rename files even across different drives/filesystems. It includes various safety mechanisms to ensure files are moved efficiently, and destination files are not overwritten unless identical to the source.

## Features

- **Safe File Move**: Moves files between paths, with support for cross-filesystem transfers.
- **Unique Destination Naming**: If the destination file already exists, appends incrementing suffixes (`-1`, `-2`, etc.) to avoid overwriting non-identical files.
- **Identical File Check**: Compares files to determine if they are identical, preventing unnecessary overwrites.
- **Path Creation**: Automatically creates directories for destination paths if they don’t exist.

## Installation
Simply include the package in your Go project:

```go
import "github.com/spf13/fileflow"
```

## Usage
### Move
Moves a file from src to dst, renaming it if necessary and ensuring cross-filesystem compatibility. If the destination file already exists, it will be renamed with an incrementing suffix.
Will rename if on same filesystem, otherwise will do a cross file system move (copy and remove original).

```go
destination, err := fileflow.Move("source.txt", "destination.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println("File moved to:", destination)
```
### Rename
Attempts to rename src to dst, adding incrementing suffixes (-1, -2, etc.) if a non-identical file already exists at the destination. Unlike Move, Rename will fail if the files are on different filesystems. Unless you want it to fail if the files are on different filesystems, use Move instead.

```go
destination, err := fileflow.Rename("source.txt", "destination.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println("File renamed to:", destination)
```

### Exists
Checks if a file exists at the specified path.

```go
exists := fileflow.Exists("path/to/file.txt")
if exists {
    fmt.Println("File exists.")
}
```

### Equal
Compares two files byte by byte to check if they are identical.

```go
identical := fileflow.Equal("file1.txt", "file2.txt")
if identical {
    fmt.Println("Files are identical.")
}
```

## Error Handling
The package includes custom error types to provide detailed error information:

* ErrFailedRemovingOriginal: Indicates failure to remove the original file after copying.
* ErrFailedCopyingFile: Indicates failure to copy a file to a new location.
* ErrFailedMovingFile: Indicates failure to move a file from the source to the destination.

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
This package is open-source and available under the MIT License.

