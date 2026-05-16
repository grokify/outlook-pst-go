package disk

import (
	"encoding/binary"
	"fmt"
)

// BlockTrailer contains the trailer at the end of each block.
// See [MS-PST] Section 2.2.2.8.1 - BLOCKTRAILER structure.
type BlockTrailer struct {
	Size      uint16 // cb - Unaligned data size
	Signature uint16 // wSig - Block signature (computed per [MS-PST] 5.5)
	CRC       uint32 // dwCRC - CRC32 of block data
	BID       uint64 // bid - Block ID
}

// ParseBlockTrailerUnicode parses a Unicode format block trailer.
func ParseBlockTrailerUnicode(data []byte) BlockTrailer {
	// Unicode trailer (16 bytes):
	// cb: 0-2 (2 bytes)
	// wSig: 2-4 (2 bytes)
	// dwCRC: 4-8 (4 bytes)
	// bid: 8-16 (8 bytes)
	return BlockTrailer{
		Size:      binary.LittleEndian.Uint16(data[0:2]),
		Signature: binary.LittleEndian.Uint16(data[2:4]),
		CRC:       binary.LittleEndian.Uint32(data[4:8]),
		BID:       binary.LittleEndian.Uint64(data[8:16]),
	}
}

// ParseBlockTrailerANSI parses an ANSI format block trailer.
func ParseBlockTrailerANSI(data []byte) BlockTrailer {
	// ANSI trailer (12 bytes):
	// cb: 0-2 (2 bytes)
	// wSig: 2-4 (2 bytes)
	// bid: 4-8 (4 bytes)
	// dwCRC: 8-12 (4 bytes)
	return BlockTrailer{
		Size:      binary.LittleEndian.Uint16(data[0:2]),
		Signature: binary.LittleEndian.Uint16(data[2:4]),
		BID:       uint64(binary.LittleEndian.Uint32(data[4:8])),
		CRC:       binary.LittleEndian.Uint32(data[8:12]),
	}
}

// ExtendedBlock represents an extended block (tree of data blocks).
// See [MS-PST] Section 2.2.2.8.3.1 - XBLOCK structure and Section 2.2.2.8.3.2 - XXBLOCK.
// Extended blocks are used when node data exceeds the maximum single block size.
type ExtendedBlock struct {
	BlockType byte     // btype - Should be 0x01 for XBLOCK/XXBLOCK
	Level     byte     // cLevel - 1 for XBLOCK (points to data), 2 for XXBLOCK (points to XBLOCKs)
	Count     uint16   // cEnt - Number of block references
	TotalSize uint32   // lcbTotal - Total logical size of all referenced data
	BIDs      []uint64 // rgbid - Array of block IDs
}

// ParseExtendedBlock parses an extended block.
func ParseExtendedBlock(data []byte, format PSTFormat) (*ExtendedBlock, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data too small for extended block header")
	}

	eb := &ExtendedBlock{
		BlockType: data[0],
		Level:     data[1],
		Count:     binary.LittleEndian.Uint16(data[2:4]),
		TotalSize: binary.LittleEndian.Uint32(data[4:8]),
	}

	if eb.BlockType != byte(BlockTypeExtended) {
		return nil, fmt.Errorf("invalid extended block type: got 0x%02X, expected 0x%02X", eb.BlockType, BlockTypeExtended)
	}

	// Parse BID array
	bidSize := 8
	if format == FormatANSI {
		bidSize = 4
	}

	eb.BIDs = make([]uint64, eb.Count)
	offset := 8
	for i := 0; i < int(eb.Count); i++ {
		if format == FormatUnicode {
			eb.BIDs[i] = binary.LittleEndian.Uint64(data[offset : offset+8])
		} else {
			eb.BIDs[i] = uint64(binary.LittleEndian.Uint32(data[offset : offset+4]))
		}
		offset += bidSize
	}

	return eb, nil
}

// SubnodeLeafEntry represents a subnode leaf entry.
type SubnodeLeafEntry struct {
	NID     uint64 // Subnode ID (node ID within this subnode tree)
	DataBID uint64 // Data block ID
	SubBID  uint64 // Sub-subnode block ID
}

// SubnodeNonleafEntry represents a subnode non-leaf entry.
type SubnodeNonleafEntry struct {
	Key    uint64 // Key value (NID)
	SubBID uint64 // Reference to child subnode block
}

