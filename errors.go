// Package outlookpst provides a Go library for reading Microsoft Outlook PST files.
package outlookpst

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when a requested item is not found.
var ErrNotFound = errors.New("not found")

// ErrInvalidFormat is returned when the file format is invalid.
var ErrInvalidFormat = errors.New("invalid format")

// ErrCorrupted is returned when the file appears to be corrupted.
var ErrCorrupted = errors.New("file corrupted")

// MessagingErrorKind represents the kind of messaging layer error.
type MessagingErrorKind int

// Messaging error kinds.
const (
	ErrKindUnknown MessagingErrorKind = iota

	// Store errors
	ErrKindStoreNotFound
	ErrKindStoreRecordKeyNotFound
	ErrKindStoreDisplayNameNotFound

	// Folder errors
	ErrKindFolderNotFound
	ErrKindFolderContentCountNotFound
	ErrKindFolderUnreadCountNotFound
	ErrKindFolderDisplayNameNotFound
	ErrKindInvalidFolderType
	ErrKindIPMSubtreeNotFound
	ErrKindRootFolderNotFound

	// Message errors
	ErrKindMessageNotFound
	ErrKindMessageClassNotFound
	ErrKindMessageSizeNotFound
	ErrKindMessageFlagsNotFound
	ErrKindInvalidMessageClass

	// Attachment errors
	ErrKindAttachmentNotFound
	ErrKindAttachmentDataNotFound
	ErrKindInvalidAttachmentMethod

	// Recipient errors
	ErrKindRecipientNotFound
	ErrKindRecipientTypeNotFound
	ErrKindRecipientNameNotFound

	// Entry ID errors
	ErrKindInvalidEntryID
	ErrKindEntryIDMismatch
)

// String returns the string representation of the error kind.
func (k MessagingErrorKind) String() string {
	switch k {
	case ErrKindStoreNotFound:
		return "store not found"
	case ErrKindStoreRecordKeyNotFound:
		return "store record key not found"
	case ErrKindStoreDisplayNameNotFound:
		return "store display name not found"
	case ErrKindFolderNotFound:
		return "folder not found"
	case ErrKindFolderContentCountNotFound:
		return "folder content count not found"
	case ErrKindFolderUnreadCountNotFound:
		return "folder unread count not found"
	case ErrKindFolderDisplayNameNotFound:
		return "folder display name not found"
	case ErrKindInvalidFolderType:
		return "invalid folder type"
	case ErrKindIPMSubtreeNotFound:
		return "IPM subtree not found"
	case ErrKindRootFolderNotFound:
		return "root folder not found"
	case ErrKindMessageNotFound:
		return "message not found"
	case ErrKindMessageClassNotFound:
		return "message class not found"
	case ErrKindMessageSizeNotFound:
		return "message size not found"
	case ErrKindMessageFlagsNotFound:
		return "message flags not found"
	case ErrKindInvalidMessageClass:
		return "invalid message class"
	case ErrKindAttachmentNotFound:
		return "attachment not found"
	case ErrKindAttachmentDataNotFound:
		return "attachment data not found"
	case ErrKindInvalidAttachmentMethod:
		return "invalid attachment method"
	case ErrKindRecipientNotFound:
		return "recipient not found"
	case ErrKindRecipientTypeNotFound:
		return "recipient type not found"
	case ErrKindRecipientNameNotFound:
		return "recipient name not found"
	case ErrKindInvalidEntryID:
		return "invalid entry ID"
	case ErrKindEntryIDMismatch:
		return "entry ID mismatch"
	default:
		return "unknown messaging error"
	}
}

// MessagingError represents an error in the messaging layer.
type MessagingError struct {
	Op         string             // Operation that failed
	Kind       MessagingErrorKind // Kind of error
	FolderName string             // Folder name (if applicable)
	NodeID     uint64             // Related node ID (if applicable)
	PropID     uint16             // Related property ID (if applicable)
	Err        error              // Underlying error
}

