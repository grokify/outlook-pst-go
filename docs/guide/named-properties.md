# Named Properties

Named properties allow applications to define custom properties beyond the standard MAPI property set. They are identified by a GUID (property set) and either a numeric ID or a string name.

## Understanding Named Properties

Standard MAPI properties use fixed property IDs (e.g., `0x0037` for Subject). Named properties extend this by mapping custom identifiers to property IDs in the range `0x8000` to `0xFFFE`.

### Property Identification

Named properties can be identified in two ways:

- **MNID_ID**: By a numeric ID within a property set GUID
- **MNID_STRING**: By a string name within a property set GUID

## Accessing the Named Property Map

```go
pst, _ := outlookpst.Open("archive.pst")
defer pst.Close()

// Get the named property map
npm, err := pst.NamedPropertyMap()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Named properties: %d\n", npm.Count())
```

## Looking Up Named Properties

### By Numeric ID

```go
// Look up a property by GUID and numeric ID
np, found := npm.Lookup(util.PSETID_Common, 0x8501)
if found {
    fmt.Printf("Property: %s\n", np.String())
    fmt.Printf("Mapped to: 0x%04X\n", np.PropID)
}
```

### By String Name

```go
// Look up a property by GUID and string name
np, found := npm.LookupByName(util.PS_PUBLIC_STRINGS, "Keywords")
if found {
    fmt.Printf("Property: %s\n", np.String())
}
```

### By Property ID

```go
// Look up by the mapped property ID
np, found := npm.LookupByPropID(0x8005)
if found {
    fmt.Printf("GUID: %s\n", np.GUID.String())
    if np.Kind == ltp.MNID_ID {
        fmt.Printf("ID: 0x%04X\n", np.ID)
    } else {
        fmt.Printf("Name: %s\n", np.Name)
    }
}
```

## Well-Known Property Sets

The library includes common property set GUIDs:

| Variable | Description |
|----------|-------------|
| `util.PS_MAPI` | Standard MAPI properties |
| `util.PS_PUBLIC_STRINGS` | Public string properties |
| `util.PS_INTERNET_HEADERS` | Internet mail headers |
| `util.PSETID_Common` | Common message properties |
| `util.PSETID_Address` | Address book properties |
| `util.PSETID_Appointment` | Calendar/appointment properties |
| `util.PSETID_Meeting` | Meeting request properties |
| `util.PSETID_Task` | Task properties |
| `util.PSETID_Note` | Note properties |

## Reading Named Property Values

Once you have the mapped property ID, use it with the PropertyBag:

```go
// Get the mapped property ID
propID, found := npm.GetPropID(util.PSETID_Common, 0x8501)
if !found {
    log.Fatal("Property not found")
}

// Use it to read the property value
msg, _ := folder.Messages().Next()
bag := msg.PropertyBag()

if bag.Exists(propID) {
    value, _ := bag.GetString(propID)
    fmt.Printf("Value: %s\n", value)
}
```

## Listing All Named Properties

```go
npm, _ := pst.NamedPropertyMap()

for _, np := range npm.Entries() {
    if np.Kind == ltp.MNID_ID {
        fmt.Printf("{%s}:0x%04X -> 0x%04X\n",
            np.GUID.String(), np.ID, np.PropID)
    } else {
        fmt.Printf("{%s}:%q -> 0x%04X\n",
            np.GUID.String(), np.Name, np.PropID)
    }
}
```

## Common Named Properties

Some frequently used named properties:

| Property Set | ID/Name | Description |
|--------------|---------|-------------|
| PSETID_Common | 0x8501 | Reminder set flag |
| PSETID_Common | 0x8502 | Reminder time |
| PSETID_Common | 0x8503 | Reminder delta |
| PSETID_Appointment | 0x8205 | Appointment duration |
| PSETID_Appointment | 0x8208 | Location |
| PSETID_Task | 0x8101 | Task status |
| PS_PUBLIC_STRINGS | "Keywords" | Categories |

## Example: Reading Categories

```go
func getCategories(pst *outlookpst.PST, msg *outlookpst.Message) ([]string, error) {
    npm, err := pst.NamedPropertyMap()
    if err != nil {
        return nil, err
    }

    // Keywords/Categories is stored as a multi-value string
    propID, found := npm.GetPropIDByName(util.PS_PUBLIC_STRINGS, "Keywords")
    if !found {
        return nil, nil
    }

    bag := msg.PropertyBag()
    if !bag.Exists(propID) {
        return nil, nil
    }

    return bag.GetStringSlice(propID)
}
```

## Specification Reference

Named properties are defined in:

- [MS-PST] Section 2.4.7 - Name-to-ID Map
- [MS-OXPROPS] - Property definitions
