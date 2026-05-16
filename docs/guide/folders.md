# Working with Folders

## Folder Properties

```go
// Get folder name
name, err := folder.Name()

// Get message count
count, err := folder.ContentCount()

// Get unread count
unread, err := folder.UnreadCount()

// Check for subfolders
hasSub, err := folder.HasSubfolders()

// Get container class (e.g., "IPF.Note", "IPF.Appointment")
class, err := folder.ContainerClass()
```

## Iterating Subfolders

Use Go 1.23+ range-over-func iterators:

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

## Finding a Subfolder by Name

```go
inbox, err := folder.FindSubfolder("Inbox")
if err != nil {
    if errors.Is(err, outlookpst.ErrNotFound) {
        fmt.Println("Inbox not found")
    } else {
        log.Fatal(err)
    }
}
```

## Counting Subfolders

```go
count, err := folder.SubfolderCount()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Has %d subfolders\n", count)
```

## Recursive Folder Traversal

```go
func printFolderTree(folder *outlookpst.Folder, indent string) {
    name, _ := folder.Name()
    count, _ := folder.ContentCount()
    fmt.Printf("%s[%s] (%d items)\n", indent, name, count)

    for subfolder, err := range folder.Subfolders() {
        if err != nil {
            continue
        }
        printFolderTree(subfolder, indent+"  ")
    }
}

// Usage
root, _ := pst.RootFolder()
printFolderTree(root, "")
```

## Accessing Folder Tables

For advanced use cases, you can access the underlying tables:

```go
// Hierarchy table (subfolders)
hierarchyTable, err := folder.HierarchyTable()
if err == nil {
    for row, err := range hierarchyTable.Rows() {
        if err != nil {
            continue
        }
        // Access row data directly
        rowID := row.RowID()
        fmt.Printf("Row ID: %d\n", rowID)
    }
}

// Contents table (messages)
contentsTable, err := folder.ContentsTable()
if err == nil {
    rowCount := contentsTable.RowCount()
    fmt.Printf("Messages: %d\n", rowCount)
}
```

## Accessing Raw Properties

```go
bag := folder.PropertyBag()

// List all properties
for _, propID := range bag.Properties() {
    propType, _ := bag.GetType(propID)
    fmt.Printf("Property 0x%04X (type 0x%04X)\n", propID, propType)
}

// Read specific properties
displayName, err := bag.GetString(ltp.PidTagDisplayName)
```

## Common Folder Names

Standard Outlook folders have well-known names:

| Folder | Description |
|--------|-------------|
| Inbox | Incoming mail |
| Sent Items | Sent mail |
| Deleted Items | Trash |
| Drafts | Unsent messages |
| Outbox | Outgoing queue |
| Calendar | Appointments |
| Contacts | Address book |
| Tasks | To-do items |
| Notes | Sticky notes |
| Journal | Activity log |

!!! note "Localized Names"
    Folder names may be localized. The actual names depend on the Outlook language settings when the PST was created.
