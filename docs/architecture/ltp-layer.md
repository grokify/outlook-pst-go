# LTP Layer

The Lists, Tables, and Properties layer (`pkg/ltp/`) provides structured data access.

## Components

### Heap-on-Node (`heap.go`)

Manages allocations within a node's data blocks:

```go
heap, err := ltp.NewHeapOnNode(node)
if err != nil {
    log.Fatal(err)
}

// Read an allocation
data, err := heap.Read(heapID)

// Get root allocation
rootID := heap.RootID()
```

### Heap Structure

Each node's data blocks contain:

```
Block 0:
┌─────────────────────────────────┐
│ Heap First Header (12 bytes)   │
│ - Page map offset              │
│ - Signature (0xEC)             │
│ - Client signature             │
│ - Root HID                     │
├─────────────────────────────────┤
│ Allocations...                 │
├─────────────────────────────────┤
│ Page Map                       │
│ - Num allocations              │
│ - Allocation offsets           │
└─────────────────────────────────┘
```

### Client Signatures

The client signature identifies the heap's purpose:

| Signature | Value | Description |
|-----------|-------|-------------|
| TC | 0x7C | Table Context |
| PC | 0xBC | Property Context |
| BTH | 0xB5 | BTree-on-Heap |

### BTree-on-Heap (`bth.go`)

A B-tree structure stored within heap allocations:

```go
bth, err := ltp.NewBTH(heap, rootHID)
if err != nil {
    log.Fatal(err)
}

// Lookup by key
value, err := bth.LookupUint16(key)

// Iterate all entries
for entry, err := range bth.Entries() {
    // entry.Key, entry.Value
}
```

### BTH Structure

```
┌─────────────────────┐
│ BTH Header          │
│ - Signature (0xB5)  │
│ - Key size          │
│ - Entry size        │
│ - Num levels        │
│ - Root HID          │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│ Non-leaf Node       │
│ Key1 → HID1         │
│ Key2 → HID2         │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│ Leaf Node           │
│ Key1 → Value1       │
│ Key2 → Value2       │
└─────────────────────┘
```

### Property Context (`propbag.go`)

Key-value storage for properties:

```go
bag, err := ltp.NewPropertyBag(node)
if err != nil {
    log.Fatal(err)
}

// Check existence
if bag.Exists(ltp.PidTagSubject) {
    subject, _ := bag.GetString(ltp.PidTagSubject)
}

// List all properties
for _, propID := range bag.Properties() {
    propType, _ := bag.GetType(propID)
}
```

### Property Types

| Type | Value | Size | Description |
|------|-------|------|-------------|
| Int16 | 0x0002 | 2 | Signed 16-bit |
| Int32 | 0x0003 | 4 | Signed 32-bit |
| Bool | 0x000B | 2 | Boolean |
| Int64 | 0x0014 | 8 | Signed 64-bit |
| String8 | 0x001E | Variable | ANSI string |
| String | 0x001F | Variable | Unicode string |
| SysTime | 0x0040 | 8 | FILETIME |
| Binary | 0x0102 | Variable | Binary data |

### Table Context (`table.go`)

Tabular data with rows and columns:

```go
table, err := ltp.NewTable(node)
if err != nil {
    log.Fatal(err)
}

// Get column info
for _, col := range table.Columns() {
    fmt.Printf("Column: 0x%04X\n", col.PropID)
}

// Iterate rows
for row, err := range table.Rows() {
    rowID := row.RowID()
    if row.HasProperty(ltp.PidTagDisplayName) {
        name, _ := row.GetString(ltp.PidTagDisplayName)
    }
}
```

### Table Structure

```
┌─────────────────────────────────┐
│ TC Header                       │
│ - Signature (0x7C)              │
│ - Num columns                   │
│ - Size offsets                  │
│ - Row BTH ID                    │
│ - Row matrix HNID               │
│ - Column descriptors            │
└─────────────────────────────────┘
         │               │
         │               │
         ▼               ▼
┌─────────────┐  ┌─────────────────┐
│ Row Index   │  │ Row Matrix       │
│ BTH         │  │ (row data)       │
│ RowID → Idx │  │                  │
└─────────────┘  └─────────────────┘
```

## Heap IDs vs Node IDs

The HNID (HeapNodeID) type can reference either:

- **Heap ID**: Allocation within the current heap
- **Node ID**: Subnode with its own data

```go
if hnid.IsHeapID() {
    data, _ := heap.Read(hnid.ToHeapID())
} else {
    subnode, _ := node.LookupSubnode(hnid.ToNodeID())
    data, _ := subnode.ReadAll()
}
```

This allows properties to be stored inline (small values) or in subnodes (large values).

## Property Access

Properties can be accessed by type:

```go
// Fixed-size properties
int16Val, _ := bag.GetInt16(propID)
int32Val, _ := bag.GetInt32(propID)
int64Val, _ := bag.GetInt64(propID)
boolVal, _ := bag.GetBool(propID)

// Variable-size properties
stringVal, _ := bag.GetString(propID)
binaryVal, _ := bag.GetBinary(propID)

// Time properties
fileTime, _ := bag.GetTime(propID)
goTime := ltp.FileTimeToTime(fileTime)
```
