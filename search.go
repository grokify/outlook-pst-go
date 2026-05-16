package outlookpst

import (
	"fmt"
	"sync"

	"github.com/grokify/outlook-pst-go/pkg/ltp"
	"github.com/grokify/outlook-pst-go/pkg/ndb"
	"github.com/grokify/outlook-pst-go/pkg/util"
)

// SearchFolder represents a search folder in a PST file.
// Search folders are virtual folders whose contents are determined by search criteria.
// See [MS-PST] Section 2.4.8.
type SearchFolder struct {
	*Folder
	criteriaOnce sync.Once
	criteria     *SearchCriteria
	criteriaErr  error
}

// newSearchFolder creates a SearchFolder from a Folder.
func newSearchFolder(f *Folder) *SearchFolder {
	return &SearchFolder{Folder: f}
}

// SearchCriteria returns the search criteria for this search folder.
// See [MS-PST] Section 2.4.8.6.2.
func (sf *SearchFolder) SearchCriteria() (*SearchCriteria, error) {
	sf.criteriaOnce.Do(func() {
		// Search criteria is stored in a subnode
		criteriaNode, err := sf.node.LookupSubnode(util.MakeNID(util.NIDTypeSearchCriteriaObj, sf.node.ID().Index()))
		if err != nil {
			sf.criteriaErr = fmt.Errorf("failed to get search criteria node: %w", err)
			return
		}

		sf.criteria, err = newSearchCriteria(criteriaNode)
		if err != nil {
			sf.criteriaErr = fmt.Errorf("failed to parse search criteria: %w", err)
			return
		}
	})

	return sf.criteria, sf.criteriaErr
}

// SearchCriteria represents the search criteria for a search folder.
// See [MS-PST] Section 2.4.8.6.2.
type SearchCriteria struct {
	bag *ltp.PropertyBag

	// Restriction is the raw restriction data defining the search filter.
	Restriction []byte

	// FolderEntryIDs contains the entry IDs of folders to search.
	FolderEntryIDs [][]byte

	// SearchFlags contains flags controlling search behavior.
	SearchFlags uint32
}

// Search criteria property IDs.
const (
	PidTagSearchFolderRestriction   ltp.PropID = 0x6842 // Search restriction
	PidTagSearchFolderEntryIdList   ltp.PropID = 0x6843 // List of folder entry IDs
	PidTagSearchFolderFlags         ltp.PropID = 0x6848 // Search flags
	PidTagSearchFolderLastUsed      ltp.PropID = 0x6834 // Last used time
	PidTagSearchFolderTemplateId    ltp.PropID = 0x6841 // Template ID
)

// Search folder flags.
const (
	SearchFlagForeground   uint32 = 0x00000001 // Foreground search
	SearchFlagRecursive    uint32 = 0x00000002 // Search subfolders recursively
	SearchFlagContentIndex uint32 = 0x00000004 // Use content indexing
	SearchFlagStatic       uint32 = 0x00000008 // Static search (no updates)
	SearchFlagMaybeStatic  uint32 = 0x00000010 // Search might be static
)

// newSearchCriteria creates SearchCriteria from a node.
func newSearchCriteria(node *ndb.Node) (*SearchCriteria, error) {
	bag, err := ltp.NewPropertyBag(node)
	if err != nil {
		return nil, fmt.Errorf("failed to create property bag: %w", err)
	}

	sc := &SearchCriteria{bag: bag}

	// Read restriction
	sc.Restriction, _ = bag.GetBinary(PidTagSearchFolderRestriction)

	// Read folder entry ID list
	entryIDData, _ := bag.GetBinary(PidTagSearchFolderEntryIdList)
	if len(entryIDData) > 0 {
		sc.FolderEntryIDs = parseEntryIDList(entryIDData)
	}

	// Read flags
	flags, err := bag.GetInt32(PidTagSearchFolderFlags)
	if err == nil {
		sc.SearchFlags = uint32(flags)
	}

	return sc, nil
}