// Error implements the error interface.
func (e *MessagingError) Error() string {
	msg := fmt.Sprintf("messaging: %s: %s", e.Op, e.Kind.String())

	if e.FolderName != "" {
		msg += fmt.Sprintf(" (folder=%q)", e.FolderName)
	}
	if e.NodeID != 0 {
		msg += fmt.Sprintf(" (node=0x%X)", e.NodeID)
	}
	if e.PropID != 0 {
		msg += fmt.Sprintf(" (prop=0x%04X)", e.PropID)
	}
	if e.Err != nil {
		msg += fmt.Sprintf(": %v", e.Err)
	}

	return msg
}

// Unwrap returns the underlying error.
func (e *MessagingError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches a target.
func (e *MessagingError) Is(target error) bool {
	t, ok := target.(*MessagingError)
	if !ok {
		// Check sentinel errors
		switch e.Kind {
		case ErrKindFolderNotFound, ErrKindMessageNotFound, ErrKindAttachmentNotFound, ErrKindRecipientNotFound:
			return target == ErrNotFound
		case ErrKindInvalidEntryID, ErrKindInvalidFolderType, ErrKindInvalidMessageClass:
			return target == ErrInvalidFormat
		}
		return false
	}
	return e.Kind == t.Kind
}

// NewFolderNotFoundError creates a new folder not found error.
func NewFolderNotFoundError(op, name string) *MessagingError {
	return &MessagingError{
		Op:         op,
		Kind:       ErrKindFolderNotFound,
		FolderName: name,
	}
}

// NewMessageNotFoundError creates a new message not found error.
func NewMessageNotFoundError(op string, nodeID uint64) *MessagingError {
	return &MessagingError{
		Op:     op,
		Kind:   ErrKindMessageNotFound,
		NodeID: nodeID,
	}
}

// NewAttachmentNotFoundError creates a new attachment not found error.
func NewAttachmentNotFoundError(op string, nodeID uint64) *MessagingError {
	return &MessagingError{
		Op:     op,
		Kind:   ErrKindAttachmentNotFound,
		NodeID: nodeID,
	}
}

// NodeNotFoundError is returned when a node is not found.
type NodeNotFoundError struct {
	NodeID uint64
}

func (e *NodeNotFoundError) Error() string {
	return fmt.Sprintf("node not found: 0x%X", e.NodeID)
}

func (e *NodeNotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// BlockNotFoundError is returned when a block is not found.
type BlockNotFoundError struct {
	BlockID uint64
}

func (e *BlockNotFoundError) Error() string {
	return fmt.Sprintf("block not found: 0x%X", e.BlockID)
}

func (e *BlockNotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// PropertyNotFoundError is returned when a property is not found.
type PropertyNotFoundError struct {
	PropID uint16
}

func (e *PropertyNotFoundError) Error() string {
	return fmt.Sprintf("property not found: 0x%04X", e.PropID)
}

func (e *PropertyNotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// FolderNotFoundError is returned when a folder is not found.
type FolderNotFoundError struct {
	Name string
}

func (e *FolderNotFoundError) Error() string {
	return fmt.Sprintf("folder not found: %s", e.Name)
}

func (e *FolderNotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// InvalidHeaderError is returned when the PST header is invalid.
type InvalidHeaderError struct {
	Message string
}

func (e *InvalidHeaderError) Error() string {
	return fmt.Sprintf("invalid header: %s", e.Message)
}

func (e *InvalidHeaderError) Is(target error) bool {
	return target == ErrInvalidFormat
}

// CRCError is returned when a CRC check fails.
type CRCError struct {
	Expected uint32
	Actual   uint32
}

func (e *CRCError) Error() string {
	return fmt.Sprintf("CRC mismatch: expected 0x%08X, got 0x%08X", e.Expected, e.Actual)
}

func (e *CRCError) Is(target error) bool {
	return target == ErrCorrupted
}

// SignatureError is returned when a signature check fails.
type SignatureError struct {
	Expected uint16
	Actual   uint16
}

func (e *SignatureError) Error() string {
	return fmt.Sprintf("signature mismatch: expected 0x%04X, got 0x%04X", e.Expected, e.Actual)
}

func (e *SignatureError) Is(target error) bool {
	return target == ErrCorrupted
}
