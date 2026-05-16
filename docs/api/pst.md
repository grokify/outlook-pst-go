# PST

The `PST` type is the main entry point for accessing PST files.

## Opening a PST File

### Open (Read-Only)

```go
func Open(filename string) (*PST, error)
```

Opens a PST file for reading only.

```go
pst, err := outlookpst.Open("archive.pst")
if err != nil {
    log.Fatal(err)
}
defer pst.Close()
```

### OpenReadWrite

```go
func OpenReadWrite(filename string) (*PST, error)
```

Opens an existing PST file for reading and writing.

```go
pst, err := outlookpst.OpenReadWrite("archive.pst")
if err != nil {
    log.Fatal(err)
}
defer pst.Close()

// Check if writable
if pst.IsReadOnly() {
    log.Fatal("File is read-only")
}
```

## Creating a PST File

### Create

```go
func Create(filename string, format disk.PSTFormat) (*PST, error)
```

Creates a new, empty PST file.

```go
pst, err := outlookpst.Create("new.pst", disk.FormatUnicode)
if err != nil {
    log.Fatal(err)
}
defer pst.Close()
```

### CreateWithOptions

```go
func CreateWithOptions(filename string, opts CreateOptions) (*PST, error)
```

Creates a new PST file with custom options.

```go
pst, err := outlookpst.CreateWithOptions("new.pst", outlookpst.CreateOptions{
    Format:      disk.FormatUnicode,
    CryptMethod: disk.CryptMethodPermute,
    DisplayName: "My Archive",
})
```

### CreateOptions

```go
type CreateOptions struct {
    Format      disk.PSTFormat   // FormatUnicode or FormatANSI
    CryptMethod disk.CryptMethod // CryptMethodNone, CryptMethodPermute, CryptMethodCyclic
    DisplayName string           // Display name for the PST
}
```

## Closing

```go
func (p *PST) Close() error
```

Closes the PST file and releases resources.

## Format Information

### Format

```go
func (p *PST) Format() disk.PSTFormat
```

Returns the PST format (ANSI or Unicode).

```go
fmt.Printf("Format: %s\n", pst.Format())
// Output: "Unicode" or "ANSI"
```

### IsUnicode

```go
func (p *PST) IsUnicode() bool
```

Returns true if the PST uses Unicode format (64-bit addresses).

### IsANSI

```go
func (p *PST) IsANSI() bool
```

Returns true if the PST uses ANSI format (32-bit addresses).

### IsPST

```go
func (p *PST) IsPST() bool
```

Returns true if this is a PST file (not OST).

### IsOST

```go
func (p *PST) IsOST() bool
```

Returns true if this is an OST file.

### CryptMethod

```go
func (p *PST) CryptMethod() disk.CryptMethod
```

Returns the encryption method used in the file.

```go
switch pst.CryptMethod() {
case disk.CryptMethodNone:
    fmt.Println("No encryption")
case disk.CryptMethodPermute:
    fmt.Println("Permute encryption")
case disk.CryptMethodCyclic:
    fmt.Println("Cyclic encryption")
}
```

## Properties

### Name

```go
func (p *PST) Name() (string, error)
```

Returns the display name of the PST file.

```go
name, err := pst.Name()
if err == nil {
    fmt.Printf("PST Name: %s\n", name)
}
```

## Folder Access

### RootFolder

```go
func (p *PST) RootFolder() (*Folder, error)
```

Returns the root folder of the PST.

```go
root, err := pst.RootFolder()
if err != nil {
    log.Fatal(err)
}

for subfolder, _ := range root.Subfolders() {
    // ...
}
```

### OpenFolder

```go
func (p *PST) OpenFolder(name string) (*Folder, error)
```

Opens a folder by name, searching from the root.

```go
inbox, err := pst.OpenFolder("Inbox")
if err != nil {
    if errors.Is(err, outlookpst.ErrNotFound) {
        fmt.Println("Inbox not found")
    }
}
```

## Advanced Access

### Database

```go
func (p *PST) Database() *ndb.Database
```

Returns the underlying NDB database for advanced operations.

```go
db := pst.Database()

// Access nodes directly
node, err := db.GetNode(nodeID)
```

### MessageStore

```go
func (p *PST) MessageStore() (*ltp.PropertyBag, error)
```

Returns the message store property bag for PST-level properties.

```go
store, err := pst.MessageStore()
if err == nil {
    props := store.Properties()
    fmt.Printf("Message store has %d properties\n", len(props))
}
```

