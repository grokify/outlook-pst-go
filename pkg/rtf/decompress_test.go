package rtf

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestParseHeader(t *testing.T) {
	// Create a valid header
	data := make([]byte, 16)
	binary.LittleEndian.PutUint32(data[0:4], 100)             // CompSize
	binary.LittleEndian.PutUint32(data[4:8], 50)              // RawSize
	binary.LittleEndian.PutUint32(data[8:12], MagicCompressed) // CompType
	binary.LittleEndian.PutUint32(data[12:16], 0x12345678)    // CRC

	header, err := ParseHeader(data)
	if err != nil {
		t.Fatalf("ParseHeader failed: %v", err)
	}

	if header.CompSize != 100 {
		t.Errorf("CompSize = %d, want 100", header.CompSize)
	}
	if header.RawSize != 50 {
		t.Errorf("RawSize = %d, want 50", header.RawSize)
	}
	if !header.IsCompressed() {
		t.Error("IsCompressed() = false, want true")
	}
	if header.CRC != 0x12345678 {
		t.Errorf("CRC = 0x%08X, want 0x12345678", header.CRC)
	}
}

func TestParseHeaderTooShort(t *testing.T) {
	data := make([]byte, 10) // Too short
	_, err := ParseHeader(data)
	if err != ErrInvalidHeader {
		t.Errorf("ParseHeader() error = %v, want ErrInvalidHeader", err)
	}
}

func TestIsCompressed(t *testing.T) {
	tests := []struct {
		name     string
		magic    uint32
		expected bool
	}{
		{"compressed", MagicCompressed, true},
		{"uncompressed", MagicUncompressed, false},
		{"invalid", 0x00000000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, 16)
			binary.LittleEndian.PutUint32(data[8:12], tt.magic)

			if got := IsCompressed(data); got != tt.expected {
				t.Errorf("IsCompressed() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDecompressUncompressed(t *testing.T) {
	// Create uncompressed RTF data
	rtfContent := []byte("{\\rtf1 Hello World}")

	data := make([]byte, 16+len(rtfContent))
	binary.LittleEndian.PutUint32(data[0:4], uint32(len(data)-4)) // CompSize
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(rtfContent))) // RawSize
	binary.LittleEndian.PutUint32(data[8:12], MagicUncompressed) // CompType
	binary.LittleEndian.PutUint32(data[12:16], 0) // CRC
	copy(data[16:], rtfContent)

	result, err := Decompress(data)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	if !bytes.Equal(result, rtfContent) {
		t.Errorf("Decompress() = %q, want %q", result, rtfContent)
	}
}

func TestPreloadedDictionary(t *testing.T) {
	// Verify the preloaded dictionary has the expected size
	if len(preloadedDictionary) != preloadSize {
		t.Errorf("preloadedDictionary length = %d, want %d", len(preloadedDictionary), preloadSize)
	}

	// Verify it starts with expected content
	if !bytes.HasPrefix(preloadedDictionary, []byte("{\\rtf1\\ansi")) {
		t.Error("preloadedDictionary should start with {\\rtf1\\ansi")
	}
}

func TestComputeCRC(t *testing.T) {
	// Test with known values
	data := []byte("test")
	crc := ComputeCRC(data)

	// CRC should be non-zero for non-empty data
	if crc == 0 {
		t.Error("ComputeCRC() returned 0 for non-empty data")
	}

	// Same data should produce same CRC
	crc2 := ComputeCRC(data)
	if crc != crc2 {
		t.Errorf("ComputeCRC() inconsistent: %d != %d", crc, crc2)
	}

	// Different data should produce different CRC
	crc3 := ComputeCRC([]byte("different"))
	if crc == crc3 {
		t.Error("ComputeCRC() produced same value for different data")
	}
}

func TestCRCTableSize(t *testing.T) {
	if len(crcTable) != 256 {
		t.Errorf("crcTable length = %d, want 256", len(crcTable))
	}
}
