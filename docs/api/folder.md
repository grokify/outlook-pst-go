# Folder

The `Folder` type represents a folder in a PST file.

## Properties

### ID

```go
func (f *Folder) ID() util.NodeID
```

Returns the folder's node ID.

### Name

```go
func (f *Folder) Name() (string, error)
```

Returns the display name of the folder.

### ContentCount

```go
func (f *Folder) ContentCount() (int32, error)
```

Returns the number of messages in the folder.

### UnreadCount

```go
func (f *Folder) UnreadCount() (int32, error)
```

Returns the number of unread messages.

### HasSubfolders

```go
func (f *Folder) HasSubfolders() (bool, error)
```

Returns true if the folder has subfolders.

### ContainerClass

```go
func (f *Folder) ContainerClass() (string, error)
```

Returns the folder's container class (e.g., "IPF.Note", "IPF.Appointment").

## Subfolder Navigation

### Subfolders

```go
func (f *Folder) Subfolders() iter.Seq2[*Folder, error]
```

Returns an iterator over subfolders.

```go
for subfolder, err := range folder.Subfolders() {
    if err != nil {
        log.Printf("Error: %v", err)
        continue
    }
    name, _ := subfolder.Name()
    fmt.Printf("Subfolder: %s\n", name)
}
```

### SubfolderCount

```go
func (f *Folder) SubfolderCount() (int, error)
```

Returns the number of subfolders.

### FindSubfolder

```go
func (f *Folder) FindSubfolder(name string) (*Folder, error)
```

Finds a subfolder by name.

```go
inbox, err := folder.FindSubfolder("Inbox")
if err != nil {
    if errors.Is(err, outlookpst.ErrNotFound) {
        fmt.Println("Not found")
    }
}
```

## Message Access

### Messages

```go
func (f *Folder) Messages() iter.Seq2[*Message, error]
```

Returns an iterator over messages in the folder.

```go
for msg, err := range folder.Messages() {
    if err != nil {
        continue
    }
    subject, _ := msg.Subject()
    fmt.Printf("Message: %s\n", subject)
}
```

### MessageCount

```go
func (f *Folder) MessageCount() (int, error)
```

Returns the number of messages in the folder.

## Search Folders

### IsSearchFolder

```go
func (f *Folder) IsSearchFolder() bool
```

Returns true if this is a search folder. Search folders are virtual folders whose contents are determined by search criteria.

### AsSearchFolder

```go
func (f *Folder) AsSearchFolder() *SearchFolder
```

Converts to a `SearchFolder` if this is a search folder. Returns `nil` if this is not a search folder.

```go
if folder.IsSearchFolder() {
    sf := folder.AsSearchFolder()
    criteria, _ := sf.SearchCriteria()
    fmt.Printf("Recursive: %v\n", criteria.IsRecursive())
}
```

## Advanced Access

### PropertyBag

```go
func (f *Folder) PropertyBag() *ltp.PropertyBag
```

Returns the folder's property bag for advanced property access.

```go
bag := folder.PropertyBag()
for _, propID := range bag.Properties() {
    fmt.Printf("Property: 0x%04X\n", propID)
}
```

### HierarchyTable

```go
func (f *Folder) HierarchyTable() (*ltp.Table, error)
```

Returns the hierarchy table (subfolders) for advanced access.

### ContentsTable

```go
func (f *Folder) ContentsTable() (*ltp.Table, error)
```

Returns the contents table (messages) for advanced access.

## Example

```go
func printFolderInfo(folder *outlookpst.Folder, indent string) {
    name, _ := folder.Name()
    contentCount, _ := folder.ContentCount()
    unreadCount, _ := folder.UnreadCount()
    containerClass, _ := folder.ContainerClass()

    fmt.Printf("%s[%s]\n", indent, name)
    fmt.Printf("%s  Type: %s\n", indent, containerClass)
    fmt.Printf("%s  Items: %d (%d unread)\n", indent, contentCount, unreadCount)

    // Recurse into subfolders
    for subfolder, err := range folder.Subfolders() {
        if err != nil {
            continue
        }
        printFolderInfo(subfolder, indent+"  ")
    }
}
```
