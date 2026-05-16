# SearchFolder

The `SearchFolder` type represents a search folder in a PST file.

## Type Definition

```go
type SearchFolder struct {
    *Folder  // Embeds Folder
}
```

SearchFolder embeds `Folder`, so all Folder methods are available.

## Creating a SearchFolder

SearchFolders are created from regular Folders:

```go
if folder.IsSearchFolder() {
    searchFolder := folder.AsSearchFolder()
}
```

## Methods

### SearchCriteria

```go
func (sf *SearchFolder) SearchCriteria() (*SearchCriteria, error)
```

Returns the search criteria for this search folder.

## SearchCriteria Type

```go
type SearchCriteria struct {
    Restriction    []byte      // Raw restriction data
    FolderEntryIDs [][]byte    // Entry IDs of folders to search
    SearchFlags    uint32      // Search behavior flags
}
```

### SearchCriteria Methods

#### IsRecursive

```go
func (sc *SearchCriteria) IsRecursive() bool
```

Returns true if the search includes subfolders.

#### IsForeground

```go
func (sc *SearchCriteria) IsForeground() bool
```

Returns true if this is a foreground search.

#### IsStatic

```go
func (sc *SearchCriteria) IsStatic() bool
```

Returns true if this is a static search (not updated).

#### UsesContentIndex

```go
func (sc *SearchCriteria) UsesContentIndex() bool
```

Returns true if the search uses content indexing.

#### PropertyBag

```go
func (sc *SearchCriteria) PropertyBag() *ltp.PropertyBag
```

Returns the underlying property bag for advanced access.

## Search Flags

| Constant | Value | Description |
|----------|-------|-------------|
| `SearchFlagForeground` | 0x00000001 | Foreground search |
| `SearchFlagRecursive` | 0x00000002 | Search subfolders |
| `SearchFlagContentIndex` | 0x00000004 | Use content indexing |
| `SearchFlagStatic` | 0x00000008 | Static search |
| `SearchFlagMaybeStatic` | 0x00000010 | Might be static |

## Example

```go
func analyzeSearchFolder(folder *outlookpst.Folder) {
    if !folder.IsSearchFolder() {
        fmt.Println("Not a search folder")
        return
    }

    sf := folder.AsSearchFolder()
    name, _ := sf.Name()
    fmt.Printf("Search Folder: %s\n", name)

    criteria, err := sf.SearchCriteria()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Flags:\n")
    fmt.Printf("  Recursive: %v\n", criteria.IsRecursive())
    fmt.Printf("  Static: %v\n", criteria.IsStatic())
    fmt.Printf("  Content Index: %v\n", criteria.UsesContentIndex())
    fmt.Printf("Folders searched: %d\n", len(criteria.FolderEntryIDs))
    fmt.Printf("Restriction size: %d bytes\n", len(criteria.Restriction))

    // Count matching messages
    count := 0
    for _, err := range sf.Messages() {
        if err == nil {
            count++
        }
    }
    fmt.Printf("Matching messages: %d\n", count)
}
```

## SearchUpdateQueue

The `SearchUpdateQueue` tracks pending updates for search folders.

### Type Definition

```go
type SearchUpdateQueue struct {
    // internal fields
}

type SearchUpdateEntry struct {
    FolderNID  util.NodeID
    MessageNID util.NodeID
    Flags      uint32
}
```

### Accessing the Queue

```go
queue, err := pst.SearchUpdateQueue()
if err != nil {
    log.Fatal(err)
}
```

### Methods

#### Entries

```go
func (suq *SearchUpdateQueue) Entries() []SearchUpdateEntry
```

Returns all entries in the queue.

#### Count

```go
func (suq *SearchUpdateQueue) Count() int
```

Returns the number of entries.

#### IsEmpty

```go
func (suq *SearchUpdateQueue) IsEmpty() bool
```

Returns true if the queue is empty.

## Specification Reference

- [MS-PST] Section 2.4.8 - Search
- [MS-PST] Section 2.4.8.6 - Search Update Queue