// SubnodeBlock represents a subnode block.
type SubnodeBlock struct {
	BlockType byte   // Should be BlockTypeSubnode (0x02)
	Level     byte   // 0 = leaf, >0 = non-leaf
	Count     uint16 // Number of entries

	// For leaf blocks
	LeafEntries []SubnodeLeafEntry

	// For non-leaf blocks
	NonleafEntries []SubnodeNonleafEntry
}

// ParseSubnodeBlock parses a subnode block.
func ParseSubnodeBlock(data []byte, format PSTFormat) (*SubnodeBlock, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too small for subnode block header")
	}

	sb := &SubnodeBlock{
		BlockType: data[0],
		Level:     data[1],
		Count:     binary.LittleEndian.Uint16(data[2:4]),
	}

	if sb.BlockType != byte(BlockTypeSubnode) {
		return nil, fmt.Errorf("invalid subnode block type: got 0x%02X, expected 0x%02X", sb.BlockType, BlockTypeSubnode)
	}

	offset := 4
	// Padding to 8-byte boundary
	if format == FormatUnicode {
		offset = 8
	}

	if sb.Level == 0 {
		// Leaf entries
		sb.LeafEntries = make([]SubnodeLeafEntry, sb.Count)
		for i := 0; i < int(sb.Count); i++ {
			if format == FormatUnicode {
				// Unicode leaf entry (24 bytes):
				// nid: 0-8 (8 bytes) - but only lower 32 bits used as node_id
				// bidData: 8-16 (8 bytes)
				// bidSub: 16-24 (8 bytes)
				sb.LeafEntries[i] = SubnodeLeafEntry{
					NID:     binary.LittleEndian.Uint64(data[offset : offset+8]),
					DataBID: binary.LittleEndian.Uint64(data[offset+8 : offset+16]),
					SubBID:  binary.LittleEndian.Uint64(data[offset+16 : offset+24]),
				}
				offset += 24
			} else {
				// ANSI leaf entry (12 bytes):
				// nid: 0-4 (4 bytes)
				// bidData: 4-8 (4 bytes)
				// bidSub: 8-12 (4 bytes)
				sb.LeafEntries[i] = SubnodeLeafEntry{
					NID:     uint64(binary.LittleEndian.Uint32(data[offset : offset+4])),
					DataBID: uint64(binary.LittleEndian.Uint32(data[offset+4 : offset+8])),
					SubBID:  uint64(binary.LittleEndian.Uint32(data[offset+8 : offset+12])),
				}
				offset += 12
			}
		}
	} else {
		// Non-leaf entries
		sb.NonleafEntries = make([]SubnodeNonleafEntry, sb.Count)
		for i := 0; i < int(sb.Count); i++ {
			if format == FormatUnicode {
				// Unicode non-leaf entry (16 bytes):
				// key: 0-8 (8 bytes)
				// bidSub: 8-16 (8 bytes)
				sb.NonleafEntries[i] = SubnodeNonleafEntry{
					Key:    binary.LittleEndian.Uint64(data[offset : offset+8]),
					SubBID: binary.LittleEndian.Uint64(data[offset+8 : offset+16]),
				}
				offset += 16
			} else {
				// ANSI non-leaf entry (8 bytes):
				// key: 0-4 (4 bytes)
				// bidSub: 4-8 (4 bytes)
				sb.NonleafEntries[i] = SubnodeNonleafEntry{
					Key:    uint64(binary.LittleEndian.Uint32(data[offset : offset+4])),
					SubBID: uint64(binary.LittleEndian.Uint32(data[offset+4 : offset+8])),
				}
				offset += 8
			}
		}
	}

	return sb, nil
}

// IsLeaf returns true if this is a leaf subnode block.
func (sb *SubnodeBlock) IsLeaf() bool {
	return sb.Level == 0
}

// BlockDataSize calculates the size needed for block data including trailer.
func BlockDataSize(dataSize uint16, format PSTFormat) uint64 {
	alignedSize := AlignDisk(uint64(dataSize))
	if format == FormatUnicode {
		return alignedSize + BlockTrailerSizeUnicode
	}
	return alignedSize + BlockTrailerSizeANSI
}
