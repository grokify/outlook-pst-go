# Creating Messages

This guide covers how to create email messages with recipients, attachments, and various properties.

## Basic Message Creation

Use `CreateMessage()` with the fluent builder pattern:

```go
ctx, _ := pst.BeginWrite()
inbox, _ := root.FindSubfolder("Inbox")

msg, err := ctx.CreateMessage(inbox).
    SetSubject("Hello World").
    SetBody("This is the message body.").
    Build()

if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

ctx.Commit()
```

## MessageBuilder Methods

The `MessageBuilder` provides a fluent interface:

### Subject and Body

```go
builder := ctx.CreateMessage(folder)

// Set subject
builder.SetSubject("Meeting Tomorrow")

// Set plain text body
builder.SetBody("Please join us for a meeting.")

// Set HTML body (optional)
builder.SetHTMLBody("<h1>Meeting</h1><p>Please join us.</p>")

// Set RTF body (optional, compressed)
builder.SetRTFBody(rtfData)
```

### Sender Information

```go
builder.SetFrom("John Doe", "john@example.com")
```

### Recipients

```go
// Add TO recipients
builder.AddTo("Alice", "alice@example.com")
builder.AddTo("Bob", "bob@example.com")

// Add CC recipients
builder.AddCC("Manager", "manager@example.com")

// Add BCC recipients
builder.AddBCC("Archive", "archive@example.com")

// Or use the generic method with recipient type
builder.AddRecipient("Alice", "alice@example.com", outlookpst.RecipientTo)
builder.AddRecipient("Manager", "manager@example.com", outlookpst.RecipientCc)
```

### Timestamps

```go
import "time"

// Set sent time
builder.SetSentTime(time.Now())

// Set a specific date
sendDate := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
builder.SetSentTime(sendDate)
```

### Attachments

```go
// Simple attachment
data := []byte("File contents here")
builder.AddAttachment("document.txt", data)

// Attachment with MIME type
builder.AddAttachmentWithMime("image.png", imageData, "image/png")
```

### Custom Properties

```go
import "github.com/grokify/outlook-pst-go/pkg/ltp"

// Set custom properties
builder.SetProperty(ltp.PidTagImportance, int32(2)) // High importance
builder.SetProperty(ltp.PidTagSensitivity, int32(2)) // Private
```

## Complete Message Example

```go
ctx, _ := pst.BeginWrite()
inbox, _ := root.FindSubfolder("Inbox")

msg, err := ctx.CreateMessage(inbox).
    SetSubject("Q4 Report").
    SetBody("Please find the Q4 report attached.").
    SetHTMLBody("<h1>Q4 Report</h1><p>Please find the report attached.</p>").
    SetFrom("Finance Team", "finance@company.com").
    AddTo("CEO", "ceo@company.com").
    AddTo("CFO", "cfo@company.com").
    AddCC("Board", "board@company.com").
    SetSentTime(time.Now()).
    AddAttachment("q4-report.pdf", pdfData).
    AddAttachmentWithMime("chart.png", chartData, "image/png").
    Build()

if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

ctx.Commit()
```

## Recipient Types

| Type | Constant | Description |
|------|----------|-------------|
| To | `RecipientTo` | Primary recipient |
| CC | `RecipientCc` | Carbon copy |
| BCC | `RecipientBcc` | Blind carbon copy |

## Attachment Methods

| Method | Constant | Description |
|--------|----------|-------------|
| By Value | `AttachMethodByValue` | Attachment data stored in PST (default) |
| By Reference | `AttachMethodByReference` | Link to external file |
| Embedded | `AttachMethodEmbedded` | Embedded message (email within email) |

## Bulk Message Import

Import multiple messages efficiently:

```go
func importMessages(pst *outlookpst.PST, folder *outlookpst.Folder, emails []Email) error {
    ctx, err := pst.BeginWrite()
    if err != nil {
        return err
    }

    for _, email := range emails {
        builder := ctx.CreateMessage(folder).
            SetSubject(email.Subject).
            SetBody(email.Body).
            SetFrom(email.FromName, email.FromEmail).
            SetSentTime(email.SentTime)

        for _, to := range email.To {
            builder.AddTo(to.Name, to.Email)
        }

        for _, att := range email.Attachments {
            builder.AddAttachment(att.Filename, att.Data)
        }

        if _, err := builder.Build(); err != nil {
            ctx.Rollback()
            return fmt.Errorf("failed to import %s: %w", email.Subject, err)
        }
    }

    return ctx.Commit()
}
```

## Message Properties Reference

Common properties you can set:

| Property | Method | Type |
|----------|--------|------|
| Subject | `SetSubject()` | string |
| Body | `SetBody()` | string |
| HTML Body | `SetHTMLBody()` | string |
| RTF Body | `SetRTFBody()` | []byte |
| Sent Time | `SetSentTime()` | time.Time |
| From | `SetFrom()` | name, email |
| Importance | `SetProperty(PidTagImportance, ...)` | int32 (0=low, 1=normal, 2=high) |
| Sensitivity | `SetProperty(PidTagSensitivity, ...)` | int32 (0=normal, 1=personal, 2=private, 3=confidential) |

## Error Handling

```go
ctx, err := pst.BeginWrite()
if err != nil {
    return err
}

committed := false
defer func() {
    if !committed {
        ctx.Rollback()
    }
}()

msg, err := ctx.CreateMessage(folder).
    SetSubject("Test").
    SetBody("Test body").
    Build()
if err != nil {
    return fmt.Errorf("failed to create message: %w", err)
}

if err := ctx.Commit(); err != nil {
    return fmt.Errorf("failed to commit: %w", err)
}
committed = true

return nil
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
    "github.com/grokify/outlook-pst-go/pkg/ltp"
)

func main() {
    // Create PST with folders
    pst, _ := outlookpst.Create("messages.pst", disk.FormatUnicode)
    defer pst.Close()

    ctx, _ := pst.BeginWrite()
    root, _ := pst.RootFolder()
    inbox, _ := ctx.CreateFolder(root, "Inbox")
    sent, _ := ctx.CreateFolder(root, "Sent Items")

    // Create a received message
    ctx.CreateMessage(inbox).
        SetSubject("Welcome to the Team!").
        SetBody("We're excited to have you join us.").
        SetFrom("HR Department", "hr@company.com").
        AddTo("New Employee", "newbie@company.com").
        SetSentTime(time.Now().Add(-24 * time.Hour)).
        SetProperty(ltp.PidTagImportance, int32(2)). // High importance
        Build()

    // Create a sent message with attachment
    reportData := []byte("Sales figures for Q4...")
    ctx.CreateMessage(sent).
        SetSubject("Q4 Sales Report").
        SetBody("Please review the attached report.").
        SetHTMLBody("<p>Please review the <b>attached report</b>.</p>").
        SetFrom("Me", "me@company.com").
        AddTo("Manager", "manager@company.com").
        AddCC("Team", "team@company.com").
        SetSentTime(time.Now()).
        AddAttachment("q4-sales.txt", reportData).
        Build()

    // Create a draft message
    drafts, _ := ctx.CreateFolder(root, "Drafts")
    ctx.CreateMessage(drafts).
        SetSubject("Meeting Notes - Draft").
        SetBody("TODO: Add meeting notes here").
        Build()

    ctx.Commit()

    fmt.Println("Messages created successfully")

    // Verify
    for msg, _ := range inbox.Messages() {
        subject, _ := msg.Subject()
        fmt.Printf("Inbox: %s\n", subject)
    }
}
```

## Next Steps

- [Modifying Content](modifying-content.md) - Update and delete messages
- [Attachments](attachments.md) - Reading attachment data
