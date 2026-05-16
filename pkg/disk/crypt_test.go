package disk

import (
	"bytes"
	"testing"
)

func TestPermuteCipher(t *testing.T) {
	// Test that encode followed by decode returns original data
	original := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xAB, 0xCD, 0xEF}

	encoded := PermuteEncode(original)
	decoded := PermuteDecode(encoded)

	if !bytes.Equal(original, decoded) {
		t.Errorf("Permute roundtrip failed: got %v, want %v", decoded, original)
	}
}

func TestPermuteDecodeInPlace(t *testing.T) {
	original := []byte{0x00, 0x01, 0x02, 0x03, 0xFF}

	// First encode, then decode in place
	encoded := PermuteEncode(original)
	PermuteDecodeInPlace(encoded)

	if !bytes.Equal(encoded, original) {
		t.Errorf("PermuteDecodeInPlace failed: got %v, want %v", encoded, original)
	}
}

func TestCyclicDeterminism(t *testing.T) {
	// Test that cyclic encoding is deterministic
	key := uint32(0x12345678)
	original := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xAB, 0xCD, 0xEF}

	encoded1 := CyclicEncode(original, key)
	encoded2 := CyclicEncode(original, key)

	if !bytes.Equal(encoded1, encoded2) {
		t.Errorf("Cyclic encoding not deterministic")
	}

	// Encoded should be different from original
	if bytes.Equal(original, encoded1) {
		t.Errorf("Cyclic encoding did not change data")
	}
}

func TestCyclicDecodeMatchesEncode(t *testing.T) {
	// CyclicDecode and CyclicEncode should produce the same result
	// (per MS-PST, the transformation is the same for both)
	key := uint32(0xABCDEF01)
	data := []byte{0x10, 0x20, 0x30, 0x40, 0x50}

	encoded := CyclicEncode(data, key)
	decoded := CyclicDecode(data, key)

	if !bytes.Equal(encoded, decoded) {
		t.Errorf("CyclicEncode and CyclicDecode produce different results")
	}
}

func TestCyclicDecodeInPlaceConsistency(t *testing.T) {
	// Test that in-place decoding matches the non-in-place version
	key := uint32(0xABCDEF01)
	data1 := []byte{0x10, 0x20, 0x30, 0x40, 0x50}
	data2 := make([]byte, len(data1))
	copy(data2, data1)

	decoded := CyclicDecode(data1, key)
	CyclicDecodeInPlace(data2, key)

	if !bytes.Equal(decoded, data2) {
		t.Errorf("CyclicDecodeInPlace inconsistent with CyclicDecode")
	}
}

func TestDecryptBlockNone(t *testing.T) {
	original := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	result := DecryptBlock(original, CryptMethodNone, 0x1234)

	if !bytes.Equal(result, original) {
		t.Errorf("DecryptBlock with CryptMethodNone modified data")
	}
}

func TestDecryptBlockPermute(t *testing.T) {
	original := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	encrypted := PermuteEncode(original)

	decrypted := DecryptBlock(encrypted, CryptMethodPermute, 0x5678)

	if !bytes.Equal(decrypted, original) {
		t.Errorf("DecryptBlock with CryptMethodPermute failed roundtrip")
	}
}

func TestDecryptBlockCyclic(t *testing.T) {
	// Test that DecryptBlock returns consistent results
	blockID := uint64(0x9ABC)
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	result1 := DecryptBlock(data, CryptMethodCyclic, blockID)
	result2 := DecryptBlock(data, CryptMethodCyclic, blockID)

	if !bytes.Equal(result1, result2) {
		t.Errorf("DecryptBlock with CryptMethodCyclic not deterministic")
	}
}

func TestComputeCRC(t *testing.T) {
	// Test CRC with known values
	data := []byte{0x01, 0x02, 0x03, 0x04}
	crc := ComputeCRC(data)

	// CRC should be deterministic
	crc2 := ComputeCRC(data)
	if crc != crc2 {
		t.Errorf("CRC not deterministic: got %08X and %08X", crc, crc2)
	}

	// Different data should produce different CRC
	data2 := []byte{0x01, 0x02, 0x03, 0x05}
	crc3 := ComputeCRC(data2)
	if crc == crc3 {
		t.Error("Different data produced same CRC")
	}
}

func TestComputeSignature(t *testing.T) {
	// Test signature computation
	sig := ComputeSignature(uint64(0x12345678), uint64(0xABCDEF00))

	// Signature should be deterministic
	sig2 := ComputeSignature(uint64(0x12345678), uint64(0xABCDEF00))
	if sig != sig2 {
		t.Errorf("Signature not deterministic: got %04X and %04X", sig, sig2)
	}

	// Test with 32-bit values
	sig32 := ComputeSignature(uint32(0x1234), uint32(0x5678))
	sig32_2 := ComputeSignature(uint32(0x1234), uint32(0x5678))
	if sig32 != sig32_2 {
		t.Errorf("32-bit signature not deterministic")
	}
}

func TestAlignDisk(t *testing.T) {
	tests := []struct {
		input    uint64
		expected uint64
	}{
		{0, 0},
		{1, 64},
		{63, 64},
		{64, 64},
		{65, 128},
		{128, 128},
		{129, 192},
	}

	for _, tc := range tests {
		result := AlignDisk(tc.input)
		if result != tc.expected {
			t.Errorf("AlignDisk(%d) = %d, want %d", tc.input, result, tc.expected)
		}
	}
}

func TestCryptMethodString(t *testing.T) {
	tests := []struct {
		method   CryptMethod
		expected string
	}{
		{CryptMethodNone, "None"},
		{CryptMethodPermute, "Permute"},
		{CryptMethodCyclic, "Cyclic"},
		{CryptMethod(99), "Unknown"},
	}

	for _, tc := range tests {
		got := tc.method.String()
		if got != tc.expected {
			t.Errorf("CryptMethod(%d).String() = %q, want %q", tc.method, got, tc.expected)
		}
	}
}

func TestPageTypeString(t *testing.T) {
	tests := []struct {
		pageType PageType
		expected string
	}{
		{PageTypeBBT, "BBT"},
		{PageTypeNBT, "NBT"},
		{PageTypeAMap, "AMap"},
		{PageType(0x99), "Unknown"},
	}

	for _, tc := range tests {
		got := tc.pageType.String()
		if got != tc.expected {
			t.Errorf("PageType(0x%02X).String() = %q, want %q", tc.pageType, got, tc.expected)
		}
	}
}

func TestBlockTypeString(t *testing.T) {
	tests := []struct {
		blockType BlockType
		expected  string
	}{
		{BlockTypeExternal, "External"},
		{BlockTypeExtended, "Extended"},
		{BlockTypeSubnode, "Subnode"},
		{BlockType(0x99), "Unknown"},
	}

	for _, tc := range tests {
		got := tc.blockType.String()
		if got != tc.expected {
			t.Errorf("BlockType(0x%02X).String() = %q, want %q", tc.blockType, got, tc.expected)
		}
	}
}

func TestPSTFormatString(t *testing.T) {
	tests := []struct {
		format   PSTFormat
		expected string
	}{
		{FormatANSI, "ANSI"},
		{FormatUnicode, "Unicode"},
		{FormatUnknown, "Unknown"},
	}

	for _, tc := range tests {
		got := tc.format.String()
		if got != tc.expected {
			t.Errorf("PSTFormat(%d).String() = %q, want %q", tc.format, got, tc.expected)
		}
	}
}
