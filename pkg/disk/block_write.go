package disk

import (
	"encoding/binary"
	"fmt"
)

// BuildBlock creates a complete block with data, padding, and trailer.
// The returned bytes are ready to be written to disk.
// Parameters:
//   - data: the block payload (will be encrypted if method != None and not internal)
//   - bid: the block ID
//   - offset: the file offset where the block will be written (for signature)
//   - format: PST format (Unicode or ANSI)
//   - crypt: encryption method
func BuildBlock(data []byte, bid uint64, offset uint64, format PSTFormat, crypt CryptMethod) ([]byte, error) {
	dataSize := len(data)
	if dataSize > int(MaxDataBlockSizeUnicode) {
		return nil, fmt.Errorf("block data too large: %d bytes (max %d)", dataSize, MaxDataBlockSizeUnicode)
	}

	// Determine trailer size
	trailerSize := BlockTrailerSizeUnicode
	if format == FormatANSI {
		trailerSize = BlockTrailerSizeANSI
	}

	// Calculate aligned size
	alignedDataSize := AlignDisk(uint64(dataSize))
	totalSize := alignedDataSize + uint64(trailerSize)

	// Allocate buffer
	buf := make([]byte, totalSize)

	// Copy and optionally encrypt data
	// Only external blocks are encrypted (internal blocks have bit 1 set in BID)
	isInternal := bid&0x2 != 0
	if crypt != CryptMethodNone && !isInternal {
		encrypted := EncryptBlock(data, crypt, bid)
		copy(buf, encrypted)
	} else {
		copy(buf, data)
	}

	// Padding between data and trailer is zeros (already zero-initialized)

	// Compute CRC of encrypted/plain data (before trailer)
	crc := ComputeCRC(buf[:alignedDataSize])

	// Compute signature
	sig := ComputeSignature(bid, offset)

	// Write trailer
	trailer := buf[alignedDataSize:]
	if format == FormatUnicode {
		// Unicode trailer (16 bytes):
		// cb: 0-2 (2 bytes) - unaligned data size
		// wSig: 2-4 (2 bytes)
		// dwCRC: 4-8 (4 bytes)
		// bid: 8-16 (8 bytes)
		binary.LittleEndian.PutUint16(trailer[0:2], uint16(dataSize))
		binary.LittleEndian.PutUint16(trailer[2:4], sig)
		binary.LittleEndian.PutUint32(trailer[4:8], crc)
		binary.LittleEndian.PutUint64(trailer[8:16], bid)
	} else {
		// ANSI trailer (12 bytes):
		// cb: 0-2 (2 bytes)
		// wSig: 2-4 (2 bytes)
		// bid: 4-8 (4 bytes)
		// dwCRC: 8-12 (4 bytes)
		binary.LittleEndian.PutUint16(trailer[0:2], uint16(dataSize))
		binary.LittleEndian.PutUint16(trailer[2:4], sig)
		binary.LittleEndian.PutUint32(trailer[4:8], uint32(bid)) //nolint:gosec // G115: ANSI format BID is 32-bit per MS-PST spec
		binary.LittleEndian.PutUint32(trailer[8:12], crc)
	}

	return buf, nil
}

// EncryptBlock encrypts block data using the specified method.
func EncryptBlock(data []byte, method CryptMethod, blockID uint64) []byte {
	switch method {
	case CryptMethodNone:
		result := make([]byte, len(data))
		copy(result, data)
		return result
	case CryptMethodPermute:
		return PermuteEncode(data)
	case CryptMethodCyclic:
		key := uint32(blockID) //nolint:gosec // G115: cyclic encryption uses 32-bit key by design
		return CyclicEncode(data, key)
	default:
		result := make([]byte, len(data))
		copy(result, data)
		return result
	}
}

// BuildExtendedBlock creates an extended block (XBLOCK or XXBLOCK).
// Extended blocks contain references to other blocks for large data.
// Parameters:
//   - level: 1 for XBLOCK (points to data blocks), 2 for XXBLOCK (points to XBLOCKs)
//   - bids: array of block IDs this extended block references
//   - totalSize: total logical size of all referenced data
//   - bid: this block's ID
//   - offset: file offset for signature
//   - format: PST format
func BuildExtendedBlock(level byte, bids []uint64, totalSize uint32, bid uint64, offset uint64, format PSTFormat) ([]byte, error) {
	// Validate
	maxEntries := MaxXBlockEntriesUnicode
	if format == FormatANSI {
		maxEntries = MaxXBlockEntriesANSI
	}
	if len(bids) > maxEntries {
		return nil, fmt.Errorf("too many BIDs for extended block: %d (max %d)", len(bids), maxEntries)
	}

	// Calculate data size
	bidSize := 8
	if format == FormatANSI {
		bidSize = 4
	}
	dataSize := 8 + len(bids)*bidSize // 8-byte header + BID array

	// Build data portion
	data := make([]byte, dataSize)
	data[0] = byte(BlockTypeExtended)                           // btype = 0x01
	data[1] = level                                             // cLevel
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(bids))) //nolint:gosec // G115: entry count bounded by MaxXBlockEntries
	binary.LittleEndian.PutUint32(data[4:8], totalSize)         // lcbTotal

	// Write BID array
	offset_pos := 8
	for _, refBID := range bids {
		if format == FormatUnicode {
			binary.LittleEndian.PutUint64(data[offset_pos:offset_pos+8], refBID)
		} else {
			binary.LittleEndian.PutUint32(data[offset_pos:offset_pos+4], uint32(refBID)) //nolint:gosec // G115: ANSI format BID is 32-bit
		}
		offset_pos += bidSize
	}

	// Build complete block (extended blocks are internal, so no encryption)
	return BuildBlock(data, bid|0x2, offset, format, CryptMethodNone)
}

