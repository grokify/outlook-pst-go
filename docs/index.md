# outlook-pst-go

A pure Go library for reading and writing Microsoft Outlook PST (Personal Storage Table) files.

## Features

- **Read and write** PST files
- **Create new PST files** from scratch
- **Both ANSI and Unicode** PST formats supported
- **All encryption methods** (none, permute, cyclic)
- **Pure Go** - no CGO dependencies
- **Transaction safety** with two-phase commit
- **Lazy loading** for efficient memory usage
- **Go 1.23+ iterators** for idiomatic iteration
- **RTF decompression** for message bodies
- **Named properties** for custom MAPI properties
- **Search folder** support
- **Granular error types** for precise error handling

## Quick Example

### Reading a PST File

```go
pst, err := outlookpst.Open("archive.pst")
if err != nil {
    log.Fatal(err)
}
defer pst.Close()

root, _ := pst.RootFolder()
for folder, _ := range root.Subfolders() {
    name, _ := folder.Name()
    fmt.Printf("Folder: %s\n", name)

    for msg, _ := range folder.Messages() {
        subject, _ := msg.Subject()
        fmt.Printf("  - %s\n", subject)
    }
}
```

### Creating a New PST File

```go
// Create a new PST file
pst, err := outlookpst.Create("new-archive.pst", disk.FormatUnicode)
if err != nil {
    log.Fatal(err)
}
defer pst.Close()

// Start a write transaction
ctx, err := pst.BeginWrite()
if err != nil {
    log.Fatal(err)
}

// Create a folder
root, _ := pst.RootFolder()
inbox, err := ctx.CreateFolder(root, "Inbox")
if err != nil {
    ctx.Rollback()
    log.Fatal(err)
}

// Create a message
_, err = ctx.CreateMessage(inbox).
    SetSubject("Hello World").
    SetBody("This is a test message.").
    SetFrom("sender@example.com", "Sender Name").
    AddTo("recipient@example.com", "Recipient").
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

## Supported Formats

| Format | Version | Status |
|--------|---------|--------|
| ANSI PST | 14-15 | ✅ Supported |
| Unicode PST | 20-23 | ✅ Supported |
| OST | Any | ✅ Supported (read-only) |

## Encryption Support

| Method | Status |
|--------|--------|
| None | ✅ Full support |
| Permute | ✅ Full support |
| Cyclic | ✅ Full support |

## Feature Support

| Feature | Read | Write |
|---------|------|-------|
| Folders & Subfolders | ✅ Full | ✅ Full |
| Messages | ✅ Full | ✅ Full |
| Attachments | ✅ Full | ✅ Full |
| Recipients | ✅ Full | ✅ Full |
| RTF Decompression | ✅ Full | - |
| Named Properties | ✅ Full | ⚠️ Partial |
| Search Folders | ✅ Full | - |
| Multi-value Properties | ✅ Full | ⚠️ Partial |
| Create PST Files | - | ✅ Full |
| Transaction Safety | - | ✅ Full |
| Crash Recovery | - | ✅ Full |

## Architecture

The library implements a 4-layer architecture matching the MS-PST specification:

```
┌─────────────────────────────────────────┐
│  Messaging Layer (PST, Folder, Message) │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│  LTP Layer (PropertyBag, Table, Heap)   │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│  NDB Layer (Database, Node, Block)      │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│  Disk Layer (Header, Pages, Encryption) │
└─────────────────────────────────────────┘
```

## Getting Started

See the [Installation](getting-started/installation.md) guide to get started.

## Specification Compliance

This library implements the [MS-PST] specification from Microsoft. See the [Specification Reference](reference/specifications.md) for details.

## License

MIT License
