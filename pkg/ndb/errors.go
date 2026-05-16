package ndb

import (
	"errors"
	"fmt"
)

// NDBErrorKind represents the kind of NDB layer error.
type NDBErrorKind int

// NDB error kinds.
// See [MS-PST] Section 2.2 for NDB layer structures.
const (
	ErrKindUnknown NDBErrorKind = iota

	// Node errors
	ErrKindInvalidNodeIDType // Invalid NID_TYPE value
	ErrKindInvalidNodeIndex  // Invalid nidIndex value
	ErrKindNodeNotFound      // Node not found in NBT
	ErrKindInvalidNodeData   // Node data is corrupted

	// Block errors
	ErrKindBlockNotFound       // Block not found in BBT
	ErrKindInvalidBlockType    // Invalid block type
	ErrKindInvalidBlockSize    // Block size exceeds limits
	ErrKindInvalidBlockTrailer // Block trailer validation failed

	// Page errors
	ErrKindInvalidPageType    // Invalid ptype value
	ErrKindInvalidPageTrailer // Page trailer validation failed
	ErrKindInvalidPageLevel   // Invalid cLevel value

	// Header errors
	ErrKindInvalidMagic       // Invalid PST magic bytes
	ErrKindInvalidVersion     // Invalid NDB version
	ErrKindInvalidCryptMethod // Invalid encryption method
	ErrKindInvalidRootEntry   // Invalid ROOT structure

	// B-tree errors
	ErrKindBTreeCorrupted // B-tree structure is corrupted
	ErrKindBTreeKeyNotFound

	// Checksum errors
	ErrKindCRCMismatch       // CRC validation failed
	ErrKindSignatureMismatch // Block signature mismatch
)

// String returns the string representation of the error kind.
func (k NDBErrorKind) String() string {
	switch k {
	case ErrKindInvalidNodeIDType:
		return "invalid node ID type"
	case ErrKindInvalidNodeIndex:
		return "invalid node index"
	case ErrKindNodeNotFound:
		return "node not found"
	case ErrKindInvalidNodeData:
		return "invalid node data"
	case ErrKindBlockNotFound:
		return "block not found"
	case ErrKindInvalidBlockType:
		return "invalid block type"
	case ErrKindInvalidBlockSize:
		return "invalid block size"
	case ErrKindInvalidBlockTrailer:
		return "invalid block trailer"
	case ErrKindInvalidPageType:
		return "invalid page type"
	case ErrKindInvalidPageTrailer:
		return "invalid page trailer"
	case ErrKindInvalidPageLevel:
		return "invalid page level"
	case ErrKindInvalidMagic:
		return "invalid PST magic"
	case ErrKindInvalidVersion:
		return "invalid NDB version"
	case ErrKindInvalidCryptMethod:
		return "invalid crypt method"
	case ErrKindInvalidRootEntry:
		return "invalid ROOT entry"
	case ErrKindBTreeCorrupted:
		return "B-tree corrupted"
	case ErrKindBTreeKeyNotFound:
		return "B-tree key not found"
	case ErrKindCRCMismatch:
		return "CRC mismatch"
	case ErrKindSignatureMismatch:
		return "signature mismatch"
	default:
		return "unknown NDB error"
	}
}

// NDBError represents an error in the NDB layer.
// See [MS-PST] Section 2.2 for NDB layer structures.
type NDBError struct {
	Op      string       // Operation that failed
	Kind    NDBErrorKind // Kind of error
	NodeID  uint64       // Related node ID (if applicable)
	BlockID uint64       // Related block ID (if applicable)
	Offset  uint64       // File offset (if applicable)
	Got     interface{}  // Actual value (for validation errors)
	Want    interface{}  // Expected value (for validation errors)
	Err     error        // Underlying error
}

// Error implements the error interface.
func (e *NDBError) Error() string {
	msg := fmt.Sprintf("ndb: %s: %s", e.Op, e.Kind.String())

	if e.NodeID != 0 {
		msg += fmt.Sprintf(" (node=0x%X)", e.NodeID)
	}
	if e.BlockID != 0 {
		msg += fmt.Sprintf(" (block=0x%X)", e.BlockID)
	}
	if e.Offset != 0 {
		msg += fmt.Sprintf(" (offset=0x%X)", e.Offset)
	}
	if e.Got != nil && e.Want != nil {
		msg += fmt.Sprintf(" (got=%v, want=%v)", e.Got, e.Want)
	} else if e.Got != nil {
		msg += fmt.Sprintf(" (got=%v)", e.Got)
	}
	if e.Err != nil {
		msg += fmt.Sprintf(": %v", e.Err)
	}

	return msg
}

// Unwrap returns the underlying error.
func (e *NDBError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches a target.
func (e *NDBError) Is(target error) bool {
	t, ok := target.(*NDBError)
	if !ok {
		return false
	}
	return e.Kind == t.Kind
}

// Sentinel errors for common cases.
var (
	ErrNodeNotFound  = &NDBError{Kind: ErrKindNodeNotFound}
	ErrBlockNotFound = &NDBError{Kind: ErrKindBlockNotFound}
	ErrInvalidMagic  = &NDBError{Kind: ErrKindInvalidMagic}
	ErrCRCMismatch   = &NDBError{Kind: ErrKindCRCMismatch}
)

// IsNodeNotFound returns true if the error is a node not found error.
func IsNodeNotFound(err error) bool {
	var ndbErr *NDBError
	if errors.As(err, &ndbErr) {
		return ndbErr.Kind == ErrKindNodeNotFound
	}
	return false
}

// IsBlockNotFound returns true if the error is a block not found error.
func IsBlockNotFound(err error) bool {
	var ndbErr *NDBError
	if errors.As(err, &ndbErr) {
		return ndbErr.Kind == ErrKindBlockNotFound
	}
	return false
}

// NewNodeNotFoundError creates a new node not found error.
func NewNodeNotFoundError(op string, nodeID uint64) *NDBError {
	return &NDBError{
		Op:     op,
		Kind:   ErrKindNodeNotFound,
		NodeID: nodeID,
	}
}

// NewBlockNotFoundError creates a new block not found error.
func NewBlockNotFoundError(op string, blockID uint64) *NDBError {
	return &NDBError{
		Op:      op,
		Kind:    ErrKindBlockNotFound,
		BlockID: blockID,
	}
}

// NewInvalidPageError creates a new invalid page error.
func NewInvalidPageError(op string, offset uint64, got, want interface{}) *NDBError {
	return &NDBError{
		Op:     op,
		Kind:   ErrKindInvalidPageType,
		Offset: offset,
		Got:    got,
		Want:   want,
	}
}

// NewCRCError creates a new CRC mismatch error.
func NewCRCError(op string, offset uint64, got, want uint32) *NDBError {
	return &NDBError{
		Op:     op,
		Kind:   ErrKindCRCMismatch,
		Offset: offset,
		Got:    got,
		Want:   want,
	}
}
