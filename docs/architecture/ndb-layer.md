# NDB Layer

The Node Database layer (`pkg/ndb/`) provides the node abstraction used by higher layers.

## Components

### Database (`database.go`)

The entry point for accessing nodes and blocks:

```go
db, err := ndb.Open("archive.pst")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Get a node by ID
node, err := db.GetNode(nid)

// Direct block access
blockData, err := db.ReadBlockData(bid)
```

### Node (`node.go`)

Represents a single node in the PST:

```go
// Read all data
data, err := node.ReadAll()

// Read partial data
data, err := node.Read(offset, size)

// Check for subnodes
if node.HasSubnodes() {
    subnode, err := node.LookupSubnode(subnodeNID)
}
```

## Node IDs

Node IDs encode both type and index:

```go
// Create a node ID
nid := util.MakeNID(util.NIDTypeNormalFolder, 0x122)

// Extract components
nodeType := nid.Type()   // NIDTypeNormalFolder
nodeIndex := nid.Index() // 0x122
```

### Node Types

| Type | Value | Description |
|------|-------|-------------|
| HID | 0x00 | Heap node |
| Internal | 0x01 | Internal node |
| NormalFolder | 0x02 | Normal folder |
| SearchFolder | 0x03 | Search folder |
| NormalMessage | 0x04 | Normal message |
| Attachment | 0x05 | Attachment |
| HierarchyTable | 0x0D | Folder hierarchy table |
| ContentsTable | 0x0E | Folder contents table |
| AttachmentTable | 0x11 | Message attachment table |
| RecipientTable | 0x12 | Message recipient table |

### Well-Known Node IDs

```go
var (
    NIDMessageStore = util.MakeNID(util.NIDTypeInternal, 0x1)
    NIDNameIDMap    = util.MakeNID(util.NIDTypeInternal, 0x3)
    NIDRootFolder   = util.MakeNID(util.NIDTypeNormalFolder, 0x122)
)
```

## B-Trees

### Node B-Tree (NBT)

Maps node IDs to node information:

```
Node ID → (Data Block ID, Subnode Block ID, Parent Node ID)
```

### Block B-Tree (BBT)

Maps block IDs to block locations:

```
Block ID → (File Offset, Size, Reference Count)
```

## Block Reading

Blocks may span multiple disk blocks using extended blocks:

```
┌─────────────────────┐
│   Extended Block    │
│   (Level 1)         │
├─────────────────────┤
│ BID 1 │ BID 2 │ ... │
└───┬───────┬─────────┘
    │       │
    ▼       ▼
┌───────┐ ┌───────┐
│ Data  │ │ Data  │
│ Block │ │ Block │
└───────┘ └───────┘
```

The node abstraction handles this automatically:

```go
// Reads across all data blocks seamlessly
allData, err := node.ReadAll()
```

## Subnodes

Nodes can contain private subnode hierarchies:

```
┌─────────────────────┐
│      Top Node       │
│  (DataBID, SubBID)  │
└─────────┬───────────┘
          │ SubBID
          ▼
┌─────────────────────┐
│   Subnode Block     │
├─────────────────────┤
│ Entry 1 │ Entry 2   │
└───┬───────────┬─────┘
    │           │
    ▼           ▼
┌───────┐   ┌───────┐
│Subnode│   │Subnode│
│  Data │   │  Data │
└───────┘   └───────┘
```

Access subnodes by ID:

```go
if node.HasSubnodes() {
    subnode, err := node.LookupSubnode(subnodeNID)
    subData, _ := subnode.ReadAll()
}
```

## Caching

The database caches node and block lookups:

```go
// First lookup reads from disk
node1, _ := db.GetNode(nid)

// Second lookup uses cache
node2, _ := db.GetNode(nid) // Fast
```

## Node Reader

For streaming access to large nodes:

```go
reader, err := ndb.NewNodeReader(node)
if err != nil {
    log.Fatal(err)
}

// Use as io.Reader
buf := make([]byte, 1024)
n, err := reader.Read(buf)

// Seekable
reader.Seek(100, io.SeekStart)
```
