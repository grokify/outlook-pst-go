# MessageBuilder

The `MessageBuilder` provides a fluent interface for creating email messages in PST files.

## Creating a MessageBuilder

```go
builder := ctx.CreateMessage(folder)
```

Or equivalently:

```go
builder := outlookpst.NewMessageBuilder(ctx, folder)
```

## Subject and Body

### SetSubject

```go
func (b *MessageBuilder) SetSubject(subject string) *MessageBuilder
```

Sets the message subject.

```go
builder.SetSubject("Meeting Tomorrow")
```

### SetBody

```go
func (b *MessageBuilder) SetBody(body string) *MessageBuilder
```

Sets the plain text message body.

```go
builder.SetBody("Please join us for the meeting.")
```

### SetHTMLBody

```go
func (b *MessageBuilder) SetHTMLBody(html string) *MessageBuilder
```

Sets the HTML message body.

```go
builder.SetHTMLBody("<h1>Meeting</h1><p>Please join us.</p>")
```

### SetRTFBody

```go
func (b *MessageBuilder) SetRTFBody(rtf []byte) *MessageBuilder
```

Sets the compressed RTF message body.

```go
builder.SetRTFBody(compressedRTFData)
```

## Sender

### SetFrom

```go
func (b *MessageBuilder) SetFrom(name, email string) *MessageBuilder
```

Sets the sender's display name and email address.

```go
builder.SetFrom("John Doe", "john@example.com")
```

## Recipients

### AddRecipient

```go
func (b *MessageBuilder) AddRecipient(name, email string, rtype RecipientType) *MessageBuilder
```

Adds a recipient with the specified type.

```go
builder.AddRecipient("Alice", "alice@example.com", outlookpst.RecipientTo)
builder.AddRecipient("Bob", "bob@example.com", outlookpst.RecipientCc)
```

### AddTo

```go
func (b *MessageBuilder) AddTo(name, email string) *MessageBuilder
```

Adds a TO recipient. Shorthand for `AddRecipient(name, email, RecipientTo)`.

```go
builder.AddTo("Alice", "alice@example.com")
```

### AddCC

```go
func (b *MessageBuilder) AddCC(name, email string) *MessageBuilder
```

Adds a CC recipient. Shorthand for `AddRecipient(name, email, RecipientCc)`.

```go
builder.AddCC("Manager", "manager@example.com")
```

### AddBCC

```go
func (b *MessageBuilder) AddBCC(name, email string) *MessageBuilder
```

Adds a BCC recipient. Shorthand for `AddRecipient(name, email, RecipientBcc)`.

```go
builder.AddBCC("Archive", "archive@example.com")
```

## Recipient Types

```go
const (
    RecipientTo  RecipientType = 1  // Primary recipient
    RecipientCc  RecipientType = 2  // Carbon copy
    RecipientBcc RecipientType = 3  // Blind carbon copy
)
```

## Timestamps

### SetSentTime

```go
func (b *MessageBuilder) SetSentTime(t time.Time) *MessageBuilder
```

Sets the message sent time.

```go
builder.SetSentTime(time.Now())

// Or a specific time
sentTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
builder.SetSentTime(sentTime)
```

## Attachments

### AddAttachment

```go
func (b *MessageBuilder) AddAttachment(filename string, data []byte) *MessageBuilder
```

Adds a file attachment with automatic MIME type detection.

```go
data, _ := os.ReadFile("report.pdf")
builder.AddAttachment("report.pdf", data)
```

### AddAttachmentWithMime

```go
func (b *MessageBuilder) AddAttachmentWithMime(filename string, data []byte, mimeType string) *MessageBuilder
```

Adds a file attachment with an explicit MIME type.

```go
builder.AddAttachmentWithMime("image.png", imageData, "image/png")
builder.AddAttachmentWithMime("data.json", jsonData, "application/json")
```

## Attachment Methods

```go
const (
    AttachMethodNone         AttachMethod = 0  // No attachment data
    AttachMethodByValue      AttachMethod = 1  // Data stored in PST (default)
    AttachMethodByReference  AttachMethod = 2  // External file link
    AttachMethodByRefResolve AttachMethod = 3  // External with resolution
    AttachMethodByRefOnly    AttachMethod = 4  // Reference only
    AttachMethodEmbedded     AttachMethod = 5  // Embedded message
    AttachMethodOLE          AttachMethod = 6  // OLE object
)
```

## Custom Properties

### SetProperty

