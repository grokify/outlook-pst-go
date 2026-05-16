// Package ltp provides the Lists, Tables, and Properties layer for PST files.
// Property types and IDs defined here correspond to MAPI specifications.
// See [MS-OXCDATA] for property type definitions and [MS-OXPROPS] for property ID definitions.
//
// [MS-OXCDATA]: https://docs.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxcdata/
// [MS-OXPROPS]: https://docs.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxprops/
package ltp

import (
	"time"
)

// PropID represents a property identifier.
// See [MS-OXCDATA] Section 2.9 - Property Tags.
// Property IDs 0x0001-0x7FFF are defined by MAPI, 0x8000-0xFFFE are named properties.
type PropID uint16

// PropType represents a property type.
// See [MS-OXCDATA] Section 2.11.1 - Property Data Types.
type PropType uint16

// MAPI property types from [MS-OXCDATA] Section 2.11.1.
const (
	PropTypeUnspecified PropType = 0x0000 // PtypUnspecified
	PropTypeNull        PropType = 0x0001 // PtypNull
	PropTypeInt16       PropType = 0x0002 // PtypInteger16
	PropTypeInt32       PropType = 0x0003 // PtypInteger32
	PropTypeFloat32     PropType = 0x0004 // PtypFloating32
	PropTypeFloat64     PropType = 0x0005 // PtypFloating64
	PropTypeCurrency    PropType = 0x0006 // PtypCurrency - 64-bit signed integer (1/10000 units)
	PropTypeAppTime     PropType = 0x0007 // PtypFloatingTime - double (days since Dec 30, 1899)
	PropTypeError       PropType = 0x000A // PtypErrorCode
	PropTypeBool        PropType = 0x000B // PtypBoolean
	PropTypeObject      PropType = 0x000D // PtypObject - embedded object
	PropTypeInt64       PropType = 0x0014 // PtypInteger64
	PropTypeString8     PropType = 0x001E // PtypString8 - ANSI string
	PropTypeString      PropType = 0x001F // PtypString - Unicode string
	PropTypeSysTime     PropType = 0x0040 // PtypTime - FILETIME (100-ns since 1601-01-01)
	PropTypeGUID        PropType = 0x0048 // PtypGuid - 16-byte GUID
	PropTypeBinary      PropType = 0x0102 // PtypBinary

	// Multi-valued types (base type | 0x1000). See [MS-OXCDATA] Section 2.11.1.
	PropTypeMVInt16    PropType = 0x1002 // PtypMultipleInteger16
	PropTypeMVInt32    PropType = 0x1003 // PtypMultipleInteger32
	PropTypeMVFloat32  PropType = 0x1004 // PtypMultipleFloating32
	PropTypeMVFloat64  PropType = 0x1005 // PtypMultipleFloating64
	PropTypeMVCurrency PropType = 0x1006 // PtypMultipleCurrency
	PropTypeMVAppTime  PropType = 0x1007 // PtypMultipleFloatingTime
	PropTypeMVInt64    PropType = 0x1014 // PtypMultipleInteger64
	PropTypeMVString8  PropType = 0x101E // PtypMultipleString8
	PropTypeMVString   PropType = 0x101F // PtypMultipleString
	PropTypeMVSysTime  PropType = 0x1040 // PtypMultipleTime
	PropTypeMVGUID     PropType = 0x1048 // PtypMultipleGuid
	PropTypeMVBinary   PropType = 0x1102 // PtypMultipleBinary
)

// IsMultiValued returns true if the property type is multi-valued.
func (pt PropType) IsMultiValued() bool {
	return pt&0x1000 != 0
}

// BaseType returns the base type (without multi-value flag).
func (pt PropType) BaseType() PropType {
	return pt &^ 0x1000
}

// FixedSize returns the size of a fixed-size property type, or 0 if variable.
func (pt PropType) FixedSize() int {
	switch pt.BaseType() {
	case PropTypeInt16, PropTypeBool, PropTypeError:
		return 2
	case PropTypeInt32, PropTypeFloat32:
		return 4
	case PropTypeFloat64, PropTypeCurrency, PropTypeAppTime, PropTypeInt64, PropTypeSysTime:
		return 8
	case PropTypeGUID:
		return 16
	default:
		return 0 // Variable size
	}
}

// IsVariable returns true if the property type is variable-length.
func (pt PropType) IsVariable() bool {
	return pt.FixedSize() == 0
}

// IsFixedSize returns true if the property type is fixed-size.
func (pt PropType) IsFixedSize() bool {
	return pt.FixedSize() > 0
}

