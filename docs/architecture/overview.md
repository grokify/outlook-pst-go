# Architecture Overview

The library implements the [MS-PST] specification using a four-layer architecture that mirrors the structure of PST files.

## Layer Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      PST Layer                               │
│         (PST, Folder, Message, Attachment, Recipient)        │
├─────────────────────────────────────────────────────────────┤
│                      LTP Layer                               │
│         (Heap, BTH, PropertyBag, Table)                      │
├─────────────────────────────────────────────────────────────┤
│                      NDB Layer                               │
│         (Database, Node, Block, B-tree)                      │
├─────────────────────────────────────────────────────────────┤
│                     Disk Layer                               │
│         (Headers, Pages, Blocks, Encryption)                 │
└─────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### Disk Layer (`pkg/disk/`)

The lowest layer handles binary file format details:

- **Header parsing**: ANSI vs Unicode format detection
- **Page structures**: 512-byte pages for B-trees and allocation maps
- **Block structures**: Data blocks up to 8KB
- **Encryption**: Permute and Cyclic decryption

### NDB Layer (`pkg/ndb/`)

The Node Database layer provides node abstraction:

- **Database**: File access, B-tree root management
- **Node B-tree (NBT)**: Maps node IDs to block references
- **Block B-tree (BBT)**: Maps block IDs to file locations
- **Node**: Unified interface for reading node data

### LTP Layer (`pkg/ltp/`)

The Lists, Tables, and Properties layer provides structured data:

- **Heap-on-Node (HN)**: Allocations within a node's data
- **BTree-on-Heap (BTH)**: B-tree structure within heap
- **Property Context (PC)**: Key-value property storage
- **Table Context (TC)**: Tabular data with rows and columns

### PST Layer (root package)

The high-level API for application developers:

- **PST**: Entry point, file metadata
- **Folder**: Folder navigation and properties
- **Message**: Email content and metadata
- **Attachment**: File attachments
- **Recipient**: Email recipients

## Data Flow

```
Application
    │
    ▼
PST.Open()
    │
    ▼
NDB.Database ─────► Disk.ReadHeader()
    │
    ▼
NDB.GetNode() ────► NBT Lookup ────► BBT Lookup ────► Disk.ReadBlock()
    │
    ▼
LTP.PropertyBag() ─► Heap ─► BTH ─► Property values
    │
    ▼
PST.Folder/Message ► Typed accessors (Subject, Body, etc.)
```

## Key Design Decisions

### 1. Lazy Loading

Data is loaded on demand to minimize memory usage:

```go
// Node data is not read until accessed
node, _ := db.GetNode(nid)
// First access triggers disk read
data, _ := node.ReadAll()
```

### 2. Format Abstraction

ANSI and Unicode formats are handled transparently:

```go
// Same API regardless of format
pst, _ := outlookpst.Open("any.pst")
// Library detects format automatically
fmt.Printf("Format: %s\n", pst.Format())
```

### 3. Error Propagation

Errors are returned, not panicked:

```go
// Every operation that can fail returns an error
subject, err := msg.Subject()
if err != nil {
    // Handle error
}
```

### 4. Iterator Pattern

Go 1.23+ iterators for efficient traversal:

```go
// Memory-efficient iteration
for folder, err := range root.Subfolders() {
    // ...
}
```

## Package Dependencies

```
outlookpst (PST layer)
    │
    ├── pkg/ltp (Properties & Tables)
    │       │
    │       ├── pkg/ndb (Node Database)
    │       │       │
    │       │       └── pkg/disk (Binary Format)
    │       │
    │       └── pkg/util (Primitives)
    │
    └── pkg/util (Primitives)
```

## File Structure

```
outlook-pst-go/
├── pst.go              # PST type
├── folder.go           # Folder type
├── message.go          # Message, Attachment, Recipient
├── errors.go           # Error types
├── pkg/
│   ├── disk/           # Disk layer
│   ├── ndb/            # NDB layer
│   ├── ltp/            # LTP layer
│   └── util/           # Utilities
├── cmd/pstinfo/        # CLI tool
└── docs/               # Documentation
```
