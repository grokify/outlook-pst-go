package ltp

import (
	"encoding/binary"
	"fmt"

	"github.com/grokify/outlook-pst-go/pkg/ndb"
	"github.com/grokify/outlook-pst-go/pkg/util"
)

// NamedPropertyKind represents the type of named property identifier.
// See [MS-PST] Section 2.4.7 and [MS-OXPROPS].
type NamedPropertyKind uint8

const (
	// MNID_ID indicates the property is identified by a numeric ID.
	MNID_ID NamedPropertyKind = 0x00
	// MNID_STRING indicates the property is identified by a string name.
	MNID_STRING NamedPropertyKind = 0x01
)

// String returns a string representation of the kind.
func (k NamedPropertyKind) String() string {
	switch k {
	case MNID_ID:
		return "MNID_ID"
	case MNID_STRING:
		return "MNID_STRING"
	default:
		return fmt.Sprintf("Unknown(0x%02X)", uint8(k))
	}
}

// NamedProperty represents a named MAPI property.
// Named properties allow applications to define custom properties
// beyond the standard MAPI property set.
// See [MS-PST] Section 2.4.7.
type NamedProperty struct {
	// GUID is the property set GUID.
	GUID util.GUID
	// Kind indicates whether the property is identified by ID or string.
	Kind NamedPropertyKind
	// ID is the numeric identifier (when Kind == MNID_ID).
	ID uint32
	// Name is the string identifier (when Kind == MNID_STRING).
	Name string
	// PropID is the mapped property ID in the range 0x8000-0xFFFE.
	PropID PropID
}

// String returns a string representation of the named property.
func (np *NamedProperty) String() string {
	if np.Kind == MNID_ID {
		return fmt.Sprintf("{%s}:0x%04X -> 0x%04X", np.GUID.String(), np.ID, np.PropID)
	}
	return fmt.Sprintf("{%s}:%q -> 0x%04X", np.GUID.String(), np.Name, np.PropID)
}

// NamedPropertyMap maps named properties to property IDs.
// See [MS-PST] Section 2.4.7.
type NamedPropertyMap struct {
	// entries maps property IDs (0x8000+) to named properties
	entries map[PropID]*NamedProperty
	// byGUIDAndID maps (GUID, numeric ID) to named properties
	byGUIDAndID map[util.GUID]map[uint32]*NamedProperty
	// byGUIDAndName maps (GUID, string name) to named properties
	byGUIDAndName map[util.GUID]map[string]*NamedProperty
}

// NewNamedPropertyMap creates a NamedPropertyMap from the name-to-ID map node.
// The node should be NID_NAME_TO_ID_MAP (0x61).
// See [MS-PST] Section 2.4.7.
func NewNamedPropertyMap(node *ndb.Node) (*NamedPropertyMap, error) {
	npm := &NamedPropertyMap{
		entries:       make(map[PropID]*NamedProperty),
		byGUIDAndID:   make(map[util.GUID]map[uint32]*NamedProperty),
		byGUIDAndName: make(map[util.GUID]map[string]*NamedProperty),
	}

	// Create property bag for the node
	pb, err := NewPropertyBag(node)
	if err != nil {
		return nil, fmt.Errorf("failed to create property bag: %w", err)
	}

	// Read the GUID stream (property 0x0002)
	// This contains all the GUIDs used by named properties
	guidData, err := pb.GetBinary(0x0002)
	if err != nil {
		// No named properties
		return npm, nil
	}

	guids, err := parseGUIDStream(guidData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GUID stream: %w", err)
	}

	// Read the entry stream (property 0x0003)
	// This contains NAMEID structures
	entryData, err := pb.GetBinary(0x0003)
	if err != nil {
		return npm, nil
	}

	// Read the string stream (property 0x0004)
	// This contains string names for MNID_STRING properties
	stringData, _ := pb.GetBinary(0x0004)

	// Parse entries
	err = npm.parseEntries(entryData, guids, stringData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse named property entries: %w", err)
	}

	return npm, nil
}

// parseGUIDStream parses the GUID stream.
// The first 3 GUIDs are always PS_MAPI, PS_PUBLIC_STRINGS, and PS_NONE.
func parseGUIDStream(data []byte) ([]util.GUID, error) {
	// Each GUID is 16 bytes
	if len(data)%16 != 0 {
		return nil, fmt.Errorf("invalid GUID stream length: %d", len(data))
	}

	count := len(data) / 16
	guids := make([]util.GUID, count)

	for i := 0; i < count; i++ {
		guids[i] = util.GUIDFromBytes(data[i*16:])
	}

	return guids, nil
}

