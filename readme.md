![fileflow](https://github.com/user-attachments/assets/1b9c44b9-7433-45d2-9096-68e0374fcf1b)


# FileFlow Package

The `fileflow` package provides a robust set of utilities to safely move, copy, and rename files even across different drives/filesystems. It includes various safety mechanisms to ensure files are moved efficiently, and destination files are not overwritten unless identical to the source.

## Features

- **Safe File Move**: Moves files between paths, with support for cross-filesystem transfers.
- **Unique Destination Naming**: If the destination file already exists, appends incrementing suffixes (`-1`, `-2`, etc.) to avoid overwriting non-identical files.
- **Identical File Check**: Compares files to determine if they are identical, preventing unnecessary overwrites.
- **Path Creation**: Automatically creates directories for destination paths if they don't exist.
- **Customizable Naming Strategy**: Flexible naming strategy for handling file conflicts through customizable functions.

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

### FindAvailableName
The package provides flexible naming strategies for handling file conflicts through the `FindAvailableName` variable. This variable holds a function that determines how to generate alternative filenames when a conflict occurs. The package includes two built-in implementations:

#### FindAvailableNameInc (Default)
The default implementation that appends incrementing numbers to filenames:
- For a file "document.txt", generates: "document-1.txt", "document-2.txt", etc.

#### FindAvailableNameTS
An alternative implementation that appends timestamps to filenames:
- For a file "document.txt", generates: "document-20230615-143022.123456789.txt"

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
This package is open-source and available under the Apache 2.0 License.
