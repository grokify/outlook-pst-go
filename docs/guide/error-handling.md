# Error Handling

The library provides granular, typed errors for precise error handling and debugging.

## Error Hierarchy

Errors are organized by layer:

```
Root Package (Messaging Layer)
├── MessagingError
├── ErrNotFound
├── ErrInvalidFormat
└── ErrCorrupted

pkg/ltp (LTP Layer)
└── LTPError

pkg/ndb (NDB Layer)
└── NDBError
```

## Messaging Layer Errors

### MessagingError

```go
type MessagingError struct {
    Op         string             // Operation that failed
    Kind       MessagingErrorKind // Error kind
    FolderName string             // Related folder (if applicable)
    NodeID     uint64             // Related node ID
    PropID     uint16             // Related property ID
    Err        error              // Underlying error
}
```

### Error Kinds

| Kind | Description |
|------|-------------|
| `ErrKindFolderNotFound` | Folder not found |
| `ErrKindMessageNotFound` | Message not found |
| `ErrKindAttachmentNotFound` | Attachment not found |
| `ErrKindRecipientNotFound` | Recipient not found |
| `ErrKindInvalidEntryID` | Invalid entry ID |
| `ErrKindInvalidMessageClass` | Invalid message class |

### Checking Error Types

```go
folder, err := pst.OpenFolder("NonExistent")
if err != nil {
    var msgErr *outlookpst.MessagingError
    if errors.As(err, &msgErr) {
        switch msgErr.Kind {
        case outlookpst.ErrKindFolderNotFound:
            fmt.Printf("Folder %q not found\n", msgErr.FolderName)
        default:
            fmt.Printf("Messaging error: %v\n", err)
        }
    }
}
```

### Sentinel Errors

```go
// Check for not found errors
if errors.Is(err, outlookpst.ErrNotFound) {
    fmt.Println("Item not found")
}

// Check for format errors
if errors.Is(err, outlookpst.ErrInvalidFormat) {
    fmt.Println("Invalid format")
}

// Check for corruption
if errors.Is(err, outlookpst.ErrCorrupted) {
    fmt.Println("File corrupted")
}
```

## LTP Layer Errors

### LTPError

```go
type LTPError struct {
    Op       string       // Operation
    Kind     LTPErrorKind // Error kind
    PropID   PropID       // Property ID
    PropType PropType     // Property type
    HeapID   uint32       // Heap ID
    RowIndex int          // Row index
    ColIndex int          // Column index
    Got      interface{}  // Actual value
    Want     interface{}  // Expected value
    Err      error        // Underlying error
}
```

### LTP Error Kinds

| Kind | Description |
|------|-------------|
| `ErrKindInvalidHeapSignature` | Invalid heap signature |
| `ErrKindHeapAllocationNotFound` | Heap allocation not found |
| `ErrKindPropertyNotFound` | Property not found |
| `ErrKindPropertyTypeMismatch` | Property type mismatch |
| `ErrKindInvalidTCSignature` | Invalid table signature |
| `ErrKindInvalidRowIndex` | Row index out of bounds |

### Checking LTP Errors

```go
import "github.com/grokify/outlook-pst-go/pkg/ltp"

value, err := bag.GetString(propID)
if err != nil {
    if ltp.IsPropertyNotFound(err) {
        fmt.Println("Property not present")
    } else {
        var ltpErr *ltp.LTPError
        if errors.As(err, &ltpErr) {
            fmt.Printf("LTP error in %s: %s\n", ltpErr.Op, ltpErr.Kind)
        }
    }
}
```

## NDB Layer Errors

### NDBError

```go
type NDBError struct {
    Op      string       // Operation
    Kind    NDBErrorKind // Error kind
    NodeID  uint64       // Node ID
    BlockID uint64       // Block ID
    Offset  uint64       // File offset
    Got     interface{}  // Actual value
    Want    interface{}  // Expected value
    Err     error        // Underlying error
}
```

### NDB Error Kinds

| Kind | Description |
|------|-------------|
| `ErrKindNodeNotFound` | Node not found in NBT |
| `ErrKindBlockNotFound` | Block not found in BBT |
| `ErrKindInvalidPageType` | Invalid page type |
| `ErrKindInvalidMagic` | Invalid PST magic bytes |
| `ErrKindCRCMismatch` | CRC validation failed |
| `ErrKindSignatureMismatch` | Signature validation failed |

### Checking NDB Errors

```go
import "github.com/grokify/outlook-pst-go/pkg/ndb"

node, err := db.GetNode(nodeID)
if err != nil {
    if ndb.IsNodeNotFound(err) {
        fmt.Printf("Node 0x%X not found\n", nodeID)
    } else if ndb.IsBlockNotFound(err) {
        fmt.Println("Block not found")
    }
}
```

## Best Practices

### 1. Use errors.Is for Sentinel Errors

```go
if errors.Is(err, outlookpst.ErrNotFound) {
    // Handle not found
}
```

### 2. Use errors.As for Typed Errors

```go
var ndbErr *ndb.NDBError
if errors.As(err, &ndbErr) {
    // Access error details
    fmt.Printf("Node: 0x%X\n", ndbErr.NodeID)
}
```

### 3. Check Layer-Specific Helpers

```go
// LTP layer
if ltp.IsPropertyNotFound(err) { ... }

// NDB layer
if ndb.IsNodeNotFound(err) { ... }
if ndb.IsBlockNotFound(err) { ... }
```

### 4. Log Full Error Details

```go
if err != nil {
    // Full error message includes context
    log.Printf("Error: %v", err)
    // Output: ndb: LookupNode: node not found (node=0x2442)
}
```

## Error Wrapping

Errors include context and can be unwrapped:

```go
// Unwrap to get underlying error
if underlying := errors.Unwrap(err); underlying != nil {
    fmt.Printf("Underlying: %v\n", underlying)
}

// Check for wrapped errors
if errors.Is(err, io.EOF) {
    fmt.Println("End of file")
}
```

## Example: Comprehensive Error Handling

```go
func processMessage(pst *outlookpst.PST, nodeID util.NodeID) error {
    node, err := pst.Database().GetNode(nodeID)
    if err != nil {
        var ndbErr *ndb.NDBError
        if errors.As(err, &ndbErr) {
            switch ndbErr.Kind {
            case ndb.ErrKindNodeNotFound:
                return fmt.Errorf("message node 0x%X not found", nodeID)
            case ndb.ErrKindCRCMismatch:
                return fmt.Errorf("message data corrupted at offset 0x%X", ndbErr.Offset)
            }
        }
        return fmt.Errorf("failed to get message node: %w", err)
    }

    msg, err := outlookpst.NewMessage(pst, node)
    if err != nil {
        var ltpErr *ltp.LTPError
        if errors.As(err, &ltpErr) {
            return fmt.Errorf("message structure invalid: %s", ltpErr.Kind)
        }
        return fmt.Errorf("failed to parse message: %w", err)
    }

    subject, err := msg.Subject()
    if err != nil {
        if ltp.IsPropertyNotFound(err) {
            subject = "(no subject)"
        } else {
            return fmt.Errorf("failed to get subject: %w", err)
        }
    }

    fmt.Printf("Subject: %s\n", subject)
    return nil
}
```
