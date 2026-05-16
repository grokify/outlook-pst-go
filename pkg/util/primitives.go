// Package util provides primitive types and utilities for the PST SDK.
// These types correspond to fundamental identifiers defined in the MS-PST specification.
//
// [MS-PST]: https://docs.microsoft.com/en-us/openspecs/office_file_formats/ms-pst/
package util

// NodeID represents a node identifier in the PST file.
// See [MS-PST] Section 2.2.2.1 - NID structure.
// Lower 5 bits contain the node type (nidType), upper 27 bits contain the index (nidIndex).
type NodeID uint32

// BlockID represents a block identifier in the PST file.
// See [MS-PST] Section 2.2.2.2 - BID structure.
// The low 2 bits contain flags (A=attached, I=internal), remaining bits are the counter.
type BlockID uint64

// PageID represents a page identifier (used internally).
type PageID uint64

// HeapID represents a heap allocation identifier.
// See [MS-PST] Section 2.3.1.1 - HID structure.
// Bits 0-4: reserved (0), Bits 5-15: block index (hidBlockIndex), Bits 16-31: allocation index (hidIndex)
type HeapID uint32

// HeapNodeID represents a heap or subnode identifier.
// If high bit is set, it's a node ID (subnode); otherwise it's a heap ID.
type HeapNodeID uint32

// PropID represents a property identifier.
type PropID uint16

// RowID represents a table row identifier.
type RowID uint32

// PropType represents a MAPI property type.
type PropType uint16

// NIDType represents the type portion of a node ID (nidType).
// See [MS-PST] Section 2.2.2.1 - nidType field.
type NIDType uint8

// Node type constants from [MS-PST] Section 2.4.1.
const (
	NIDTypeHID                 NIDType = 0x00 // NID_TYPE_HID - Heap node
	NIDTypeInternal            NIDType = 0x01 // NID_TYPE_INTERNAL - Internal node
	NIDTypeNormalFolder        NIDType = 0x02 // NID_TYPE_NORMAL_FOLDER - Normal folder. See [MS-PST] Section 2.4.4.
	NIDTypeSearchFolder        NIDType = 0x03 // NID_TYPE_SEARCH_FOLDER - Search folder. See [MS-PST] Section 2.4.8.
	NIDTypeNormalMessage       NIDType = 0x04 // NID_TYPE_NORMAL_MESSAGE - Normal message. See [MS-PST] Section 2.4.5.
	NIDTypeAttachment          NIDType = 0x05 // NID_TYPE_ATTACHMENT - Attachment. See [MS-PST] Section 2.4.6.
	NIDTypeSearchUpdateQueue   NIDType = 0x06 // NID_TYPE_SEARCH_UPDATE_QUEUE - Search update queue
	NIDTypeSearchCriteriaObj   NIDType = 0x07 // NID_TYPE_SEARCH_CRITERIA_OBJECT - Search criteria object
	NIDTypeAssocMessage        NIDType = 0x08 // NID_TYPE_ASSOC_MESSAGE - Associated (FAI) message
	NIDTypeContentsTableIdx    NIDType = 0x0A // NID_TYPE_CONTENTS_TABLE_INDEX - Contents table index
	NIDTypeReceiveFolderTbl    NIDType = 0x0B // NID_TYPE_RECEIVE_FOLDER_TABLE - Receive folder table
	NIDTypeOutgoingQueueTbl    NIDType = 0x0C // NID_TYPE_OUTGOING_QUEUE_TABLE - Outgoing queue table
	NIDTypeHierarchyTable      NIDType = 0x0D // NID_TYPE_HIERARCHY_TABLE - Hierarchy table (subfolders). See [MS-PST] Section 2.4.4.4.
	NIDTypeContentsTable       NIDType = 0x0E // NID_TYPE_CONTENTS_TABLE - Contents table (messages). See [MS-PST] Section 2.4.4.5.
	NIDTypeAssocContentsTable  NIDType = 0x0F // NID_TYPE_ASSOC_CONTENTS_TABLE - Associated contents table (FAI). See [MS-PST] Section 2.4.4.6.
	NIDTypeSearchContentsTable NIDType = 0x10 // NID_TYPE_SEARCH_CONTENTS_TABLE - Search contents table
	NIDTypeAttachmentTable     NIDType = 0x11 // NID_TYPE_ATTACHMENT_TABLE - Attachment table. See [MS-PST] Section 2.4.6.1.
	NIDTypeRecipientTable      NIDType = 0x12 // NID_TYPE_RECIPIENT_TABLE - Recipient table. See [MS-PST] Section 2.4.5.3.
	NIDTypeSearchTableIndex    NIDType = 0x13 // NID_TYPE_SEARCH_TABLE_INDEX - Search table index
	NIDTypeLTP                 NIDType = 0x1F // NID_TYPE_LTP - LTP
)

