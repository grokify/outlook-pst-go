# Case Study: Implementing PST Parsing in Go

This document chronicles the bugs discovered and fixed while implementing a Go library for reading Microsoft Outlook PST files. The implementation was based primarily on Microsoft's [MS-PST specification](https://docs.microsoft.com/en-us/openspecs/office_file_formats/ms-pst/) and validated against a real PST file.

## Background

PST (Personal Storage Table) files are complex binary formats with a layered architecture:

```
┌─────────────────────────────────────┐
│         Messaging Layer             │  ← Folders, Messages, Attachments
├─────────────────────────────────────┤
│    LTP (Lists, Tables, Properties)  │  ← Heap-on-Node, BTH, PC, TC
├─────────────────────────────────────┤
│      NDB (Node Database Layer)      │  ← Nodes, Blocks, B-trees
├─────────────────────────────────────┤
│           Disk Layer                │  ← Header, Pages, Encryption
└─────────────────────────────────────┘
```

Each layer depends on correct implementation of the layers below it. A subtle bug in bit manipulation at the disk layer can cascade into complete failure at the messaging layer.

## The Bugs

### Bug 1: HeapID Bit Field Extraction

**Location**: `pkg/util/primitives.go`

**Symptom**: `heap alloc index out of bounds: 32 >= 2`

**Root Cause**: The HeapID (HID) structure uses specific bit fields that were being extracted incorrectly.

According to [MS-PST] Section 2.3.1.1, the HID structure is:

```
Bits 0-4   (5 bits):  hidType - Reserved, must be 0
Bits 5-15  (11 bits): hidIndex - 1-based allocation index within block
Bits 16-31 (16 bits): hidBlockIndex - Block index within data tree
```

**Incorrect Implementation**:
```go
// WRONG: Treated bits 0-15 as allocation index
func (hid HeapID) AllocIndex() uint16 {
    return uint16(hid & 0xFFFF)
}
```

**Correct Implementation**:
```go
const (
    heapIDBlockIndexShift = 16
    heapIDAllocIndexShift = 5
    heapIDAllocIndexMask  = 0x7FF // 11 bits
)

func (hid HeapID) AllocIndex() uint16 {
    // Extract bits 5-15, convert from 1-based to 0-based
    index := uint16((hid >> heapIDAllocIndexShift) & heapIDAllocIndexMask)
    if index == 0 {
        return 0 // Invalid HID
    }
    return index - 1
}

func (hid HeapID) BlockIndex() uint16 {
    return uint16(hid >> heapIDBlockIndexShift)
}
```

**Lesson**: When working with packed bit fields, carefully document the bit layout and verify against hex dumps of real data.

---

### Bug 2: Folder Table Lookup Strategy

**Location**: `folder.go`

**Symptom**: `hierarchy table not found: subnode not found: 0x12D`

**Root Cause**: Misunderstanding of how associated tables are stored in the PST structure.

The MS-PST specification describes that folder nodes (NID type 0x02) have associated hierarchy tables (NID type 0x0D) and contents tables (NID type 0x0E). The original implementation assumed these were stored as subnodes of the folder node.

**Incorrect Implementation**:
```go
// WRONG: Looking for table as subnode
hierarchyNode, err := f.node.LookupSubnode(hierarchyNID)
```

**Correct Implementation**:
```go
// CORRECT: Tables are separate NBT entries with derived NIDs
// Hierarchy table NID = folder index | NIDTypeHierarchyTable
hierarchyNID := util.MakeNID(util.NIDTypeHierarchyTable, folderNID.Index())
hierarchyNode, err := f.pst.db.GetNode(hierarchyNID)
```

**Lesson**: The relationship between nodes isn't always parent-child. Some related structures share a common index but have different type codes, making them siblings in the NBT rather than nested subnodes.

---

### Bug 3: BTH Empty Detection

**Location**: `pkg/ltp/bth.go`

**Symptom**: BTH iteration returned no entries even when data existed

**Root Cause**: Misinterpretation of the `bIdxLevels` field in the BTH header.

According to [MS-PST] Section 2.3.2.1, `bIdxLevels` indicates the number of **intermediate** levels:

- `bIdxLevels = 0`: Root allocation contains leaf records directly
- `bIdxLevels = 1`: One intermediate level, root points to leaf pages
- `bIdxLevels = 2`: Two intermediate levels, etc.

**Incorrect Implementation**:
```go
// WRONG: Treated bIdxLevels=0 as empty
func (b *BTH) IsEmpty() bool {
    return b.header.NumLevels == 0
}
```

**Correct Implementation**:
```go
// CORRECT: Empty only if root HID is zero
func (b *BTH) IsEmpty() bool {
    return b.header.RootID == 0
}
```

