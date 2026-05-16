package outlookpst

import (
	"fmt"
	"sync"

	"github.com/grokify/outlook-pst-go/pkg/disk"
	"github.com/grokify/outlook-pst-go/pkg/ltp"
	"github.com/grokify/outlook-pst-go/pkg/ndb"
	"github.com/grokify/outlook-pst-go/pkg/util"
)

// PST represents an open PST file.
type PST struct {
	db *ndb.Database

	// Message store property bag (lazy loaded)
	storeOnce sync.Once
	storeBag  *ltp.PropertyBag
	storeErr  error

	// Root folder (lazy loaded)
	rootOnce   sync.Once
	rootFolder *Folder
	rootErr    error

	// Named property map (lazy loaded)
	namedPropOnce sync.Once
	namedPropMap  *ltp.NamedPropertyMap
	namedPropErr  error
}

// Open opens a PST file for reading.
func Open(filename string) (*PST, error) {
	db, err := ndb.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open PST: %w", err)
	}

	return &PST{db: db}, nil
}

// Close closes the PST file.
func (p *PST) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Format returns the PST format (ANSI or Unicode).
func (p *PST) Format() disk.PSTFormat {
	return p.db.Format()
}

// IsUnicode returns true if the PST uses Unicode format.
func (p *PST) IsUnicode() bool {
	return p.db.Format() == disk.FormatUnicode
}

// IsANSI returns true if the PST uses ANSI format.
func (p *PST) IsANSI() bool {
	return p.db.Format() == disk.FormatANSI
}

// IsPST returns true if this is a PST file (not OST).
func (p *PST) IsPST() bool {
	return p.db.Header().IsPST()
}

// IsOST returns true if this is an OST file.
func (p *PST) IsOST() bool {
	return p.db.Header().IsOST()
}

// CryptMethod returns the encryption method used.
func (p *PST) CryptMethod() disk.CryptMethod {
	return p.db.CryptMethod()
}

// loadMessageStore loads the message store property bag.
func (p *PST) loadMessageStore() error {
	p.storeOnce.Do(func() {
		node, err := p.db.GetNode(util.NIDMessageStore)
		if err != nil {
			p.storeErr = fmt.Errorf("failed to get message store node: %w", err)
			return
		}

		p.storeBag, err = ltp.NewPropertyBag(node)
		if err != nil {
			p.storeErr = fmt.Errorf("failed to create message store property bag: %w", err)
			return
		}
	})
	return p.storeErr
}

// Name returns the display name of the PST file.
func (p *PST) Name() (string, error) {
	if err := p.loadMessageStore(); err != nil {
		return "", err
	}

	return p.storeBag.GetString(ltp.PidTagDisplayName)
}

// RootFolder returns the root folder of the PST.
func (p *PST) RootFolder() (*Folder, error) {
	p.rootOnce.Do(func() {
		node, err := p.db.GetNode(util.NIDRootFolder)
		if err != nil {
			p.rootErr = fmt.Errorf("failed to get root folder node: %w", err)
			return
		}

		p.rootFolder, err = newFolder(p, node)
		if err != nil {
			p.rootErr = fmt.Errorf("failed to create root folder: %w", err)
			return
		}
	})

	return p.rootFolder, p.rootErr
}

// OpenFolder opens a folder by name (searches from root).
func (p *PST) OpenFolder(name string) (*Folder, error) {
	root, err := p.RootFolder()
	if err != nil {
		return nil, err
	}

	return root.FindSubfolder(name)
}

// Database returns the underlying NDB database.
// This is useful for advanced operations.
func (p *PST) Database() *ndb.Database {
	return p.db
}

// MessageStore returns the message store property bag.
// This provides access to PST-level properties.
func (p *PST) MessageStore() (*ltp.PropertyBag, error) {
	if err := p.loadMessageStore(); err != nil {
		return nil, err
	}
	return p.storeBag, nil
}

// NamedPropertyMap returns the named property map.
// Named properties allow custom properties beyond the standard MAPI set.
// See [MS-PST] Section 2.4.7.
func (p *PST) NamedPropertyMap() (*ltp.NamedPropertyMap, error) {
	p.namedPropOnce.Do(func() {
		node, err := p.db.GetNode(util.NIDNameToIDMap)
		if err != nil {
			p.namedPropErr = fmt.Errorf("failed to get name-to-ID map node: %w", err)
			return
		}

		p.namedPropMap, err = ltp.NewNamedPropertyMap(node)
		if err != nil {
			p.namedPropErr = fmt.Errorf("failed to create named property map: %w", err)
			return
		}
	})

	return p.namedPropMap, p.namedPropErr
}
