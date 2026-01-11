//go:build darwin

package bridge

/*
#cgo CFLAGS: -x objective-c -fmodules -fblocks
#cgo LDFLAGS: -framework Foundation -framework AppKit -framework UniformTypeIdentifiers
#include "bridge.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unsafe"
)

// BridgeError represents an error from the macOS bridge layer
type BridgeError struct {
	Code    int
	Message string
}

func (e *BridgeError) Error() string {
	return fmt.Sprintf("bridge error (code %d): %s", e.Code, e.Message)
}

// Error codes matching bridge.h
const (
	ErrOK            = C.BRIDGE_OK
	ErrInvalidApp    = C.BRIDGE_ERROR_INVALID_APP
	ErrInvalidUTI    = C.BRIDGE_ERROR_INVALID_UTI
	ErrInvalidScheme = C.BRIDGE_ERROR_INVALID_SCHEME
	ErrSystem        = C.BRIDGE_ERROR_SYSTEM
	ErrUserDeclined  = C.BRIDGE_ERROR_USER_DECLINED
	ErrNotFound      = C.BRIDGE_ERROR_NOT_FOUND
)

// Common errors
var (
	ErrInvalidParameters = errors.New("invalid parameters")
	ErrMemoryAllocation  = errors.New("memory allocation failed")
)

// Helper function to convert C error to Go error
func cErrorToGoError(code C.int, cError *C.char) error {
	if code == C.BRIDGE_OK {
		return nil
	}

	var message string
	if cError != nil {
		message = C.GoString(cError)
		C.FreeCString(cError)
	} else {
		message = "unknown error"
	}

	return &BridgeError{
		Code:    int(code),
		Message: message,
	}
}

// GetDefaultAppForUTI returns the default application path for a UTI
//
// Parameters:
//   - uti: The Uniform Type Identifier (e.g., "public.plain-text")
//
// Returns:
//   - appPath: Full path to the default application bundle
//   - error: Error if any
func GetDefaultAppForUTI(uti string) (string, error) {
	if uti == "" {
		return "", ErrInvalidParameters
	}

	cUTI := C.CString(uti)
	defer C.free(unsafe.Pointer(cUTI))

	var cAppPath *C.char
	var cError *C.char

	code := C.GetDefaultAppForUTI(cUTI, &cAppPath, &cError)

	if code != C.BRIDGE_OK {
		return "", cErrorToGoError(code, cError)
	}

	if cAppPath == nil {
		return "", &BridgeError{
			Code:    int(ErrNotFound),
			Message: fmt.Sprintf("no default app found for UTI: %s", uti),
		}
	}

	appPath := C.GoString(cAppPath)
	C.FreeCString(cAppPath)

	return appPath, nil
}

// GetDefaultAppForScheme returns the default application path for a URL scheme
//
// Parameters:
//   - scheme: The URL scheme (e.g., "http", "mailto")
//
// Returns:
//   - appPath: Full path to the default application bundle
//   - error: Error if any
func GetDefaultAppForScheme(scheme string) (string, error) {
	if scheme == "" {
		return "", ErrInvalidParameters
	}

	cScheme := C.CString(scheme)
	defer C.free(unsafe.Pointer(cScheme))

	var cAppPath *C.char
	var cError *C.char

	code := C.GetDefaultAppForScheme(cScheme, &cAppPath, &cError)

	if code != C.BRIDGE_OK {
		return "", cErrorToGoError(code, cError)
	}

	if cAppPath == nil {
		return "", &BridgeError{
			Code:    int(ErrNotFound),
			Message: fmt.Sprintf("no default app found for scheme: %s", scheme),
		}
	}

	appPath := C.GoString(cAppPath)
	C.FreeCString(cAppPath)

	return appPath, nil
}

// SetDefaultForUTI sets the default application for a UTI
//
// Parameters:
//   - appPath: Full path to the application bundle
//   - uti: The Uniform Type Identifier
//
// Returns:
//   - error: Error if any
func SetDefaultForUTI(appPath, uti string) error {
	if appPath == "" || uti == "" {
		return ErrInvalidParameters
	}

	cAppPath := C.CString(appPath)
	defer C.free(unsafe.Pointer(cAppPath))

	cUTI := C.CString(uti)
	defer C.free(unsafe.Pointer(cUTI))

	var cError *C.char

	code := C.SetDefaultForUTI(cAppPath, cUTI, &cError)

	return cErrorToGoError(code, cError)
}

// SetDefaultForScheme sets the default application for a URL scheme
//
// Parameters:
//   - appPath: Full path to the application bundle
//   - scheme: The URL scheme
//
// Returns:
//   - error: Error if any
func SetDefaultForScheme(appPath, scheme string) error {
	if appPath == "" || scheme == "" {
		return ErrInvalidParameters
	}

	cAppPath := C.CString(appPath)
	defer C.free(unsafe.Pointer(cAppPath))

	cScheme := C.CString(scheme)
	defer C.free(unsafe.Pointer(cScheme))

	var cError *C.char

	code := C.SetDefaultForScheme(cAppPath, cScheme, &cError)

	return cErrorToGoError(code, cError)
}

// ResolveUTIsForExtension resolves a file extension to one or more UTIs
//
// Parameters:
//   - extension: File extension without dot (e.g., "txt", "md")
//
// Returns:
//   - utis: Slice of UTI identifiers
//   - error: Error if any
func ResolveUTIsForExtension(extension string) ([]string, error) {
	if extension == "" {
		return nil, ErrInvalidParameters
	}

	cExt := C.CString(extension)
	defer C.free(unsafe.Pointer(cExt))

	var cUTIs **C.char
	var count C.int
	var cError *C.char

	code := C.ResolveUTIsForExtension(cExt, &cUTIs, &count, &cError)

	if code != C.BRIDGE_OK {
		return nil, cErrorToGoError(code, cError)
	}

	if count == 0 {
		return []string{}, nil
	}

	// Convert C array to Go slice
	utis := make([]string, int(count))
	cUTIsSlice := (*[1 << 28]*C.char)(unsafe.Pointer(cUTIs))[:count:count]

	for i := 0; i < int(count); i++ {
		utis[i] = C.GoString(cUTIsSlice[i])
	}

	C.FreeCStringArray(cUTIs, count)

	return utis, nil
}

// ResolveExtensionsForUTI returns all file extensions associated with a UTI
//
// Parameters:
//   - uti: The UTI string (e.g., "public.plain-text", "public.html")
//
// Returns:
//   - extensions: Slice of file extensions (without dots)
//   - error: Error if any
func ResolveExtensionsForUTI(uti string) ([]string, error) {
	if uti == "" {
		return nil, ErrInvalidParameters
	}

	cUTI := C.CString(uti)
	defer C.free(unsafe.Pointer(cUTI))

	var cExtensions **C.char
	var count C.int
	var cError *C.char

	code := C.GetExtensionsForUTI(cUTI, &cExtensions, &count, &cError)

	if code != C.BRIDGE_OK {
		return nil, cErrorToGoError(code, cError)
	}

	if count == 0 {
		return []string{}, nil
	}

	// Convert C array to Go slice
	extensions := make([]string, int(count))
	cExtensionsSlice := (*[1 << 28]*C.char)(unsafe.Pointer(cExtensions))[:count:count]

	for i := 0; i < int(count); i++ {
		extensions[i] = C.GoString(cExtensionsSlice[i])
	}

	C.FreeCStringArray(cExtensions, count)

	return extensions, nil
}

// ListAppsForUTI returns all applications that can open a UTI
//
// Parameters:
//   - uti: The Uniform Type Identifier
//
// Returns:
//   - appPaths: Slice of application bundle paths
//   - error: Error if any
func ListAppsForUTI(uti string) ([]string, error) {
	if uti == "" {
		return nil, ErrInvalidParameters
	}

	cUTI := C.CString(uti)
	defer C.free(unsafe.Pointer(cUTI))

	var cAppPaths **C.char
	var count C.int
	var cError *C.char

	code := C.ListAppsForUTI(cUTI, &cAppPaths, &count, &cError)

	if code != C.BRIDGE_OK {
		return nil, cErrorToGoError(code, cError)
	}

	if count == 0 {
		return []string{}, nil
	}

	// Convert C array to Go slice
	appPaths := make([]string, int(count))
	cAppPathsSlice := (*[1 << 28]*C.char)(unsafe.Pointer(cAppPaths))[:count:count]

	for i := 0; i < int(count); i++ {
		appPaths[i] = C.GoString(cAppPathsSlice[i])
	}

	C.FreeCStringArray(cAppPaths, count)

	return appPaths, nil
}

// ListAppsForScheme returns all applications that can handle a URL scheme
//
// Parameters:
//   - scheme: The URL scheme
//
// Returns:
//   - appPaths: Slice of application bundle paths
//   - error: Error if any
func ListAppsForScheme(scheme string) ([]string, error) {
	if scheme == "" {
		return nil, ErrInvalidParameters
	}

	cScheme := C.CString(scheme)
	defer C.free(unsafe.Pointer(cScheme))

	var cAppPaths **C.char
	var count C.int
	var cError *C.char

	code := C.ListAppsForScheme(cScheme, &cAppPaths, &count, &cError)

	if code != C.BRIDGE_OK {
		return nil, cErrorToGoError(code, cError)
	}

	if count == 0 {
		return []string{}, nil
	}

	// Convert C array to Go slice
	appPaths := make([]string, int(count))
	cAppPathsSlice := (*[1 << 28]*C.char)(unsafe.Pointer(cAppPaths))[:count:count]

	for i := 0; i < int(count); i++ {
		appPaths[i] = C.GoString(cAppPathsSlice[i])
	}

	C.FreeCStringArray(cAppPaths, count)

	return appPaths, nil
}

// AppInfo represents an installed application with its metadata
type AppInfo struct {
	Name     string // Application display name
	Path     string // Full path to application bundle
	BundleID string // Bundle identifier (e.g., "com.apple.Safari")
}

// DocumentType represents a document type that an application can handle
type DocumentType struct {
	TypeName    string   // Human-readable name (e.g., "JPEG Image", "PDF Document")
	Role        string   // Role: "Editor", "Viewer", "Shell", "None"
	HandlerRank string   // Handler rank: "Owner", "Default", "Alternate", "None", or empty if not specified
	UTIs        []string // Array of UTI identifiers
	Extensions  []string // Array of file extensions
	IsPackage   bool     // true if this is a package/bundle type
}

// ListAllApplications returns all installed applications on the system
//
// Returns:
//   - apps: Slice of AppInfo structures containing app metadata
//   - error: Error if any
func ListAllApplications() ([]AppInfo, error) {
	var cApps **C.AppInfo
	var count C.int
	var cError *C.char

	code := C.ListAllApplications(&cApps, &count, &cError)

	if code != C.BRIDGE_OK {
		return nil, cErrorToGoError(code, cError)
	}

	if count == 0 {
		return []AppInfo{}, nil
	}

	// Convert C array to Go slice
	apps := make([]AppInfo, int(count))
	cAppsSlice := (*[1 << 28]*C.AppInfo)(unsafe.Pointer(cApps))[:count:count]

	for i := 0; i < int(count); i++ {
		cAppInfo := cAppsSlice[i]
		apps[i] = AppInfo{
			Name:     C.GoString(cAppInfo.name),
			Path:     C.GoString(cAppInfo.path),
			BundleID: C.GoString(cAppInfo.bundleID),
		}
	}

	C.FreeAppInfoArray(cApps, count)

	return apps, nil
}

// getExtensionsForUTIs returns all file extensions for the given UTIs
func getExtensionsForUTIs(utis []string) []string {
	extensionsSet := make(map[string]bool)

	for _, uti := range utis {
		extensions, err := ResolveExtensionsForUTI(uti)
		if err != nil {
			continue
		}
		for _, ext := range extensions {
			extensionsSet[ext] = true
		}
	}

	// Convert set to sorted slice
	var result []string
	for ext := range extensionsSet {
		result = append(result, ext)
	}
	sort.Strings(result)

	return result
}

// ListDefaultDocumentTypes returns all document types where the given application is the system default
//
// This function checks which document types the app supports AND is actually set as the default handler for.
//
// Parameters:
//   - appPath: Full path to the application bundle
//
// Returns:
//   - docTypes: Slice of DocumentType structures where this app is the system default
//   - error: Error if any
func ListDefaultDocumentTypes(appPath string) ([]DocumentType, error) {
	if appPath == "" {
		return nil, ErrInvalidParameters
	}

	// Get all document types this app supports
	allDocTypes, err := ListSupportedDocumentTypes(appPath)
	if err != nil {
		return nil, err
	}

	// Filter to only those where this app is the default
	var defaultDocTypes []DocumentType

	// Normalize app path for comparison
	cleanAppPath := appPath
	if resolved, err := filepath.EvalSymlinks(appPath); err == nil {
		cleanAppPath = resolved
	}

	for _, docType := range allDocTypes {
		// Collect only the UTIs where this app is actually the default
		var matchingUTIs []string

		for _, uti := range docType.UTIs {
			defaultApp, err := GetDefaultAppForUTI(uti)
			if err != nil {
				// Skip UTIs that have no default or error
				continue
			}

			// Normalize default app path
			cleanDefaultApp := defaultApp
			if resolved, err := filepath.EvalSymlinks(defaultApp); err == nil {
				cleanDefaultApp = resolved
			}

			// Check if this app is the default for this specific UTI
			if cleanAppPath == cleanDefaultApp || strings.EqualFold(cleanAppPath, cleanDefaultApp) {
				matchingUTIs = append(matchingUTIs, uti)
			}
		}

		// Only include if at least one UTI matches
		if len(matchingUTIs) > 0 {
			// Derive extensions from matching UTIs
			matchingExtensions := getExtensionsForUTIs(matchingUTIs)

			// Create filtered document type
			filteredDocType := DocumentType{
				TypeName:    docType.TypeName,
				Role:        docType.Role,
				HandlerRank: docType.HandlerRank,
				UTIs:        matchingUTIs,
				Extensions:  matchingExtensions,
				IsPackage:   docType.IsPackage,
			}

			defaultDocTypes = append(defaultDocTypes, filteredDocType)
		}
	}

	return defaultDocTypes, nil
}

// ListSupportedDocumentTypes returns all document types that an application can handle
//
// This returns what the app CLAIMS it can handle, not what it's the default for.
//
// Parameters:
//   - appPath: Full path to the application bundle
//
// Returns:
//   - docTypes: Slice of DocumentType structures containing detailed info about supported file types
//   - error: Error if any
func ListSupportedDocumentTypes(appPath string) ([]DocumentType, error) {
	if appPath == "" {
		return nil, ErrInvalidParameters
	}

	cAppPath := C.CString(appPath)
	defer C.free(unsafe.Pointer(cAppPath))

	var cDocTypes **C.DocumentType
	var count C.int
	var cError *C.char

	code := C.GetSupportedDocumentTypesForApp(cAppPath, &cDocTypes, &count, &cError)

	if code != C.BRIDGE_OK {
		return nil, cErrorToGoError(code, cError)
	}

	if count == 0 {
		return []DocumentType{}, nil
	}

	// Convert C array to Go slice
	docTypes := make([]DocumentType, int(count))
	cDocTypesSlice := (*[1 << 28]*C.DocumentType)(unsafe.Pointer(cDocTypes))[:count:count]

	for i := 0; i < int(count); i++ {
		cDocType := cDocTypesSlice[i]

		// Convert UTIs array
		utis := make([]string, int(cDocType.utiCount))
		if cDocType.utiCount > 0 && cDocType.utis != nil {
			cUTIsSlice := (*[1 << 28]*C.char)(unsafe.Pointer(cDocType.utis))[:cDocType.utiCount:cDocType.utiCount]
			for j := 0; j < int(cDocType.utiCount); j++ {
				utis[j] = C.GoString(cUTIsSlice[j])
			}
		}

		// Convert extensions array
		extensions := make([]string, int(cDocType.extensionCount))
		if cDocType.extensionCount > 0 && cDocType.extensions != nil {
			cExtensionsSlice := (*[1 << 28]*C.char)(unsafe.Pointer(cDocType.extensions))[:cDocType.extensionCount:cDocType.extensionCount]
			for j := 0; j < int(cDocType.extensionCount); j++ {
				extensions[j] = C.GoString(cExtensionsSlice[j])
			}
		}

		// Get handler rank (may be NULL)
		handlerRank := ""
		if cDocType.handlerRank != nil {
			handlerRank = C.GoString(cDocType.handlerRank)
		}

		docTypes[i] = DocumentType{
			TypeName:    C.GoString(cDocType.typeName),
			Role:        C.GoString(cDocType.role),
			HandlerRank: handlerRank,
			UTIs:        utis,
			Extensions:  extensions,
			IsPackage:   cDocType.isPackage != 0,
		}
	}

	C.FreeDocumentTypeArray(cDocTypes, count)

	return docTypes, nil
}