**Lesson**: "Levels" in B-tree terminology often refers to intermediate levels only. A single-level tree (leaf at root) has 0 intermediate levels but is not empty.

---

### Bug 4: BTH Entry Size Interpretation

**Location**: `pkg/ltp/bth.go`

**Symptom**: `property entry too small: 4 bytes`

**Root Cause**: The BTH header's `cbEnt` field was misinterpreted.

Per [MS-PST] Section 2.3.2.1:

- `cbKey`: Size of the key in bytes
- `cbEnt`: Size of the **data value** in bytes (not the full entry!)

**Incorrect Implementation**:
```go
// WRONG: Treated cbEnt as full entry size
func (b *BTH) searchLeaf(data []byte, key []byte) ([]byte, error) {
    entrySize := int(b.header.EntrySize) // This is just value size!
    // ...
    return data[offset+keySize : offset+entrySize], nil // Wrong slice
}
```

**Correct Implementation**:
```go
func (b *BTH) searchLeaf(data []byte, key []byte) ([]byte, error) {
    keySize := int(b.header.KeySize)
    valueSize := int(b.header.EntrySize) // cbEnt is VALUE size
    entrySize := keySize + valueSize      // Full entry = key + value

    // ...
    return data[offset+keySize : offset+keySize+valueSize], nil
}
```

**Lesson**: Field names in specifications can be misleading. "EntrySize" sounds like it should be the full entry, but it's actually just the value portion. Always verify against the spec text and sample data.

---

### Bug 5: Table Context CEB Offset

**Location**: `pkg/ltp/table.go`

**Symptom**: `property not present in row: 0x67F2` for properties that existed

**Root Cause**: Incorrect calculation of the Cell Existence Bitmap (CEB) offset within row data.

The Table Context row layout per [MS-PST] Section 2.3.4.4:

```
┌────────────────┬────────────────┬────────────────┬─────────┐
│ 4-byte columns │ 2-byte columns │ 1-byte columns │   CEB   │
└────────────────┴────────────────┴────────────────┴─────────┘
     0 → rgib[0]    rgib[0] → rgib[1]  rgib[1] → rgib[2]  rgib[2] → rgib[3]
```

Where `rgib` is the `SizeOffsets` array in the TC header.

**Incorrect Implementation**:
```go
// WRONG: Used SizeOffsets[3] which is the END of CEB (total row size)
bitmapOffset := int(r.table.header.SizeOffsets[3])
```

**Correct Implementation**:
```go
// CORRECT: CEB starts at SizeOffsets[2] (end of 1-byte columns)
bitmapOffset := int(r.table.header.SizeOffsets[2])
```

**Lesson**: Array indices for "size offsets" represent cumulative boundaries. The CEB starts where the 1-byte columns end (`SizeOffsets[2]`), not at `SizeOffsets[3]` which is past the CEB.

---

### Bug 6: Row ID Source

**Location**: `folder.go`

**Symptom**: `property not present in row: 0x0000` when reading `PidTagLtpRowId`

**Root Cause**: Attempting to read the row ID from row data when it's stored in the BTH key.

In a Table Context, rows are indexed via a BTH where:

- **Key**: `dwRowID` (the row identifier, e.g., subfolder NID)
- **Value**: Row index within the RowMatrix

The `PidTagLtpRowId` (0x67F2) property exists in the column descriptors but its CEB bit is typically 0 because the value comes from the BTH key, not the row data.

**Incorrect Implementation**:
```go
// WRONG: Trying to read row ID from row data
nidData, err := row.GetRaw(ltp.PidTagLtpRowId)
subfolderNID := util.NodeID(binary.LittleEndian.Uint32(nidData))
```

**Correct Implementation**:
```go
// CORRECT: Row ID comes from BTH key
subfolderNID := util.NodeID(row.RowID())
```

**Lesson**: Some "properties" in table contexts are metadata about the row structure itself, not actual data stored in the row. The row ID is inherently part of the BTH index, not a column value.

---

### Bug 7: BBT B-tree Search Algorithm

**Location**: `pkg/ndb/database.go`

**Symptom**: `block not found: 0x40` for blocks that existed in the BBT

**Root Cause**: The BBT search used incorrect child selection logic compared to the NBT search.

In both NBT and BBT, intermediate entries have a `Key` field representing the **minimum** key in that subtree. To find the correct child:

1. Find the **last** entry where `entry.Key <= target`
2. Follow that entry's child reference

