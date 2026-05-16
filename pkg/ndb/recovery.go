package ndb

import (
	"fmt"

	"github.com/grokify/outlook-pst-go/pkg/disk"
)

// RecoveryResult contains information about a recovery operation.
type RecoveryResult struct {
	// WasCorrupted indicates whether the file needed recovery.
	WasCorrupted bool
	// PreviousStatus is the AMap status before recovery.
	PreviousStatus disk.AMapStatus
	// RecoveredBlocks is the number of blocks recovered from the BBT.
	RecoveredBlocks int
	// Error contains any non-fatal error encountered during recovery.
	Error error
}

// CheckAndRecover checks if the PST file needs recovery and performs it if necessary.
// This should be called when opening a file in read-write mode.
//
// Recovery is needed when:
//   - AMapStatus is Invalid (crash during phase 1 of commit)
//   - AMapStatus is Valid2 (successful commit, needs normalization)
//
// The recovery process rebuilds the allocation map from the B-trees.
func CheckAndRecover(db *Database) (*RecoveryResult, error) {
	if db.readOnly {
		return nil, fmt.Errorf("cannot recover read-only database")
	}

	result := &RecoveryResult{
		PreviousStatus: db.header.GetAMapStatus(),
	}

	switch result.PreviousStatus {
	case disk.AMapStatusValid:
		// File is in consistent state
		result.WasCorrupted = false
		return result, nil

	case disk.AMapStatusValid2:
		// File was properly committed but not yet normalized
		// Just normalize the status
		result.WasCorrupted = false
		if err := normalizeAMapStatus(db); err != nil {
			return result, fmt.Errorf("failed to normalize AMap status: %w", err)
		}
		return result, nil

	case disk.AMapStatusInvalid:
		// Crash during write - need full recovery
		result.WasCorrupted = true
		if err := performRecovery(db, result); err != nil {
			return result, fmt.Errorf("recovery failed: %w", err)
		}
		return result, nil

	default:
		// Unknown status - treat as corruption
		result.WasCorrupted = true
		result.Error = fmt.Errorf("unknown AMap status: 0x%02X", result.PreviousStatus)
		if err := performRecovery(db, result); err != nil {
			return result, fmt.Errorf("recovery failed: %w", err)
		}
		return result, nil
	}
}

// normalizeAMapStatus changes Valid2 to Valid after successful commit.
func normalizeAMapStatus(db *Database) error {
	db.header.SetAMapStatus(disk.AMapStatusValid)
	return disk.WriteHeader(db.file, db.header)
}

