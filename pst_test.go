package outlookpst

import (
	"os"
	"testing"

	"github.com/grokify/outlook-pst-go/pkg/disk"
)

const samplePSTPath = "sample.pst"

func TestOpen(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	// Test basic properties
	if pst.db == nil {
		t.Error("database is nil")
	}
}

func TestFormat(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	format := pst.Format()
	if format != disk.FormatUnicode && format != disk.FormatANSI {
		t.Errorf("unexpected format: %v", format)
	}

	// Verify consistent with IsUnicode/IsANSI
	if format == disk.FormatUnicode {
		if !pst.IsUnicode() {
			t.Error("IsUnicode() should return true for Unicode format")
		}
		if pst.IsANSI() {
			t.Error("IsANSI() should return false for Unicode format")
		}
	} else {
		if pst.IsUnicode() {
			t.Error("IsUnicode() should return false for ANSI format")
		}
		if !pst.IsANSI() {
			t.Error("IsANSI() should return true for ANSI format")
		}
	}
}

func TestCryptMethod(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	crypt := pst.CryptMethod()
	// Valid methods: None (0), Permute (1), Cyclic (2)
	if crypt > disk.CryptMethodCyclic {
		t.Errorf("unexpected crypt method: %v", crypt)
	}
	t.Logf("Crypt method: %v", crypt)
}

func TestIsPSTOrOST(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	// Should be exactly one of PST or OST
	if pst.IsPST() == pst.IsOST() {
		t.Error("IsPST() and IsOST() should not return the same value")
	}
}

func TestName(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	name, err := pst.Name()
	if err != nil {
		// Property may not exist in some PST files
		t.Logf("PST name not found (may be expected): %v", err)
	} else {
		t.Logf("PST name: %q", name)
	}
}

func TestRootFolder(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	root, err := pst.RootFolder()
	if err != nil {
		t.Fatalf("failed to get root folder: %v", err)
	}

	if root == nil {
		t.Fatal("root folder is nil")
	}

	// Test root folder ID is valid
	if root.ID() == 0 {
		t.Error("root folder ID is zero")
	}
}

func TestSubfolders(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	root, err := pst.RootFolder()
	if err != nil {
		t.Fatalf("failed to get root folder: %v", err)
	}

	// Count subfolders
	count := 0
	for subfolder, err := range root.Subfolders() {
		if err != nil {
			t.Errorf("error iterating subfolders: %v", err)
			break
		}
		name, _ := subfolder.Name()
		t.Logf("Subfolder: %q (ID: %d)", name, subfolder.ID())
		count++
	}
	t.Logf("Total subfolders: %d", count)
}

func TestSubfolderCount(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	root, err := pst.RootFolder()
	if err != nil {
		t.Fatalf("failed to get root folder: %v", err)
	}

	count, err := root.SubfolderCount()
	if err != nil {
		t.Errorf("failed to get subfolder count: %v", err)
	}
	t.Logf("Subfolder count: %d", count)

	// Verify count matches iteration
	iterCount := 0
	for _, iterErr := range root.Subfolders() {
		if iterErr != nil {
			break
		}
		iterCount++
	}
	if count != iterCount {
		t.Errorf("SubfolderCount() = %d, but iteration found %d", count, iterCount)
	}
}

func TestFolderProperties(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	root, err := pst.RootFolder()
	if err != nil {
		t.Fatalf("failed to get root folder: %v", err)
	}

	// Test ContentCount
	contentCount, err := root.ContentCount()
	if err != nil {
		t.Logf("ContentCount error (may be expected): %v", err)
	} else {
		t.Logf("Content count: %d", contentCount)
	}

	// Test UnreadCount
	unreadCount, err := root.UnreadCount()
	if err != nil {
		t.Logf("UnreadCount error (may be expected): %v", err)
	} else {
		t.Logf("Unread count: %d", unreadCount)
	}

	// Test HasSubfolders
	hasSubfolders, err := root.HasSubfolders()
	if err != nil {
		t.Logf("HasSubfolders error (may be expected): %v", err)
	} else {
		t.Logf("Has subfolders: %v", hasSubfolders)
	}

	// Test ContainerClass
	containerClass, err := root.ContainerClass()
	if err != nil {
		t.Logf("ContainerClass error (may be expected): %v", err)
	} else {
		t.Logf("Container class: %q", containerClass)
	}
}

func TestMessageStore(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	store, err := pst.MessageStore()
	if err != nil {
		t.Fatalf("failed to get message store: %v", err)
	}

	if store == nil {
		t.Fatal("message store is nil")
	}
}

func TestDatabase(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	db := pst.Database()
	if db == nil {
		t.Fatal("database is nil")
	}
}

func TestOpenFolder(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	// Get root folder first to find a subfolder name
	root, err := pst.RootFolder()
	if err != nil {
		t.Fatalf("failed to get root folder: %v", err)
	}

	// Find first subfolder name
	var firstSubfolderName string
	for subfolder, err := range root.Subfolders() {
		if err != nil {
			break
		}
		firstSubfolderName, _ = subfolder.Name()
		break
	}

	if firstSubfolderName == "" {
		t.Skip("no subfolders found")
	}

	// Test OpenFolder
	folder, err := pst.OpenFolder(firstSubfolderName)
	if err != nil {
		t.Errorf("failed to open folder %q: %v", firstSubfolderName, err)
	} else {
		name, _ := folder.Name()
		t.Logf("Successfully opened folder: %q", name)
	}
}

func TestOpenFolderNotFound(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	_, err = pst.OpenFolder("NonExistentFolder_12345")
	if err == nil {
		t.Error("expected error for non-existent folder")
	}
}

func TestMessages(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}
	defer func() { _ = pst.Close() }()

	root, err := pst.RootFolder()
	if err != nil {
		t.Fatalf("failed to get root folder: %v", err)
	}

	// Find a folder with messages
	var folderWithMessages *Folder
	for subfolder, err := range root.Subfolders() {
		if err != nil {
			continue
		}
		count, _ := subfolder.ContentCount()
		if count > 0 {
			folderWithMessages = subfolder
			break
		}
	}

	if folderWithMessages == nil {
		t.Skip("no folder with messages found")
	}

	name, _ := folderWithMessages.Name()
	t.Logf("Testing messages in folder: %q", name)

	// Iterate messages
	msgCount := 0
	for msg, err := range folderWithMessages.Messages() {
		if err != nil {
			t.Errorf("error iterating messages: %v", err)
			break
		}
		subject, _ := msg.Subject()
		t.Logf("Message: %q", subject)
		msgCount++
		if msgCount >= 5 {
			t.Logf("(showing first 5 messages only)")
			break
		}
	}
	t.Logf("Found %d messages", msgCount)
}

func TestClose(t *testing.T) {
	if _, err := os.Stat(samplePSTPath); os.IsNotExist(err) {
		t.Skip("sample.pst not found")
	}

	pst, err := Open(samplePSTPath)
	if err != nil {
		t.Fatalf("failed to open PST: %v", err)
	}

	// Close should not return error
	if err := pst.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Double close may return an error (file already closed), which is acceptable
	_ = pst.Close()
}

func TestOpenNonExistent(t *testing.T) {
	_, err := Open("nonexistent_file_12345.pst")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
