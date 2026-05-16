package disk

import (
	"encoding/binary"
	"fmt"
)

// PageTrailer contains the trailer at the end of each page.
// See [MS-PST] Section 2.2.2.7.1 - PAGETRAILER structure.
type PageTrailer struct {
	PageType       PageType // ptype - Type of page
	PageTypeRepeat PageType // ptypeRepeat - Validation copy of page type
	Signature      uint16   // wSig - Computed signature
	CRC            uint32   // dwCRC - CRC32 of page data (4 bytes in Unicode, computed per [MS-PST] 5.3)
	BID            uint64   // bid - Block ID of the page
}

// ParsePageTrailerUnicode parses a Unicode format page trailer.
func ParsePageTrailerUnicode(data []byte) PageTrailer {
	// Unicode trailer (16 bytes):
	// ptype: 0 (1 byte)
	// ptypeRepeat: 1 (1 byte)
	// wSig: 2-4 (2 bytes)
	// dwCRC: 4-8 (4 bytes)
	// bid: 8-16 (8 bytes)
	return PageTrailer{
		PageType:       PageType(data[0]),
		PageTypeRepeat: PageType(data[1]),
		Signature:      binary.LittleEndian.Uint16(data[2:4]),
		CRC:            binary.LittleEndian.Uint32(data[4:8]),
		BID:            binary.LittleEndian.Uint64(data[8:16]),
	}
}

// ParsePageTrailerANSI parses an ANSI format page trailer.
func ParsePageTrailerANSI(data []byte) PageTrailer {
	// ANSI trailer (12 bytes):
	// ptype: 0 (1 byte)
	// ptypeRepeat: 1 (1 byte)
	// wSig: 2-4 (2 bytes)
	// bid: 4-8 (4 bytes)
	// dwCRC: 8-12 (4 bytes)
	return PageTrailer{
		PageType:       PageType(data[0]),
		PageTypeRepeat: PageType(data[1]),
		Signature:      binary.LittleEndian.Uint16(data[2:4]),
		BID:            uint64(binary.LittleEndian.Uint32(data[4:8])),
		CRC:            binary.LittleEndian.Uint32(data[8:12]),
	}
}

// BTPageHeader contains the header for a B-tree page.
// See [MS-PST] Section 2.2.2.7.7.1 - BTPAGE structure (first 4 bytes after entries).
type BTPageHeader struct {
	NumEntries    byte // cEnt - Number of entries
	NumEntriesMax byte // cEntMax - Maximum entries
	EntrySize     byte // cbEnt - Size of each entry in bytes
	Level         byte // cLevel - Tree level (0 = leaf)
}

// NBTLeafEntry represents an entry in the Node B-tree leaf page.
// See [MS-PST] Section 2.2.2.7.7.4 - NBTENTRY structure.
type NBTLeafEntry struct {
	NID       uint64 // nid - Node ID
	DataBID   uint64 // bidData - Data block ID
	SubBID    uint64 // bidSub - Subnode block ID
	ParentNID uint32 // nidParent - Parent node ID (only lower 32 bits used)
}

// BBTLeafEntry represents an entry in the Block B-tree leaf page.
// See [MS-PST] Section 2.2.2.7.7.3 - BBTENTRY structure.
type BBTLeafEntry struct {
	BRef     BlockReference // BREF - Block reference (bid + ib)
	Size     uint16         // cb - Unaligned data size
	RefCount uint16         // cRef - Reference count
}

// BTNonleafEntry represents an entry in a B-tree non-leaf (internal) page.
// See [MS-PST] Section 2.2.2.7.7.2 - BTENTRY structure.
type BTNonleafEntry struct {
	Key uint64         // btkey - Key value (max key in subtree)
	Ref BlockReference // BREF - Reference to child page
}

// BTPage represents a parsed B-tree page.
type BTPage struct {
	Trailer    PageTrailer
	Header     BTPageHeader
	Level      int

	// For leaf pages
	NBTEntries []NBTLeafEntry
	BBTEntries []BBTLeafEntry

	// For non-leaf pages
	NonleafEntries []BTNonleafEntry
}

