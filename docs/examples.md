# Examples

## Basic Usage

### Open and Inspect a PST File

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

    fmt.Printf("Format: %s\n", pst.Format())
    fmt.Printf("Encryption: %s\n", pst.CryptMethod())

    name, _ := pst.Name()
    fmt.Printf("Name: %s\n", name)
}
```

### List All Folders

```go
func listFolders(folder *outlookpst.Folder, indent string) {
    name, _ := folder.Name()
    count, _ := folder.ContentCount()
    fmt.Printf("%s[%s] (%d items)\n", indent, name, count)

    for subfolder, err := range folder.Subfolders() {
        if err != nil {
            continue
        }
        listFolders(subfolder, indent+"  ")
    }
}

func main() {
    pst, _ := outlookpst.Open("archive.pst")
    defer pst.Close()

    root, _ := pst.RootFolder()
    listFolders(root, "")
}
```

### Read Messages from Inbox

```go
func main() {
    pst, _ := outlookpst.Open("archive.pst")
    defer pst.Close()

    inbox, err := pst.OpenFolder("Inbox")
    if err != nil {
        log.Fatal(err)
    }

    for msg, err := range inbox.Messages() {
        if err != nil {
            continue
        }

        subject, _ := msg.Subject()
        sender, _ := msg.SenderName()
        deliveryTime, _ := msg.DeliveryTime()

        fmt.Printf("From: %s\n", sender)
        fmt.Printf("Subject: %s\n", subject)
        fmt.Printf("Date: %s\n", deliveryTime.Format("2006-01-02 15:04"))
        fmt.Println("---")
    }
}
```

## Email Processing

### Export Messages to Text Files

```go
func exportMessages(folder *outlookpst.Folder, outputDir string) error {
    os.MkdirAll(outputDir, 0755)

    count := 0
    for msg, err := range folder.Messages() {
        if err != nil {
            continue
        }

        subject, _ := msg.Subject()
        body, _ := msg.Body()
        sender, _ := msg.SenderName()
        deliveryTime, _ := msg.DeliveryTime()

        // Create filename from subject
        filename := fmt.Sprintf("%04d_%s.txt", count,
            sanitizeFilename(subject))
        filepath := path.Join(outputDir, filename)

        content := fmt.Sprintf("From: %s\nDate: %s\nSubject: %s\n\n%s",
            sender, deliveryTime.Format(time.RFC1123), subject, body)

        os.WriteFile(filepath, []byte(content), 0644)
        count++
    }

    fmt.Printf("Exported %d messages to %s\n", count, outputDir)
    return nil
}
```

### Search Messages by Subject

```go
func searchBySubject(folder *outlookpst.Folder, query string) []*outlookpst.Message {
    var matches []*outlookpst.Message
    query = strings.ToLower(query)

    for msg, err := range folder.Messages() {
        if err != nil {
            continue
        }

        subject, _ := msg.Subject()
        if strings.Contains(strings.ToLower(subject), query) {
            matches = append(matches, msg)
        }
    }

    return matches
}
```

## Attachment Handling

### Extract All Attachments

```go
func extractAttachments(folder *outlookpst.Folder, outputDir string) error {
    os.MkdirAll(outputDir, 0755)

    for msg, err := range folder.Messages() {
        if err != nil {
            continue
        }

        for att, err := range msg.Attachments() {
            if err != nil {
                continue
            }

            filename, _ := att.Filename()
            if filename == "" {
                continue
            }

            data, err := att.Data()
            if err != nil {
                log.Printf("Failed to read %s: %v", filename, err)
                continue
            }

            // Handle duplicate filenames
            outputPath := uniquePath(outputDir, filename)
            os.WriteFile(outputPath, data, 0644)

            size, _ := att.Size()
            fmt.Printf("Extracted: %s (%d bytes)\n", filename, size)
        }
    }

    return nil
}
```

### Process Embedded Messages

```go
func processEmbeddedMessages(msg *outlookpst.Message, depth int) {
    indent := strings.Repeat("  ", depth)
    subject, _ := msg.Subject()
    fmt.Printf("%sMessage: %s\n", indent, subject)

    for att, err := range msg.Attachments() {
        if err != nil {
            continue
        }

        isEmbedded, _ := att.IsEmbeddedMessage()
        if isEmbedded {
            embeddedMsg, err := att.OpenAsMessage()
            if err != nil {
                continue
            }
            // Recursively process
            processEmbeddedMessages(embeddedMsg, depth+1)
        }
    }
}
```

## Statistics

### Generate Folder Statistics

```go
type FolderStats struct {
    Name         string
    MessageCount int
    TotalSize    int64
    Subfolders   int
}