// BuildSubnodeLeafBlock creates a subnode leaf block (SLBLOCK).
// Subnode blocks contain a B-tree of subnodes for a node.
func BuildSubnodeLeafBlock(entries []SubnodeLeafEntry, bid uint64, offset uint64, format PSTFormat) ([]byte, error) {
	maxEntries := MaxSubnodeLeafEntriesUnicode
	if format == FormatANSI {
		maxEntries = MaxSubnodeLeafEntriesANSI
	}
	if len(entries) > maxEntries {
		return nil, fmt.Errorf("too many subnode entries: %d (max %d)", len(entries), maxEntries)
	}

	// Calculate entry size
	entrySize := 24 // Unicode
	if format == FormatANSI {
		entrySize = 12
	}

	// Calculate data size (header + entries)
	headerSize := 8 // Unicode
	if format == FormatANSI {
		headerSize = 4
	}
	dataSize := headerSize + len(entries)*entrySize

	// Build data
	data := make([]byte, dataSize)
	data[0] = byte(BlockTypeSubnode)                               // btype = 0x02
	data[1] = 0                                                    // cLevel = 0 (leaf)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(entries))) //nolint:gosec // G115: entry count bounded by block capacity

	// Write entries
	pos := headerSize
	for _, entry := range entries {
		if format == FormatUnicode {
			binary.LittleEndian.PutUint64(data[pos:pos+8], entry.NID)
			binary.LittleEndian.PutUint64(data[pos+8:pos+16], entry.DataBID)
			binary.LittleEndian.PutUint64(data[pos+16:pos+24], entry.SubBID)
		} else {
			// ANSI format uses 32-bit values per MS-PST spec
			binary.LittleEndian.PutUint32(data[pos:pos+4], uint32(entry.NID))       //nolint:gosec // G115
			binary.LittleEndian.PutUint32(data[pos+4:pos+8], uint32(entry.DataBID)) //nolint:gosec // G115
			binary.LittleEndian.PutUint32(data[pos+8:pos+12], uint32(entry.SubBID)) //nolint:gosec // G115
		}
		pos += entrySize
	}

	// Build complete block (subnode blocks are internal)
	return BuildBlock(data, bid|0x2, offset, format, CryptMethodNone)
}

// BuildSubnodeNonleafBlock creates a subnode non-leaf block (SIBLOCK).
func BuildSubnodeNonleafBlock(entries []SubnodeNonleafEntry, level byte, bid uint64, offset uint64, format PSTFormat) ([]byte, error) {
	// Calculate entry size
	entrySize := 16 // Unicode: key(8) + bid(8)
	if format == FormatANSI {
		entrySize = 8 // ANSI: key(4) + bid(4)
	}

	// Calculate data size
	headerSize := 8
	if format == FormatANSI {
		headerSize = 4
	}
	dataSize := headerSize + len(entries)*entrySize

	// Build data
	data := make([]byte, dataSize)
	data[0] = byte(BlockTypeSubnode)
	data[1] = level
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(entries))) //nolint:gosec // G115: entry count bounded by block capacity

	// Write entries
	pos := headerSize
	for _, entry := range entries {
		if format == FormatUnicode {
			binary.LittleEndian.PutUint64(data[pos:pos+8], entry.Key)
			binary.LittleEndian.PutUint64(data[pos+8:pos+16], entry.SubBID)
		} else {
			// ANSI format uses 32-bit values per MS-PST spec
			binary.LittleEndian.PutUint32(data[pos:pos+4], uint32(entry.Key))      //nolint:gosec // G115
			binary.LittleEndian.PutUint32(data[pos+4:pos+8], uint32(entry.SubBID)) //nolint:gosec // G115
		}
		pos += entrySize
	}

	return BuildBlock(data, bid|0x2, offset, format, CryptMethodNone)
}

// CalculateBlockDiskSize returns the total disk size for a block with given data size.
func CalculateBlockDiskSize(dataSize uint64, format PSTFormat) uint64 {
	alignedSize := AlignDisk(dataSize)
	if format == FormatUnicode {
		return alignedSize + BlockTrailerSizeUnicode
	}
	return alignedSize + BlockTrailerSizeANSI
}

// ValidateBlockTrailer validates a block trailer against computed values.
func ValidateBlockTrailer(data []byte, trailer *BlockTrailer, bid uint64, offset uint64) error {
	// Compute expected CRC
	alignedSize := AlignDisk(uint64(trailer.Size))
	expectedCRC := ComputeCRC(data[:alignedSize])
	if trailer.CRC != expectedCRC {
		return fmt.Errorf("CRC mismatch: got 0x%08X, expected 0x%08X", trailer.CRC, expectedCRC)
	}

	// Compute expected signature
	expectedSig := ComputeSignature(bid, offset)
	if trailer.Signature != expectedSig {
		return fmt.Errorf("signature mismatch: got 0x%04X, expected 0x%04X", trailer.Signature, expectedSig)
	}

	// Verify BID matches
	if trailer.BID != bid {
		return fmt.Errorf("BID mismatch: got 0x%X, expected 0x%X", trailer.BID, bid)
	}

	return nil
}