// parseEntries parses the NAMEID entry stream.
func (npm *NamedPropertyMap) parseEntries(entryData []byte, guids []util.GUID, stringData []byte) error {
	// Each NAMEID entry is 8 bytes:
	// dwPropertyID (4 bytes) - numeric ID or string offset
	// wGuid (2 bytes) - GUID index
	// wPropIdx (2 bytes) - property index (maps to 0x8000 + wPropIdx)

	if len(entryData)%8 != 0 {
		return fmt.Errorf("invalid entry stream length: %d", len(entryData))
	}

	count := len(entryData) / 8

	for i := 0; i < count; i++ {
		offset := i * 8
		entry := entryData[offset : offset+8]

		dwPropertyID := binary.LittleEndian.Uint32(entry[0:4])
		wGuid := binary.LittleEndian.Uint16(entry[4:6])
		wPropIdx := binary.LittleEndian.Uint16(entry[6:8])

		// Determine the kind based on wGuid bit 0
		kind := NamedPropertyKind(wGuid & 0x01)
		guidIndex := int(wGuid >> 1)

		// The first index (0) refers to PS_MAPI
		// Index 1 refers to PS_PUBLIC_STRINGS
		// Index 2+ refers to guids[index-2]
		var guid util.GUID
		switch guidIndex {
		case 0:
			guid = util.PS_MAPI
		case 1:
			guid = util.PS_PUBLIC_STRINGS
		default:
			if guidIndex-2 >= len(guids) {
				continue // Skip invalid entries
			}
			guid = guids[guidIndex-2]
		}

		np := &NamedProperty{
			GUID:   guid,
			Kind:   kind,
			PropID: PropID(0x8000 + wPropIdx),
		}

		if kind == MNID_ID {
			np.ID = dwPropertyID
		} else {
			// dwPropertyID is an offset into the string stream
			if stringData != nil && int(dwPropertyID) < len(stringData) {
				np.Name = readStringFromStream(stringData, int(dwPropertyID))
			}
		}

		npm.addEntry(np)
	}

	return nil
}

// readStringFromStream reads a null-terminated Unicode string from the string stream.
func readStringFromStream(data []byte, offset int) string {
	if offset >= len(data) {
		return ""
	}

	// First 4 bytes at offset is the string length in bytes
	if offset+4 > len(data) {
		return ""
	}
	strLen := binary.LittleEndian.Uint32(data[offset:])

	start := offset + 4
	end := start + int(strLen)
	if end > len(data) {
		end = len(data)
	}

	// UTF-16LE string
	strData := data[start:end]
	return decodeUTF16LE(strData)
}

// addEntry adds a named property to the map.
func (npm *NamedPropertyMap) addEntry(np *NamedProperty) {
	npm.entries[np.PropID] = np

	if np.Kind == MNID_ID {
		if npm.byGUIDAndID[np.GUID] == nil {
			npm.byGUIDAndID[np.GUID] = make(map[uint32]*NamedProperty)
		}
		npm.byGUIDAndID[np.GUID][np.ID] = np
	} else {
		if npm.byGUIDAndName[np.GUID] == nil {
			npm.byGUIDAndName[np.GUID] = make(map[string]*NamedProperty)
		}
		npm.byGUIDAndName[np.GUID][np.Name] = np
	}
}

// Lookup looks up a named property by GUID and numeric ID.
func (npm *NamedPropertyMap) Lookup(guid util.GUID, id uint32) (*NamedProperty, bool) {
	if guidMap, ok := npm.byGUIDAndID[guid]; ok {
		if np, ok := guidMap[id]; ok {
			return np, true
		}
	}
	return nil, false
}

// LookupByName looks up a named property by GUID and string name.
func (npm *NamedPropertyMap) LookupByName(guid util.GUID, name string) (*NamedProperty, bool) {
	if guidMap, ok := npm.byGUIDAndName[guid]; ok {
		if np, ok := guidMap[name]; ok {
			return np, true
		}
	}
	return nil, false
}

// LookupByPropID looks up a named property by its mapped property ID.
func (npm *NamedPropertyMap) LookupByPropID(propID PropID) (*NamedProperty, bool) {
	np, ok := npm.entries[propID]
	return np, ok
}

// GetPropID returns the mapped property ID for a named property.
func (npm *NamedPropertyMap) GetPropID(guid util.GUID, id uint32) (PropID, bool) {
	np, ok := npm.Lookup(guid, id)
	if !ok {
		return 0, false
	}
	return np.PropID, true
}

// GetPropIDByName returns the mapped property ID for a named property by name.
func (npm *NamedPropertyMap) GetPropIDByName(guid util.GUID, name string) (PropID, bool) {
	np, ok := npm.LookupByName(guid, name)
	if !ok {
		return 0, false
	}
	return np.PropID, true
}

// Entries returns all named property entries.
func (npm *NamedPropertyMap) Entries() []*NamedProperty {
	entries := make([]*NamedProperty, 0, len(npm.entries))
	for _, np := range npm.entries {
		entries = append(entries, np)
	}
	return entries
}

// Count returns the number of named properties.
func (npm *NamedPropertyMap) Count() int {
	return len(npm.entries)
}
