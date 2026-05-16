# Attachment

The `Attachment` type represents a file attachment on a message.

## Properties

### Filename

```go
func (a *Attachment) Filename() (string, error)
```

Returns the attachment filename. Prefers the long filename, falls back to short filename.

### Extension

```go
func (a *Attachment) Extension() (string, error)
```

Returns the file extension.

### Size

```go
func (a *Attachment) Size() (int32, error)
```

Returns the attachment size in bytes.

### Method

```go
func (a *Attachment) Method() (int32, error)
```

Returns the attachment method.

| Value | Constant | Description |
|-------|----------|-------------|
| 0 | AttachMethodNone | No attachment |
| 1 | AttachMethodByValue | Binary data stored in PST |
| 2 | AttachMethodByRef | Reference to external file |
| 4 | AttachMethodByRefRes | Reference (resolved) |
| 5 | AttachMethodEmbedded | Embedded message |
| 6 | AttachMethodOLE | OLE object |

### MimeType

```go
func (a *Attachment) MimeType() (string, error)
```

Returns the MIME type.

### ContentID

```go
func (a *Attachment) ContentID() (string, error)
```

Returns the Content-ID (for inline attachments in HTML).

## Data Access

### Data

```go
func (a *Attachment) Data() ([]byte, error)
```

Returns the attachment data as bytes.

```go
data, err := att.Data()
if err != nil {
    log.Printf("Failed to read attachment: %v", err)
    return
}

// Save to file
filename, _ := att.Filename()
err = os.WriteFile(filename, data, 0644)
```

## Embedded Messages

### IsEmbeddedMessage

```go
func (a *Attachment) IsEmbeddedMessage() (bool, error)
```

Returns true if the attachment is an embedded message (email within email).

### OpenAsMessage

```go
func (a *Attachment) OpenAsMessage() (*Message, error)
```

Opens an embedded message attachment as a Message.

```go
isEmbedded, _ := att.IsEmbeddedMessage()
if isEmbedded {
    embeddedMsg, err := att.OpenAsMessage()
    if err != nil {
        log.Printf("Failed to open: %v", err)
        return
    }

    subject, _ := embeddedMsg.Subject()
    fmt.Printf("Embedded: %s\n", subject)
}
```

## Advanced Access

### PropertyBag

```go
func (a *Attachment) PropertyBag() *ltp.PropertyBag
```

Returns the attachment's property bag for advanced property access.

## Example

```go
func extractAttachment(att *outlookpst.Attachment, outputDir string) error {
    // Check if embedded message
    isEmbedded, _ := att.IsEmbeddedMessage()
    if isEmbedded {
        embeddedMsg, err := att.OpenAsMessage()
        if err != nil {
            return err
        }

        // Handle embedded message differently
        subject, _ := embeddedMsg.Subject()
        fmt.Printf("Embedded message: %s\n", subject)
        return nil
    }

    // Get filename
    filename, _ := att.Filename()
    if filename == "" {
        filename = "unnamed"
    }

    // Get data
    data, err := att.Data()
    if err != nil {
        return fmt.Errorf("failed to read data: %w", err)
    }

    // Save file
    outputPath := filepath.Join(outputDir, filename)
    err = os.WriteFile(outputPath, data, 0644)
    if err != nil {
        return fmt.Errorf("failed to write file: %w", err)
    }

    size, _ := att.Size()
    mimeType, _ := att.MimeType()
    fmt.Printf("Saved: %s (%d bytes, %s)\n", filename, size, mimeType)

    return nil
}
```
