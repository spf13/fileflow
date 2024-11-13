# FileFlow Package

The `fileflow` package provides a robust set of utilities to safely move, copy, and rename files in a filesystem, even across different drives. It includes various safety mechanisms to ensure files are moved efficiently, and destination files are not overwritten unless identical to the source.

## Features

- **Safe File Move**: Moves files between paths, with support for cross-filesystem transfers.
- **Unique Destination Naming**: If the destination file already exists, appends incrementing suffixes (`-1`, `-2`, etc.) to avoid overwriting non-identical files.
- **Identical File Check**: Compares files byte-by-byte to determine if they are identical, preventing unnecessary overwrites.
- **Path Creation**: Automatically creates directories for destination paths if they don’t exist.

## Installation

Simply include the package in your Go project:

```go
import "github.com/spf13/fileflow"
```

## Usage
### SafeMoveFileEfficient
Moves a file from src to dst, renaming it if necessary and ensuring cross-filesystem compatibility.

```go
destination, err := fileflow.SafeMoveFileEfficient("source.txt", "destination.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println("File moved to:", destination)
```
### SafeRename
Attempts to rename src to dst, adding incrementing suffixes (-1, -2, etc.) if a non-identical file already exists at the destination.

```go
destination, err := fileflow.SafeRename("source.txt", "destination.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println("File renamed to:", destination)
```

### SafeMoveFile 
Moves a file with additional safety checks, ensuring that the destination file is not overwritten unless it’s identical to the source file.

```go
destination, err := fileflow.SafeMoveFile("source.txt", "destination.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println("File moved safely to:", destination)
```

### FileExists
Checks if a file exists at the specified path.

```go
exists := fileflow.FileExists("path/to/file.txt")
if exists {
    fmt.Println("File exists.")
}
```

### IdenticalFiles
Compares two files byte by byte to check if they are identical.

```go
identical := fileflow.IdenticalFiles("file1.txt", "file2.txt")
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

    movedFile, err := fileflow.SafeMoveFileEfficient(src, dst)
    if err != nil {
        log.Fatalf("Failed to move file: %v", err)
    }

    fmt.Printf("File successfully moved to %s\n", movedFile)
}
```

## License
This package is open-source and available under the MIT License.

