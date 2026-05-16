# Reading Messages

## Iterating Messages

```go
for msg, err := range folder.Messages() {
    if err != nil {
        log.Printf("Error: %v", err)
        continue
    }

    subject, _ := msg.Subject()
    fmt.Printf("Subject: %s\n", subject)
}
```

## Message Properties

### Basic Information

```go
subject, _ := msg.Subject()
normalizedSubject, _ := msg.NormalizedSubject() // Without "Re:", "Fwd:", etc.
messageClass, _ := msg.MessageClass()           // e.g., "IPM.Note"
```

### Body Content

```go
// Plain text body
body, _ := msg.Body()

// HTML body
htmlBody, _ := msg.HTMLBody()

// RTF body (compressed, raw bytes)
rtfBody, _ := msg.RTFBody()
```

### Sender Information

```go
senderName, _ := msg.SenderName()
senderEmail, _ := msg.SenderEmail()

// "Sent on behalf of" information
representingName, _ := msg.SentRepresentingName()
representingEmail, _ := msg.SentRepresentingEmail()
```

### Recipients

```go
// As formatted strings
displayTo, _ := msg.DisplayTo()   // "John Doe; Jane Smith"
displayCc, _ := msg.DisplayCc()
displayBcc, _ := msg.DisplayBcc()

// As individual recipients (see Recipients guide)
for recip, err := range msg.Recipients() {
    // ...
}
```

### Timestamps

```go
import "time"

deliveryTime, _ := msg.DeliveryTime()
submitTime, _ := msg.SubmitTime()
creationTime, _ := msg.CreationTime()
lastModified, _ := msg.LastModificationTime()

fmt.Printf("Received: %s\n", deliveryTime.Format(time.RFC3339))
```

### Message Attributes

```go
// Size in bytes
size, _ := msg.MessageSize()

// Importance (0=low, 1=normal, 2=high)
importance, _ := msg.Importance()

// Priority (0=non-urgent, 1=normal, 2=urgent)
priority, _ := msg.Priority()

// Sensitivity (0=normal, 1=personal, 2=private, 3=confidential)
sensitivity, _ := msg.Sensitivity()

// Has attachments?
hasAttachments, _ := msg.HasAttachments()
```

### Internet Headers

```go
// RFC 822 Message-ID
messageID, _ := msg.InternetMessageID()

// Conversation tracking
topic, _ := msg.ConversationTopic()
index, _ := msg.ConversationIndex()
```

## Counting Messages

```go
count, err := folder.MessageCount()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Folder has %d messages\n", count)
```

## Accessing Raw Properties

```go
bag := msg.PropertyBag()

// Check if a property exists
if bag.Exists(ltp.PidTagBody) {
    body, _ := bag.GetString(ltp.PidTagBody)
    fmt.Println(body)
}

// Get property type
propType, _ := bag.GetType(ltp.PidTagSubject)

// Get raw bytes
raw, _ := bag.GetRaw(ltp.PidTagConversationIndex)
```

## Message Classes

Common message classes:

| Class | Description |
|-------|-------------|
| `IPM.Note` | Email message |
| `IPM.Appointment` | Calendar appointment |
| `IPM.Contact` | Contact card |
| `IPM.Task` | Task item |
| `IPM.StickyNote` | Sticky note |
| `IPM.Activity` | Journal entry |
| `IPM.Schedule.Meeting.Request` | Meeting request |
| `IPM.Schedule.Meeting.Canceled` | Meeting cancellation |

## Complete Example

```go
func printMessage(msg *outlookpst.Message) {
    subject, _ := msg.Subject()
    sender, _ := msg.SenderName()
    deliveryTime, _ := msg.DeliveryTime()
    body, _ := msg.Body()

    fmt.Println("=" + strings.Repeat("=", 60))
    fmt.Printf("Subject: %s\n", subject)
    fmt.Printf("From: %s\n", sender)
    fmt.Printf("Date: %s\n", deliveryTime.Format("2006-01-02 15:04"))
    fmt.Println("-" + strings.Repeat("-", 60))

    // Truncate body for display
    if len(body) > 500 {
        body = body[:500] + "..."
    }
    fmt.Println(body)

    // List attachments
    hasAtt, _ := msg.HasAttachments()
    if hasAtt {
        fmt.Println("\nAttachments:")
        for att, err := range msg.Attachments() {
            if err != nil {
                continue
            }
            name, _ := att.Filename()
            size, _ := att.Size()
            fmt.Printf("  - %s (%d bytes)\n", name, size)
        }
    }
}
```