```go
func (b *MessageBuilder) SetProperty(propID ltp.PropID, value interface{}) *MessageBuilder
```

Sets a custom property on the message.

```go
import "github.com/grokify/outlook-pst-go/pkg/ltp"

// Set importance (0=low, 1=normal, 2=high)
builder.SetProperty(ltp.PidTagImportance, int32(2))

// Set sensitivity (0=normal, 1=personal, 2=private, 3=confidential)
builder.SetProperty(ltp.PidTagSensitivity, int32(2))

// Set priority (0=non-urgent, 1=normal, 2=urgent)
builder.SetProperty(ltp.PidTagPriority, int32(2))
```

Supported value types:

| Type | Example |
|------|---------|
| `string` | `"text value"` |
| `int32` | `int32(42)` |
| `int64` | `int64(123456789)` |
| `bool` | `true` |
| `time.Time` | `time.Now()` |
| `[]byte` | `[]byte{1, 2, 3}` |

## Building

### Build

```go
func (b *MessageBuilder) Build() (*Message, error)
```

Creates the message in the PST file. Must be called within an active transaction.

```go
msg, err := builder.Build()
if err != nil {
    ctx.Rollback()
    return err
}
```

## Chaining

All setter methods return the builder, allowing method chaining:

```go
msg, err := ctx.CreateMessage(folder).
    SetSubject("Important Update").
    SetBody("Please read this message.").
    SetHTMLBody("<h1>Important</h1><p>Please read.</p>").
    SetFrom("Sender", "sender@example.com").
    AddTo("Recipient 1", "r1@example.com").
    AddTo("Recipient 2", "r2@example.com").
    AddCC("Manager", "manager@example.com").
    SetSentTime(time.Now()).
    AddAttachment("doc.pdf", pdfData).
    SetProperty(ltp.PidTagImportance, int32(2)).
    Build()
```

## Complete Example

```go
package main

import (
    "log"
    "os"
    "time"

    outlookpst "github.com/grokify/outlook-pst-go"
    "github.com/grokify/outlook-pst-go/pkg/disk"
    "github.com/grokify/outlook-pst-go/pkg/ltp"
)

func main() {
    pst, _ := outlookpst.Create("messages.pst", disk.FormatUnicode)
    defer pst.Close()

    ctx, _ := pst.BeginWrite()
    root, _ := pst.RootFolder()
    inbox, _ := ctx.CreateFolder(root, "Inbox")

    // Read attachment data
    reportData, err := os.ReadFile("quarterly-report.pdf")
    if err != nil {
        log.Fatal(err)
    }

    // Create message with all features
    msg, err := ctx.CreateMessage(inbox).
        SetSubject("Q4 Financial Report").
        SetBody("Please find the Q4 report attached.\n\nBest regards,\nFinance Team").
        SetHTMLBody(`
            <html>
            <body>
            <h1>Q4 Financial Report</h1>
            <p>Please find the Q4 report attached.</p>
            <p>Best regards,<br>Finance Team</p>
            </body>
            </html>
        `).
        SetFrom("Finance Team", "finance@company.com").
        AddTo("CEO", "ceo@company.com").
        AddTo("CFO", "cfo@company.com").
        AddCC("Board Members", "board@company.com").
        AddBCC("Legal Archive", "legal@company.com").
        SetSentTime(time.Now()).
        AddAttachmentWithMime("Q4-Report.pdf", reportData, "application/pdf").
        SetProperty(ltp.PidTagImportance, int32(2)). // High importance
        SetProperty(ltp.PidTagSensitivity, int32(3)). // Confidential
        Build()

    if err != nil {
        ctx.Rollback()
        log.Fatal(err)
    }

    ctx.Commit()

    // Verify
    subject, _ := msg.Subject()
    log.Printf("Created message: %s", subject)
}
```

## Common Property IDs

| Property | ID | Type | Description |
|----------|----|----|-------------|
| `PidTagImportance` | 0x0017 | int32 | 0=low, 1=normal, 2=high |
| `PidTagPriority` | 0x0026 | int32 | 0=non-urgent, 1=normal, 2=urgent |
| `PidTagSensitivity` | 0x0036 | int32 | 0=normal, 1=personal, 2=private, 3=confidential |
| `PidTagMessageFlags` | 0x0E07 | int32 | Message flags |

## See Also

- [WriteContext](write-context.md) - Transaction management
- [Creating Messages Guide](../guide/creating-messages.md) - Detailed examples
