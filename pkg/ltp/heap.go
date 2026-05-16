// Package ltp implements the LTP (Lists, Tables, and Properties) layer of the PST format.
// This layer sits on top of the NDB layer and provides property access and table structures.
// See [MS-PST] Section 2.3 for the LTP layer specification.
//
// [MS-PST]: https://docs.microsoft.com/en-us/openspecs/office_file_formats/ms-pst/
package ltp

import (
	"encoding/binary"
	"fmt"

	"github.com/grokify/outlook-pst-go/pkg/disk"
	"github.com/grokify/outlook-pst-go/pkg/ndb"
	"github.com/grokify/outlook-pst-go/pkg/util"
)

// HeapOnNode provides access to heap allocations within a node.
// See [MS-PST] Section 2.3.1 - Heap-on-Node (HN).
// HN is a memory allocation mechanism built on top of NDB nodes that allows
// variable-size allocations to be stored efficiently.
type HeapOnNode struct {
	node   *ndb.Node
	blocks []heapBlock
}

// heapBlock represents a single heap block (page).
type heapBlock struct {
	data        []byte
	pageMapOff  uint16
	numAllocs   uint16
	allocations []uint16 // Allocation offsets
}

// HeapFirstHeader represents the header of the first heap block.
// See [MS-PST] Section 2.3.1.2 - HNHDR structure.
type HeapFirstHeader struct {
	PageMapOffset   uint16      // ibHnpm - Offset to page map
	Signature       byte        // bSig - Must be 0xEC (HN signature)
	ClientSignature byte        // bClientSig - Client signature (PC=0xBC, TC=0x7C, BTH=0xB5)
	RootID          util.HeapID // hidUserRoot - HID of root allocation
	FillLevels      [4]byte     // rgbFillLevel - Fill level indicators
}

// NewHeapOnNode creates a HeapOnNode from a node.
func NewHeapOnNode(node *ndb.Node) (*HeapOnNode, error) {
	count, err := node.BlockCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get block count: %w", err)
	}

	h := &HeapOnNode{
		node:   node,
		blocks: make([]heapBlock, count),
	}

	// Load all blocks
	for i := 0; i < count; i++ {
		data, err := node.GetBlock(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get block %d: %w", i, err)
		}
		if err := h.parseBlock(i, data); err != nil {
			return nil, fmt.Errorf("failed to parse heap block %d: %w", i, err)
		}
	}

	return h, nil
}

// parseBlock parses a heap block.
func (h *HeapOnNode) parseBlock(index int, data []byte) error {
	block := &h.blocks[index]
	block.data = data

	if len(data) < 4 {
		return fmt.Errorf("heap block too small: %d bytes", len(data))
	}

	if index == 0 {
		// First block has extended header
		if len(data) < 12 {
			return fmt.Errorf("first heap block too small: %d bytes", len(data))
		}

		header := HeapFirstHeader{
			PageMapOffset:   binary.LittleEndian.Uint16(data[0:2]),
			Signature:       data[2],
			ClientSignature: data[3],
			RootID:          util.HeapID(binary.LittleEndian.Uint32(data[4:8])),
		}
		copy(header.FillLevels[:], data[8:12])

		if header.Signature != disk.HeapSignature {
			return fmt.Errorf("invalid heap signature: got 0x%02X, expected 0x%02X", header.Signature, disk.HeapSignature)
		}

		block.pageMapOff = header.PageMapOffset
	} else {
		// Subsequent blocks have just page map offset
		block.pageMapOff = binary.LittleEndian.Uint16(data[0:2])
	}

	// Parse page map
	if int(block.pageMapOff) >= len(data)-2 {
		return fmt.Errorf("page map offset out of bounds: %d >= %d", block.pageMapOff, len(data)-2)
	}

	pageMap := data[block.pageMapOff:]
	if len(pageMap) < 4 {
		return fmt.Errorf("page map too small")
	}

	block.numAllocs = binary.LittleEndian.Uint16(pageMap[0:2])
	// numFrees at pageMap[2:4] - we don't need it for reading

	// Parse allocation offsets
	// There are numAllocs+1 entries (last one is end offset for calculating size)
	numOffsets := int(block.numAllocs) + 1
	if len(pageMap) < 4+numOffsets*2 {
		return fmt.Errorf("page map truncated: need %d bytes, have %d", 4+numOffsets*2, len(pageMap))
	}

	block.allocations = make([]uint16, numOffsets)
	for i := 0; i < numOffsets; i++ {
		block.allocations[i] = binary.LittleEndian.Uint16(pageMap[4+i*2 : 4+i*2+2])
	}

	return nil
}

// RootID returns the root heap ID from the first block header.
func (h *HeapOnNode) RootID() util.HeapID {
	if len(h.blocks) == 0 || len(h.blocks[0].data) < 8 {
		return 0
	}
	return util.HeapID(binary.LittleEndian.Uint32(h.blocks[0].data[4:8]))
}

// ClientSignature returns the client signature from the first block header.
func (h *HeapOnNode) ClientSignature() byte {
	if len(h.blocks) == 0 || len(h.blocks[0].data) < 4 {
		return 0
	}
	return h.blocks[0].data[3]
}

// Read reads data from a heap allocation.
func (h *HeapOnNode) Read(hid util.HeapID) ([]byte, error) {
	if hid == 0 {
		return nil, nil
	}

	pageIndex := hid.PageIndex()
	allocIndex := hid.AllocIndex()

	if int(pageIndex) >= len(h.blocks) {
		return nil, fmt.Errorf("heap page index out of bounds: %d >= %d", pageIndex, len(h.blocks))
	}

	block := &h.blocks[pageIndex]

	if int(allocIndex) >= int(block.numAllocs) {
		return nil, fmt.Errorf("heap alloc index out of bounds: %d >= %d", allocIndex, block.numAllocs)
	}

	// Get allocation start and end offsets
	start := block.allocations[allocIndex]
	end := block.allocations[allocIndex+1]

	if start > end || int(end) > len(block.data) {
		return nil, fmt.Errorf("invalid allocation bounds: start=%d, end=%d, block size=%d", start, end, len(block.data))
	}

	return block.data[start:end], nil
}

// Size returns the size of a heap allocation.
func (h *HeapOnNode) Size(hid util.HeapID) (int, error) {
	if hid == 0 {
		return 0, nil
	}

	pageIndex := hid.PageIndex()
	allocIndex := hid.AllocIndex()

	if int(pageIndex) >= len(h.blocks) {
		return 0, fmt.Errorf("heap page index out of bounds: %d >= %d", pageIndex, len(h.blocks))
	}

	block := &h.blocks[pageIndex]

	if int(allocIndex) >= int(block.numAllocs) {
		return 0, fmt.Errorf("heap alloc index out of bounds: %d >= %d", allocIndex, block.numAllocs)
	}

	start := block.allocations[allocIndex]
	end := block.allocations[allocIndex+1]

	return int(end - start), nil
}

// Node returns the underlying node.
func (h *HeapOnNode) Node() *ndb.Node {
	return h.node
}
