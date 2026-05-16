package ltp

import (
	"testing"
	"time"
)

func TestPropTypeIsMultiValued(t *testing.T) {
	tests := []struct {
		propType    PropType
		multiValued bool
	}{
		{PropTypeInt32, false},
		{PropTypeString, false},
		{PropTypeBinary, false},
		{PropTypeMVInt32, true},
		{PropTypeMVString, true},
		{PropTypeMVBinary, true},
	}

	for _, tc := range tests {
		got := tc.propType.IsMultiValued()
		if got != tc.multiValued {
			t.Errorf("PropType(0x%04X).IsMultiValued() = %v, want %v", tc.propType, got, tc.multiValued)
		}
	}
}

func TestPropTypeBaseType(t *testing.T) {
	tests := []struct {
		propType PropType
		baseType PropType
	}{
		{PropTypeInt32, PropTypeInt32},
		{PropTypeMVInt32, PropTypeInt32},
		{PropTypeString, PropTypeString},
		{PropTypeMVString, PropTypeString},
	}

	for _, tc := range tests {
		got := tc.propType.BaseType()
		if got != tc.baseType {
			t.Errorf("PropType(0x%04X).BaseType() = 0x%04X, want 0x%04X", tc.propType, got, tc.baseType)
		}
	}
}

func TestPropTypeFixedSize(t *testing.T) {
	tests := []struct {
		propType  PropType
		fixedSize int
	}{
		{PropTypeInt16, 2},
		{PropTypeBool, 2},
		{PropTypeInt32, 4},
		{PropTypeFloat32, 4},
		{PropTypeFloat64, 8},
		{PropTypeInt64, 8},
		{PropTypeSysTime, 8},
		{PropTypeGUID, 16},
		{PropTypeString, 0},  // Variable
		{PropTypeBinary, 0},  // Variable
		{PropTypeString8, 0}, // Variable
	}

	for _, tc := range tests {
		got := tc.propType.FixedSize()
		if got != tc.fixedSize {
			t.Errorf("PropType(0x%04X).FixedSize() = %d, want %d", tc.propType, got, tc.fixedSize)
		}
	}
}

func TestPropTypeIsVariable(t *testing.T) {
	if PropTypeInt32.IsVariable() {
		t.Error("PropTypeInt32 should not be variable")
	}
	if !PropTypeString.IsVariable() {
		t.Error("PropTypeString should be variable")
	}
	if !PropTypeBinary.IsVariable() {
		t.Error("PropTypeBinary should be variable")
	}
}

func TestPropIDIsNamedProperty(t *testing.T) {
	tests := []struct {
		propID  PropID
		isNamed bool
	}{
		{PidTagSubject, false},
		{PidTagBody, false},
		{0x8000, true}, // Start of named range
		{0x8001, true},
		{0xFFFE, true}, // End of named range
		{0x7FFF, false},
		{0xFFFF, false}, // Above named range
	}

	for _, tc := range tests {
		got := tc.propID.IsNamedProperty()
		if got != tc.isNamed {
			t.Errorf("PropID(0x%04X).IsNamedProperty() = %v, want %v", tc.propID, got, tc.isNamed)
		}
	}
}

func TestFileTimeConversion(t *testing.T) {
	// Test roundtrip conversion
	original := time.Date(2020, 6, 15, 12, 30, 45, 0, time.UTC)

	ft := TimeToFileTime(original)
	roundtrip := FileTimeToTime(ft)

	// Allow for precision loss (100ns resolution)
	diff := roundtrip.Sub(original)
	if diff < -time.Microsecond || diff > time.Microsecond {
		t.Errorf("Time roundtrip failed: original=%v, roundtrip=%v, diff=%v", original, roundtrip, diff)
	}
}

func TestFileTimeToTimeZero(t *testing.T) {
	// Zero FILETIME should return zero time
	got := FileTimeToTime(0)
	if !got.IsZero() {
		t.Errorf("FileTimeToTime(0) should return zero time, got %v", got)
	}

	// FILETIME before Unix epoch (but after Windows epoch)
	got = FileTimeToTime(100)
	if !got.IsZero() {
		t.Errorf("FileTimeToTime(100) should return zero time (before Unix epoch), got %v", got)
	}
}

func TestTimeToFileTimeZero(t *testing.T) {
	// Zero time should return 0
	got := TimeToFileTime(time.Time{})
	if got != 0 {
		t.Errorf("TimeToFileTime(zero) = %d, want 0", got)
	}
}

func TestPropIDString(t *testing.T) {
	// Test that well-known property IDs have string representations
	if PidTagSubject.String() != "PidTagSubject" {
		t.Errorf("PidTagSubject.String() = %q, want %q", PidTagSubject.String(), "PidTagSubject")
	}

	// Unknown property IDs should return empty string
	unknown := PropID(0x1234)
	if unknown.String() != "" {
		t.Errorf("Unknown PropID.String() = %q, want empty", unknown.String())
	}
}

func TestFileTimeKnownValue(t *testing.T) {
	// Test with a known FILETIME value
	// January 1, 2000, 00:00:00 UTC = FILETIME 125911584000000000
	ft := uint64(125911584000000000)
	expected := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	got := FileTimeToTime(ft)

	// Allow small difference due to precision
	diff := got.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("FileTimeToTime(%d) = %v, want ~%v (diff: %v)", ft, got, expected, diff)
	}
}
