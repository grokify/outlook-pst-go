package ndb

import (
	"fmt"
	"io"
	"iter"
	"sync"

	"github.com/grokify/outlook-pst-go/pkg/disk"
	"github.com/grokify/outlook-pst-go/pkg/util"
)

// Node represents a node in the PST file.
// A node is the basic unit of storage in the NDB layer, identified by a NodeID.
// Each node can have associated data (stored in data blocks) and subnodes.
// See [MS-PST] Section 2.2.2.7.7.4 for NBTENTRY (node information) and
// Section 2.2.2.8 for block structures used by nodes.
type Node struct {
	db   *Database
	info *NodeInfo

	// Data blocks (lazy loaded) - See [MS-PST] Section 2.2.2.8.1 for data blocks
	dataOnce   sync.Once
	dataBlocks [][]byte
	dataSize   uint64
	dataErr    error

	// Subnode block (lazy loaded) - See [MS-PST] Section 2.2.2.8.3.3 for subnode blocks
	subOnce  sync.Once
	subBlock *disk.SubnodeBlock
	subErr   error
}

// newNode creates a new Node from NodeInfo.
func newNode(db *Database, info *NodeInfo) *Node {
	return &Node{
		db:   db,
		info: info,
	}
}

// ID returns the node ID.
func (n *Node) ID() util.NodeID {
	return n.info.NID
}

// Type returns the node type.
func (n *Node) Type() util.NIDType {
	return n.info.NID.Type()
}

// ParentID returns the parent node ID.
func (n *Node) ParentID() util.NodeID {
	return n.info.ParentNID
}

// HasData returns true if the node has a data block.
func (n *Node) HasData() bool {
	return n.info.DataBID != 0
}

// HasSubnodes returns true if the node has subnodes.
func (n *Node) HasSubnodes() bool {
	return n.info.SubBID != 0
}

// loadData loads the data blocks for this node.
func (n *Node) loadData() error {
	n.dataOnce.Do(func() {
		if n.info.DataBID == 0 {
			n.dataBlocks = [][]byte{}
			n.dataSize = 0
			return
		}

		blocks, size, err := n.db.readNodeData(n.info.DataBID)
		if err != nil {
			n.dataErr = err
			return
		}
		n.dataBlocks = blocks
		n.dataSize = size
	})
	return n.dataErr
}

// readNodeData reads the data for a node, handling extended blocks.
func (db *Database) readNodeData(bid util.BlockID) ([][]byte, uint64, error) {
	info, err := db.LookupBlock(bid)
	if err != nil {
		return nil, 0, err
	}

	data, err := db.ReadBlockDataFromInfo(info)
	if err != nil {
		return nil, 0, err
	}

	// Check if this is an extended block (internal)
	if bid.IsInternal() {
		eb, err := disk.ParseExtendedBlock(data, db.header.Format)
		if err != nil {
			return nil, 0, err
		}

		if eb.Level == 0 {
			// Level 0: BIDs point to data blocks
			var blocks [][]byte
			var totalSize uint64
			for _, childBID := range eb.BIDs {
				childData, err := db.ReadBlockData(util.BlockID(childBID))
				if err != nil {
					return nil, 0, err
				}
				blocks = append(blocks, childData)
				totalSize += uint64(len(childData))
			}
			return blocks, totalSize, nil
		}

		// Level 1+: BIDs point to extended blocks
		var blocks [][]byte
		var totalSize uint64
		for _, childBID := range eb.BIDs {
			childBlocks, childSize, err := db.readNodeData(util.BlockID(childBID))
			if err != nil {
				return nil, 0, err
			}
			blocks = append(blocks, childBlocks...)
			totalSize += childSize
		}
		return blocks, totalSize, nil
	}

	// Simple data block
	return [][]byte{data}, uint64(len(data)), nil
}

// Size returns the total data size for this node.
func (n *Node) Size() (uint64, error) {
	if err := n.loadData(); err != nil {
		return 0, err
	}
	return n.dataSize, nil
}

// BlockCount returns the number of data blocks.
func (n *Node) BlockCount() (int, error) {
	if err := n.loadData(); err != nil {
		return 0, err
	}
	return len(n.dataBlocks), nil
}

// ReadAll reads all data from the node.
func (n *Node) ReadAll() ([]byte, error) {
	if err := n.loadData(); err != nil {
		return nil, err
	}

	if len(n.dataBlocks) == 0 {
		return []byte{}, nil
	}
	if len(n.dataBlocks) == 1 {
		return n.dataBlocks[0], nil
	}

	// Concatenate all blocks
	result := make([]byte, 0, n.dataSize)
	for _, block := range n.dataBlocks {
		result = append(result, block...)
	}
	return result, nil
}

// Read reads data from the node at the given offset.
func (n *Node) Read(offset, size uint64) ([]byte, error) {
	if err := n.loadData(); err != nil {
		return nil, err
	}

	if offset >= n.dataSize {
		return []byte{}, nil
	}
	if offset+size > n.dataSize {
		size = n.dataSize - offset
	}

	result := make([]byte, 0, size)
	remaining := size
	currentOffset := uint64(0)

	for _, block := range n.dataBlocks {
		blockSize := uint64(len(block))

		// Skip blocks before the offset
		if currentOffset+blockSize <= offset {
			currentOffset += blockSize
			continue
		}

		// Calculate how much to read from this block
		blockStart := uint64(0)
		if currentOffset < offset {
			blockStart = offset - currentOffset
		}

		blockEnd := blockSize
		if blockStart+remaining < blockSize {
			blockEnd = blockStart + remaining
		}

		result = append(result, block[blockStart:blockEnd]...)
		remaining -= (blockEnd - blockStart)

		if remaining == 0 {
			break
		}

		currentOffset += blockSize
	}

	return result, nil
}

