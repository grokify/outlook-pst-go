package ltp

import (
	"encoding/binary"
	"fmt"

	"github.com/grokify/outlook-pst-go/pkg/disk"
	"github.com/grokify/outlook-pst-go/pkg/util"
)

// TableWriter builds a Table Context structure for writing.
// See [MS-PST] Section 2.3.4 for the Table Context specification.
type TableWriter struct {
	heap      *HeapWriter
	rowBTH    *BTHWriter
	format    disk.PSTFormat
	columns   []ColumnDef
	rows      []*tableRow
	nextRowID uint32
}

// ColumnDef defines a column in the table.
type ColumnDef struct {
	PropID   PropID
	PropType PropType
	Size     int    // Size in bytes (0 = variable)
	Offset   uint16 // Offset in row data (calculated during build)
}

// tableRow holds a row's data during building.
type tableRow struct {
	rowID  uint32
	values map[PropID][]byte
}

// NewTableWriter creates a new table writer.
func NewTableWriter(format disk.PSTFormat) *TableWriter {
	heap := CreateTableContextHeap(format)
	rowBTH := CreateRowIndexBTH(heap, format)

	return &TableWriter{
		heap:      heap,
		rowBTH:    rowBTH,
		format:    format,
		nextRowID: 1, // Row IDs start at 1
	}
}

// AddColumn adds a column definition to the table.
func (w *TableWriter) AddColumn(propID PropID, propType PropType) {
	size := propType.FixedSize()
	if size == 0 {
		size = 4 // Variable-size columns store HNID (4 bytes)
	}

	w.columns = append(w.columns, ColumnDef{
		PropID:   propID,
		PropType: propType,
		Size:     size,
	})
}

// AddRow adds a new row and returns its row ID.
func (w *TableWriter) AddRow() uint32 {
	rowID := w.nextRowID
	w.nextRowID++

	w.rows = append(w.rows, &tableRow{
		rowID:  rowID,
		values: make(map[PropID][]byte),
	})

	return rowID
}

// SetRowValue sets a value in a row.
func (w *TableWriter) SetRowValue(rowID uint32, propID PropID, value []byte) error {
	for _, row := range w.rows {
		if row.rowID == rowID {
			row.values[propID] = value
			return nil
		}
	}
	return fmt.Errorf("row %d not found", rowID)
}

// SetRowInt32 sets an int32 value in a row.
func (w *TableWriter) SetRowInt32(rowID uint32, propID PropID, value int32) error {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(value)) //nolint:gosec // G115: signed/unsigned reinterpretation
	return w.SetRowValue(rowID, propID, data)
}

// SetRowInt64 sets an int64 value in a row.
func (w *TableWriter) SetRowInt64(rowID uint32, propID PropID, value int64) error {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(value)) //nolint:gosec // G115: signed/unsigned reinterpretation
	return w.SetRowValue(rowID, propID, data)
}

// SetRowString sets a string value in a row.
func (w *TableWriter) SetRowString(rowID uint32, propID PropID, value string) error {
	data := encodeUTF16LE(value)
	return w.SetRowValue(rowID, propID, data)
}

// SetRowBinary sets a binary value in a row.
func (w *TableWriter) SetRowBinary(rowID uint32, propID PropID, value []byte) error {
	return w.SetRowValue(rowID, propID, value)
}

// DeleteRow removes a row.
func (w *TableWriter) DeleteRow(rowID uint32) {
	for i, row := range w.rows {
		if row.rowID == rowID {
			w.rows = append(w.rows[:i], w.rows[i+1:]...)
			return
		}
	}
}

