# Disk Layer

The disk layer (`pkg/disk/`) handles the binary file format defined in [MS-PST].

## Components

### Header (`header.go`)

Parses the PST file header to determine format and locate B-tree roots.

```go
header, err := disk.ReadHeader(file)
if err != nil {
    log.Fatal(err)
}

// Format detection
if header.IsUnicode() {
    // 64-bit addresses
} else {
    // 32-bit addresses (ANSI)
}

// B-tree roots
nbtRoot := header.NBTRoot()
bbtRoot := header.BBTRoot()
```

### Header Structure

| Field | ANSI Offset | Unicode Offset | Description |
|-------|-------------|----------------|-------------|
| dwMagic | 0 | 0 | Magic number (0x4E444221) |
| wMagicClient | 8 | 8 | PST (0x4D53) or OST (0x4F53) |
| wVer | 10 | 10 | Format version |
| bCryptMethod | 461 | 509 | Encryption method |
| root | 164 | 172 | B-tree root info |

### Pages (`page.go`)

Parses 512-byte B-tree pages:

```go
pageData := make([]byte, disk.PageSize)
file.ReadAt(pageData, offset)

page, err := disk.ParseBTPage(pageData, format, disk.PageTypeNBT)
if page.IsLeaf() {
    for _, entry := range page.NBTEntries {
        // Process leaf entries
    }
}
```

### Page Types

| Type | Value | Description |
|------|-------|-------------|
| BBT | 0x80 | Block B-tree |
| NBT | 0x81 | Node B-tree |
| FMap | 0x82 | Free Map (deprecated) |
| PMap | 0x83 | Page Map (deprecated) |
| AMap | 0x84 | Allocation Map |
| FPMap | 0x85 | Free Page Map (deprecated) |
| DList | 0x86 | Density List |

### Blocks (`block.go`)

Parses data blocks (up to 8KB):

```go
// Extended blocks contain references to other blocks
eb, err := disk.ParseExtendedBlock(data, format)
for _, bid := range eb.BIDs {
    // Read child blocks
}

// Subnode blocks contain private node hierarchies
sb, err := disk.ParseSubnodeBlock(data, format)
for _, entry := range sb.LeafEntries {
    // Process subnode entries
}
```

### Block Types

| Type | Value | Description |
|------|-------|-------------|
| External | 0x00 | Data block |
| Extended | 0x01 | Multi-block reference tree |
| Subnode | 0x02 | Subnode container |

### Encryption (`crypt.go`)

Handles decryption of external blocks:

```go
// Automatic decryption
data := disk.DecryptBlock(encryptedData, method, blockID)

// Or specific methods
decoded := disk.PermuteDecode(data)
decoded := disk.CyclicDecode(data, key)
```

### Encryption Methods

| Method | Value | Description |
|--------|-------|-------------|
| None | 0 | No encryption |
| Permute | 1 | Byte substitution cipher |
| Cyclic | 2 | XOR-based cipher |

## Constants (`constants.go`)

Key constants from the MS-PST specification:

```go
const (
    PageSize         = 512   // All pages are 512 bytes
    MaxBlockDiskSize = 8192  // Max block size (8KB)
    BytesPerSlot     = 64    // AMap allocation unit
)
```

## ANSI vs Unicode

The format affects address sizes throughout the file:

| Aspect | ANSI | Unicode |
|--------|------|---------|
| Header size | 512 bytes | 568 bytes |
| Address size | 4 bytes | 8 bytes |
| Page trailer | 12 bytes | 16 bytes |
| Block trailer | 12 bytes | 16 bytes |
| NBT entry | 16 bytes | 32 bytes |
| BBT entry | 12 bytes | 24 bytes |
| Max file size | ~2GB | No limit |

## CRC and Signature

Data integrity is verified using:

```go
// CRC-32 checksum
crc := disk.ComputeCRC(data)

// Signature from ID and address
sig := disk.ComputeSignature(id, address)
```
