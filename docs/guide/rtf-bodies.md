# RTF Message Bodies

Email messages can have multiple body formats: plain text, HTML, and RTF. The RTF body is typically stored in compressed format using the LZFu algorithm.

## Body Formats

Messages can contain up to three body formats:

| Method | Property | Description |
|--------|----------|-------------|
| `Body()` | PidTagBody | Plain text body |
| `HTMLBody()` | PidTagHtmlBody | HTML formatted body |
| `RTFBody()` | PidTagRtfCompressed | Compressed RTF body |
| `RTFBodyDecompressed()` | - | Decompressed RTF body |

## Reading the RTF Body

### Compressed (Raw)

```go
msg, _ := folder.Messages().Next()

// Get compressed RTF data
compressed, err := msg.RTFBody()
if err != nil {
    log.Printf("No RTF body: %v", err)
} else {
    fmt.Printf("Compressed size: %d bytes\n", len(compressed))
}
```

### Decompressed

```go
// Get decompressed RTF as string
rtfContent, err := msg.RTFBodyDecompressed()
if err != nil {
    log.Printf("Failed to decompress: %v", err)
} else {
    fmt.Printf("RTF content:\n%s\n", rtfContent)
}
```

## RTF Compression Format

The RTF body uses Microsoft's compressed RTF format (MS-OXRTFCP):

- **LZFu compression**: LZ77-based algorithm with preloaded dictionary
- **MELA format**: Uncompressed RTF with header wrapper

### Compression Header

```go
import "github.com/grokify/outlook-pst-go/pkg/rtf"

compressed, _ := msg.RTFBody()

// Parse the header
header, err := rtf.ParseHeader(compressed)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Compressed size: %d\n", header.CompSize)
fmt.Printf("Raw size: %d\n", header.RawSize)
fmt.Printf("Is compressed: %v\n", header.IsCompressed())
fmt.Printf("CRC: 0x%08X\n", header.CRC)
```

### Direct Decompression

```go
import "github.com/grokify/outlook-pst-go/pkg/rtf"

// Check if data is compressed
if rtf.IsCompressed(data) {
    decompressed, err := rtf.Decompress(data)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(decompressed))
}
```

## Handling Missing Bodies

Not all messages have all body formats. Always handle errors:

```go
func getMessageBody(msg *outlookpst.Message) string {
    // Try plain text first (most common)
    if body, err := msg.Body(); err == nil && body != "" {
        return body
    }

    // Try HTML
    if html, err := msg.HTMLBody(); err == nil && html != "" {
        return html
    }

    // Try RTF (decompressed)
    if rtf, err := msg.RTFBodyDecompressed(); err == nil && rtf != "" {
        return rtf
    }

    return "(no body)"
}
```

## RTF to Text Conversion

The library provides raw RTF content. For text extraction, you may need additional processing:

```go
func extractTextFromRTF(rtfContent string) string {
    // Simple approach: strip RTF control words
    // For production use, consider a dedicated RTF parser

    var result strings.Builder
    inControl := false

    for _, r := range rtfContent {
        if r == '\\' {
            inControl = true
            continue
        }
        if inControl {
            if r == ' ' || r == '\n' {
                inControl = false
            }
            continue
        }
        if r == '{' || r == '}' {
            continue
        }
        result.WriteRune(r)
    }

    return result.String()
}
```

## Example: Export All Body Formats

```go
func exportMessageBodies(msg *outlookpst.Message, basePath string) error {
    subject, _ := msg.Subject()
    safeName := sanitizeFilename(subject)

    // Plain text
    if body, err := msg.Body(); err == nil && body != "" {
        path := filepath.Join(basePath, safeName+".txt")
        os.WriteFile(path, []byte(body), 0644)
    }

    // HTML
    if html, err := msg.HTMLBody(); err == nil && html != "" {
        path := filepath.Join(basePath, safeName+".html")
        os.WriteFile(path, []byte(html), 0644)
    }

    // RTF
    if rtf, err := msg.RTFBodyDecompressed(); err == nil && rtf != "" {
        path := filepath.Join(basePath, safeName+".rtf")
        os.WriteFile(path, []byte(rtf), 0644)
    }

    return nil
}
```

## Specification Reference

RTF compression is defined in:

- [MS-OXRTFCP] - Rich Text Format (RTF) Compression Algorithm
