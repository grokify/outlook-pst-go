# Specification Reference

This library implements the Microsoft PST file format as defined in the MS-PST specification and related documents.

## Inline Code References

All major types and functions in the codebase include inline comments referencing the relevant MS-PST specification sections. For example:

```go
// Node represents a node in the PST file.
// See [MS-PST] Section 2.2.2.7.7.4 for NBTENTRY (node information).
type Node struct { ... }
```

This allows developers to quickly look up the specification when debugging or extending the library.

## Primary Specifications

### MS-PST: Outlook Personal Folders (.pst) File Format

The main specification defining the PST file format.

**URL**: [MS-PST](https://docs.microsoft.com/en-us/openspecs/office_file_formats/ms-pst/)

| Section | Topic | Go Package/Type |
|---------|-------|-----------------|
| 2.2 | NDB Layer | `pkg/ndb` |
| 2.2.2.1 | NID (Node ID) | `util.NodeID` |
| 2.2.2.2 | BID (Block ID) | `util.BlockID` |
| 2.2.2.6 | HEADER | `disk.Header` |
| 2.2.2.7 | Pages | `disk.BTPage` |
| 2.2.2.8 | Blocks | `disk.Block` |
| 2.3 | LTP Layer | `pkg/ltp` |
| 2.3.1 | HN (Heap-on-Node) | `ltp.HeapOnNode` |
| 2.3.2 | BTH (BTree-on-Heap) | `ltp.BTH` |
| 2.3.3 | PC (Property Context) | `ltp.PropertyBag` |
| 2.3.4 | TC (Table Context) | `ltp.Table` |
| 2.4 | Messaging Layer | Root package |
| 2.4.3 | Message Store | `PST.MessageStore()` |
| 2.4.4 | Folder Objects | `Folder` |
| 2.4.5 | Message Objects | `Message` |
| 2.4.6 | Attachment Objects | `Attachment` |
| 2.4.7 | Name-to-ID Map | `ltp.NamedPropertyMap` |
| 2.4.8 | Search | `SearchFolder` |
| 2.5 | Calculated CRC | `disk.ComputeCRC` |

### MS-OXRTFCP: RTF Compression Algorithm

Defines the compression algorithm used for RTF message bodies.

**URL**: [MS-OXRTFCP](https://docs.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxrtfcp/)

| Section | Topic | Go Package/Type |
|---------|-------|-----------------|
| 2.1 | LZFu Algorithm | `pkg/rtf` |
| 2.2 | Compressed RTF | `rtf.Decompress` |
| 2.2.2 | Header | `rtf.Header` |

### MS-OXPROPS: Exchange Server Protocols Property Sets

Defines MAPI property IDs and named property sets.

**URL**: [MS-OXPROPS](https://docs.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxprops/)

| Topic | Go Constants |
|-------|--------------|
| Property IDs | `ltp.PidTag*` |
| Property Types | `ltp.PropType*` |
| Property Sets | `util.PSETID_*`, `util.PS_*` |

### MS-OXCDATA: Data Structures

Defines common data structures used across Exchange protocols.

**URL**: [MS-OXCDATA](https://docs.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxcdata/)

| Topic | Relevance |
|-------|-----------|
| Property Values | Multi-value properties |
| Restrictions | Search folder criteria |
| Entry IDs | Folder/message references |

## Layer Mapping

### Disk Layer (pkg/disk)

| MS-PST Section | Implementation |
|----------------|----------------|
| 2.2.2.6 HEADER | `header.go` |
| 2.2.2.7 Pages | `page.go` |
| 2.2.2.8 Blocks | `block.go` |
| 2.5 CRC | `crypt.go` |

**Encryption Methods** (Section 2.2.2.6):

| Value | Constant | Description |
|-------|----------|-------------|
| 0x00 | `CryptNone` | No encryption |
| 0x01 | `CryptPermute` | Permutative encoding |
| 0x02 | `CryptCyclic` | Cyclic encoding |

### NDB Layer (pkg/ndb)

| MS-PST Section | Implementation |
|----------------|----------------|
| 2.2.2.1 NID | `util.NodeID` |
| 2.2.2.2 BID | `util.BlockID` |
| 2.2.2.7.7 BTPage | `database.go` |
| 2.2.2.8 Data Blocks | `node.go` |

**Node Types** (Section 2.2.2.1):

| Value | Constant | Description |
|-------|----------|-------------|
| 0x00 | `NIDTypeHID` | Heap node |
| 0x01 | `NIDTypeInternal` | Internal node |
| 0x02 | `NIDTypeNormalFolder` | Normal folder |
| 0x03 | `NIDTypeSearchFolder` | Search folder |
| 0x04 | `NIDTypeNormalMessage` | Normal message |
| 0x05 | `NIDTypeAttachment` | Attachment |

### LTP Layer (pkg/ltp)

| MS-PST Section | Implementation |
|----------------|----------------|
| 2.3.1 HN | `heap.go` |
| 2.3.2 BTH | `bth.go` |
| 2.3.3 PC | `propbag.go` |
| 2.3.4 TC | `table.go` |

**Property Types** (MS-OXCDATA):

| Value | Constant | Size |
|-------|----------|------|
| 0x0002 | `PropTypeInt16` | 2 bytes |
| 0x0003 | `PropTypeInt32` | 4 bytes |
| 0x0004 | `PropTypeFloat32` | 4 bytes |
| 0x0005 | `PropTypeFloat64` | 8 bytes |
| 0x000B | `PropTypeBool` | 2 bytes |
| 0x0014 | `PropTypeInt64` | 8 bytes |
| 0x001F | `PropTypeString` | Variable |
| 0x0040 | `PropTypeSysTime` | 8 bytes |
| 0x0102 | `PropTypeBinary` | Variable |

### Messaging Layer (Root Package)

| MS-PST Section | Implementation |
|----------------|----------------|
| 2.4.3 Message Store | `pst.go` |
| 2.4.4 Folder | `folder.go` |
| 2.4.5 Message | `message.go` |
| 2.4.6 Attachment | `message.go` |
| 2.4.7 Name-to-ID Map | `ltp/namedprop.go` |
| 2.4.8 Search | `search.go` |

## Version Support

### PST Formats

| Format | Version | Supported |
|--------|---------|-----------|
| ANSI | 14-15 | ✅ |
| Unicode | 23 | ✅ |

### File Types

| Type | Constant | Supported |
|------|----------|-----------|
| PST | `TypePST` | ✅ |
| OST | `TypeOST` | ✅ |

## Implementation Notes

For a detailed analysis of specification nuances and common implementation pitfalls discovered while building this library, see:

- [Case Study: Porting PST Parsing from C++ to Rust to Go](../CASE_STUDY_PORTING.md)

This document covers subtle bugs related to bit field extraction, B-tree search semantics, and structural relationships that can trip up implementers.

## Additional Resources

- [PST File Format (Wikipedia)](https://en.wikipedia.org/wiki/Personal_Storage_Table)
- [libpst (Open Source)](https://www.five-ten-sg.com/libpst/)
- [Microsoft PST SDK (C++)](https://github.com/enrondata/microsoft-pst-sdk)
- [Microsoft PST SDK (Rust)](https://github.com/microsoft/outlook-pst-rs)
