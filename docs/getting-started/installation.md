# Installation

## Requirements

- Go 1.23 or later (for iterator support)

## Install the Library

```bash
go get github.com/grokify/outlook-pst-go
```

## Install the CLI Tool

```bash
go install github.com/grokify/outlook-pst-go/cmd/pstinfo@latest
```

## Import in Your Code

```go
import outlookpst "github.com/grokify/outlook-pst-go"
```

## Verify Installation

Create a simple test program:

```go
package main

import (
    "fmt"
    "log"

    outlookpst "github.com/grokify/outlook-pst-go"
)

func main() {
    pst, err := outlookpst.Open("test.pst")
    if err != nil {
        log.Fatal(err)
    }
    defer pst.Close()

    fmt.Printf("Format: %s\n", pst.Format())
    fmt.Printf("Encryption: %s\n", pst.CryptMethod())
}
```

Run it:

```bash
go run main.go
```