// parseEntryIDList parses a list of entry IDs from binary data.
func parseEntryIDList(data []byte) [][]byte {
	var entryIDs [][]byte

	// Entry ID list format: count (4 bytes) followed by entry IDs
	// Each entry ID is preceded by its length (4 bytes)
	if len(data) < 4 {
		return entryIDs
	}

	offset := 0
	for offset+4 <= len(data) {
		length := int(data[offset]) | int(data[offset+1])<<8 |
			int(data[offset+2])<<16 | int(data[offset+3])<<24
		offset += 4

		if length <= 0 || offset+length > len(data) {
			break
		}

		entryID := make([]byte, length)
		copy(entryID, data[offset:offset+length])
		entryIDs = append(entryIDs, entryID)
		offset += length
	}

	return entryIDs
}

// IsRecursive returns true if the search includes subfolders.
func (sc *SearchCriteria) IsRecursive() bool {
	return sc.SearchFlags&SearchFlagRecursive != 0
}

// IsForeground returns true if this is a foreground search.
func (sc *SearchCriteria) IsForeground() bool {
	return sc.SearchFlags&SearchFlagForeground != 0
}

// IsStatic returns true if this is a static search (not updated).
func (sc *SearchCriteria) IsStatic() bool {
	return sc.SearchFlags&SearchFlagStatic != 0
}

// UsesContentIndex returns true if the search uses content indexing.
func (sc *SearchCriteria) UsesContentIndex() bool {
	return sc.SearchFlags&SearchFlagContentIndex != 0
}

// PropertyBag returns the underlying property bag for advanced access.
func (sc *SearchCriteria) PropertyBag() *ltp.PropertyBag {
	return sc.bag
}

// SearchUpdateQueue represents the queue of pending search folder updates.
// See [MS-PST] Section 2.4.8.6.
type SearchUpdateQueue struct {
	pst     *PST
	node    *ndb.Node
	entries []SearchUpdateEntry
}

// SearchUpdateEntry represents an entry in the search update queue.
type SearchUpdateEntry struct {
	// FolderNID is the node ID of the folder containing the message.
	FolderNID util.NodeID
	// MessageNID is the node ID of the message to update.
	MessageNID util.NodeID
	// Flags contains update flags.
	Flags uint32
}

// SearchUpdateQueue returns the search update queue.
// See [MS-PST] Section 2.4.8.6.
func (p *PST) SearchUpdateQueue() (*SearchUpdateQueue, error) {
	node, err := p.db.GetNode(util.NIDSearchUpdateQueue)
	if err != nil {
		return nil, fmt.Errorf("failed to get search update queue node: %w", err)
	}

	return newSearchUpdateQueue(p, node)
}

// newSearchUpdateQueue creates a SearchUpdateQueue from a node.
func newSearchUpdateQueue(pst *PST, node *ndb.Node) (*SearchUpdateQueue, error) {
	suq := &SearchUpdateQueue{
		pst:  pst,
		node: node,
	}

	// The search update queue is stored as a table
	table, err := ltp.NewTable(node)
	if err != nil {
		// No entries or invalid table
		return suq, nil
	}

	// Parse entries from the table
	for row, err := range table.Rows() {
		if err != nil {
			break
		}

		entry := SearchUpdateEntry{}

		// Read folder NID
		if folderData, err := row.GetRaw(ltp.PidTagLtpRowId); err == nil && len(folderData) >= 4 {
			entry.FolderNID = util.NodeID(folderData[0]) | util.NodeID(folderData[1])<<8 |
				util.NodeID(folderData[2])<<16 | util.NodeID(folderData[3])<<24
		}

		suq.entries = append(suq.entries, entry)
	}

	return suq, nil
}

// Entries returns all entries in the search update queue.
func (suq *SearchUpdateQueue) Entries() []SearchUpdateEntry {
	return suq.entries
}

// Count returns the number of entries in the queue.
func (suq *SearchUpdateQueue) Count() int {
	return len(suq.entries)
}

// IsEmpty returns true if the queue is empty.
func (suq *SearchUpdateQueue) IsEmpty() bool {
	return len(suq.entries) == 0
}
