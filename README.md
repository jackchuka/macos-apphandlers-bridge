# macos-apphandlers-bridge

[![Go Reference](https://pkg.go.dev/badge/github.com/jackchuka/macos-apphandlers-bridge.svg)](https://pkg.go.dev/github.com/jackchuka/macos-apphandlers-bridge)
[![Test](https://github.com/jackchuka/macos-apphandlers-bridge/workflows/Test/badge.svg)](https://github.com/jackchuka/macos-apphandlers-bridge/actions)

CGO bridge to macOS NSWorkspace APIs for managing application handlers (UTIs and URL schemes).

Provides Go bindings to:

- Query and set default applications for UTIs (file types)
- Query and set default applications for URL schemes
- Resolve file extensions to UTIs
- List all capable application handlers

## Requirements

- macOS 12.0 (Monterey) or later
- CGO enabled (`CGO_ENABLED=1`)
- Xcode Command Line Tools installed
- Go 1.21 or later

## Installation

```bash
go get github.com/jackchuka/macos-apphandlers-bridge
```

## Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/jackchuka/macos-apphandlers-bridge"
)

func main() {
    // Get the default application for plain text files
    appPath, err := bridge.GetDefaultAppForUTI("public.plain-text")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Default app for text files: %s\n", appPath)

    // Get the default browser for HTTP URLs
    browser, err := bridge.GetDefaultAppForScheme("http")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Default browser: %s\n", browser)

    // Resolve file extension to UTIs
    utis, err := bridge.ResolveUTIsForExtension("txt")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("UTIs for .txt: %v\n", utis)

    // List all apps that can open HTML files
    apps, err := bridge.ListAppsForUTI("public.html")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Apps that can open HTML: %v\n", apps)

    // Set default application for a UTI
    err = bridge.SetDefaultForUTI("/Applications/Visual Studio Code.app", "public.plain-text")
    if err != nil {
        log.Fatal(err)
    }

    // Set default application for a URL scheme
    err = bridge.SetDefaultForScheme("/Applications/Firefox.app", "http")
    if err != nil {
        log.Fatal(err)
    }
}
```

## API Reference

### Query Functions

#### `GetDefaultAppForUTI(uti string) (string, error)`

Returns the default application path for a given UTI (Uniform Type Identifier).

**Parameters:**

- `uti` - The Uniform Type Identifier (e.g., "public.plain-text", "public.jpeg")

**Returns:**

- Full path to the default application bundle
- Error if UTI is invalid or no default app is set

**Example:**

```go
appPath, err := bridge.GetDefaultAppForUTI("public.plain-text")
// Returns: "/System/Applications/TextEdit.app"
```

#### `GetDefaultAppForScheme(scheme string) (string, error)`

Returns the default application path for a given URL scheme.

**Parameters:**

- `scheme` - The URL scheme (e.g., "http", "https", "mailto")

**Returns:**

- Full path to the default application bundle
- Error if scheme is invalid or no default app is set

**Example:**

```go
appPath, err := bridge.GetDefaultAppForScheme("http")
// Returns: "/Applications/Safari.app"
```

#### `ResolveUTIsForExtension(extension string) ([]string, error)`

Resolves a file extension to one or more UTI identifiers.

**Parameters:**

- `extension` - File extension without the dot (e.g., "txt", "jpg", "html")

**Returns:**

- Slice of UTI identifiers that match the extension
- Error if extension is invalid

**Example:**

```go
utis, err := bridge.ResolveUTIsForExtension("txt")
// Returns: ["public.plain-text", "public.text"]
```

#### `ListAppsForUTI(uti string) ([]string, error)`

Returns all applications capable of opening a given UTI.

**Parameters:**

- `uti` - The Uniform Type Identifier

**Returns:**

- Slice of application bundle paths
- Error if UTI is invalid

**Example:**

```go
apps, err := bridge.ListAppsForUTI("public.html")
// Returns: ["/Applications/Safari.app", "/Applications/Google Chrome.app", ...]
```

#### `ListAppsForScheme(scheme string) ([]string, error)`

Returns all applications capable of handling a given URL scheme.

**Parameters:**

- `scheme` - The URL scheme

**Returns:**

- Slice of application bundle paths
- Error if scheme is invalid

**Example:**

```go
apps, err := bridge.ListAppsForScheme("http")
// Returns: ["/Applications/Safari.app", "/Applications/Firefox.app", ...]
```

### Setter Functions

#### `SetDefaultForUTI(appPath, uti string) error`

Sets the default application for a given UTI. This operation may prompt the user for confirmation.

**Parameters:**

- `appPath` - Full path to the application bundle
- `uti` - The Uniform Type Identifier

**Returns:**

- Error if operation fails or user declines the change

**Example:**

```go
err := bridge.SetDefaultForUTI("/Applications/TextEdit.app", "public.plain-text")
```

#### `SetDefaultForScheme(appPath, scheme string) error`

Sets the default application for a given URL scheme. This operation may prompt the user for confirmation.

**Parameters:**

- `appPath` - Full path to the application bundle
- `scheme` - The URL scheme

**Returns:**

- Error if operation fails or user declines the change

**Example:**

```go
err := bridge.SetDefaultForScheme("/Applications/Firefox.app", "http")
```

## Error Handling

The package provides structured error types:

```go
type BridgeError struct {
    Code    int
    Message string
}
```

**Error Codes:**

- `ErrOK` - Success (0)
- `ErrInvalidApp` - Invalid application path
- `ErrInvalidUTI` - Invalid or unknown UTI
- `ErrInvalidScheme` - Invalid URL scheme
- `ErrSystem` - System error occurred
- `ErrUserDeclined` - User declined the permission prompt
- `ErrNotFound` - No handler found

**Common Errors:**

- `ErrInvalidParameters` - Invalid input parameters
- `ErrMemoryAllocation` - Memory allocation failed

**Example:**

```go
appPath, err := bridge.GetDefaultAppForUTI("invalid.uti")
if err != nil {
    if bridgeErr, ok := err.(*bridge.BridgeError); ok {
        switch bridgeErr.Code {
        case bridge.ErrInvalidUTI:
            fmt.Println("Invalid UTI provided")
        case bridge.ErrNotFound:
            fmt.Println("No default app found")
        default:
            fmt.Printf("Error: %s\n", bridgeErr.Message)
        }
    }
}
```

## Common UTI Examples

| File Type   | UTI Identifier                |
| ----------- | ----------------------------- |
| Plain Text  | `public.plain-text`           |
| JPEG Image  | `public.jpeg`                 |
| PNG Image   | `public.png`                  |
| PDF         | `com.adobe.pdf`               |
| HTML        | `public.html`                 |
| Markdown    | `net.daringfireball.markdown` |
| JSON        | `public.json`                 |
| ZIP Archive | `public.zip-archive`          |
| MP3 Audio   | `public.mp3`                  |
| MPEG Video  | `public.mpeg`                 |

## Common URL Schemes

| Scheme           | Purpose        |
| ---------------- | -------------- |
| `http` / `https` | Web browsers   |
| `mailto`         | Email clients  |
| `tel`            | Phone/FaceTime |
| `ftp`            | FTP clients    |
| `ssh`            | SSH clients    |
| `file`           | File handlers  |

## Building

The package uses CGO and requires the following frameworks:

- Foundation
- AppKit
- UniformTypeIdentifiers

Build with:

```bash
CGO_ENABLED=1 go build
```

## Testing

Run tests:

```bash
go test -v ./...
```

Note: Some tests (especially `SetDefaultFor*`) may trigger macOS permission prompts.

## Platform Support

This package is macOS-only and uses the `//go:build darwin` build constraint.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please ensure:

- Tests pass on macOS
- Code follows Go conventions
- CGO memory management is correct
- Changes maintain backward compatibility
