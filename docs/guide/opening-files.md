# Opening PST Files

## Basic Usage

```go
pst, err := outlookpst.Open("archive.pst")
if err != nil {
    log.Fatal(err)
}
defer pst.Close()
```

!!! warning "Always Close the PST"
    Always use `defer pst.Close()` to ensure the file is properly closed.

## Checking File Format

PST files come in two formats:

- **ANSI** (older, 32-bit): Limited to ~2GB file size
- **Unicode** (newer, 64-bit): No practical size limit

```go
if pst.IsUnicode() {
    fmt.Println("Unicode format (64-bit)")
} else {
    fmt.Println("ANSI format (32-bit)")
}

// Or get the format directly
fmt.Printf("Format: %s\n", pst.Format())
```

## PST vs OST

```go
if pst.IsPST() {
    fmt.Println("Personal Storage Table (PST)")
} else if pst.IsOST() {
    fmt.Println("Offline Storage Table (OST)")
}
```

## Encryption

PST files can use different encryption methods:

| Method | Description |
|--------|-------------|
| None | No encryption |
| Permute | Simple byte substitution |
| Cyclic | XOR-based encoding |

```go
switch pst.CryptMethod() {
case disk.CryptMethodNone:
    fmt.Println("No encryption")
case disk.CryptMethodPermute:
    fmt.Println("Permute encryption")
case disk.CryptMethodCyclic:
    fmt.Println("Cyclic encryption")
}
```

!!! note "Automatic Decryption"
    The library automatically handles decryption. You don't need to do anything special to read encrypted PST files.

## Error Handling

```go
pst, err := outlookpst.Open("archive.pst")
if err != nil {
    if errors.Is(err, os.ErrNotExist) {
        log.Fatal("File not found")
    }
    if errors.Is(err, outlookpst.ErrInvalidFormat) {
        log.Fatal("Not a valid PST file")
    }
    log.Fatalf("Failed to open: %v", err)
}
```

## Getting PST Properties

```go
// Display name (from message store)
name, err := pst.Name()
if err == nil {
    fmt.Printf("PST Name: %s\n", name)
}

// Access the message store property bag directly
store, err := pst.MessageStore()
if err == nil {
    // Read any property
    props := store.Properties()
    fmt.Printf("Message store has %d properties\n", len(props))
}
```

## Accessing the Root Folder

```go
root, err := pst.RootFolder()
if err != nil {
    log.Fatal(err)
}

name, _ := root.Name()
fmt.Printf("Root folder: %s\n", name)
```

## Opening a Specific Folder

```go
// Open folder by name (searches from root)
inbox, err := pst.OpenFolder("Inbox")
if err != nil {
    log.Fatal(err)
}
```
