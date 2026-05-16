package util

import (
	"bytes"
	"testing"
)

func TestParseGUID(t *testing.T) {
	// Test parsing a GUID from bytes
	data := []byte{
		0x28, 0x03, 0x02, 0x00, // Data1: 0x00020328
		0x00, 0x00, // Data2: 0x0000
		0x00, 0x00, // Data3: 0x0000
		0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46, // Data4
	}

	guid, err := ParseGUID(data)
	if err != nil {
		t.Fatalf("ParseGUID failed: %v", err)
	}

	if guid.Data1 != 0x00020328 {
		t.Errorf("Data1 = 0x%08X, want 0x00020328", guid.Data1)
	}
	if guid.Data2 != 0x0000 {
		t.Errorf("Data2 = 0x%04X, want 0x0000", guid.Data2)
	}
	if guid.Data3 != 0x0000 {
		t.Errorf("Data3 = 0x%04X, want 0x0000", guid.Data3)
	}

	expectedData4 := [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}
	if guid.Data4 != expectedData4 {
		t.Errorf("Data4 = %v, want %v", guid.Data4, expectedData4)
	}
}

func TestGUIDBytes(t *testing.T) {
	guid := GUID{
		Data1: 0x00020328,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}

	b := guid.Bytes()

	expected := []byte{
		0x28, 0x03, 0x02, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46,
	}

	if !bytes.Equal(b, expected) {
		t.Errorf("Bytes() = %v, want %v", b, expected)
	}
}

func TestGUIDRoundtrip(t *testing.T) {
	original := GUID{
		Data1: 0x12345678,
		Data2: 0xABCD,
		Data3: 0xEF01,
		Data4: [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	}

	b := original.Bytes()
	parsed, err := ParseGUID(b)
	if err != nil {
		t.Fatalf("ParseGUID failed: %v", err)
	}

	if !original.Equal(parsed) {
		t.Errorf("Roundtrip failed: got %v, want %v", parsed, original)
	}
}

func TestGUIDString(t *testing.T) {
	guid := PS_MAPI

	str := guid.String()
	expected := "{00020328-0000-0000-C000-000000000046}"

	if str != expected {
		t.Errorf("String() = %s, want %s", str, expected)
	}
}

func TestGUIDIsZero(t *testing.T) {
	zero := GUID{}
	if !zero.IsZero() {
		t.Error("Zero GUID should return IsZero() = true")
	}

	nonZero := PS_MAPI
	if nonZero.IsZero() {
		t.Error("Non-zero GUID should return IsZero() = false")
	}
}

func TestGUIDEqual(t *testing.T) {
	g1 := PS_MAPI
	g2 := PS_MAPI
	g3 := PS_PUBLIC_STRINGS

	if !g1.Equal(g2) {
		t.Error("Same GUIDs should be equal")
	}

	if g1.Equal(g3) {
		t.Error("Different GUIDs should not be equal")
	}
}

func TestParseGUIDError(t *testing.T) {
	// Too short
	_, err := ParseGUID([]byte{0x01, 0x02, 0x03})
	if err == nil {
		t.Error("ParseGUID should fail with insufficient data")
	}
}

func TestWellKnownGUIDs(t *testing.T) {
	// Verify well-known GUIDs are not zero
	guids := []struct {
		name string
		guid GUID
	}{
		{"PS_MAPI", PS_MAPI},
		{"PS_PUBLIC_STRINGS", PS_PUBLIC_STRINGS},
		{"PS_INTERNET_HEADERS", PS_INTERNET_HEADERS},
		{"PSETID_Common", PSETID_Common},
		{"PSETID_Address", PSETID_Address},
		{"PSETID_Appointment", PSETID_Appointment},
		{"PSETID_Meeting", PSETID_Meeting},
		{"PSETID_Task", PSETID_Task},
		{"PSETID_Note", PSETID_Note},
	}

	for _, tc := range guids {
		if tc.guid.IsZero() {
			t.Errorf("%s should not be zero", tc.name)
		}
	}

	// GUID_NULL should be zero
	if !GUID_NULL.IsZero() {
		t.Error("GUID_NULL should be zero")
	}
}