// ParseBTPage parses a B-tree page from raw data.
func ParseBTPage(data []byte, format PSTFormat, pageType PageType) (*BTPage, error) {
	if len(data) != PageSize {
		return nil, fmt.Errorf("invalid page size: got %d, expected %d", len(data), PageSize)
	}

	page := &BTPage{}

	// Parse trailer (at end of page)
	var trailerSize int
	if format == FormatUnicode {
		trailerSize = PageTrailerSizeUnicode
		page.Trailer = ParsePageTrailerUnicode(data[PageSize-trailerSize:])
	} else {
		trailerSize = PageTrailerSizeANSI
		page.Trailer = ParsePageTrailerANSI(data[PageSize-trailerSize:])
	}

	// Validate page type
	if page.Trailer.PageType != pageType {
		return nil, fmt.Errorf("unexpected page type: got %v, expected %v", page.Trailer.PageType, pageType)
	}
	if page.Trailer.PageType != page.Trailer.PageTypeRepeat {
		return nil, fmt.Errorf("page type mismatch: %v != %v", page.Trailer.PageType, page.Trailer.PageTypeRepeat)
	}

	// Parse header (just before trailer)
	// For Unicode BT pages, layout is: entries | header(4 bytes) | dwPadding(4 bytes) | trailer
	// For ANSI BT pages, layout is: entries | header(4 bytes) | trailer
	// See [MS-PST] Section 2.2.2.7.7.1 - BTPAGE structure
	var headerOffset int
	if format == FormatUnicode {
		// Unicode: header is before 4-byte padding and 16-byte trailer
		headerOffset = PageSize - trailerSize - 4 - 4
	} else {
		// ANSI: header is directly before trailer
		headerOffset = PageSize - trailerSize - 4
	}
	page.Header = BTPageHeader{
		NumEntries:    data[headerOffset],
		NumEntriesMax: data[headerOffset+1],
		EntrySize:     data[headerOffset+2],
		Level:         data[headerOffset+3],
	}
	page.Level = int(page.Header.Level)

	// Parse entries
	if page.Level == 0 {
		// Leaf page
		switch pageType {
		case PageTypeNBT:
			return parseNBTLeafPage(page, data, format)
		case PageTypeBBT:
			return parseBBTLeafPage(page, data, format)
		}
	} else {
		// Non-leaf page
		return parseBTNonleafPage(page, data, format)
	}

	return page, nil
}

// parseNBTLeafPage parses NBT leaf entries.
func parseNBTLeafPage(page *BTPage, data []byte, format PSTFormat) (*BTPage, error) {
	entrySize := int(page.Header.EntrySize)
	numEntries := int(page.Header.NumEntries)

	page.NBTEntries = make([]NBTLeafEntry, numEntries)

	for i := 0; i < numEntries; i++ {
		offset := i * entrySize
		if format == FormatUnicode {
			// Unicode entry (32 bytes):
			// nid: 0-8 (8 bytes)
			// bidData: 8-16 (8 bytes)
			// bidSub: 16-24 (8 bytes)
			// nidParent: 24-28 (4 bytes)
			// dwPadding: 28-32 (4 bytes)
			page.NBTEntries[i] = NBTLeafEntry{
				NID:       binary.LittleEndian.Uint64(data[offset : offset+8]),
				DataBID:   binary.LittleEndian.Uint64(data[offset+8 : offset+16]),
				SubBID:    binary.LittleEndian.Uint64(data[offset+16 : offset+24]),
				ParentNID: binary.LittleEndian.Uint32(data[offset+24 : offset+28]),
			}
		} else {
			// ANSI entry (16 bytes):
			// nid: 0-4 (4 bytes)
			// bidData: 4-8 (4 bytes)
			// bidSub: 8-12 (4 bytes)
			// nidParent: 12-16 (4 bytes)
			page.NBTEntries[i] = NBTLeafEntry{
				NID:       uint64(binary.LittleEndian.Uint32(data[offset : offset+4])),
				DataBID:   uint64(binary.LittleEndian.Uint32(data[offset+4 : offset+8])),
				SubBID:    uint64(binary.LittleEndian.Uint32(data[offset+8 : offset+12])),
				ParentNID: binary.LittleEndian.Uint32(data[offset+12 : offset+16]),
			}
		}
	}

	return page, nil
}

