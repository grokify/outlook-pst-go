# Search Folders

Search folders are virtual folders whose contents are determined by search criteria rather than containing actual messages. They provide dynamic views of messages matching specific conditions.

## Understanding Search Folders

Search folders differ from regular folders:

| Aspect | Regular Folder | Search Folder |
|--------|----------------|---------------|
| Contents | Messages stored in folder | Messages matching criteria |
| NID Type | `NIDTypeNormalFolder` | `NIDTypeSearchFolder` |
| Updates | Manual | Automatic (when criteria matches) |
| Storage | Contains message references | Contains search criteria |

## Detecting Search Folders

```go
pst, _ := outlookpst.Open("archive.pst")
defer pst.Close()

root, _ := pst.RootFolder()

for folder, err := range root.Subfolders() {
    if err != nil {
        continue
    }

    name, _ := folder.Name()

    if folder.IsSearchFolder() {
        fmt.Printf("[Search] %s\n", name)
    } else {
        fmt.Printf("[Normal] %s\n", name)
    }
}
```

## Working with Search Folders

### Convert to SearchFolder

```go
if folder.IsSearchFolder() {
    searchFolder := folder.AsSearchFolder()

    // Access search-specific features
    criteria, err := searchFolder.SearchCriteria()
    if err != nil {
        log.Printf("Failed to get criteria: %v", err)
    }
}
```

### Read Search Criteria

```go
searchFolder := folder.AsSearchFolder()
criteria, _ := searchFolder.SearchCriteria()

// Check search flags
fmt.Printf("Recursive: %v\n", criteria.IsRecursive())
fmt.Printf("Static: %v\n", criteria.IsStatic())
fmt.Printf("Uses content index: %v\n", criteria.UsesContentIndex())

// Get folder entry IDs being searched
fmt.Printf("Searching %d folders\n", len(criteria.FolderEntryIDs))
```

## Search Criteria Flags

| Flag | Method | Description |
|------|--------|-------------|
| Foreground | `IsForeground()` | Search runs in foreground |
| Recursive | `IsRecursive()` | Search includes subfolders |
| ContentIndex | `UsesContentIndex()` | Uses content indexing |
| Static | `IsStatic()` | Search results don't update |

## Search Update Queue

The search update queue tracks pending updates for search folders:

```go
queue, err := pst.SearchUpdateQueue()
if err != nil {
    log.Printf("No search update queue: %v", err)
    return
}

fmt.Printf("Pending updates: %d\n", queue.Count())

if !queue.IsEmpty() {
    for _, entry := range queue.Entries() {
        fmt.Printf("Folder: 0x%X, Message: 0x%X\n",
            entry.FolderNID, entry.MessageNID)
    }
}
```

## Example: List All Search Folders

```go
func listSearchFolders(folder *outlookpst.Folder, indent string) {
    for subfolder, err := range folder.Subfolders() {
        if err != nil {
            continue
        }

        name, _ := subfolder.Name()

        if subfolder.IsSearchFolder() {
            sf := subfolder.AsSearchFolder()
            criteria, _ := sf.SearchCriteria()

            flags := []string{}
            if criteria != nil {
                if criteria.IsRecursive() {
                    flags = append(flags, "recursive")
                }
                if criteria.IsStatic() {
                    flags = append(flags, "static")
                }
            }

            fmt.Printf("%s[Search] %s (%s)\n", indent, name, strings.Join(flags, ", "))
        } else {
            fmt.Printf("%s[Folder] %s\n", indent, name)
        }

        listSearchFolders(subfolder, indent+"  ")
    }
}
```

## Common Search Folders

Outlook typically creates these search folders:

- **Unread Mail**: Messages with unread flag
- **Flagged Items**: Messages with follow-up flags
- **Important**: High importance messages
- **Large Mail**: Messages exceeding size threshold
- **Categorized Mail**: Messages with specific categories

## Restrictions

The search restriction data (`SearchCriteria.Restriction`) contains the binary-encoded search filter. Parsing this requires understanding the MAPI restriction format defined in [MS-OXCDATA].

```go
criteria, _ := searchFolder.SearchCriteria()

// Raw restriction data
if len(criteria.Restriction) > 0 {
    fmt.Printf("Restriction size: %d bytes\n", len(criteria.Restriction))
    // Advanced: parse according to MS-OXCDATA
}
```

## Iterating Search Folder Contents

Search folders support the same iteration methods as regular folders:

```go
if folder.IsSearchFolder() {
    // Iterate messages matching the search criteria
    for msg, err := range folder.Messages() {
        if err != nil {
            continue
        }

        subject, _ := msg.Subject()
        fmt.Printf("Match: %s\n", subject)
    }
}
```

## Specification Reference

Search folders are defined in:

- [MS-PST] Section 2.4.8 - Search
- [MS-OXCDATA] - Restriction structures