// Build builds the table context and returns the heap data.
func (w *TableWriter) Build() ([]byte, error) {
	if len(w.columns) == 0 {
		return nil, fmt.Errorf("table must have at least one column")
	}

	// Calculate row layout and offsets
	w.calculateColumnOffsets()

	// Calculate row size
	rowSize := w.calculateRowSize()

	// Build existence bitmap size (1 bit per column, rounded up to bytes)
	existenceBitmapSize := (len(w.columns) + 7) / 8

	// Build row matrix
	rowMatrix := make([]byte, len(w.rows)*(rowSize+existenceBitmapSize))

	for i, row := range w.rows {
		rowOffset := i * (rowSize + existenceBitmapSize)

		// Build existence bitmap
		existenceBitmap := make([]byte, existenceBitmapSize)

		// Fill in column values
		for colIdx, col := range w.columns {
			value, exists := row.values[col.PropID]
			if exists {
				// Set existence bit
				byteIdx := colIdx / 8
				bitIdx := colIdx % 8
				existenceBitmap[byteIdx] |= (1 << bitIdx)

				// Write value
				valueOffset := rowOffset + existenceBitmapSize + int(col.Offset)

				if col.PropType.IsFixedSize() {
					// Fixed-size: copy value directly
					copyLen := col.Size
					if len(value) < copyLen {
						copyLen = len(value)
					}
					copy(rowMatrix[valueOffset:], value[:copyLen])
				} else {
					// Variable-size: allocate in heap and store HNID
					hid, err := w.heap.Allocate(value)
					if err != nil {
						return nil, fmt.Errorf("failed to allocate row value: %w", err)
					}
					binary.LittleEndian.PutUint32(rowMatrix[valueOffset:], uint32(hid))
				}
			}
		}

		// Write existence bitmap
		copy(rowMatrix[rowOffset:], existenceBitmap)

		// Add row to BTH
		indexData := make([]byte, 4)
		binary.LittleEndian.PutUint32(indexData, uint32(i))
		if err := w.rowBTH.InsertUint32Key(row.rowID, indexData); err != nil {
			return nil, fmt.Errorf("failed to insert row %d into BTH: %w", row.rowID, err)
		}
	}

	// Allocate row matrix in heap (or as subnode if large)
	var rowMatrixHNID util.HeapNodeID
	if len(rowMatrix) <= disk.HeapMaxAllocSize {
		hid, err := w.heap.Allocate(rowMatrix)
		if err != nil {
			return nil, fmt.Errorf("failed to allocate row matrix: %w", err)
		}
		rowMatrixHNID = util.HeapNodeID(hid)
	} else {
		// Large row matrix - would need subnode allocation
		// For now, return error
		return nil, fmt.Errorf("row matrix too large for heap: %d bytes", len(rowMatrix))
	}

	// Build BTH
	rowBTHHID, err := w.rowBTH.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build row BTH: %w", err)
	}

	// Build column descriptors
	colDescData := w.buildColumnDescriptors()
	colDescHID, err := w.heap.Allocate(colDescData)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate column descriptors: %w", err)
	}

	// Build TCINFO header
	tcInfo := w.buildTCInfo(colDescHID, rowBTHHID, rowMatrixHNID, rowSize+existenceBitmapSize)
	tcInfoHID, err := w.heap.Allocate(tcInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate TCINFO: %w", err)
	}

	w.heap.SetRoot(tcInfoHID)

	// Build heap
	return w.heap.Build()
}

// calculateColumnOffsets calculates the offset of each column in a row.
func (w *TableWriter) calculateColumnOffsets() {
	offset := uint16(0)
	for i := range w.columns {
		w.columns[i].Offset = offset
		if w.columns[i].Size == 0 {
			w.columns[i].Size = 4 // HNID for variable columns
		}
		offset += uint16(w.columns[i].Size) //nolint:gosec // G115: column size bounded
	}
}

// calculateRowSize returns the size of a row (excluding existence bitmap).
func (w *TableWriter) calculateRowSize() int {
	size := 0
	for _, col := range w.columns {
		size += col.Size
	}
	return size
}

// buildColumnDescriptors builds the column descriptor array.
func (w *TableWriter) buildColumnDescriptors() []byte {
	// Each column descriptor is 8 bytes:
	// PropType: 2 bytes
	// PropID: 2 bytes
	// ibData: 2 bytes (offset in row)
	// cbData: 1 byte (size)
	// iBit: 1 byte (bit index in existence bitmap)

	data := make([]byte, len(w.columns)*8)
	for i, col := range w.columns {
		offset := i * 8
		binary.LittleEndian.PutUint16(data[offset:offset+2], uint16(col.PropType))
		binary.LittleEndian.PutUint16(data[offset+2:offset+4], uint16(col.PropID))
		binary.LittleEndian.PutUint16(data[offset+4:offset+6], col.Offset)
		data[offset+6] = byte(col.Size) //nolint:gosec // G115: column size bounded by row size
		data[offset+7] = byte(i)        // iBit
	}
	return data
}

