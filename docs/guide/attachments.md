# Attachments

## Iterating Attachments

```go
for att, err := range msg.Attachments() {
    if err != nil {
        log.Printf("Error: %v", err)
        continue
    }

    filename, _ := att.Filename()
    fmt.Printf("Attachment: %s\n", filename)
}
```

## Attachment Properties

```go
// Filename (tries long name first, falls back to short name)
filename, _ := att.Filename()

// File extension
extension, _ := att.Extension()

// Size in bytes
size, _ := att.Size()

// MIME type
mimeType, _ := att.MimeType()

// Content ID (for inline attachments in HTML)
contentID, _ := att.ContentID()
```

## Reading Attachment Data

```go
data, err := att.Data()
if err != nil {
    log.Printf("Failed to read attachment: %v", err)
    continue
}

// Save to file
err = os.WriteFile(filename, data, 0644)
```

## Attachment Methods

The attachment method indicates how the attachment is stored:

```go
method, _ := att.Method()

switch method {
case ltp.AttachMethodByValue:
    // Binary data stored in the attachment
    data, _ := att.Data()
case ltp.AttachMethodEmbedded:
    // Embedded message (email within email)
    embeddedMsg, _ := att.OpenAsMessage()
case ltp.AttachMethodByRef:
    // Reference to external file
case ltp.AttachMethodOLE:
    // OLE object
}
```

## Embedded Messages

Attachments can contain entire email messages:

```go
isEmbedded, _ := att.IsEmbeddedMessage()
if isEmbedded {
    embeddedMsg, err := att.OpenAsMessage()
    if err != nil {
        log.Printf("Failed to open embedded message: %v", err)
        continue
    }

    subject, _ := embeddedMsg.Subject()
    fmt.Printf("Embedded message: %s\n", subject)

    // Process like any other message
    for innerAtt, _ := range embeddedMsg.Attachments() {
        // Recursive attachments...
    }
}
```

## Counting Attachments

```go
count, _ := msg.AttachmentCount()
fmt.Printf("Message has %d attachments\n", count)
```

## Complete Example: Extract All Attachments

```go
func extractAttachments(msg *outlookpst.Message, outputDir string) error {
    for att, err := range msg.Attachments() {
        if err != nil {
            return err
        }

        filename, _ := att.Filename()
        if filename == "" {
            filename = "unnamed_attachment"
        }

        // Handle embedded messages
        isEmbedded, _ := att.IsEmbeddedMessage()
        if isEmbedded {
            embeddedMsg, err := att.OpenAsMessage()
            if err != nil {
                continue
            }
            // Recursively extract from embedded message
            subDir := filepath.Join(outputDir, filename)
            os.MkdirAll(subDir, 0755)
            extractAttachments(embeddedMsg, subDir)
            continue
        }

        // Save regular attachment
        data, err := att.Data()
        if err != nil {
            log.Printf("Failed to read %s: %v", filename, err)
            continue
        }

        outputPath := filepath.Join(outputDir, filename)
        err = os.WriteFile(outputPath, data, 0644)
        if err != nil {
            log.Printf("Failed to write %s: %v", filename, err)
            continue
        }

        fmt.Printf("Saved: %s (%d bytes)\n", filename, len(data))
    }
    return nil
}
```

## Accessing Raw Properties

```go
bag := att.PropertyBag()

// Get all property IDs
props := bag.Properties()
fmt.Printf("Attachment has %d properties\n", len(props))

// Read specific properties
renderingPosition, _ := bag.GetInt32(ltp.PidTagRenderingPosition)
```
