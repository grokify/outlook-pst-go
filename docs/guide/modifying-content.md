# Modifying Content

This guide covers how to update and delete existing folders and messages.

## Deleting Messages

Delete a message from a folder:

```go
ctx, _ := pst.BeginWrite()

// Find the message to delete
for msg, err := range folder.Messages() {
    if err != nil {
        continue
    }

    subject, _ := msg.Subject()
    if subject == "Old Message" {
        if err := ctx.DeleteMessage(msg); err != nil {
            ctx.Rollback()
            log.Fatal(err)
        }
        break
    }
}

ctx.Commit()
```

Or using the helper function:

```go
err := outlookpst.DeleteMessage(ctx, message)
```

!!! danger "Permanent Deletion"
    `DeleteMessage()` permanently removes the message. This cannot be undone.

## Moving Messages

Move a message to a different folder:

```go
ctx, _ := pst.BeginWrite()

inbox, _ := root.FindSubfolder("Inbox")
archive, _ := root.FindSubfolder("Archive")

for msg, _ := range inbox.Messages() {
    subject, _ := msg.Subject()
    if strings.Contains(subject, "2023") {
        // Move to archive
        if err := outlookpst.MoveMessage(ctx, msg, archive); err != nil {
            log.Printf("Failed to move: %v", err)
        }
    }
}

ctx.Commit()
```

## Copying Messages

Copy a message to another folder (original remains):

```go
ctx, _ := pst.BeginWrite()

source, _ := root.FindSubfolder("Templates")
inbox, _ := root.FindSubfolder("Inbox")

// Copy all template messages to inbox
for msg, _ := range source.Messages() {
    if err := outlookpst.CopyMessage(ctx, msg, inbox); err != nil {
        log.Printf("Copy failed: %v", err)
    }
}

ctx.Commit()
```

## Updating Message Properties

Update properties on an existing message:

```go
ctx, _ := pst.BeginWrite()

updates := map[ltp.PropID]interface{}{
    ltp.PidTagSubject: "Updated Subject",
    ltp.PidTagImportance: int32(2), // High
}

err := outlookpst.UpdateMessage(ctx, message, updates)
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

ctx.Commit()
```

## Marking Messages Read/Unread

```go
ctx, _ := pst.BeginWrite()

// Mark as read
err := outlookpst.MarkAsRead(ctx, message)

// Mark as unread
err := outlookpst.MarkAsUnread(ctx, message)

ctx.Commit()
```

## Deleting Folders

Delete a folder and all its contents:

```go
ctx, _ := pst.BeginWrite()

folder, _ := root.FindSubfolder("Old Archive")

if err := ctx.DeleteFolder(folder); err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

ctx.Commit()
```

!!! danger "Recursive Deletion"
    `DeleteFolder()` deletes the folder and ALL contents recursively, including:

    - All messages in the folder
    - All attachments
    - All subfolders and their contents

## Renaming Folders

```go
ctx, _ := pst.BeginWrite()

folder, _ := root.FindSubfolder("Old Name")

err := outlookpst.RenameFolder(ctx, folder, "New Name")
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

ctx.Commit()
```

## Moving Folders

Move a folder to a new parent:

```go
ctx, _ := pst.BeginWrite()

projects, _ := root.FindSubfolder("Projects")
archive, _ := root.FindSubfolder("Archive")

// Move Projects folder to Archive
err := outlookpst.MoveFolder(ctx, projects, archive)
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

ctx.Commit()
```

## Emptying Folders

Remove all contents without deleting the folder:

```go
ctx, _ := pst.BeginWrite()

trash, _ := root.FindSubfolder("Deleted Items")

// Empty the trash
err := outlookpst.EmptyFolder(ctx, trash)
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

ctx.Commit()
```

## Updating Folder Properties

Update a folder's properties:

```go
ctx, _ := pst.BeginWrite()

// Update display name
err := outlookpst.UpdateFolderProperty(ctx, folder, ltp.PidTagDisplayName, "New Name")

ctx.Commit()
```

## Bulk Operations

Process multiple items efficiently:

```go
func archiveOldMessages(pst *outlookpst.PST, olderThan time.Time) error {
    ctx, err := pst.BeginWrite()
    if err != nil {
        return err
    }

    inbox, _ := pst.RootFolder()
    inbox, _ = inbox.FindSubfolder("Inbox")
    archive, _ := outlookpst.GetOrCreateFolder(ctx, inbox, "Archive")

    count := 0
    for msg, err := range inbox.Messages() {
        if err != nil {
            continue
        }

        sentTime, err := msg.SubmitTime()
        if err != nil {
            continue
        }

        if sentTime.Before(olderThan) {
            if err := outlookpst.MoveMessage(ctx, msg, archive); err != nil {
                log.Printf("Failed to archive message: %v", err)
                continue
            }
            count++
        }
    }

    if err := ctx.Commit(); err != nil {
        return err
    }

    fmt.Printf("Archived %d messages\n", count)
    return nil
}
```

## Compacting PST Files

Remove unused space after deleting content:

```go
// Create a compacted copy
err := outlookpst.Compact("original.pst", "compacted.pst", outlookpst.CompactOptions{
    RemoveDeletedItems: true,  // Skip "Deleted Items" folder
    DefragmentBlocks:   true,  // Consolidate fragmented blocks
})
```

## PST Statistics

Get information about PST contents:

```go
stats, err := pst.GetStatistics()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Format: %s\n", stats.Format)
fmt.Printf("File Size: %d bytes\n", stats.FileSize)
fmt.Printf("Free Space: %d bytes\n", stats.FreeSpace)
fmt.Printf("Folders: %d\n", stats.FolderCount)
fmt.Printf("Messages: %d\n", stats.MessageCount)
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "time"

    outlookpst "github.com/grokify/outlook-pst-go"
)

func main() {
    pst, err := outlookpst.OpenReadWrite("archive.pst")
    if err != nil {
        log.Fatal(err)
    }
    defer pst.Close()

    ctx, _ := pst.BeginWrite()
    root, _ := pst.RootFolder()

    // Create archive folder if needed
    archive, _ := outlookpst.GetOrCreateFolder(ctx, root, "Archive")

    // Find inbox
    inbox, _ := root.FindSubfolder("Inbox")

    // Archive messages older than 30 days
    cutoff := time.Now().AddDate(0, 0, -30)
    archived := 0

    for msg, err := range inbox.Messages() {
        if err != nil {
            continue
        }

        sentTime, _ := msg.SubmitTime()
        if sentTime.Before(cutoff) {
            if err := outlookpst.MoveMessage(ctx, msg, archive); err == nil {
                archived++
            }
        }
    }

    // Delete empty subfolders
    for folder, _ := range inbox.Subfolders() {
        count, _ := folder.ContentCount()
        subCount, _ := folder.SubfolderCount()

        if count == 0 && subCount == 0 {
            name, _ := folder.Name()
            fmt.Printf("Deleting empty folder: %s\n", name)
            ctx.DeleteFolder(folder)
        }
    }

    if err := ctx.Commit(); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Archived %d messages\n", archived)
}
```

## Next Steps

- [Error Handling](error-handling.md) - Handle errors gracefully
- [Transactions](transactions.md) - Transaction best practices