// performRecovery performs full recovery by rebuilding the allocation map.
func performRecovery(db *Database, result *RecoveryResult) error {
	// Load B-tree roots
	if err := db.loadNBTRoot(); err != nil {
		return fmt.Errorf("failed to load NBT root: %w", err)
	}
	if err := db.loadBBTRoot(); err != nil {
		return fmt.Errorf("failed to load BBT root: %w", err)
	}

	// Collect all allocated blocks from BBT
	allocatedBlocks, err := collectAllocatedBlocks(db, db.bbtRoot)
	if err != nil {
		return fmt.Errorf("failed to collect allocated blocks: %w", err)
	}
	result.RecoveredBlocks = len(allocatedBlocks)

	// Also collect page allocations (B-tree pages)
	pageBlocks, err := collectPageAllocations(db)
	if err != nil {
		return fmt.Errorf("failed to collect page allocations: %w", err)
	}
	allocatedBlocks = append(allocatedBlocks, pageBlocks...)

	// Create new AMap manager and rebuild
	amap := disk.CreateNewAMapManager(db.Format())
	if err := amap.RebuildFromBTrees(allocatedBlocks); err != nil {
		return fmt.Errorf("failed to rebuild AMap: %w", err)
	}

	// Write recovered AMap pages
	for _, page := range getAMapPages(amap) {
		pageData, err := page.Serialize()
		if err != nil {
			return fmt.Errorf("failed to serialize AMap page: %w", err)
		}
		if err := db.WritePage(page.Offset, pageData); err != nil {
			return fmt.Errorf("failed to write AMap page: %w", err)
		}
	}

	// Update header with recovered state
	db.header.Root.CBAMapFree = amap.FreeSpace()
	db.header.Root.IBFileEOF = amap.FileSize()
	db.header.SetAMapStatus(disk.AMapStatusValid)

	if err := disk.WriteHeader(db.file, db.header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Sync to ensure recovery is persisted
	if err := db.Sync(); err != nil {
		return fmt.Errorf("failed to sync after recovery: %w", err)
	}

	return nil
}

// collectAllocatedBlocks recursively collects all block allocations from BBT.
func collectAllocatedBlocks(db *Database, page *disk.BTPage) ([]disk.BlockAllocation, error) {
	var allocs []disk.BlockAllocation

	if page.IsLeaf() {
		for _, entry := range page.BBTEntries {
			diskSize := disk.CalculateBlockDiskSize(uint64(entry.Size), db.Format())
			allocs = append(allocs, disk.BlockAllocation{
				Offset: entry.BRef.IB,
				Size:   uint16(diskSize),
			})
		}
		return allocs, nil
	}

	// Non-leaf - recurse into children
	for _, child := range page.NonleafEntries {
		childData, err := db.readPage(child.Ref.IB)
		if err != nil {
			return nil, err
		}
		childPage, err := disk.ParseBTPage(childData, db.Format(), disk.PageTypeBBT)
		if err != nil {
			return nil, err
		}
		childAllocs, err := collectAllocatedBlocks(db, childPage)
		if err != nil {
			return nil, err
		}
		allocs = append(allocs, childAllocs...)
	}

	return allocs, nil
}

// collectPageAllocations collects allocations for B-tree pages themselves.
func collectPageAllocations(db *Database) ([]disk.BlockAllocation, error) {
	var allocs []disk.BlockAllocation

	// Collect NBT page allocations
	nbtAllocs, err := collectBTreePageAllocations(db, db.nbtRoot, disk.PageTypeNBT)
	if err != nil {
		return nil, err
	}
	allocs = append(allocs, nbtAllocs...)

	// Collect BBT page allocations
	bbtAllocs, err := collectBTreePageAllocations(db, db.bbtRoot, disk.PageTypeBBT)
	if err != nil {
		return nil, err
	}
	allocs = append(allocs, bbtAllocs...)

	return allocs, nil
}

// collectBTreePageAllocations collects allocations for pages in a B-tree.
func collectBTreePageAllocations(db *Database, page *disk.BTPage, pageType disk.PageType) ([]disk.BlockAllocation, error) {
	var allocs []disk.BlockAllocation

	// Add this page's allocation
	// Note: We need to track page offsets, which are stored in the parent's reference
	// For the root page, use the header's reference

	if !page.IsLeaf() {
		for _, child := range page.NonleafEntries {
			// Add child page allocation
			allocs = append(allocs, disk.BlockAllocation{
				Offset: child.Ref.IB,
				Size:   disk.PageSize,
			})

			// Recurse
			childData, err := db.readPage(child.Ref.IB)
			if err != nil {
				return nil, err
			}
			childPage, err := disk.ParseBTPage(childData, db.Format(), pageType)
			if err != nil {
				return nil, err
			}
			childAllocs, err := collectBTreePageAllocations(db, childPage, pageType)
			if err != nil {
				return nil, err
			}
			allocs = append(allocs, childAllocs...)
		}
	}

	return allocs, nil
}

// getAMapPages returns the AMap pages from the manager for writing.
// This is a helper that accesses internal state.
func getAMapPages(m *disk.AMapManager) []*disk.AMapPage {
	// Note: This would need to be exposed from the AMapManager
	// For now, return empty - the AMapManager would need a Pages() method
	return nil
}

// ValidateIntegrity performs integrity checks on the PST file.
// Returns a list of issues found.
func ValidateIntegrity(db *Database) ([]string, error) {
	var issues []string

	// Check AMap status
	status := db.header.GetAMapStatus()
	if status == disk.AMapStatusInvalid {
		issues = append(issues, "AMap status is Invalid - file may be corrupted")
	}

	// Load and validate B-trees
	if err := db.loadNBTRoot(); err != nil {
		issues = append(issues, fmt.Sprintf("Failed to load NBT root: %v", err))
	} else {
		nbtIssues := validateBTreeIntegrity(db, db.nbtRoot, disk.PageTypeNBT)
		issues = append(issues, nbtIssues...)
	}

	if err := db.loadBBTRoot(); err != nil {
		issues = append(issues, fmt.Sprintf("Failed to load BBT root: %v", err))
	} else {
		bbtIssues := validateBTreeIntegrity(db, db.bbtRoot, disk.PageTypeBBT)
		issues = append(issues, bbtIssues...)
	}

	// Verify file size
	if db.header.Root.IBFileEOF == 0 {
		issues = append(issues, "File size in header is zero")
	}

	return issues, nil
}

// validateBTreeIntegrity validates a B-tree's structure.
func validateBTreeIntegrity(db *Database, page *disk.BTPage, pageType disk.PageType) []string {
	var issues []string

	// Verify page type
	if page.Trailer.PageType != pageType {
		issues = append(issues, fmt.Sprintf("Page type mismatch: expected %v, got %v", pageType, page.Trailer.PageType))
	}

	// Verify page type repeat
	if page.Trailer.PageType != page.Trailer.PageTypeRepeat {
		issues = append(issues, fmt.Sprintf("Page type repeat mismatch: %v != %v", page.Trailer.PageType, page.Trailer.PageTypeRepeat))
	}

	// For non-leaf pages, recurse into children
	if !page.IsLeaf() {
		for i, child := range page.NonleafEntries {
			childData, err := db.readPage(child.Ref.IB)
			if err != nil {
				issues = append(issues, fmt.Sprintf("Failed to read child page %d at 0x%X: %v", i, child.Ref.IB, err))
				continue
			}
			childPage, err := disk.ParseBTPage(childData, db.Format(), pageType)
			if err != nil {
				issues = append(issues, fmt.Sprintf("Failed to parse child page %d at 0x%X: %v", i, child.Ref.IB, err))
				continue
			}

			// Verify level decreases
			if childPage.Level >= page.Level {
				issues = append(issues, fmt.Sprintf("Child level %d >= parent level %d", childPage.Level, page.Level))
			}

			childIssues := validateBTreeIntegrity(db, childPage, pageType)
			issues = append(issues, childIssues...)
		}
	}

	return issues
}

// RepairCorruption attempts to repair common corruption issues.
// This is a best-effort operation and may not fix all issues.
func RepairCorruption(db *Database) (*RecoveryResult, error) {
	// Force recovery regardless of status
	result := &RecoveryResult{
		PreviousStatus: db.header.GetAMapStatus(),
		WasCorrupted:   true,
	}

	if err := performRecovery(db, result); err != nil {
		return result, err
	}

	return result, nil
}