**Incorrect Implementation**:
```go
// WRONG: Found FIRST entry where target <= key
for i, entry := range page.NonleafEntries {
    if bid <= entry.Key {
        childRef = &page.NonleafEntries[i].Ref
        break
    }
}
```

**Correct Implementation**:
```go
// CORRECT: Find LAST entry where key <= target
childIdx := -1
for i, entry := range page.NonleafEntries {
    if entry.Key <= bid {
        childIdx = i
    } else {
        break // Keys are sorted
    }
}
childRef = &page.NonleafEntries[childIdx].Ref
```

**Lesson**: B-tree search semantics depend on whether keys represent minimums or maximums of subtrees. The MS-PST uses minimum-key semantics, requiring "find last entry ≤ target" logic.

---

## Debugging Techniques

### What Actually Worked

**The test PST file was essential.** Without real data to parse, specification ambiguities would have been impossible to resolve. The debugging loop was:

1. Run code → get error message
2. Write a debug script to dump raw bytes at that layer
3. Read the relevant MS-PST specification section
4. Compare the spec's bit/byte layout against actual hex values
5. Fix the code to match reality
6. Repeat

### 1. Hex Dump Debug Scripts

Create standalone scripts that dump raw bytes alongside parsed structures:

```go
// Debug script to examine heap allocations
data, _ := node.GetBlock(0)
fmt.Printf("Raw block: % X\n", data)
fmt.Printf("ibHnpm (page map offset): %d\n", binary.LittleEndian.Uint16(data[0:2]))
fmt.Printf("bClientSig: 0x%02X\n", data[3])
```

This was the primary debugging technique. Each bug was found by comparing what the code computed against what the hex dump showed.

### 2. Layer-by-Layer Validation

Test each layer independently before integrating:

```go
// Test disk layer: Can we read the header and pages?
// Test NDB layer: Can we find nodes and blocks in the B-trees?
// Test LTP layer: Can we parse heaps, BTHs, and property bags?
// Test messaging layer: Can we read folders and messages?
```

When the messaging layer fails, the bug is often 2-3 layers down.

### 3. Comparative Testing

When one operation works and a similar one fails, compare them:

```go
// These BIDs go through different B-tree paths
db.LookupBlock(0x40)   // Fails - goes through child 1
db.LookupBlock(0x1390) // Works - goes through child 3
// Dump both paths to find where they diverge
```

### 4. Specification as Ground Truth

The MS-PST specification was the authoritative reference. When code behavior didn't match expectations:

1. Find the exact section describing that structure
2. Read the bit/byte layout carefully
3. Verify field meanings (e.g., "cbEnt" = value size, not entry size)
4. Check for off-by-one issues (1-based vs 0-based indices)

### What Didn't Help Much

**Reference implementations (C++/Rust)**: These were not consulted during debugging. The spec + real data was sufficient. Reference code could help for complex algorithms, but most bugs were simple bit manipulation or offset errors that hex dumps revealed directly.

---

## Key Takeaways

1. **A real test file is essential**: Without actual data, specification ambiguities are unresolvable. The sample PST file was more valuable than any reference implementation.

2. **Hex dumps reveal truth**: When parsing fails, dump the raw bytes and compare against the spec. Most bugs become obvious when you see `0x40` in the data but your code computes `0x20`.

3. **Bit fields are tricky**: Always document the exact bit layout. A 1-bit shift error or wrong mask causes cascading failures.

4. **Specification wording matters**: "EntrySize" vs "full entry size" - read the actual description, not just field names.

5. **Structural relationships vary**: Parent-child, sibling-by-index, embedded subnodes - PST uses all of them. Don't assume.

6. **B-tree semantics differ**: Min-key vs max-key partitioning changes the search algorithm. Verify against actual tree structure.

7. **Test incrementally**: When the top layer fails, the bug is usually 2-3 layers down. Test each layer independently.

8. **Debug scripts are disposable**: Write quick scripts in `/tmp` to probe specific structures. Delete them when done.

---

## References

### Primary (Used Extensively)

- [MS-PST]: Office File Formats - PST File Format
  https://docs.microsoft.com/en-us/openspecs/office_file_formats/ms-pst/

  *The authoritative specification. Every bug fix required reading relevant sections.*

- [MS-OXCDATA]: Data Structures
  https://docs.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxcdata/

- [MS-OXPROPS]: Exchange Server Protocols Master Property List
  https://docs.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxprops/

### Secondary (Available for Reference)

- libpff: Library to access the Personal Folder File (PFF) format
  https://github.com/libyal/libpff

- pst-sdk (Rust): PST parsing library
  https://github.com/nicokosi/pst-sdk

  *These can help with complex algorithms but weren't needed for the bugs documented here. The spec + test file was sufficient.*
