# Quick Start

This guide walks you through the basics of reading and writing PST files.

## Opening a PST File

```go
pst, err := outlookpst.Open("archive.pst")
if err != nil {
    log.Fatal(err)
}
defer pst.Close()
```

## Checking File Properties

```go
// Check format (ANSI or Unicode)
fmt.Printf("Format: %s\n", pst.Format())

// Check if PST or OST
if pst.IsPST() {
    fmt.Println("This is a PST file")
}

// Check encryption method
fmt.Printf("Encryption: %s\n", pst.CryptMethod())

// Get display name
name, err := pst.Name()
if err == nil {
    fmt.Printf("Name: %s\n", name)
}
```

## Navigating Folders

```go
// Get the root folder
root, err := pst.RootFolder()
if err != nil {
    log.Fatal(err)
}

// Iterate through subfolders
for folder, err := range root.Subfolders() {
    if err != nil {
        log.Printf("Error: %v", err)
        continue
    }

    name, _ := folder.Name()
    count, _ := folder.ContentCount()
    fmt.Printf("Folder: %s (%d items)\n", name, count)
}
```

## Reading Messages

```go
for msg, err := range folder.Messages() {
    if err != nil {
        continue
    }

    subject, _ := msg.Subject()
    sender, _ := msg.SenderName()
    deliveryTime, _ := msg.DeliveryTime()

    fmt.Printf("Subject: %s\n", subject)
    fmt.Printf("From: %s\n", sender)
    fmt.Printf("Date: %s\n", deliveryTime.Format("2006-01-02 15:04"))
    fmt.Println()
}
```

## Complete Example

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

    root, err := pst.RootFolder()
    if err != nil {
        log.Fatal(err)
    }

    printFolder(root, "")
}

func printFolder(folder *outlookpst.Folder, indent string) {
    name, _ := folder.Name()
    count, _ := folder.ContentCount()
    fmt.Printf("%s[%s] (%d items)\n", indent, name, count)

    // Print messages
    for msg, err := range folder.Messages() {
        if err != nil {
            continue
        }
        subject, _ := msg.Subject()
        fmt.Printf("%s  - %s\n", indent, subject)
    }

    // Recurse into subfolders
    for subfolder, err := range folder.Subfolders() {
        if err != nil {
            continue
        }
        printFolder(subfolder, indent+"  ")
    }
}
```

## Creating a PST File

```go
import "github.com/grokify/outlook-pst-go/pkg/disk"

// Create a new Unicode PST file
pst, err := outlookpst.Create("new-archive.pst", disk.FormatUnicode)
if err != nil {
    log.Fatal(err)
}
defer pst.Close()
```

## Writing Content

All write operations use transactions:

```go
// Begin a write transaction
ctx, err := pst.BeginWrite()
if err != nil {
    log.Fatal(err)
}

// Create a folder
root, _ := pst.RootFolder()
inbox, err := ctx.CreateFolder(root, "Inbox")
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

// Create a message
msg, err := ctx.CreateMessage(inbox).
    SetSubject("Hello World").
    SetBody("This is a test message.").
    SetFrom("sender@example.com", "Sender Name").
    AddTo("recipient@example.com", "Recipient").
    Build()
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

// Commit the transaction
if err := ctx.Commit(); err != nil {
    log.Fatal(err)
}
```

## Complete Write Example

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
    pst, err := outlookpst.Create("my-archive.pst", disk.FormatUnicode)
    if err != nil {
        log.Fatal(err)
    }
    defer pst.Close()

    // Begin transaction
    ctx, _ := pst.BeginWrite()
    root, _ := pst.RootFolder()

    // Create folders
    inbox, _ := ctx.CreateFolder(root, "Inbox")
    sent, _ := ctx.CreateFolder(root, "Sent Items")

    // Create messages
    ctx.CreateMessage(inbox).
        SetSubject("Welcome!").
        SetBody("Welcome to your new PST file.").
        SetSentTime(time.Now()).
        Build()

    ctx.CreateMessage(sent).
        SetSubject("Test Email").
        SetBody("This is a test.").
        AddTo("friend@example.com", "Friend").
        SetSentTime(time.Now()).
        Build()

    // Commit
    ctx.Commit()

    fmt.Println("PST created successfully!")
}
```

## Next Steps

### Reading

- Learn about [Working with Folders](../guide/folders.md)
- Learn about [Reading Messages](../guide/messages.md)

### Writing

- Learn about [Creating PST Files](../guide/creating-pst-files.md)
- Learn about [Transactions](../guide/transactions.md)
- Learn about [Creating Messages](../guide/creating-messages.md)

### Reference

- Explore the [API Reference](../api/pst.md)
