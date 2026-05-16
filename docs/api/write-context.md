# WriteContext

The `WriteContext` represents an active write transaction on a PST file. All write operations must be performed within a `WriteContext`.

## Creating a WriteContext

```go
func (p *PST) BeginWrite() (*WriteContext, error)
```

Begins a new write transaction.

```go
ctx, err := pst.BeginWrite()
if err != nil {
    log.Fatal(err)
}
// Use ctx for write operations
// Then commit or rollback
```

## Transaction Control

### Commit

```go
func (w *WriteContext) Commit() error
```

Commits all pending changes to disk using a two-phase commit protocol.

```go
if err := ctx.Commit(); err != nil {
    log.Fatal(err)
}
```

After commit:

- All changes are written to disk
- The transaction is closed
- The WriteContext cannot be reused

### Rollback

```go
func (w *WriteContext) Rollback() error
```

Discards all pending changes and closes the transaction.

```go
ctx.Rollback()
```

## Folder Operations

### CreateFolder

```go
func (w *WriteContext) CreateFolder(parent *Folder, name string) (*Folder, error)
```

Creates a new subfolder in the given parent folder.

```go
inbox, err := ctx.CreateFolder(root, "Inbox")
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}
```

### DeleteFolder

```go
func (w *WriteContext) DeleteFolder(folder *Folder) error
```

Deletes a folder and all its contents recursively.

```go
err := ctx.DeleteFolder(oldFolder)
```

!!! danger "Recursive Deletion"
    This deletes ALL contents including messages, attachments, and subfolders.

## Message Operations

### CreateMessage

```go
func (w *WriteContext) CreateMessage(folder *Folder) *MessageBuilder
```

Creates a new message builder for the specified folder.

```go
builder := ctx.CreateMessage(inbox)
msg, err := builder.
    SetSubject("Hello").
    SetBody("World").
    Build()
```

See [MessageBuilder](message-builder.md) for the full builder API.

### DeleteMessage

```go
func (w *WriteContext) DeleteMessage(msg *Message) error
```

Deletes a message from its folder.

```go
err := ctx.DeleteMessage(message)
```

## Accessors

### Transaction

```go
func (w *WriteContext) Transaction() *ndb.WriteTransaction
```

Returns the underlying NDB transaction for advanced operations.

```go
txn := ctx.Transaction()
// Use txn for low-level operations
```

### PST

```go
func (w *WriteContext) PST() *PST
```

Returns the parent PST file.

```go
pst := ctx.PST()
format := pst.Format()
```

## Helper Functions

These functions operate within a WriteContext:

### Folder Operations

```go
// Create folder (same as ctx.CreateFolder)
func CreateFolder(ctx *WriteContext, parent *Folder, name string) (*Folder, error)

// Delete folder recursively
func DeleteFolder(ctx *WriteContext, folder *Folder) error

// Rename folder
func RenameFolder(ctx *WriteContext, folder *Folder, newName string) error

// Move folder to new parent
func MoveFolder(ctx *WriteContext, folder, newParent *Folder) error

// Copy folder and contents
func CopyFolder(ctx *WriteContext, folder, newParent *Folder) (*Folder, error)

// Empty folder (remove all contents)
func EmptyFolder(ctx *WriteContext, folder *Folder) error

// Check if folder exists
func FolderExists(parent *Folder, name string) (bool, error)

// Get or create folder
func GetOrCreateFolder(ctx *WriteContext, parent *Folder, name string) (*Folder, error)

// Update folder property
func UpdateFolderProperty(ctx *WriteContext, folder *Folder, propID ltp.PropID, value interface{}) error
```

### Message Operations

```go
// Delete message
func DeleteMessage(ctx *WriteContext, msg *Message) error

// Copy message to folder
func CopyMessage(ctx *WriteContext, msg *Message, destFolder *Folder) error

// Move message to folder
func MoveMessage(ctx *WriteContext, msg *Message, destFolder *Folder) error

// Update message properties
func UpdateMessage(ctx *WriteContext, msg *Message, updates map[ltp.PropID]interface{}) error

// Mark as read
func MarkAsRead(ctx *WriteContext, msg *Message) error

// Mark as unread
func MarkAsUnread(ctx *WriteContext, msg *Message) error
```

## Usage Pattern

```go
func modifyPST(pst *outlookpst.PST) error {
    ctx, err := pst.BeginWrite()
    if err != nil {
        return err
    }

    // Track commit status
    committed := false
    defer func() {
        if !committed {
            ctx.Rollback()
        }
    }()

    // Perform operations
    root, _ := pst.RootFolder()
    inbox, err := ctx.CreateFolder(root, "Inbox")
    if err != nil {
        return err
    }

    _, err = ctx.CreateMessage(inbox).
        SetSubject("Welcome").
        Build()
    if err != nil {
        return err
    }

    // Commit
    if err := ctx.Commit(); err != nil {
        return err
    }
    committed = true
    return nil
}
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "time"

    outlookpst "github.com/grokify/outlook-pst-go"
    "github.com/grokify/outlook-pst-go/pkg/disk"
)

func main() {
    pst, _ := outlookpst.Create("example.pst", disk.FormatUnicode)
    defer pst.Close()

    ctx, err := pst.BeginWrite()
    if err != nil {
        log.Fatal(err)
    }

    root, _ := pst.RootFolder()

    // Create folders
    inbox, _ := ctx.CreateFolder(root, "Inbox")
    sent, _ := ctx.CreateFolder(root, "Sent Items")
    trash, _ := ctx.CreateFolder(root, "Deleted Items")

    // Create messages
    ctx.CreateMessage(inbox).
        SetSubject("Welcome!").
        SetBody("Welcome to your new mailbox.").
        SetSentTime(time.Now()).
        Build()

    ctx.CreateMessage(sent).
        SetSubject("Hello").
        SetBody("Just saying hi.").
        AddTo("Friend", "friend@example.com").
        Build()

    // Commit all changes
    if err := ctx.Commit(); err != nil {
        log.Fatal(err)
    }

    fmt.Println("PST created successfully")
}
```

## See Also

- [MessageBuilder](message-builder.md) - Building messages
- [Transactions Guide](../guide/transactions.md) - Transaction best practices
