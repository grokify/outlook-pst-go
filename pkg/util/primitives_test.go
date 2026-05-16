package util

import "testing"

func TestMakeNID(t *testing.T) {
	tests := []struct {
		nidType NIDType
		index   uint32
	}{
		{NIDTypeNormalFolder, 0x122},
		{NIDTypeNormalMessage, 0x1},
		{NIDTypeInternal, 0x3},
		{NIDTypeAttachment, 0x100},
	}

	for _, tc := range tests {
		nid := MakeNID(tc.nidType, tc.index)

		gotType := nid.Type()
		if gotType != tc.nidType {
			t.Errorf("MakeNID(%v, %d).Type() = %v, want %v", tc.nidType, tc.index, gotType, tc.nidType)
		}

		gotIndex := nid.Index()
		if gotIndex != tc.index {
			t.Errorf("MakeNID(%v, %d).Index() = %d, want %d", tc.nidType, tc.index, gotIndex, tc.index)
		}
	}
}

func TestNodeIDConstants(t *testing.T) {
	// Verify well-known NIDs
	if NIDMessageStore.Type() != NIDTypeInternal {
		t.Errorf("NIDMessageStore type = %v, want %v", NIDMessageStore.Type(), NIDTypeInternal)
	}

	if NIDNameIDMap.Type() != NIDTypeInternal {
		t.Errorf("NIDNameIDMap type = %v, want %v", NIDNameIDMap.Type(), NIDTypeInternal)
	}

	if NIDRootFolder.Type() != NIDTypeNormalFolder {
		t.Errorf("NIDRootFolder type = %v, want %v", NIDRootFolder.Type(), NIDTypeNormalFolder)
	}
}

func TestBlockIDFlags(t *testing.T) {
	// External block (no internal bit)
	external := BlockID(0x100)
	if !external.IsExternal() {
		t.Error("External block should have IsExternal() = true")
	}
	if external.IsInternal() {
		t.Error("External block should have IsInternal() = false")
	}

	// Internal block (has internal bit set)
	internal := BlockID(0x102) // 0x100 | 0x2
	if internal.IsExternal() {
		t.Error("Internal block should have IsExternal() = false")
	}
	if !internal.IsInternal() {
		t.Error("Internal block should have IsInternal() = true")
	}
}

func TestHeapID(t *testing.T) {
	// Test HeapID construction and extraction
	pageIndex := uint16(5)
	allocIndex := uint16(3)

	hid := MakeHeapID(pageIndex, allocIndex)

	gotPage := hid.PageIndex()
	if gotPage != pageIndex {
		t.Errorf("HeapID.PageIndex() = %d, want %d", gotPage, pageIndex)
	}

	gotAlloc := hid.AllocIndex()
	if gotAlloc != allocIndex {
		t.Errorf("HeapID.AllocIndex() = %d, want %d", gotAlloc, allocIndex)
	}
}

func TestHeapNodeID(t *testing.T) {
	// Test heap ID detection
	heapHNID := HeapNodeID(0x00010020) // Page 1, alloc 32
	if !heapHNID.IsHeapID() {
		t.Error("Heap HNID should return IsHeapID() = true")
	}

	// Test subnode ID detection (has NID type bits set)
	subnodeHNID := HeapNodeID(0x00000022) // Type 0x02 (folder)
	if !subnodeHNID.IsSubnodeID() {
		t.Error("Subnode HNID should return IsSubnodeID() = true")
	}

	// Zero should be treated as heap ID
	zeroHNID := HeapNodeID(0)
	if !zeroHNID.IsHeapID() {
		t.Error("Zero HNID should return IsHeapID() = true")
	}
}

func TestNIDTypeMask(t *testing.T) {
	// Verify type mask covers all type values
	for i := NIDType(0); i <= 0x1F; i++ {
		nid := MakeNID(i, 0)
		if nid.Type() != i {
			t.Errorf("Type %d not preserved", i)
		}
	}
}
