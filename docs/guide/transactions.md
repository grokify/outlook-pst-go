# Transactions

All write operations in outlook-pst-go are performed within transactions. This ensures data integrity and allows for atomic commits or rollbacks.

## Why Transactions?

PST files have complex internal structures (B-trees, allocation maps, etc.) that must remain consistent. Transactions provide:

- **Atomicity** - All changes commit together or not at all
- **Crash Recovery** - Incomplete writes can be recovered
- **Rollback** - Discard changes if something goes wrong

## Basic Transaction Pattern

```go
// Start a write transaction
ctx, err := pst.BeginWrite()
if err != nil {
    log.Fatal(err)
}

// Perform write operations
folder, err := ctx.CreateFolder(root, "New Folder")
if err != nil {
    ctx.Rollback() // Discard changes
    log.Fatal(err)
}

// Commit all changes
if err := ctx.Commit(); err != nil {
    log.Fatal(err)
}
```

## WriteContext

The `WriteContext` object provides all write operations:

```go
ctx, err := pst.BeginWrite()
if err != nil {
    log.Fatal(err)
}

// Available operations:
// ctx.CreateFolder(parent, name)
// ctx.CreateMessage(folder)
// ctx.DeleteFolder(folder)
// ctx.DeleteMessage(msg)
```

## Committing Changes

`Commit()` writes all pending changes to disk:

```go
if err := ctx.Commit(); err != nil {
    log.Fatal(err)
}
```

After commit:

- All changes are written to disk
- The transaction is closed
- The WriteContext cannot be reused

## Rolling Back

`Rollback()` discards all pending changes:

```go
folder, err := ctx.CreateFolder(root, "Test")
if err != nil {
    ctx.Rollback() // Discard the failed operation
    log.Fatal(err)
}

// Changed your mind? Rollback before commit
ctx.Rollback()
```

## Error Handling Pattern

Use defer for cleanup:

```go
func addContent(pst *outlookpst.PST) error {
    ctx, err := pst.BeginWrite()
    if err != nil {
        return err
    }

    // Ensure cleanup on any exit path
    committed := false
    defer func() {
        if !committed {
            ctx.Rollback()
        }
    }()

    // Perform operations
    root, _ := pst.RootFolder()
    _, err = ctx.CreateFolder(root, "Archive")
    if err != nil {
        return err // Rollback will be called by defer
    }

    // Commit
    if err := ctx.Commit(); err != nil {
        return err
    }
    committed = true
    return nil
}
```

## Two-Phase Commit

Internally, transactions use a two-phase commit protocol:

1. **Phase 1 (Start)** - Mark allocation map as invalid
2. **Write** - Perform all pending writes
3. **Phase 2 (Finish)** - Mark allocation map as valid

This ensures that incomplete writes can be detected and recovered.

## Crash Recovery

If a crash occurs during a transaction:

```go
// OpenReadWrite automatically checks for recovery
pst, err := outlookpst.OpenReadWrite("archive.pst")
if err != nil {
    log.Fatal(err)
}
// Recovery is performed automatically if needed
```

You can also check recovery status explicitly:

```go
info, err := outlookpst.CheckRecovery("archive.pst")
if err != nil {
    log.Fatal(err)
}

if info.NeedsRecovery {
    fmt.Printf("Recovery needed, status: %v\n", info.Status)
}
```

## Transaction Limitations

- Only one active transaction per PST file
- Transactions should be short-lived
- Don't hold transactions across user interactions

```go
// BAD: Long-running transaction
ctx, _ := pst.BeginWrite()
// ... wait for user input ...  // DON'T DO THIS
ctx.Commit()

// GOOD: Transaction for each operation
ctx, _ := pst.BeginWrite()
ctx.CreateFolder(root, "Folder")
ctx.Commit()
```

## Nested Operations

Multiple operations can be performed in a single transaction:

```go
ctx, _ := pst.BeginWrite()

root, _ := pst.RootFolder()

// Create multiple folders
inbox, _ := ctx.CreateFolder(root, "Inbox")
sent, _ := ctx.CreateFolder(root, "Sent Items")
drafts, _ := ctx.CreateFolder(root, "Drafts")

// Create messages in folders
ctx.CreateMessage(inbox).
    SetSubject("Welcome").
    SetBody("Welcome to your new inbox!").
    Build()

// Commit all changes at once
ctx.Commit()
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
    // Create PST
    pst, err := outlookpst.Create("archive.pst", disk.FormatUnicode)
    if err != nil {
        log.Fatal(err)
    }
    defer pst.Close()

    // Begin transaction
    ctx, err := pst.BeginWrite()
    if err != nil {
        log.Fatal(err)
    }

    // Track commit status for cleanup
    committed := false
    defer func() {
        if !committed {
            fmt.Println("Rolling back transaction")
            ctx.Rollback()
        }
    }()

    // Create folder structure
    root, _ := pst.RootFolder()
    inbox, err := ctx.CreateFolder(root, "Inbox")
    if err != nil {
        log.Fatal(err)
    }

    // Add a message
    _, err = ctx.CreateMessage(inbox).
        SetSubject("Test Message").
        SetBody("This is a test.").
        SetSentTime(time.Now()).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Commit
    if err := ctx.Commit(); err != nil {
        log.Fatal(err)
    }
    committed = true

    fmt.Println("Transaction committed successfully")
}
```

## Next Steps

- [Creating Folders](creating-folders.md) - Add folders to the PST
- [Creating Messages](creating-messages.md) - Add messages with attachments
- [Modifying Content](modifying-content.md) - Update and delete items
