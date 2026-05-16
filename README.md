# Outlook PST SDK for Go

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/grokify/outlook-pst-go/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/grokify/outlook-pst-go/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/grokify/outlook-pst-go/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/grokify/outlook-pst-go/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/grokify/outlook-pst-go/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/grokify/outlook-pst-go/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/grokify/outlook-pst-go
 [goreport-url]: https://goreportcard.com/report/github.com/grokify/outlook-pst-go
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/outlook-pst-go
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/outlook-pst-go
 [viz-svg]: https://img.shields.io/badge/visualization-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=grokify%2Foutlook-pst-go
 [loc-svg]: https://tokei.rs/b1/github/grokify/outlook-pst-go
 [repo-url]: https://github.com/grokify/outlook-pst-go
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/outlook-pst-go/blob/main/LICENSE

A pure Go library for reading and writing Microsoft Outlook PST (Personal Storage Table) files.

## Features

- 📖 **Read and write** PST and OST files
- ✨ **Create new PST files** from scratch
- 🔒 **Transaction safety** with two-phase commit
- 🌐 **Both ANSI and Unicode** PST formats supported
- 🔐 **All encryption methods** (none, permute, cyclic)
- 🐹 **Pure Go** - no CGO dependencies
- ⚡ **Lazy loading** for efficient memory usage
- 🔄 **Go 1.23+ iterators** for idiomatic iteration

## Installation

```bash
go get github.com/grokify/outlook-pst-go
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    outlookpst "github.com/grokify/outlook-pst-go"
)

func main() {
    // Open the PST file
    pst, err := outlookpst.Open("archive.pst")
    if err != nil {
        log.Fatal(err)
    }
    defer pst.Close()

    // Get root folder
    root, err := pst.RootFolder()
    if err != nil {
        log.Fatal(err)
    }

    // Iterate through folders
    for folder, err := range root.Subfolders() {
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }

        name, _ := folder.Name()
        fmt.Printf("Folder: %s\n", name)

        // Iterate through messages
        for msg, err := range folder.Messages() {
            if err != nil {
                continue
            }

            subject, _ := msg.Subject()
            sender, _ := msg.SenderName()
            fmt.Printf("  - %s (from: %s)\n", subject, sender)
        }
    }
}
```

## Architecture

The library is organized in four layers, matching the [MS-PST] specification:

```
PST Layer    - High-level API (PST, Folder, Message, Attachment, Recipient)
LTP Layer    - Logical layer (Heap, BTH, PropertyBag, Table)
NDB Layer    - Node Database (Database, Node, Block, B-tree pages)
Disk Layer   - Binary format structures (Headers, Pages, Blocks, Encryption)
```

### Package Structure

```
outlook-pst-go/
├── pst.go              # Main entry point (PST type)
├── folder.go           # Folder type
├── message.go          # Message, Attachment, Recipient types
├── errors.go           # Error types
├── pkg/
│   ├── disk/           # Layer 1: Binary format
│   │   ├── constants.go    # Magic numbers, enums
│   │   ├── header.go       # File headers
│   │   ├── page.go         # Page structures
│   │   ├── block.go        # Block structures
│   │   └── crypt.go        # Encryption/decryption
│   ├── ndb/            # Layer 2: Node Database
│   │   ├── database.go     # Database interface
│   │   └── node.go         # Node abstraction
│   ├── ltp/            # Layer 3: Properties & Tables
│   │   ├── heap.go         # Heap-on-Node
│   │   ├── bth.go          # BTree-on-Heap
│   │   ├── propbag.go      # Property Context
│   │   ├── table.go        # Table Context
│   │   └── property.go     # Property types
│   └── util/           # Utilities
│       ├── primitives.go   # Type definitions
│       └── guid.go         # GUID handling
└── cmd/
    └── pstinfo/        # Sample CLI tool
```

## API Reference

### PST

```go
// Open a PST file (read-only)
pst, err := outlookpst.Open("archive.pst")
defer pst.Close()

// Open for read-write
pst, err := outlookpst.OpenReadWrite("archive.pst")

// Create a new PST file
pst, err := outlookpst.Create("new.pst", disk.FormatUnicode)

// Check format
pst.Format()      // disk.FormatUnicode or disk.FormatANSI
pst.IsUnicode()   // true for Unicode format
pst.IsPST()       // true for PST, false for OST
pst.CryptMethod() // disk.CryptMethodNone/Permute/Cyclic

// Get PST name
name, err := pst.Name()

// Get root folder
root, err := pst.RootFolder()

// Open folder by name
inbox, err := pst.OpenFolder("Inbox")
```

