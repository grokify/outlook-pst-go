# PST Layer

The PST layer (root package) provides the high-level API for applications.

## Components

### PST (`pst.go`)

The entry point for accessing PST files:

```go
pst, err := outlookpst.Open("archive.pst")
if err != nil {
    log.Fatal(err)
}
defer pst.Close()
```

### Internal Structure

```go
type PST struct {
    db *ndb.Database     // NDB layer access

    // Lazy-loaded components
    storeOnce sync.Once
    storeBag  *ltp.PropertyBag  // Message store properties

    rootOnce   sync.Once
    rootFolder *Folder          // Root folder
}
```

### Folder (`folder.go`)

Represents a folder in the PST:

```go
type Folder struct {
    pst  *PST
    node *ndb.Node
    bag  *ltp.PropertyBag

    // Lazy-loaded tables
    hierarchyTable *ltp.Table  // Subfolders
    contentsTable  *ltp.Table  // Messages
}
```

### Message (`message.go`)

Represents an email message:

```go
type Message struct {
    pst  *PST
    node *ndb.Node
    bag  *ltp.PropertyBag

    // Lazy-loaded tables
    attachmentTable *ltp.Table  // Attachments
    recipientTable  *ltp.Table  // Recipients
}
```

### Attachment

Represents a file attachment:

```go
type Attachment struct {
    pst     *PST
    message *Message
    node    *ndb.Node
    bag     *ltp.PropertyBag
}
```

### Recipient

Represents an email recipient:

```go
type Recipient struct {
    row *ltp.TableRow  // From recipient table
}
```

## Node ID Relationships

```
PST
 │
 ├── Message Store (NID: 0x21)
 │
 └── Root Folder (NID: 0x2442)
      │
      ├── Hierarchy Table (NID: 0x244D)
      │   └── Subfolder entries...
      │
      └── Contents Table (NID: 0x244E)
          └── Message entries...

Message (NID: 0x8024)
 │
 ├── Attachment Table (NID: 0x8031)
 │   └── Attachment entries...
 │
 └── Recipient Table (NID: 0x8032)
     └── Recipient entries...
```

## Table Structure

Folders and messages use tables for child items:

| Table | Purpose | Key Properties |
|-------|---------|----------------|
| Hierarchy | Subfolders | DisplayName, FolderType |
| Contents | Messages | Subject, SenderName, DeliveryTime |
| Attachment | Attachments | Filename, Size, Method |
| Recipient | Recipients | DisplayName, EmailAddress, Type |

## Lazy Loading Pattern

Components are loaded on first access:

```go
func (f *Folder) loadHierarchyTable() error {
    f.hierarchyOnce.Do(func() {
        hierarchyNID := util.MakeNID(
            util.NIDTypeHierarchyTable,
            f.node.ID().Index(),
        )

        hierarchyNode, err := f.node.LookupSubnode(hierarchyNID)
        if err != nil {
            f.hierarchyErr = err
            return
        }

        f.hierarchyTable, f.hierarchyErr = ltp.NewTable(hierarchyNode)
    })
    return f.hierarchyErr
}
```

## Iterator Pattern

Go 1.23+ iterators provide memory-efficient traversal:

```go
func (f *Folder) Subfolders() iter.Seq2[*Folder, error] {
    return func(yield func(*Folder, error) bool) {
        if err := f.loadHierarchyTable(); err != nil {
            yield(nil, err)
            return
        }

        for row, err := range f.hierarchyTable.Rows() {
            if err != nil {
                yield(nil, err)
                return
            }

            // Create Folder from row
            subfolder, err := f.createSubfolder(row)
            if !yield(subfolder, err) {
                return
            }
        }
    }
}
```

## Error Types

Specific error types for better error handling:

```go
// Not found errors
type NodeNotFoundError struct { NodeID uint64 }
type PropertyNotFoundError struct { PropID uint16 }
type FolderNotFoundError struct { Name string }

// Format errors
type InvalidHeaderError struct { Message string }

// Integrity errors
type CRCError struct { Expected, Actual uint32 }
type SignatureError struct { Expected, Actual uint16 }
```

Usage:

```go
folder, err := pst.OpenFolder("NonExistent")
if errors.Is(err, outlookpst.ErrNotFound) {
    // Handle not found
}
```

## Lower Layer Access

For advanced use cases, access lower layers:

```go
// Get NDB database
db := pst.Database()
node, _ := db.GetNode(nid)

// Get property bag
bag := msg.PropertyBag()
rawBytes, _ := bag.GetRaw(propID)

// Get tables
table, _ := folder.HierarchyTable()
for row, _ := range table.Rows() {
    // Direct row access
}
```