// Common property IDs from [MS-OXPROPS].
// These are the most commonly used properties for messages, folders, attachments, and recipients.
const (
	// Message properties - See [MS-OXPROPS] for canonical definitions
	PidTagSubject                      PropID = 0x0037
	PidTagSubjectPrefix                PropID = 0x003D
	PidTagNormalizedSubject            PropID = 0x0E1D
	PidTagBody                         PropID = 0x1000
	PidTagHtmlBody                     PropID = 0x1013
	PidTagRtfCompressed                PropID = 0x1009
	PidTagImportance                   PropID = 0x0017
	PidTagPriority                     PropID = 0x0026
	PidTagSensitivity                  PropID = 0x0036
	PidTagMessageClass                 PropID = 0x001A
	PidTagMessageFlags                 PropID = 0x0E07
	PidTagMessageSize                  PropID = 0x0E08
	PidTagMessageStatus                PropID = 0x0E17
	PidTagHasAttachments               PropID = 0x0E1B
	PidTagCreationTime                 PropID = 0x3007
	PidTagLastModificationTime         PropID = 0x3008
	PidTagMessageDeliveryTime          PropID = 0x0E06
	PidTagClientSubmitTime             PropID = 0x0039
	PidTagSentRepresentingName         PropID = 0x0042
	PidTagSentRepresentingEmailAddress PropID = 0x0065
	PidTagSenderName                   PropID = 0x0C1A
	PidTagSenderEmailAddress           PropID = 0x0C1F
	PidTagSenderAddressType            PropID = 0x0C1E
	PidTagDisplayTo                    PropID = 0x0E04
	PidTagDisplayCc                    PropID = 0x0E03
	PidTagDisplayBcc                   PropID = 0x0E02
	PidTagConversationTopic            PropID = 0x0070
	PidTagConversationIndex            PropID = 0x0071
	PidTagInternetMessageId            PropID = 0x1035

	// Folder properties
	PidTagDisplayName        PropID = 0x3001
	PidTagContentCount       PropID = 0x3602
	PidTagContentUnreadCount PropID = 0x3603
	PidTagSubfolders         PropID = 0x360A
	PidTagContainerClass     PropID = 0x3613
	PidTagFolderType         PropID = 0x3601
	PidTagDepth              PropID = 0x3005

	// Attachment properties
	PidTagAttachFilename     PropID = 0x3704
	PidTagAttachLongFilename PropID = 0x3707
	PidTagAttachExtension    PropID = 0x3703
	PidTagAttachSize         PropID = 0x0E20
	PidTagAttachMethod       PropID = 0x3705
	PidTagAttachDataBinary   PropID = 0x3701
	PidTagAttachDataObject   PropID = 0x3701 // Same as binary
	PidTagAttachMimeTag      PropID = 0x370E
	PidTagAttachContentId    PropID = 0x3712
	PidTagAttachNumber       PropID = 0x0E21
	PidTagRenderingPosition  PropID = 0x370B

	// Recipient properties
	PidTagRecipientType        PropID = 0x0C15
	PidTagEmailAddress         PropID = 0x3003
	PidTagSmtpAddress          PropID = 0x39FE
	PidTagAddressType          PropID = 0x3002
	PidTagRecipientDisplayName PropID = 0x5FF6

	// Table row properties
	PidTagRowId     PropID = 0x3000
	PidTagLtpRowId  PropID = 0x67F2
	PidTagLtpRowVer PropID = 0x67F3

	// Named property range
	NamedPropertyRangeStart PropID = 0x8000
	NamedPropertyRangeEnd   PropID = 0xFFFE
)

// String returns a string representation of the property ID.
func (id PropID) String() string {
	// Map common property IDs to names
	names := map[PropID]string{
		PidTagSubject:              "PidTagSubject",
		PidTagBody:                 "PidTagBody",
		PidTagHtmlBody:             "PidTagHtmlBody",
		PidTagMessageClass:         "PidTagMessageClass",
		PidTagCreationTime:         "PidTagCreationTime",
		PidTagLastModificationTime: "PidTagLastModificationTime",
		PidTagDisplayName:          "PidTagDisplayName",
		PidTagContentCount:         "PidTagContentCount",
		PidTagAttachFilename:       "PidTagAttachFilename",
		PidTagAttachSize:           "PidTagAttachSize",
		PidTagEmailAddress:         "PidTagEmailAddress",
		PidTagSenderName:           "PidTagSenderName",
	}
	if name, ok := names[id]; ok {
		return name
	}
	return ""
}

// IsNamedProperty returns true if this property ID is in the named property range.
func (id PropID) IsNamedProperty() bool {
	return id >= NamedPropertyRangeStart && id <= NamedPropertyRangeEnd
}

// Recipient types.
const (
	RecipientTypeOriginator = 0
	RecipientTypeTo         = 1
	RecipientTypeCc         = 2
	RecipientTypeBcc        = 3
)

// Attachment methods.
const (
	AttachMethodNone     = 0
	AttachMethodByValue  = 1
	AttachMethodByRef    = 2
	AttachMethodByRefRes = 4
	AttachMethodEmbedded = 5 // Embedded message
	AttachMethodOLE      = 6
)

// FileTimeToTime converts a Windows FILETIME to time.Time.
// FILETIME is 100-nanosecond intervals since January 1, 1601 UTC.
// See [MS-OXCDATA] Section 2.11.1 - PtypTime definition.
func FileTimeToTime(ft uint64) time.Time {
	// FILETIME epoch is January 1, 1601
	// Unix epoch is January 1, 1970
	// Difference is 116444736000000000 100-nanosecond intervals
	const filetimeEpochDiff = 116444736000000000

	if ft < filetimeEpochDiff {
		return time.Time{}
	}

	// Convert to Unix nanoseconds
	unixNano := (int64(ft) - filetimeEpochDiff) * 100 //nolint:gosec // G115: FILETIME reinterpretation
	return time.Unix(0, unixNano).UTC()
}

// TimeToFileTime converts a time.Time to Windows FILETIME.
func TimeToFileTime(t time.Time) uint64 {
	const filetimeEpochDiff = 116444736000000000

	if t.IsZero() {
		return 0
	}

	unixNano := t.UnixNano()
	return uint64(unixNano/100 + filetimeEpochDiff)
}
