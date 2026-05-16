package ltp

import (
	"errors"
	"fmt"
)

// LTPErrorKind represents the kind of LTP layer error.
type LTPErrorKind int

// LTP error kinds.
// See [MS-PST] Section 2.3 for LTP layer structures.
const (
	ErrKindUnknown LTPErrorKind = iota

	// Heap errors (Section 2.3.1)
	ErrKindInvalidHeapSignature   // Invalid HN signature (bSig)
	ErrKindInvalidHeapClientSig   // Invalid client signature
	ErrKindHeapAllocationNotFound // HID not found in heap
	ErrKindHeapCorrupted          // Heap data is corrupted

	// BTH errors (Section 2.3.2)
	ErrKindInvalidBTHSignature // Invalid BTH signature (bType)
	ErrKindBTHKeyNotFound      // Key not found in BTH
	ErrKindInvalidBTHLevel     // Invalid BTH level
	ErrKindBTHCorrupted        // BTH structure is corrupted

	// Property Context errors (Section 2.3.3)
	ErrKindInvalidPCSignature // Invalid PC signature
	ErrKindPropertyNotFound   // Property not found
	ErrKindInvalidPropertyType
	ErrKindPropertyTypeMismatch // Property type doesn't match request

	// Table Context errors (Section 2.3.4)
	ErrKindInvalidTCSignature   // Invalid TC signature
	ErrKindInvalidRowIndex      // Row index out of bounds
	ErrKindInvalidColumnIndex   // Column index out of bounds
	ErrKindInvalidColumnType    // Invalid column type
	ErrKindMissingRowMatrix     // Row matrix not found
	ErrKindTableCorrupted       // Table structure is corrupted
)

// String returns the string representation of the error kind.
func (k LTPErrorKind) String() string {
	switch k {
	case ErrKindInvalidHeapSignature:
		return "invalid heap signature"
	case ErrKindInvalidHeapClientSig:
		return "invalid heap client signature"
	case ErrKindHeapAllocationNotFound:
		return "heap allocation not found"
	case ErrKindHeapCorrupted:
		return "heap corrupted"
	case ErrKindInvalidBTHSignature:
		return "invalid BTH signature"
	case ErrKindBTHKeyNotFound:
		return "BTH key not found"
	case ErrKindInvalidBTHLevel:
		return "invalid BTH level"
	case ErrKindBTHCorrupted:
		return "BTH corrupted"
	case ErrKindInvalidPCSignature:
		return "invalid PC signature"
	case ErrKindPropertyNotFound:
		return "property not found"
	case ErrKindInvalidPropertyType:
		return "invalid property type"
	case ErrKindPropertyTypeMismatch:
		return "property type mismatch"
	case ErrKindInvalidTCSignature:
		return "invalid TC signature"
	case ErrKindInvalidRowIndex:
		return "invalid row index"
	case ErrKindInvalidColumnIndex:
		return "invalid column index"
	case ErrKindInvalidColumnType:
		return "invalid column type"
	case ErrKindMissingRowMatrix:
		return "missing row matrix"
	case ErrKindTableCorrupted:
		return "table corrupted"
	default:
		return "unknown LTP error"
	}
}

// LTPError represents an error in the LTP layer.
// See [MS-PST] Section 2.3 for LTP layer structures.
type LTPError struct {
	Op       string       // Operation that failed
	Kind     LTPErrorKind // Kind of error
	PropID   PropID       // Related property ID (if applicable)
	PropType PropType     // Related property type (if applicable)
	HeapID   uint32       // Related heap ID (if applicable)
	RowIndex int          // Row index (if applicable)
	ColIndex int          // Column index (if applicable)
	Got      interface{}  // Actual value (for validation errors)
	Want     interface{}  // Expected value (for validation errors)
	Err      error        // Underlying error
}

// Error implements the error interface.
func (e *LTPError) Error() string {
	msg := fmt.Sprintf("ltp: %s: %s", e.Op, e.Kind.String())

	if e.PropID != 0 {
		msg += fmt.Sprintf(" (propID=0x%04X)", e.PropID)
	}
	if e.PropType != 0 {
		msg += fmt.Sprintf(" (propType=0x%04X)", e.PropType)
	}
	if e.HeapID != 0 {
		msg += fmt.Sprintf(" (heapID=0x%X)", e.HeapID)
	}
	if e.RowIndex >= 0 && e.Kind == ErrKindInvalidRowIndex {
		msg += fmt.Sprintf(" (row=%d)", e.RowIndex)
	}
	if e.ColIndex >= 0 && e.Kind == ErrKindInvalidColumnIndex {
		msg += fmt.Sprintf(" (col=%d)", e.ColIndex)
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
func (e *LTPError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches a target.
func (e *LTPError) Is(target error) bool {
	t, ok := target.(*LTPError)
	if !ok {
		return false
	}
	return e.Kind == t.Kind
}

// Sentinel errors for common cases.
var (
	ErrPropertyNotFound       = &LTPError{Kind: ErrKindPropertyNotFound}
	ErrInvalidHeapSignature   = &LTPError{Kind: ErrKindInvalidHeapSignature}
	ErrInvalidPCSignature     = &LTPError{Kind: ErrKindInvalidPCSignature}
	ErrInvalidTCSignature     = &LTPError{Kind: ErrKindInvalidTCSignature}
	ErrPropertyTypeMismatch   = &LTPError{Kind: ErrKindPropertyTypeMismatch}
)

// IsPropertyNotFound returns true if the error is a property not found error.
func IsPropertyNotFound(err error) bool {
	var ltpErr *LTPError
	if errors.As(err, &ltpErr) {
		return ltpErr.Kind == ErrKindPropertyNotFound
	}
	return false
}

// NewPropertyNotFoundError creates a new property not found error.
func NewPropertyNotFoundError(op string, propID PropID) *LTPError {
	return &LTPError{
		Op:     op,
		Kind:   ErrKindPropertyNotFound,
		PropID: propID,
	}
}

// NewPropertyTypeMismatchError creates a new property type mismatch error.
func NewPropertyTypeMismatchError(op string, propID PropID, got, want PropType) *LTPError {
	return &LTPError{
		Op:       op,
		Kind:     ErrKindPropertyTypeMismatch,
		PropID:   propID,
		PropType: got,
		Got:      got,
		Want:     want,
	}
}

// NewInvalidHeapSignatureError creates a new invalid heap signature error.
func NewInvalidHeapSignatureError(op string, got, want byte) *LTPError {
	return &LTPError{
		Op:   op,
		Kind: ErrKindInvalidHeapSignature,
		Got:  got,
		Want: want,
	}
}

// NewInvalidTCSignatureError creates a new invalid TC signature error.
func NewInvalidTCSignatureError(op string, got, want byte) *LTPError {
	return &LTPError{
		Op:   op,
		Kind: ErrKindInvalidTCSignature,
		Got:  got,
		Want: want,
	}
}

// NewRowIndexError creates a new invalid row index error.
func NewRowIndexError(op string, index, max int) *LTPError {
	return &LTPError{
		Op:       op,
		Kind:     ErrKindInvalidRowIndex,
		RowIndex: index,
		Got:      index,
		Want:     fmt.Sprintf("< %d", max),
	}
}