func getFolderStats(folder *outlookpst.Folder) FolderStats {
    name, _ := folder.Name()
    count, _ := folder.MessageCount()

    var totalSize int64
    for msg, err := range folder.Messages() {
        if err != nil {
            continue
        }
        size, _ := msg.MessageSize()
        totalSize += int64(size)
    }

    subfolderCount, _ := folder.SubfolderCount()

    return FolderStats{
        Name:         name,
        MessageCount: count,
        TotalSize:    totalSize,
        Subfolders:   subfolderCount,
    }
}
```

### Count Messages by Sender

```go
func countBySender(folder *outlookpst.Folder) map[string]int {
    counts := make(map[string]int)

    for msg, err := range folder.Messages() {
        if err != nil {
            continue
        }

        sender, _ := msg.SenderEmail()
        if sender == "" {
            sender, _ = msg.SenderName()
        }
        if sender == "" {
            sender = "(unknown)"
        }

        counts[sender]++
    }

    return counts
}
```

## Export to Other Formats

### Export to JSON

```go
type MessageJSON struct {
    Subject     string    `json:"subject"`
    Sender      string    `json:"sender"`
    SenderEmail string    `json:"sender_email"`
    To          string    `json:"to"`
    Date        time.Time `json:"date"`
    Body        string    `json:"body"`
}

