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
