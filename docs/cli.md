# CLI Tool

The `pstinfo` command-line tool displays information about PST files.

## Installation

```bash
go install github.com/grokify/outlook-pst-go/cmd/pstinfo@latest
```

## Usage

```bash
pstinfo [options] <pst-file>
```

## Options

| Option | Description |
|--------|-------------|
| `-messages` | Show messages in each folder |
| `-max-messages N` | Maximum messages to show per folder (default: 10) |
| `-attachments` | Show attachment info for messages |

## Examples

### Basic Information

```bash
$ pstinfo archive.pst

PST File: archive.pst
Format: Unicode
Encryption: None
Type: PST (Personal Storage Table)
Display Name: Outlook Data File

Folder Structure:
-----------------
[Root - Mailbox] (0 items, 0 unread)
  [Deleted Items] (5 items, 0 unread)
  [Inbox] (150 items, 3 unread)
  [Outbox] (0 items, 0 unread)
  [Sent Items] (89 items, 0 unread)
  [Calendar] (42 items, 0 unread)
  [Contacts] (128 items, 0 unread)
  [Drafts] (2 items, 0 unread)
```

### Show Messages

```bash
$ pstinfo -messages archive.pst

...
[Inbox] (150 items, 3 unread)
  - Meeting reminder: Project review
    From: John Smith
    Date: 2024-01-15 09:30:00
  - Re: Budget proposal
    From: Jane Doe
    Date: 2024-01-14 16:45:00
  - Weekly report
    From: Reports System
    Date: 2024-01-14 08:00:00
  ... and 147 more messages
```

### Show Attachments

```bash
$ pstinfo -messages -attachments archive.pst

...
[Inbox] (150 items, 3 unread)
  - Budget proposal
    From: Finance Team
    Date: 2024-01-14 10:00:00
    Attachments (2):
      - budget_2024.xlsx (45678 bytes)
      - summary.pdf (12345 bytes)
```

### Limit Messages

```bash
$ pstinfo -messages -max-messages 5 archive.pst
```

## Output Format

The tool outputs:

1. **File Information**
   - File path
   - Format (ANSI/Unicode)
   - Encryption method
   - File type (PST/OST)
   - Display name

2. **Folder Tree**
   - Folder name
   - Message count
   - Unread count

3. **Messages** (with `-messages`)
   - Subject
   - Sender name
   - Delivery date

4. **Attachments** (with `-attachments`)
   - Filename
   - Size in bytes

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Error (file not found, invalid PST, etc.) |

## Source Code

The CLI tool source is at `cmd/pstinfo/main.go`:

```go
package main

import (
    "flag"
    "fmt"
    "os"

    outlookpst "github.com/grokify/outlook-pst-go"
)

func main() {
    showMessages := flag.Bool("messages", false, "Show messages")
    maxMessages := flag.Int("max-messages", 10, "Max messages per folder")
    showAttachments := flag.Bool("attachments", false, "Show attachments")
    flag.Parse()

    if flag.NArg() < 1 {
        fmt.Fprintf(os.Stderr, "Usage: pstinfo [options] <pst-file>\n")
        os.Exit(1)
    }

    pst, err := outlookpst.Open(flag.Arg(0))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer pst.Close()

    // ... display information
}
```