func exportToJSON(folder *outlookpst.Folder, outputPath string) error {
    var messages []MessageJSON

    for msg, err := range folder.Messages() {
        if err != nil {
            continue
        }

        subject, _ := msg.Subject()
        sender, _ := msg.SenderName()
        senderEmail, _ := msg.SenderEmail()
        to, _ := msg.DisplayTo()
        deliveryTime, _ := msg.DeliveryTime()
        body, _ := msg.Body()

        messages = append(messages, MessageJSON{
            Subject:     subject,
            Sender:      sender,
            SenderEmail: senderEmail,
            To:          to,
            Date:        deliveryTime,
            Body:        body,
        })
    }

    data, err := json.MarshalIndent(messages, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(outputPath, data, 0644)
}
```

### Export to CSV

```go
func exportToCSV(folder *outlookpst.Folder, outputPath string) error {
    file, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Header
    writer.Write([]string{"Subject", "From", "To", "Date", "Size"})

    for msg, err := range folder.Messages() {
        if err != nil {
            continue
        }

        subject, _ := msg.Subject()
        sender, _ := msg.SenderName()
        to, _ := msg.DisplayTo()
        deliveryTime, _ := msg.DeliveryTime()
        size, _ := msg.MessageSize()

        writer.Write([]string{
            subject,
            sender,
            to,
            deliveryTime.Format(time.RFC3339),
            fmt.Sprintf("%d", size),
        })
    }

    return nil
}
```

## Creating PST Files

### Create a New PST with Folders

```go
package main

import (
    "fmt"
    "log"

    outlookpst "github.com/grokify/outlook-pst-go"
    "github.com/grokify/outlook-pst-go/pkg/disk"
)

func main() {
    pst, err := outlookpst.Create("new-archive.pst", disk.FormatUnicode)
    if err != nil {
        log.Fatal(err)
    }
    defer pst.Close()

    ctx, _ := pst.BeginWrite()
    root, _ := pst.RootFolder()

    // Create standard folder structure
    folders := []string{"Inbox", "Sent Items", "Drafts", "Deleted Items", "Outbox"}
    for _, name := range folders {
        if _, err := ctx.CreateFolder(root, name); err != nil {
            ctx.Rollback()
            log.Fatalf("Failed to create %s: %v", name, err)
        }
    }

    ctx.Commit()
    fmt.Println("PST created with standard folders")
}
```

### Create Messages with Attachments

```go
func createMessageWithAttachments(pst *outlookpst.PST) error {
    ctx, err := pst.BeginWrite()
    if err != nil {
        return err
    }

    inbox, _ := pst.OpenFolder("Inbox")

    // Read attachment file
    pdfData, err := os.ReadFile("report.pdf")
    if err != nil {
        ctx.Rollback()
        return err
    }

    _, err = ctx.CreateMessage(inbox).
        SetSubject("Q4 Report with Attachments").
        SetBody("Please find the report attached.").
        SetFrom("Finance", "finance@company.com").
        AddTo("CEO", "ceo@company.com").
        AddCC("CFO", "cfo@company.com").
        SetSentTime(time.Now()).
        AddAttachmentWithMime("Q4-Report.pdf", pdfData, "application/pdf").
        Build()

    if err != nil {
        ctx.Rollback()
        return err
    }

    return ctx.Commit()
}
```

### Import Emails from Another Source

```go
type EmailData struct {
    Subject    string
    Body       string
    From       string
    FromEmail  string
    To         []string
    Date       time.Time
}

func importEmails(pst *outlookpst.PST, emails []EmailData) error {
    ctx, err := pst.BeginWrite()
    if err != nil {
        return err
    }

    root, _ := pst.RootFolder()
    inbox, _ := outlookpst.GetOrCreateFolder(ctx, root, "Imported")

    for _, email := range emails {
        builder := ctx.CreateMessage(inbox).
            SetSubject(email.Subject).
            SetBody(email.Body).
            SetFrom(email.From, email.FromEmail).
            SetSentTime(email.Date)

        for _, to := range email.To {
            builder.AddTo("", to)
        }

        if _, err := builder.Build(); err != nil {
            log.Printf("Failed to import '%s': %v", email.Subject, err)
        }
    }

    return ctx.Commit()
}
```

### Archive Old Messages

```go
func archiveOldMessages(pst *outlookpst.PST, olderThan time.Time) error {
    ctx, err := pst.BeginWrite()
    if err != nil {
        return err
    }

    root, _ := pst.RootFolder()
    inbox, _ := root.FindSubfolder("Inbox")
    archive, _ := outlookpst.GetOrCreateFolder(ctx, root, "Archive")

    archived := 0
    for msg, err := range inbox.Messages() {
        if err != nil {
            continue
        }

        sentTime, _ := msg.SubmitTime()
        if sentTime.Before(olderThan) {
            if err := outlookpst.MoveMessage(ctx, msg, archive); err == nil {
                archived++
            }
        }
    }

    if err := ctx.Commit(); err != nil {
        return err
    }

    fmt.Printf("Archived %d messages\n", archived)
    return nil
}
```

### Copy PST Contents

```go
func copyPSTContents(srcPath, dstPath string) error {
    // Open source
    src, err := outlookpst.Open(srcPath)
    if err != nil {
        return fmt.Errorf("failed to open source: %w", err)
    }
    defer src.Close()

    // Create destination
    dst, err := outlookpst.Create(dstPath, src.Format())
    if err != nil {
        return fmt.Errorf("failed to create destination: %w", err)
    }
    defer dst.Close()

    ctx, _ := dst.BeginWrite()
    srcRoot, _ := src.RootFolder()
    dstRoot, _ := dst.RootFolder()

    // Copy recursively
    copyFolder(ctx, srcRoot, dstRoot)

    return ctx.Commit()
}

func copyFolder(ctx *outlookpst.WriteContext, src, dst *outlookpst.Folder) {
    // Copy messages
    for msg, err := range src.Messages() {
        if err != nil {
            continue
        }

        subject, _ := msg.Subject()
        body, _ := msg.Body()
        sentTime, _ := msg.SubmitTime()

        ctx.CreateMessage(dst).
            SetSubject(subject).
            SetBody(body).
            SetSentTime(sentTime).
            Build()
    }

    // Copy subfolders
    for subfolder, err := range src.Subfolders() {
        if err != nil {
            continue
        }

        name, _ := subfolder.Name()
        newFolder, err := ctx.CreateFolder(dst, name)
        if err != nil {
            continue
        }

        copyFolder(ctx, subfolder, newFolder)
    }
}
```

### Compact a PST File

```go
func compactPST(inputPath, outputPath string) error {
    return outlookpst.Compact(inputPath, outputPath, outlookpst.CompactOptions{
        RemoveDeletedItems: true,
        DefragmentBlocks:   true,
    })
}
```

### Bulk Delete Old Messages

```go
func deleteOldMessages(pst *outlookpst.PST, folder *outlookpst.Folder, olderThan time.Time) (int, error) {
    ctx, err := pst.BeginWrite()
    if err != nil {
        return 0, err
    }

    deleted := 0
    var toDelete []*outlookpst.Message

    // Collect messages to delete (can't delete while iterating)
    for msg, err := range folder.Messages() {
        if err != nil {
            continue
        }

        sentTime, _ := msg.SubmitTime()
        if sentTime.Before(olderThan) {
            toDelete = append(toDelete, msg)
        }
    }

    // Delete collected messages
    for _, msg := range toDelete {
        if err := ctx.DeleteMessage(msg); err == nil {
            deleted++
        }
    }

    if err := ctx.Commit(); err != nil {
        return 0, err
    }

    return deleted, nil
}
```
