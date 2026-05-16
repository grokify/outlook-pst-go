# Recipient

The `Recipient` type represents an email recipient.

## Properties

### Name

```go
func (r *Recipient) Name() (string, error)
```

Returns the recipient's display name.

### Email

```go
func (r *Recipient) Email() (string, error)
```

Returns the recipient's email address. Prefers SMTP address, falls back to other address types.

### Type

```go
func (r *Recipient) Type() (RecipientType, error)
```

Returns the recipient type.

### AddressType

```go
func (r *Recipient) AddressType() (string, error)
```

Returns the address type (e.g., "SMTP", "EX").

## RecipientType

```go
type RecipientType int32

const (
    RecipientOriginator RecipientType = 0
    RecipientTo         RecipientType = 1
    RecipientCc         RecipientType = 2
    RecipientBcc        RecipientType = 3
)
```

### String

```go
func (rt RecipientType) String() string
```

Returns a string representation ("Originator", "To", "Cc", "Bcc").

## Advanced Access

### Row

```go
func (r *Recipient) Row() *ltp.TableRow
```

Returns the underlying table row for advanced property access.

```go
row := recip.Row()

// Check for specific properties
if row.HasProperty(ltp.PidTagSmtpAddress) {
    smtp, _ := row.GetString(ltp.PidTagSmtpAddress)
    fmt.Printf("SMTP: %s\n", smtp)
}
```

## Example

```go
func printRecipients(msg *outlookpst.Message) {
    count, _ := msg.RecipientCount()
    fmt.Printf("Recipients (%d):\n", count)

    for recip, err := range msg.Recipients() {
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }

        name, _ := recip.Name()
        email, _ := recip.Email()
        recipType, _ := recip.Type()
        addrType, _ := recip.AddressType()

        fmt.Printf("  %s: %s <%s> [%s]\n",
            recipType.String(), name, email, addrType)
    }
}
```

## Filtering by Type

```go
func getToRecipients(msg *outlookpst.Message) []string {
    var to []string

    for recip, err := range msg.Recipients() {
        if err != nil {
            continue
        }

        recipType, _ := recip.Type()
        if recipType != outlookpst.RecipientTo {
            continue
        }

        name, _ := recip.Name()
        email, _ := recip.Email()

        if email != "" {
            to = append(to, fmt.Sprintf("%s <%s>", name, email))
        } else {
            to = append(to, name)
        }
    }

    return to
}
```
