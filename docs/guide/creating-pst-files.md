# Creating PST Files

This guide covers how to create new PST files from scratch.

## Creating a New PST File

Use `Create()` to create a new, empty PST file:

```go
import (
    outlookpst "github.com/grokify/outlook-pst-go"
    "github.com/grokify/outlook-pst-go/pkg/disk"
)

// Create a Unicode format PST (recommended)
pst, err := outlookpst.Create("new-archive.pst", disk.FormatUnicode)
if err != nil {
    log.Fatal(err)
}
defer pst.Close()
```

## PST Formats

Two formats are supported:

| Format | Constant | Description |
|--------|----------|-------------|
| Unicode | `disk.FormatUnicode` | Modern format with 64-bit addresses. Recommended for new files. |
| ANSI | `disk.FormatANSI` | Legacy format with 32-bit addresses. Limited to ~2GB file size. |

```go
// Unicode format (recommended)
pst, err := outlookpst.Create("archive.pst", disk.FormatUnicode)

// ANSI format (legacy compatibility)
pst, err := outlookpst.Create("archive.pst", disk.FormatANSI)
```

## Create Options

For more control, use `CreateWithOptions()`:

```go
pst, err := outlookpst.CreateWithOptions("archive.pst", outlookpst.CreateOptions{
    Format:      disk.FormatUnicode,
    CryptMethod: disk.CryptMethodPermute, // Enable encryption
    DisplayName: "My Archive",            // PST display name
})
```

### Encryption Methods

| Method | Constant | Description |
|--------|----------|-------------|
| None | `disk.CryptMethodNone` | No encryption |
| Permute | `disk.CryptMethodPermute` | Basic obfuscation (default) |
| Cyclic | `disk.CryptMethodCyclic` | Cyclic encryption |

!!! note "Encryption Limitations"
    PST encryption is basic obfuscation, not security. All methods are fully supported for both reading and writing.

## Initial Structure

A newly created PST file contains:

- **Message Store** - Top-level properties
- **Root Folder** - The root of the folder hierarchy
- Empty hierarchy and contents tables

```go
pst, _ := outlookpst.Create("archive.pst", disk.FormatUnicode)
defer pst.Close()

// The root folder is automatically created
root, err := pst.RootFolder()
if err != nil {
    log.Fatal(err)
}

name, _ := root.Name()
fmt.Printf("Root folder: %s\n", name) // "Root"
```

## Opening for Read-Write

To modify an existing PST file, use `OpenReadWrite()`:

```go
// Open existing PST for modification
pst, err := outlookpst.OpenReadWrite("existing.pst")
if err != nil {
    log.Fatal(err)
}
defer pst.Close()

// Check if file is writable
if pst.IsReadOnly() {
    log.Fatal("File is read-only")
}
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"

    outlookpst "github.com/grokify/outlook-pst-go"
    "github.com/grokify/outlook-pst-go/pkg/disk"
)

func main() {
    // Create new PST
    pst, err := outlookpst.CreateWithOptions("my-archive.pst", outlookpst.CreateOptions{
        Format:      disk.FormatUnicode,
        CryptMethod: disk.CryptMethodPermute,
        DisplayName: "Personal Archive",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer pst.Close()

    // Print info
    fmt.Printf("Format: %s\n", pst.Format())
    fmt.Printf("Encryption: %s\n", pst.CryptMethod())

    root, _ := pst.RootFolder()
    name, _ := root.Name()
    fmt.Printf("Root folder: %s\n", name)

    // Now you can begin adding content using transactions
    // See the Transactions guide for details
}
```

## Next Steps

- [Transactions](transactions.md) - Learn about write transactions
- [Creating Folders](creating-folders.md) - Add folders to the PST
- [Creating Messages](creating-messages.md) - Add messages to folders
