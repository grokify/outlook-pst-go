package util

import (
	"encoding/binary"
	"fmt"
)

// GUID represents a 128-bit globally unique identifier.
// Uses the Windows GUID structure layout.
type GUID struct {
	Data1 uint32   // First 4 bytes (little-endian)
	Data2 uint16   // Next 2 bytes (little-endian)
	Data3 uint16   // Next 2 bytes (little-endian)
	Data4 [8]byte  // Last 8 bytes (big-endian/sequential)
}

// GUIDSize is the size of a GUID in bytes.
const GUIDSize = 16

// ParseGUID parses a GUID from a 16-byte slice.
func ParseGUID(data []byte) (GUID, error) {
	if len(data) < GUIDSize {
		return GUID{}, fmt.Errorf("insufficient data for GUID: need %d bytes, got %d", GUIDSize, len(data))
	}

	var g GUID
	g.Data1 = binary.LittleEndian.Uint32(data[0:4])
	g.Data2 = binary.LittleEndian.Uint16(data[4:6])
	g.Data3 = binary.LittleEndian.Uint16(data[6:8])
	copy(g.Data4[:], data[8:16])

	return g, nil
}

// GUIDFromBytes parses a GUID from a byte slice, returning zero GUID if insufficient data.
// This is a convenience function that doesn't return an error.
func GUIDFromBytes(data []byte) GUID {
	if len(data) < GUIDSize {
		return GUID{}
	}

	var g GUID
	g.Data1 = binary.LittleEndian.Uint32(data[0:4])
	g.Data2 = binary.LittleEndian.Uint16(data[4:6])
	g.Data3 = binary.LittleEndian.Uint16(data[6:8])
	copy(g.Data4[:], data[8:16])

	return g
}

// Bytes returns the GUID as a 16-byte slice.
func (g GUID) Bytes() []byte {
	data := make([]byte, GUIDSize)
	binary.LittleEndian.PutUint32(data[0:4], g.Data1)
	binary.LittleEndian.PutUint16(data[4:6], g.Data2)
	binary.LittleEndian.PutUint16(data[6:8], g.Data3)
	copy(data[8:16], g.Data4[:])
	return data
}

// String returns the standard GUID string representation.
// Format: {XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX}
func (g GUID) String() string {
	return fmt.Sprintf("{%08X-%04X-%04X-%02X%02X-%02X%02X%02X%02X%02X%02X}",
		g.Data1, g.Data2, g.Data3,
		g.Data4[0], g.Data4[1],
		g.Data4[2], g.Data4[3], g.Data4[4], g.Data4[5], g.Data4[6], g.Data4[7])
}

// IsZero returns true if the GUID is all zeros.
func (g GUID) IsZero() bool {
	return g.Data1 == 0 && g.Data2 == 0 && g.Data3 == 0 &&
		g.Data4 == [8]byte{}
}

// Equal returns true if two GUIDs are equal.
func (g GUID) Equal(other GUID) bool {
	return g.Data1 == other.Data1 &&
		g.Data2 == other.Data2 &&
		g.Data3 == other.Data3 &&
		g.Data4 == other.Data4
}

// Well-known MAPI GUIDs.
var (
	// GUID_NULL is the null GUID.
	GUID_NULL = GUID{}

	// PS_MAPI is the MAPI namespace GUID.
	PS_MAPI = GUID{
		Data1: 0x00020328,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}

	// PS_PUBLIC_STRINGS is the public strings namespace GUID.
	PS_PUBLIC_STRINGS = GUID{
		Data1: 0x00020329,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}

	// PS_INTERNET_HEADERS is the internet headers namespace GUID.
	PS_INTERNET_HEADERS = GUID{
		Data1: 0x00020386,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}

	// PSETID_Common is the common properties namespace GUID.
	PSETID_Common = GUID{
		Data1: 0x00062008,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}

	// PSETID_Address is the address book namespace GUID.
	PSETID_Address = GUID{
		Data1: 0x00062004,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}

	// PSETID_Appointment is the calendar/appointment namespace GUID.
	PSETID_Appointment = GUID{
		Data1: 0x00062002,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}

	// PSETID_Meeting is the meeting namespace GUID.
	PSETID_Meeting = GUID{
		Data1: 0x6ED8DA90,
		Data2: 0x450B,
		Data3: 0x101B,
		Data4: [8]byte{0x98, 0xDA, 0x00, 0xAA, 0x00, 0x3F, 0x13, 0x05},
	}

	// PSETID_Task is the task namespace GUID.
	PSETID_Task = GUID{
		Data1: 0x00062003,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}

	// PSETID_Note is the note namespace GUID.
	PSETID_Note = GUID{
		Data1: 0x0006200E,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}
)
