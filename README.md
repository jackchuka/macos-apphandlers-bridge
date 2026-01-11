# macos-apphandlers-bridge

[![Go Reference](https://pkg.go.dev/badge/github.com/jackchuka/macos-apphandlers-bridge.svg)](https://pkg.go.dev/github.com/jackchuka/macos-apphandlers-bridge)
[![Test](https://github.com/jackchuka/macos-apphandlers-bridge/workflows/Test/badge.svg)](https://github.com/jackchuka/macos-apphandlers-bridge/actions)

CGO bridge to macOS NSWorkspace APIs for managing application handlers (UTIs and URL schemes).

Provides Go bindings to:

- Query and set default applications for UTIs (file types)
- Query and set default applications for URL schemes
- Resolve file extensions to UTIs
- List all capable application handlers
- List all installed applications with metadata (name, path, bundle ID)

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

    // Get extensions for a UTI
    extensions, err := bridge.ResolveExtensionsForUTI("public.html")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Extensions for public.html: %v\n", extensions)

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

    // List all installed applications
    apps, err := bridge.ListAllApplications()
    if err != nil {
        log.Fatal(err)
    }
    for _, app := range apps {
        fmt.Printf("%s - %s (%s)\n", app.Name, app.Path, app.BundleID)
    }

    // Get supported document types for an application
    docTypes, err := bridge.ListSupportedDocumentTypes("/System/Applications/TextEdit.app")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("TextEdit supports %d document types\n", len(docTypes))

    // Get document types where the app is the default
    defaults, err := bridge.ListDefaultDocumentTypes("/System/Applications/TextEdit.app")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("TextEdit is default for %d document types:\n", len(defaults))
    for _, dt := range defaults {
        fmt.Printf("  %s (%s): %v\n", dt.TypeName, dt.Role, dt.Extensions)
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

#### `ResolveExtensionsForUTI(uti string) ([]string, error)`

Returns all file extensions associated with a UTI. This is the inverse of `ResolveUTIsForExtension`.

**Parameters:**

- `uti` - The UTI string (e.g., "public.plain-text", "public.html")

**Returns:**

- Slice of file extensions (without dots) associated with this UTI
- Error if UTI is invalid

**Example:**

```go
extensions, err := bridge.ResolveExtensionsForUTI("public.html")
// Returns: ["htm", "html", "shtml"]

extensions, err := bridge.ResolveExtensionsForUTI("public.plain-text")
// Returns: ["text", "txt"]

// Some UTIs may have no file extensions
extensions, err := bridge.ResolveExtensionsForUTI("public.folder")
// Returns: []
```

**Use Cases:**

- Convert UTI identifiers to user-friendly file extensions
- Display file types supported by an app in readable format
- Validate file associations

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

#### `ListAllApplications() ([]AppInfo, error)`

Returns all installed applications on the system with their metadata.

**Returns:**

- Slice of `AppInfo` structures containing application metadata
- Error if operation fails

**AppInfo Structure:**

```go
type AppInfo struct {
    Name     string // Application display name
    Path     string // Full path to application bundle
    BundleID string // Bundle identifier (e.g., "com.apple.Safari")
}
```

**Example:**

```go
apps, err := bridge.ListAllApplications()
// Returns: []AppInfo{
//   {Name: "Safari", Path: "/Applications/Safari.app", BundleID: "com.apple.Safari"},
//   {Name: "Chrome", Path: "/Applications/Google Chrome.app", BundleID: "com.google.Chrome"},
//   ...
// }
```

#### `ListSupportedDocumentTypes(appPath string) ([]DocumentType, error)`

Returns all document types that an application can handle, with rich metadata about each type.

**Parameters:**

- `appPath` - Full path to the application bundle

**Returns:**

- Slice of `DocumentType` structures containing detailed info about supported file types
- Error if app path is invalid

**DocumentType Structure:**

```go
type DocumentType struct {
    TypeName    string   // Human-readable name (e.g., "JPEG Image", "PDF Document")
    Role        string   // Role: "Editor", "Viewer", "Shell", "None"
    HandlerRank string   // Handler rank: "Owner", "Default", "Alternate", "None", or empty if not specified
    UTIs        []string // Array of UTI identifiers
    Extensions  []string // Array of file extensions (without dots)
    IsPackage   bool     // true if this is a package/bundle type
}
```

**Field Descriptions:**

- **TypeName**: Human-readable description from `CFBundleTypeName`
- **Role**: Capability level
  - `Editor` - Can create and modify files
  - `Viewer` - Can only view/display files
  - `Shell` - Can execute files
  - `None` - Minimal support
- **HandlerRank**: Priority/suitability level
  - `Owner` - This app "owns" the format
  - `Default` - Suitable default handler
  - `Alternate` - Can handle but not preferred
  - `None` - Last resort
  - Empty string if not specified
- **UTIs**: Array of Uniform Type Identifiers this document type handles
- **Extensions**: File extensions resolved from the UTIs
- **IsPackage**: Whether this is a bundle/package type (like .rtfd or .app)

**How it works:**

- Reads the application's `Info.plist` for `CFBundleDocumentTypes`
- Extracts all metadata for each document type
- Resolves UTIs to file extensions
- Also checks legacy `CFBundleTypeExtensions` key
- Filters out wildcard types (`public.item`, `public.data`) to avoid overly broad results

**Example:**

```go
docTypes, err := bridge.ListSupportedDocumentTypes("/System/Applications/TextEdit.app")
// Returns: []DocumentType{
//   {
//     TypeName: "NSRTFPboardType",
//     Role: "Editor",
//     HandlerRank: "",
//     UTIs: ["public.rtf"],
//     Extensions: ["rtf"],
//     IsPackage: false,
//   },
//   {
//     TypeName: "Apple HTML document",
//     Role: "Editor",
//     HandlerRank: "",
//     UTIs: ["public.html"],
//     Extensions: ["htm", "html", "shtm", "shtml"],
//     IsPackage: false,
//   },
//   {
//     TypeName: "Microsoft Word 2007 document",
//     Role: "Editor",
//     HandlerRank: "Alternate",
//     UTIs: ["org.openxmlformats.wordprocessingml.document"],
//     Extensions: ["docx"],
//     IsPackage: false,
//   },
//   ...
// }

// Filter to find types where app can edit (not just view)
for _, dt := range docTypes {
    if dt.Role == "Editor" {
        fmt.Printf("Can edit: %s (%v)\n", dt.TypeName, dt.Extensions)
    }
}

// Find the app's "owned" formats
for _, dt := range docTypes {
    if dt.HandlerRank == "Owner" {
        fmt.Printf("Owns format: %s\n", dt.TypeName)
    }
}
```

**Note:** Some applications (like system utilities) may not declare document types and will return an empty list. This is not an error.

#### `ListDefaultDocumentTypes(appPath string) ([]DocumentType, error)`

Returns all document types where the given application is ACTUALLY the system default.

**Key Difference from ListSupportedDocumentTypes:**
- `ListSupportedDocumentTypes`: Returns what the app CLAIMS it can handle
- `ListDefaultDocumentTypes`: Returns what the app IS THE DEFAULT for

**Parameters:**

- `appPath` - Full path to the application bundle

**Returns:**

- Slice of `DocumentType` structures where this app is the system default handler
- Error if app path is invalid

**How it works:**

1. Gets all document types the app supports (via `ListSupportedDocumentTypes`)
2. For each UTI, checks if this app is the system default (via `GetDefaultAppForUTI`)
3. Returns only document types where the app IS the default

**Example:**

```go
// Get what TextEdit SUPPORTS
supported, _ := bridge.ListSupportedDocumentTypes("/System/Applications/TextEdit.app")
fmt.Printf("TextEdit supports: %d types\n", len(supported))  // e.g., 12 types

// Get what TextEdit IS THE DEFAULT for
defaults, _ := bridge.ListDefaultDocumentTypes("/System/Applications/TextEdit.app")
fmt.Printf("TextEdit is default for: %d types\n", len(defaults))  // e.g., 4 types

// Show which types TextEdit owns
for _, dt := range defaults {
    fmt.Printf("Default for: %s (%v)\n", dt.TypeName, dt.Extensions)
}
// Output might be:
// Default for: NSRTFDPboardType ([rtfd])
// Default for: OpenDocument Text document ([odt])
// Default for: Apple SimpleText document ([])
```

**Use Cases:**

- Find which file types an app is currently handling
- Detect if another app has "stolen" file associations
- Show users which formats are controlled by an app
- Audit system-wide default handlers

**Note:** Returns an empty list if the app is not the default for any of its supported types. This is not an error.

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