### NamedPropertyMap

```go
func (p *PST) NamedPropertyMap() (*ltp.NamedPropertyMap, error)
```

Returns the named property map for custom MAPI properties. Named properties map GUIDs and IDs/names to property IDs in the 0x8000+ range.

```go
npm, err := pst.NamedPropertyMap()
if err == nil {
    fmt.Printf("Named properties: %d\n", npm.Count())

    // Look up a specific named property
    propID, found := npm.GetPropID(util.PSETID_Common, 0x8501)
    if found {
        fmt.Printf("Mapped to: 0x%04X\n", propID)
    }
}
```

### SearchUpdateQueue

```go
func (p *PST) SearchUpdateQueue() (*SearchUpdateQueue, error)
```

Returns the search update queue containing pending updates for search folders.

```go
queue, err := pst.SearchUpdateQueue()
if err == nil {
    fmt.Printf("Pending search updates: %d\n", queue.Count())
}
```

## Write Operations

### IsReadOnly

```go
func (p *PST) IsReadOnly() bool
```

Returns true if the PST was opened in read-only mode.

```go
if pst.IsReadOnly() {
    log.Fatal("Cannot modify read-only PST")
}
```

### BeginWrite

```go
func (p *PST) BeginWrite() (*WriteContext, error)
```

Begins a write transaction. All modifications must be made within a transaction.

```go
ctx, err := pst.BeginWrite()
if err != nil {
    log.Fatal(err)
}

// Perform write operations
root, _ := pst.RootFolder()
inbox, _ := ctx.CreateFolder(root, "Inbox")

// Commit or rollback
if err := ctx.Commit(); err != nil {
    log.Fatal(err)
}
```

See [WriteContext](write-context.md) for the full API.

### Save

```go
func (p *PST) Save() error
```

Saves any pending changes to disk. If no transaction is active, this is a no-op.

```go
if err := pst.Save(); err != nil {
    log.Fatal(err)
}
```

### GetStatistics

```go
func (p *PST) GetStatistics() (*Statistics, error)
```

Returns statistics about the PST file.

```go
stats, err := pst.GetStatistics()
if err == nil {
    fmt.Printf("Format: %s\n", stats.Format)
    fmt.Printf("File Size: %d bytes\n", stats.FileSize)
    fmt.Printf("Free Space: %d bytes\n", stats.FreeSpace)
    fmt.Printf("Folders: %d\n", stats.FolderCount)
    fmt.Printf("Messages: %d\n", stats.MessageCount)
}
```

### Statistics Type

```go
type Statistics struct {
    Format          disk.PSTFormat
    FileSize        uint64
    FreeSpace       uint64
    FolderCount     int
    MessageCount    int
    AttachmentCount int
}
```

## Utility Functions

### CheckRecovery

```go
func CheckRecovery(filename string) (*RecoverableInfo, error)
```

Checks if a PST file needs recovery without opening it for write.

```go
info, err := outlookpst.CheckRecovery("archive.pst")
if err != nil {
    log.Fatal(err)
}

if info.NeedsRecovery {
    fmt.Printf("Recovery needed, status: %v\n", info.Status)
}
```

### Compact

```go
func Compact(srcFilename, dstFilename string, opts CompactOptions) error
```

Creates a compacted copy of a PST file, removing unused space.

```go
err := outlookpst.Compact("original.pst", "compacted.pst", outlookpst.CompactOptions{
    RemoveDeletedItems: true,
    DefragmentBlocks:   true,
})
```

### CompactOptions

```go
type CompactOptions struct {
    RemoveDeletedItems bool // Skip "Deleted Items" folder
    DefragmentBlocks   bool // Consolidate fragmented blocks
}
```

## Example

```go
package main

import (
    "fmt"
    "log"

    outlookpst "github.com/grokify/outlook-pst-go"
)

func main() {
    pst, err := outlookpst.Open("archive.pst")
    if err != nil {
        log.Fatal(err)
    }
    defer pst.Close()

    // Print file info
    fmt.Printf("Format: %s\n", pst.Format())
    fmt.Printf("Type: %s\n", map[bool]string{true: "PST", false: "OST"}[pst.IsPST()])
    fmt.Printf("Encryption: %s\n", pst.CryptMethod())

    if name, err := pst.Name(); err == nil {
        fmt.Printf("Name: %s\n", name)
    }

    // Access root folder
    root, _ := pst.RootFolder()
    rootName, _ := root.Name()
    fmt.Printf("Root folder: %s\n", rootName)
}
```
