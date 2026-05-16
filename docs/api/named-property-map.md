# NamedPropertyMap

The `NamedPropertyMap` type maps named properties to property IDs.

## Overview

Named properties extend MAPI by allowing custom properties identified by:

- A GUID (property set)
- Either a numeric ID or string name

These are mapped to property IDs in the range `0x8000` to `0xFFFE`.

## Type Definitions

### NamedProperty

```go
type NamedProperty struct {
    GUID   util.GUID          // Property set GUID
    Kind   NamedPropertyKind  // MNID_ID or MNID_STRING
    ID     uint32             // Numeric ID (when Kind == MNID_ID)
    Name   string             // String name (when Kind == MNID_STRING)
    PropID PropID             // Mapped property ID (0x8000+)
}
```

### NamedPropertyKind

```go
type NamedPropertyKind uint8

const (
    MNID_ID     NamedPropertyKind = 0x00  // Numeric identifier
    MNID_STRING NamedPropertyKind = 0x01  // String identifier
)
```

### NamedPropertyMap

```go
type NamedPropertyMap struct {
    // internal fields
}
```

## Accessing the Map

```go
pst, _ := outlookpst.Open("archive.pst")
npm, err := pst.NamedPropertyMap()
if err != nil {
    log.Fatal(err)
}
```

## Methods

### Lookup

```go
func (npm *NamedPropertyMap) Lookup(guid util.GUID, id uint32) (*NamedProperty, bool)
```

Looks up a named property by GUID and numeric ID.

```go
np, found := npm.Lookup(util.PSETID_Common, 0x8501)
if found {
    fmt.Printf("Mapped to: 0x%04X\n", np.PropID)
}
```

### LookupByName

```go
func (npm *NamedPropertyMap) LookupByName(guid util.GUID, name string) (*NamedProperty, bool)
```

Looks up a named property by GUID and string name.

```go
np, found := npm.LookupByName(util.PS_PUBLIC_STRINGS, "Keywords")
if found {
    fmt.Printf("Mapped to: 0x%04X\n", np.PropID)
}
```

### LookupByPropID

```go
func (npm *NamedPropertyMap) LookupByPropID(propID PropID) (*NamedProperty, bool)
```

Looks up a named property by its mapped property ID.

```go
np, found := npm.LookupByPropID(0x8005)
if found {
    fmt.Printf("GUID: %s\n", np.GUID.String())
}
```

### GetPropID

```go
func (npm *NamedPropertyMap) GetPropID(guid util.GUID, id uint32) (PropID, bool)
```

Returns the mapped property ID for a GUID and numeric ID.

### GetPropIDByName

```go
func (npm *NamedPropertyMap) GetPropIDByName(guid util.GUID, name string) (PropID, bool)
```

Returns the mapped property ID for a GUID and string name.

### Entries

```go
func (npm *NamedPropertyMap) Entries() []*NamedProperty
```

Returns all named property entries.

### Count

```go
func (npm *NamedPropertyMap) Count() int
```

Returns the number of named properties.

## Well-Known GUIDs

The `util` package provides common property set GUIDs:

| Variable | GUID | Description |
|----------|------|-------------|
| `PS_MAPI` | 00020328-... | Standard MAPI |
| `PS_PUBLIC_STRINGS` | 00020329-... | Public strings |
| `PS_INTERNET_HEADERS` | 00020386-... | Internet headers |
| `PSETID_Common` | 00062008-... | Common properties |
| `PSETID_Address` | 00062004-... | Address book |
| `PSETID_Appointment` | 00062002-... | Calendar |
| `PSETID_Meeting` | 6ED8DA90-... | Meetings |
| `PSETID_Task` | 00062003-... | Tasks |
| `PSETID_Note` | 0006200E-... | Notes |

## Example

```go
func listNamedProperties(pst *outlookpst.PST) {
    npm, err := pst.NamedPropertyMap()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Total named properties: %d\n\n", npm.Count())

    // Group by GUID
    byGUID := make(map[string][]*ltp.NamedProperty)

    for _, np := range npm.Entries() {
        key := np.GUID.String()
        byGUID[key] = append(byGUID[key], np)
    }

    for guid, props := range byGUID {
        fmt.Printf("GUID: %s (%d properties)\n", guid, len(props))
        for _, np := range props {
            if np.Kind == ltp.MNID_ID {
                fmt.Printf("  0x%04X -> 0x%04X\n", np.ID, np.PropID)
            } else {
                fmt.Printf("  %q -> 0x%04X\n", np.Name, np.PropID)
            }
        }
        fmt.Println()
    }
}
```

## Using Named Properties

```go
func readNamedProperty(pst *outlookpst.PST, msg *outlookpst.Message,
    guid util.GUID, id uint32) (string, error) {

    npm, err := pst.NamedPropertyMap()
    if err != nil {
        return "", err
    }

    propID, found := npm.GetPropID(guid, id)
    if !found {
        return "", fmt.Errorf("property not found")
    }

    return msg.PropertyBag().GetString(propID)
}
```

## Specification Reference

- [MS-PST] Section 2.4.7 - Name-to-ID Map
- [MS-OXPROPS] - Property definitions