### Folder

```go
// Properties
name, err := folder.Name()
count, err := folder.ContentCount()
unread, err := folder.UnreadCount()
hasSub, err := folder.HasSubfolders()

// Iterate subfolders
for subfolder, err := range folder.Subfolders() {
    // ...
}

// Find subfolder by name
inbox, err := folder.FindSubfolder("Inbox")

// Iterate messages
for msg, err := range folder.Messages() {
    // ...
}
```

### Message

```go
// Basic properties
subject, err := msg.Subject()
body, err := msg.Body()
htmlBody, err := msg.HTMLBody()
messageClass, err := msg.MessageClass()

// Sender info
senderName, err := msg.SenderName()
senderEmail, err := msg.SenderEmail()

// Recipients as strings
displayTo, err := msg.DisplayTo()
displayCc, err := msg.DisplayCc()

// Timestamps
deliveryTime, err := msg.DeliveryTime()
submitTime, err := msg.SubmitTime()
creationTime, err := msg.CreationTime()

// Other properties
size, err := msg.MessageSize()
importance, err := msg.Importance()
hasAttachments, err := msg.HasAttachments()

// Iterate attachments
for att, err := range msg.Attachments() {
    filename, _ := att.Filename()
    data, _ := att.Data()
}

// Iterate recipients
for recip, err := range msg.Recipients() {
    name, _ := recip.Name()
    email, _ := recip.Email()
    recipType, _ := recip.Type() // To, Cc, Bcc
}
```

### Attachment

```go
filename, err := att.Filename()
size, err := att.Size()
mimeType, err := att.MimeType()
data, err := att.Data()

// Check if embedded message
isEmbedded, err := att.IsEmbeddedMessage()
if isEmbedded {
    embeddedMsg, err := att.OpenAsMessage()
}
```

### Recipient

```go
name, err := recip.Name()
email, err := recip.Email()
recipType, err := recip.Type() // RecipientTo, RecipientCc, RecipientBcc
```

### Writing PST Files

```go
import "github.com/grokify/outlook-pst-go/pkg/disk"

// Create a new PST file
pst, err := outlookpst.Create("archive.pst", disk.FormatUnicode)
if err != nil {
    log.Fatal(err)
}
defer pst.Close()

// Begin a write transaction
ctx, err := pst.BeginWrite()
if err != nil {
    log.Fatal(err)
}

// Create folders
root, _ := pst.RootFolder()
inbox, err := ctx.CreateFolder(root, "Inbox")
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

// Create a message
msg, err := ctx.CreateMessage(inbox).
    SetSubject("Hello World").
    SetBody("This is a test message.").
    SetFrom("Sender", "sender@example.com").
    AddTo("Recipient", "recipient@example.com").
    SetSentTime(time.Now()).
    Build()
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

// Commit the transaction
if err := ctx.Commit(); err != nil {
    log.Fatal(err)
}
```

## Command-Line Tool

The `pstinfo` command displays information about a PST file:

```bash
go install github.com/grokify/outlook-pst-go/cmd/pstinfo@latest

pstinfo archive.pst
pstinfo -messages archive.pst
pstinfo -messages -attachments archive.pst
```

## Supported Formats

| Format | Version | Status |
|--------|---------|--------|
| ANSI PST | 14-15 | Supported |
| Unicode PST | 20-23 | Supported |
| OST | Any | Supported (read-only) |

## Encryption Support

| Method | Status |
|--------|--------|
| None | Full support |
| Permute | Full support |
| Cyclic | Full support |

## Limitations

- **No password-protected PST**: Password-encrypted PST files are not supported
- **No RTF decompression**: RTF compression is not decompressed (raw bytes returned)
- **OST files are read-only**: Write support is for PST files only

## References

- [MS-PST]: Outlook Personal Folders (.pst) File Format
- [MS-OXPROPS]: Exchange Server Protocols Master Property List

## License

MIT License