// buildTCInfo builds the TCINFO header structure.
func (w *TableWriter) buildTCInfo(_ /* colDescHID */, rowBTHHID util.HeapID, rowMatrixHNID util.HeapNodeID, rowSize int) []byte {
	// TCINFO structure - See [MS-PST] Section 2.3.4.1
	// bType: 1 byte (0x7C for TC)
	// cCols: 1 byte (column count)
	// rgib: 8 bytes (4 uint16 offsets for TCI_4b, TCI_2b, TCI_1b, TCI_bm)
	// hidRowIndex: 4 bytes (HID of row BTH)
	// hnidRows: 4 bytes (HNID of row matrix)
	// hidIndex: 4 bytes (deprecated, 0)

	data := make([]byte, 22)
	data[0] = disk.HeapSigTC       // 0x7C
	data[1] = byte(len(w.columns)) //nolint:gosec // G115: column count bounded by table capacity

	// rgib offsets - simplified: all columns at same boundary
	rowSizeWithBitmap := rowSize
	binary.LittleEndian.PutUint16(data[2:4], 0)                          // TCI_4b
	binary.LittleEndian.PutUint16(data[4:6], uint16(rowSizeWithBitmap))  //nolint:gosec // G115: row size bounded by heap
	binary.LittleEndian.PutUint16(data[6:8], uint16(rowSizeWithBitmap))  //nolint:gosec // G115: row size bounded by heap
	binary.LittleEndian.PutUint16(data[8:10], uint16(rowSizeWithBitmap)) //nolint:gosec // G115: row size bounded by heap

	binary.LittleEndian.PutUint32(data[10:14], uint32(rowBTHHID))
	binary.LittleEndian.PutUint32(data[14:18], uint32(rowMatrixHNID))
	// hidIndex at 18:22 left as 0

	return data
}

// RowCount returns the number of rows.
func (w *TableWriter) RowCount() int {
	return len(w.rows)
}

// ColumnCount returns the number of columns.
func (w *TableWriter) ColumnCount() int {
	return len(w.columns)
}

// CreateHierarchyTable creates a table for folder hierarchy (subfolders).
func CreateHierarchyTable(format disk.PSTFormat) *TableWriter {
	w := NewTableWriter(format)

	// Standard hierarchy table columns
	w.AddColumn(PidTagRowId, PropTypeInt32)
	w.AddColumn(PidTagDisplayName, PropTypeString)
	w.AddColumn(PidTagContentCount, PropTypeInt32)
	w.AddColumn(PidTagContentUnreadCount, PropTypeInt32)
	w.AddColumn(PidTagSubfolders, PropTypeBool)
	w.AddColumn(PidTagDepth, PropTypeInt32)

	return w
}

// CreateContentsTable creates a table for folder contents (messages).
func CreateContentsTable(format disk.PSTFormat) *TableWriter {
	w := NewTableWriter(format)

	// Standard contents table columns
	w.AddColumn(PidTagRowId, PropTypeInt32)
	w.AddColumn(PidTagSubject, PropTypeString)
	w.AddColumn(PidTagMessageClass, PropTypeString)
	w.AddColumn(PidTagClientSubmitTime, PropTypeSysTime)
	w.AddColumn(PidTagMessageSize, PropTypeInt32)
	w.AddColumn(PidTagMessageFlags, PropTypeInt32)
	w.AddColumn(PidTagHasAttachments, PropTypeBool)

	return w
}

// CreateRecipientTable creates a table for message recipients.
func CreateRecipientTable(format disk.PSTFormat) *TableWriter {
	w := NewTableWriter(format)

	// Standard recipient table columns
	w.AddColumn(PidTagRowId, PropTypeInt32)
	w.AddColumn(PidTagRecipientType, PropTypeInt32)
	w.AddColumn(PidTagDisplayName, PropTypeString)
	w.AddColumn(PidTagEmailAddress, PropTypeString)
	w.AddColumn(PidTagAddressType, PropTypeString)

	return w
}

// CreateAttachmentTable creates a table for message attachments.
func CreateAttachmentTable(format disk.PSTFormat) *TableWriter {
	w := NewTableWriter(format)

	// Standard attachment table columns
	w.AddColumn(PidTagRowId, PropTypeInt32)
	w.AddColumn(PidTagAttachNumber, PropTypeInt32)
	w.AddColumn(PidTagAttachMethod, PropTypeInt32)
	w.AddColumn(PidTagAttachFilename, PropTypeString)
	w.AddColumn(PidTagAttachSize, PropTypeInt32)
	w.AddColumn(PidTagAttachMimeTag, PropTypeString)

	return w
}