// Node type mask for extracting type from NodeID.
const NIDTypeMask = 0x1F

// MakeNID creates a NodeID from a type and index.
func MakeNID(nidType NIDType, index uint32) NodeID {
	return NodeID(uint32(nidType&NIDTypeMask) | (index << 5))
}

// Type returns the node type from a NodeID.
func (nid NodeID) Type() NIDType {
	return NIDType(nid & NIDTypeMask)
}

// Index returns the index portion of a NodeID.
func (nid NodeID) Index() uint32 {
	return uint32(nid >> 5)
}

// Predefined (special) node IDs.
// See [MS-PST] Section 2.4.1 - Special Internal NIDs.
var (
	NIDMessageStore      = MakeNID(NIDTypeInternal, 0x1)             // NID_MESSAGE_STORE - Message store (NID 0x21). See [MS-PST] Section 2.4.3.
	NIDNameIDMap         = MakeNID(NIDTypeInternal, 0x3)             // NID_NAME_TO_ID_MAP - Named property map (NID 0x61). See [MS-PST] Section 2.4.7.
	NIDNameToIDMap       = NIDNameIDMap                              // Alias for named property map
	NIDRootFolder        = NodeID(0x122)                             // NID_ROOT_FOLDER - Root folder. See [MS-PST] Section 2.4.4.1.
	NIDSearchFolder      = NodeID(0x123)                             // Search folder root
	NIDSearchUpdateQueue = MakeNID(NIDTypeSearchUpdateQueue, 0x1)    // NID_SEARCH_MANAGEMENT_QUEUE (NID 0xC1). See [MS-PST] Section 2.4.8.6.
)

// BlockID flag constants.
const (
	BlockIDAttachedBit  BlockID = 0x1 // Block is attached (in memory)
	BlockIDInternalBit  BlockID = 0x2 // Block is internal (extended/subnode)
	BlockIDIncrement    BlockID = 0x4 // Counter increment per block
)

// IsInternal returns true if the block ID refers to an internal block.
func (bid BlockID) IsInternal() bool {
	return bid&BlockIDInternalBit != 0
}

// IsExternal returns true if the block ID refers to an external (data) block.
func (bid BlockID) IsExternal() bool {
	return bid&BlockIDInternalBit == 0
}

// HeapID extraction constants.
// See [MS-PST] Section 2.3.1.1 - HID structure:
// - Bits 0-4 (5 bits): hidType - reserved, must be 0
// - Bits 5-15 (11 bits): hidIndex - 1-based allocation index
// - Bits 16-31 (16 bits): hidBlockIndex - block index
const (
	heapIDBlockIndexShift = 16
	heapIDAllocIndexShift = 5
	heapIDAllocIndexMask  = 0x7FF // 11 bits
)

// BlockIndex returns the block index from a HeapID (hidBlockIndex).
func (hid HeapID) BlockIndex() uint16 {
	return uint16(hid >> heapIDBlockIndexShift)
}

// PageIndex returns the page/block index from a HeapID (alias for BlockIndex).
func (hid HeapID) PageIndex() uint16 {
	return hid.BlockIndex()
}

// AllocIndex returns the 0-based allocation index within the block.
// The hidIndex field is 1-based, so we subtract 1 to get a 0-based index.
func (hid HeapID) AllocIndex() uint16 {
	// Extract bits 5-15 (shift right by 5, mask 11 bits)
	index := uint16((hid >> heapIDAllocIndexShift) & heapIDAllocIndexMask)
	if index == 0 {
		return 0 // Invalid HID, return 0
	}
	return index - 1 // Convert from 1-based to 0-based
}

// MakeHeapID creates a HeapID from block index and 0-based allocation index.
// The allocIndex is converted to 1-based hidIndex for storage.
func MakeHeapID(blockIndex, allocIndex uint16) HeapID {
	hidIndex := allocIndex + 1 // Convert to 1-based
	return HeapID(uint32(blockIndex)<<heapIDBlockIndexShift | uint32(hidIndex)<<heapIDAllocIndexShift)
}

// IsHeapID returns true if the HeapNodeID is a heap allocation.
func (hnid HeapNodeID) IsHeapID() bool {
	// If high bit in lower word is clear, it's a heap ID
	return hnid&0x1F == 0 || hnid == 0
}

// IsSubnodeID returns true if the HeapNodeID refers to a subnode.
func (hnid HeapNodeID) IsSubnodeID() bool {
	// If it has a NID type, it's a subnode reference
	return hnid != 0 && (hnid&0x1F) != 0
}

// ToHeapID converts HeapNodeID to HeapID.
func (hnid HeapNodeID) ToHeapID() HeapID {
	return HeapID(hnid)
}

// ToNodeID converts HeapNodeID to NodeID (for subnode lookup).
func (hnid HeapNodeID) ToNodeID() NodeID {
	return NodeID(hnid)
}
