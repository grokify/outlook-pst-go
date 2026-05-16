# Creating Folders

This guide covers how to create, rename, move, copy, and delete folders in PST files.

## Creating a Folder

Use `CreateFolder()` within a transaction:

```go
ctx, err := pst.BeginWrite()
if err != nil {
    log.Fatal(err)
}

root, _ := pst.RootFolder()

// Create a new folder
inbox, err := ctx.CreateFolder(root, "Inbox")
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

ctx.Commit()
```

## Creating Nested Folders

Create folders at any depth:

```go
ctx, _ := pst.BeginWrite()
root, _ := pst.RootFolder()

// Create hierarchy
archive, _ := ctx.CreateFolder(root, "Archive")
archive2023, _ := ctx.CreateFolder(archive, "2023")
archive2024, _ := ctx.CreateFolder(archive, "2024")

// Create subfolders
ctx.CreateFolder(archive2023, "Q1")
ctx.CreateFolder(archive2023, "Q2")
ctx.CreateFolder(archive2023, "Q3")
ctx.CreateFolder(archive2023, "Q4")

ctx.Commit()
```

## Creating Standard Folders

Create a typical email folder structure:

```go
func createStandardFolders(pst *outlookpst.PST) error {
    ctx, err := pst.BeginWrite()
    if err != nil {
        return err
    }

    root, _ := pst.RootFolder()

    folders := []string{
        "Inbox",
        "Drafts",
        "Sent Items",
        "Deleted Items",
        "Outbox",
        "Junk Email",
    }

    for _, name := range folders {
        if _, err := ctx.CreateFolder(root, name); err != nil {
            ctx.Rollback()
            return fmt.Errorf("failed to create %s: %w", name, err)
        }
    }

    return ctx.Commit()
}
```

## Get or Create Pattern

Check if a folder exists before creating:

```go
func getOrCreateFolder(ctx *outlookpst.WriteContext, parent *outlookpst.Folder, name string) (*outlookpst.Folder, error) {
    // Try to find existing folder
    folder, err := parent.FindSubfolder(name)
    if err == nil {
        return folder, nil
    }

    // Create if not found
    return ctx.CreateFolder(parent, name)
}
```

Or use the built-in helper:

```go
// outlookpst.GetOrCreateFolder handles the check automatically
folder, err := outlookpst.GetOrCreateFolder(ctx, parent, "Archive")
```

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

folder, _ := root.FindSubfolder("Projects")
archive, _ := root.FindSubfolder("Archive")

// Move Projects folder under Archive
err := outlookpst.MoveFolder(ctx, folder, archive)
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

ctx.Commit()
```

!!! warning "Move Restrictions"
    You cannot move a folder into itself or one of its subfolders.

## Copying Folders

Copy a folder and all its contents:

```go
ctx, _ := pst.BeginWrite()

source, _ := root.FindSubfolder("Template Folder")
destination, _ := root.FindSubfolder("Archive")

// Copy folder (including all messages and subfolders)
newFolder, err := outlookpst.CopyFolder(ctx, source, destination)
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

ctx.Commit()
```

## Deleting Folders

Delete a folder and all its contents recursively:

```go
ctx, _ := pst.BeginWrite()

folder, _ := root.FindSubfolder("Old Archive")

// Delete folder and all contents
err := ctx.DeleteFolder(folder)
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

ctx.Commit()
```

!!! danger "Permanent Deletion"
    `DeleteFolder()` permanently removes the folder and all its contents (messages, attachments, subfolders). This cannot be undone.

## Emptying Folders

Remove all contents without deleting the folder:

```go
ctx, _ := pst.BeginWrite()

folder, _ := root.FindSubfolder("Trash")

// Remove all messages and subfolders
err := outlookpst.EmptyFolder(ctx, folder)
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

ctx.Commit()
```

## Checking Folder Existence

```go
exists, err := outlookpst.FolderExists(parent, "Inbox")
if err != nil {
    log.Fatal(err)
}

if exists {
    fmt.Println("Inbox exists")
} else {
    fmt.Println("Inbox does not exist")
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
    // Create PST
    pst, err := outlookpst.Create("folders.pst", disk.FormatUnicode)
    if err != nil {
        log.Fatal(err)
    }
    defer pst.Close()

    // Create folder structure
    ctx, _ := pst.BeginWrite()
    root, _ := pst.RootFolder()

    // Create main folders
    inbox, _ := ctx.CreateFolder(root, "Inbox")
    sent, _ := ctx.CreateFolder(root, "Sent Items")
    drafts, _ := ctx.CreateFolder(root, "Drafts")
    trash, _ := ctx.CreateFolder(root, "Deleted Items")

    // Create archive structure
    archive, _ := ctx.CreateFolder(root, "Archive")
    ctx.CreateFolder(archive, "2023")
    ctx.CreateFolder(archive, "2024")

    // Create project folders in Inbox
    projects, _ := ctx.CreateFolder(inbox, "Projects")
    ctx.CreateFolder(projects, "Project A")
    ctx.CreateFolder(projects, "Project B")

    ctx.Commit()

    // Print folder tree
    fmt.Println("Created folder structure:")
    printFolderTree(root, "")
}

func printFolderTree(folder *outlookpst.Folder, indent string) {
    name, _ := folder.Name()
    fmt.Printf("%s- %s\n", indent, name)

    for subfolder, err := range folder.Subfolders() {
        if err != nil {
            continue
        }
        printFolderTree(subfolder, indent+"  ")
    }
}
```

## Next Steps

- [Creating Messages](creating-messages.md) - Add messages to folders
- [Modifying Content](modifying-content.md) - Update and delete items