// parseBBTLeafPage parses BBT leaf entries.
func parseBBTLeafPage(page *BTPage, data []byte, format PSTFormat) (*BTPage, error) {
	entrySize := int(page.Header.EntrySize)
	numEntries := int(page.Header.NumEntries)

	page.BBTEntries = make([]BBTLeafEntry, numEntries)

	for i := 0; i < numEntries; i++ {
		offset := i * entrySize
		if format == FormatUnicode {
			// Unicode entry (24 bytes):
			// bref.bid: 0-8 (8 bytes)
			// bref.ib: 8-16 (8 bytes)
			// cb: 16-18 (2 bytes)
			// cRef: 18-20 (2 bytes)
			// dwPadding: 20-24 (4 bytes)
			page.BBTEntries[i] = BBTLeafEntry{
				BRef: BlockReference{
					BID: binary.LittleEndian.Uint64(data[offset : offset+8]),
					IB:  binary.LittleEndian.Uint64(data[offset+8 : offset+16]),
				},
				Size:     binary.LittleEndian.Uint16(data[offset+16 : offset+18]),
				RefCount: binary.LittleEndian.Uint16(data[offset+18 : offset+20]),
			}
		} else {
			// ANSI entry (12 bytes):
			// bref.bid: 0-4 (4 bytes)
			// bref.ib: 4-8 (4 bytes)
			// cb: 8-10 (2 bytes)
			// cRef: 10-12 (2 bytes)
			page.BBTEntries[i] = BBTLeafEntry{
				BRef: BlockReference{
					BID: uint64(binary.LittleEndian.Uint32(data[offset : offset+4])),
					IB:  uint64(binary.LittleEndian.Uint32(data[offset+4 : offset+8])),
				},
				Size:     binary.LittleEndian.Uint16(data[offset+8 : offset+10]),
				RefCount: binary.LittleEndian.Uint16(data[offset+10 : offset+12]),
			}
		}
	}

	return page, nil
}

// parseBTNonleafPage parses non-leaf B-tree entries.
func parseBTNonleafPage(page *BTPage, data []byte, format PSTFormat) (*BTPage, error) {
	entrySize := int(page.Header.EntrySize)
	numEntries := int(page.Header.NumEntries)

	page.NonleafEntries = make([]BTNonleafEntry, numEntries)

	for i := 0; i < numEntries; i++ {
		offset := i * entrySize
		if format == FormatUnicode {
			// Unicode non-leaf entry (24 bytes):
			// key: 0-8 (8 bytes)
			// bref.bid: 8-16 (8 bytes)
			// bref.ib: 16-24 (8 bytes)
			page.NonleafEntries[i] = BTNonleafEntry{
				Key: binary.LittleEndian.Uint64(data[offset : offset+8]),
				Ref: BlockReference{
					BID: binary.LittleEndian.Uint64(data[offset+8 : offset+16]),
					IB:  binary.LittleEndian.Uint64(data[offset+16 : offset+24]),
				},
			}
		} else {
			// ANSI non-leaf entry (12 bytes):
			// key: 0-4 (4 bytes)
			// bref.bid: 4-8 (4 bytes)
			// bref.ib: 8-12 (4 bytes)
			page.NonleafEntries[i] = BTNonleafEntry{
				Key: uint64(binary.LittleEndian.Uint32(data[offset : offset+4])),
				Ref: BlockReference{
					BID: uint64(binary.LittleEndian.Uint32(data[offset+4 : offset+8])),
					IB:  uint64(binary.LittleEndian.Uint32(data[offset+8 : offset+12])),
				},
			}
		}
	}

	return page, nil
}

// IsLeaf returns true if this is a leaf page.
func (p *BTPage) IsLeaf() bool {
	return p.Level == 0
}