// GetBlock returns a specific data block by index.
func (n *Node) GetBlock(index int) ([]byte, error) {
	if err := n.loadData(); err != nil {
		return nil, err
	}

	if index < 0 || index >= len(n.dataBlocks) {
		return nil, fmt.Errorf("block index out of range: %d (have %d blocks)", index, len(n.dataBlocks))
	}

	return n.dataBlocks[index], nil
}

// loadSubnodes loads the subnode block.
func (n *Node) loadSubnodes() error {
	n.subOnce.Do(func() {
		if n.info.SubBID == 0 {
			return
		}

		data, err := n.db.ReadBlockData(n.info.SubBID)
		if err != nil {
			n.subErr = err
			return
		}

		n.subBlock, err = disk.ParseSubnodeBlock(data, n.db.header.Format)
		if err != nil {
			n.subErr = err
			return
		}
	})
	return n.subErr
}

// LookupSubnode looks up a subnode by ID.
func (n *Node) LookupSubnode(nid util.NodeID) (*Node, error) {
	if err := n.loadSubnodes(); err != nil {
		return nil, err
	}

	if n.subBlock == nil {
		return nil, fmt.Errorf("subnode not found: 0x%X (no subnodes)", nid)
	}

	info, err := n.searchSubnode(n.subBlock, uint64(nid))
	if err != nil {
		return nil, err
	}

	return &Node{
		db:   n.db,
		info: info,
	}, nil
}

// searchSubnode searches for a subnode in a subnode block.
func (n *Node) searchSubnode(block *disk.SubnodeBlock, nid uint64) (*NodeInfo, error) {
	if block.IsLeaf() {
		for _, entry := range block.LeafEntries {
			if entry.NID == nid {
				return &NodeInfo{
					NID:     util.NodeID(entry.NID),
					DataBID: util.BlockID(entry.DataBID),
					SubBID:  util.BlockID(entry.SubBID),
				}, nil
			}
		}
		return nil, fmt.Errorf("subnode not found: 0x%X", nid)
	}

	// Non-leaf: find appropriate child
	var childBID uint64
	for _, entry := range block.NonleafEntries {
		if nid <= entry.Key {
			childBID = entry.SubBID
			break
		}
	}
	if childBID == 0 && len(block.NonleafEntries) > 0 {
		childBID = block.NonleafEntries[len(block.NonleafEntries)-1].SubBID
	}
	if childBID == 0 {
		return nil, fmt.Errorf("subnode not found: 0x%X", nid)
	}

	// Read child block
	childData, err := n.db.ReadBlockData(util.BlockID(childBID))
	if err != nil {
		return nil, err
	}
	childBlock, err := disk.ParseSubnodeBlock(childData, n.db.header.Format)
	if err != nil {
		return nil, err
	}

	return n.searchSubnode(childBlock, nid)
}

// Subnodes returns an iterator over all subnodes.
func (n *Node) Subnodes() iter.Seq2[*Node, error] {
	return func(yield func(*Node, error) bool) {
		if err := n.loadSubnodes(); err != nil {
			yield(nil, err)
			return
		}

		if n.subBlock == nil {
			return
		}

		n.iterateSubnodes(n.subBlock, yield)
	}
}

// iterateSubnodes iterates through all entries in a subnode block.
func (n *Node) iterateSubnodes(block *disk.SubnodeBlock, yield func(*Node, error) bool) bool {
	if block.IsLeaf() {
		for _, entry := range block.LeafEntries {
			info := &NodeInfo{
				NID:     util.NodeID(entry.NID),
				DataBID: util.BlockID(entry.DataBID),
				SubBID:  util.BlockID(entry.SubBID),
			}
			subnode := &Node{db: n.db, info: info}
			if !yield(subnode, nil) {
				return false
			}
		}
		return true
	}

	// Non-leaf: recurse into children
	for _, entry := range block.NonleafEntries {
		childData, err := n.db.ReadBlockData(util.BlockID(entry.SubBID))
		if err != nil {
			yield(nil, err)
			return false
		}
		childBlock, err := disk.ParseSubnodeBlock(childData, n.db.header.Format)
		if err != nil {
			yield(nil, err)
			return false
		}
		if !n.iterateSubnodes(childBlock, yield) {
			return false
		}
	}
	return true
}

// NodeReader provides an io.Reader interface for node data.
type NodeReader struct {
	node   *Node
	offset uint64
}

// NewNodeReader creates a new reader for the node's data.
func NewNodeReader(n *Node) (*NodeReader, error) {
	if err := n.loadData(); err != nil {
		return nil, err
	}
	return &NodeReader{node: n}, nil
}

// Read implements io.Reader.
func (r *NodeReader) Read(p []byte) (int, error) {
	size, err := r.node.Size()
	if err != nil {
		return 0, err
	}

	if r.offset >= size {
		return 0, io.EOF
	}

	data, err := r.node.Read(r.offset, uint64(len(p)))
	if err != nil {
		return 0, err
	}

	n := copy(p, data)
	r.offset += uint64(n)

	if n == 0 {
		return 0, io.EOF
	}

	return n, nil
}

// Seek implements io.Seeker.
func (r *NodeReader) Seek(offset int64, whence int) (int64, error) {
	size, err := r.node.Size()
	if err != nil {
		return 0, err
	}

	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = int64(r.offset) + offset
	case io.SeekEnd:
		newOffset = int64(size) + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if newOffset < 0 {
		return 0, fmt.Errorf("negative seek position: %d", newOffset)
	}

	r.offset = uint64(newOffset)
	return newOffset, nil
}
