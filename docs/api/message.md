# Message

The `Message` type represents an email message in a PST file.

## Identification

### ID

```go
func (m *Message) ID() util.NodeID
```

Returns the message's node ID.

## Content

### Subject

```go
func (m *Message) Subject() (string, error)
```

Returns the message subject.

### NormalizedSubject

```go
func (m *Message) NormalizedSubject() (string, error)
```

Returns the normalized subject (without "Re:", "Fwd:" prefixes).

### Body

```go
func (m *Message) Body() (string, error)
```

Returns the plain text body.

### HTMLBody

```go
func (m *Message) HTMLBody() (string, error)
```

Returns the HTML body.

### RTFBody

```go
func (m *Message) RTFBody() ([]byte, error)
```

Returns the compressed RTF body (raw bytes).

### RTFBodyDecompressed

```go
func (m *Message) RTFBodyDecompressed() (string, error)
```

Returns the decompressed RTF body as a string. Uses the LZFu decompression algorithm per MS-OXRTFCP.

### MessageClass

```go
func (m *Message) MessageClass() (string, error)
```

Returns the message class (e.g., "IPM.Note").

## Sender Information

### SenderName

```go
func (m *Message) SenderName() (string, error)
```

Returns the sender's display name.

### SenderEmail

```go
func (m *Message) SenderEmail() (string, error)
```

Returns the sender's email address.

### SentRepresentingName

```go
func (m *Message) SentRepresentingName() (string, error)
```

Returns the "sent on behalf of" display name.

### SentRepresentingEmail

```go
func (m *Message) SentRepresentingEmail() (string, error)
```

Returns the "sent on behalf of" email address.

## Recipients

### DisplayTo

```go
func (m *Message) DisplayTo() (string, error)
```

Returns the To recipients as a formatted string.

### DisplayCc

```go
func (m *Message) DisplayCc() (string, error)
```

Returns the Cc recipients as a formatted string.

### DisplayBcc

```go
func (m *Message) DisplayBcc() (string, error)
```

Returns the Bcc recipients as a formatted string.

### Recipients

```go
func (m *Message) Recipients() iter.Seq2[*Recipient, error]
```

Returns an iterator over individual recipients.

### RecipientCount

```go
func (m *Message) RecipientCount() (int, error)
```

Returns the number of recipients.

## Timestamps

### DeliveryTime

```go
func (m *Message) DeliveryTime() (time.Time, error)
```

Returns when the message was delivered.

### SubmitTime

```go
func (m *Message) SubmitTime() (time.Time, error)
```

Returns when the message was submitted.

### CreationTime

```go
func (m *Message) CreationTime() (time.Time, error)
```

Returns when the message was created.

### LastModificationTime

```go
func (m *Message) LastModificationTime() (time.Time, error)
```

Returns when the message was last modified.

## Attributes

### MessageSize

```go
func (m *Message) MessageSize() (int32, error)
```

Returns the message size in bytes.

### Importance

```go
func (m *Message) Importance() (int32, error)
```

Returns the importance level (0=low, 1=normal, 2=high).

### Priority

```go
func (m *Message) Priority() (int32, error)
```

Returns the priority (0=non-urgent, 1=normal, 2=urgent).

### Sensitivity

```go
func (m *Message) Sensitivity() (int32, error)
```

Returns the sensitivity (0=normal, 1=personal, 2=private, 3=confidential).

### HasAttachments

```go
func (m *Message) HasAttachments() (bool, error)
```

Returns true if the message has attachments.

## Internet Headers

### InternetMessageID

```go
func (m *Message) InternetMessageID() (string, error)
```

Returns the RFC 822 Message-ID.

### ConversationTopic

```go
func (m *Message) ConversationTopic() (string, error)
```

Returns the conversation topic.

### ConversationIndex

```go
func (m *Message) ConversationIndex() ([]byte, error)
```

Returns the conversation index (binary).

## Attachments

### Attachments

```go
func (m *Message) Attachments() iter.Seq2[*Attachment, error]
```

Returns an iterator over attachments.

### AttachmentCount

```go
func (m *Message) AttachmentCount() (int, error)
```

Returns the number of attachments.

## Advanced Access

### PropertyBag

```go
func (m *Message) PropertyBag() *ltp.PropertyBag
```

Returns the message's property bag.

### AttachmentTable

```go
func (m *Message) AttachmentTable() (*ltp.Table, error)
```

Returns the attachment table.

### RecipientTable

```go
func (m *Message) RecipientTable() (*ltp.Table, error)
```

Returns the recipient table.

## Example

```go
func printMessage(msg *outlookpst.Message) {
    subject, _ := msg.Subject()
    sender, _ := msg.SenderName()
    senderEmail, _ := msg.SenderEmail()
    deliveryTime, _ := msg.DeliveryTime()
    body, _ := msg.Body()

    fmt.Printf("Subject: %s\n", subject)
    fmt.Printf("From: %s <%s>\n", sender, senderEmail)
    fmt.Printf("Date: %s\n", deliveryTime.Format(time.RFC1123))
    fmt.Printf("\n%s\n", body)

    // Print attachments
    for att, _ := range msg.Attachments() {
        filename, _ := att.Filename()
        size, _ := att.Size()
        fmt.Printf("Attachment: %s (%d bytes)\n", filename, size)
    }
}
```
