# Recipients

## Iterating Recipients

```go
for recip, err := range msg.Recipients() {
    if err != nil {
        log.Printf("Error: %v", err)
        continue
    }

    name, _ := recip.Name()
    email, _ := recip.Email()
    fmt.Printf("%s <%s>\n", name, email)
}
```

## Recipient Properties

```go
// Display name
name, _ := recip.Name()

// Email address (tries SMTP first, falls back to other types)
email, _ := recip.Email()

// Address type (e.g., "SMTP", "EX")
addrType, _ := recip.AddressType()

// Recipient type
recipType, _ := recip.Type()
```

## Recipient Types

```go
switch recipType {
case outlookpst.RecipientTo:
    fmt.Println("To recipient")
case outlookpst.RecipientCc:
    fmt.Println("Cc recipient")
case outlookpst.RecipientBcc:
    fmt.Println("Bcc recipient")
case outlookpst.RecipientOriginator:
    fmt.Println("Originator")
}
```

Or use the string representation:

```go
fmt.Printf("Type: %s\n", recipType.String())
// Output: "To", "Cc", "Bcc", or "Originator"
```

## Counting Recipients

```go
count, _ := msg.RecipientCount()
fmt.Printf("Message has %d recipients\n", count)
```

## Filtering Recipients by Type

```go
func getRecipientsByType(msg *outlookpst.Message, targetType outlookpst.RecipientType) []string {
    var recipients []string

    for recip, err := range msg.Recipients() {
        if err != nil {
            continue
        }

        recipType, _ := recip.Type()
        if recipType == targetType {
            name, _ := recip.Name()
            email, _ := recip.Email()

            if email != "" {
                recipients = append(recipients, fmt.Sprintf("%s <%s>", name, email))
            } else {
                recipients = append(recipients, name)
            }
        }
    }

    return recipients
}

// Usage
toRecipients := getRecipientsByType(msg, outlookpst.RecipientTo)
ccRecipients := getRecipientsByType(msg, outlookpst.RecipientCc)
```

## Display Recipients vs. Recipient Objects

Messages have two ways to access recipients:

### Display Strings (Quick)

```go
// Formatted strings from message properties
displayTo, _ := msg.DisplayTo()   // "John Doe; Jane Smith"
displayCc, _ := msg.DisplayCc()
displayBcc, _ := msg.DisplayBcc()
```

### Recipient Objects (Detailed)

```go
// Individual recipient objects with full details
for recip, _ := range msg.Recipients() {
    name, _ := recip.Name()
    email, _ := recip.Email()
    addrType, _ := recip.AddressType()
    recipType, _ := recip.Type()
    // ...
}
```

## Accessing Raw Properties

```go
row := recip.Row()

// Access table row data directly
rowID := row.RowID()

// Check for specific properties
if row.HasProperty(ltp.PidTagSmtpAddress) {
    smtp, _ := row.GetString(ltp.PidTagSmtpAddress)
    fmt.Printf("SMTP: %s\n", smtp)
}
```

## Complete Example

```go
func printRecipients(msg *outlookpst.Message) {
    fmt.Println("Recipients:")
    fmt.Println("-----------")

    var to, cc, bcc []string

    for recip, err := range msg.Recipients() {
        if err != nil {
            continue
        }

        name, _ := recip.Name()
        email, _ := recip.Email()
        recipType, _ := recip.Type()

        formatted := name
        if email != "" && email != name {
            formatted = fmt.Sprintf("%s <%s>", name, email)
        }

        switch recipType {
        case outlookpst.RecipientTo:
            to = append(to, formatted)
        case outlookpst.RecipientCc:
            cc = append(cc, formatted)
        case outlookpst.RecipientBcc:
            bcc = append(bcc, formatted)
        }
    }

    if len(to) > 0 {
        fmt.Printf("To: %s\n", strings.Join(to, "; "))
    }
    if len(cc) > 0 {
        fmt.Printf("Cc: %s\n", strings.Join(cc, "; "))
    }
    if len(bcc) > 0 {
        fmt.Printf("Bcc: %s\n", strings.Join(bcc, "; "))
    }
}
```
