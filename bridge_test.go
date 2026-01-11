//go:build darwin

package bridge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Common macOS applications for testing
const (
	textEditPath = "/System/Applications/TextEdit.app"
)

// TestGetDefaultAppForUTI tests reading default app for various UTIs
func TestGetDefaultAppForUTI(t *testing.T) {
	tests := []struct {
		name    string
		uti     string
		wantErr bool
	}{
		{
			name:    "plain text",
			uti:     "public.plain-text",
			wantErr: false,
		},
		{
			name:    "jpeg image",
			uti:     "public.jpeg",
			wantErr: false,
		},
		{
			name:    "html",
			uti:     "public.html",
			wantErr: false,
		},
		{
			name:    "invalid UTI",
			uti:     "com.example.nonexistent",
			wantErr: true,
		},
		{
			name:    "empty UTI",
			uti:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appPath, err := GetDefaultAppForUTI(tt.uti)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetDefaultAppForUTI() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetDefaultAppForUTI() error = %v", err)
				return
			}

			if appPath == "" {
				t.Errorf("GetDefaultAppForUTI() returned empty path")
				return
			}

			// Verify the path exists
			if _, err := os.Stat(appPath); os.IsNotExist(err) {
				t.Errorf("GetDefaultAppForUTI() returned non-existent path: %s", appPath)
			}

			t.Logf("Default app for %s: %s", tt.uti, appPath)
		})
	}
}

// TestGetDefaultAppForScheme tests reading default app for URL schemes
func TestGetDefaultAppForScheme(t *testing.T) {
	tests := []struct {
		name    string
		scheme  string
		wantErr bool
	}{
		{
			name:    "http",
			scheme:  "http",
			wantErr: false,
		},
		{
			name:    "https",
			scheme:  "https",
			wantErr: false,
		},
		{
			name:    "mailto",
			scheme:  "mailto",
			wantErr: false,
		},
		{
			name:    "empty scheme",
			scheme:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appPath, err := GetDefaultAppForScheme(tt.scheme)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetDefaultAppForScheme() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetDefaultAppForScheme() error = %v", err)
				return
			}

			if appPath == "" {
				t.Errorf("GetDefaultAppForScheme() returned empty path")
				return
			}

			t.Logf("Default app for %s: %s", tt.scheme, appPath)
		})
	}
}

// TestResolveExtensionsForUTI tests resolving file extensions for a UTI
func TestResolveExtensionsForUTI(t *testing.T) {
	tests := []struct {
		name        string
		uti         string
		wantErr     bool
		minExtCount int
		mustContain []string
	}{
		{
			name:        "HTML UTI",
			uti:         "public.html",
			minExtCount: 1,
			mustContain: []string{"html"},
			wantErr:     false,
		},
		{
			name:        "Plain text UTI",
			uti:         "public.plain-text",
			minExtCount: 1,
			mustContain: []string{"txt"},
			wantErr:     false,
		},
		{
			name:        "JPEG UTI",
			uti:         "public.jpeg",
			minExtCount: 1,
			mustContain: []string{"jpg"},
			wantErr:     false,
		},
		{
			name:    "Empty UTI",
			uti:     "",
			wantErr: true,
		},
		{
			name:        "Folder UTI (no extensions)",
			uti:         "public.folder",
			minExtCount: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extensions, err := ResolveExtensionsForUTI(tt.uti)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveExtensionsForUTI() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveExtensionsForUTI() error = %v", err)
				return
			}

			if len(extensions) < tt.minExtCount {
				t.Errorf("ResolveExtensionsForUTI() got %d extensions, want at least %d", len(extensions), tt.minExtCount)
			}

			// Check that required extensions are present
			for _, required := range tt.mustContain {
				found := false
				for _, ext := range extensions {
					if ext == required {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ResolveExtensionsForUTI() missing required extension %s in result %v", required, extensions)
				}
			}

			t.Logf("Extensions for %s: %v", tt.uti, extensions)
		})
	}
}

// TestResolveUTIsForExtension tests extension to UTI resolution
func TestResolveUTIsForExtension(t *testing.T) {
	tests := []struct {
		name        string
		extension   string
		wantUTIs    []string // Expected UTIs (can be subset)
		wantErr     bool
		minUTICount int // Minimum expected UTI count
	}{
		{
			name:        "txt",
			extension:   "txt",
			wantUTIs:    []string{"public.plain-text"},
			minUTICount: 1,
			wantErr:     false,
		},
		{
			name:        "md",
			extension:   "md",
			wantUTIs:    []string{"net.daringfireball.markdown"},
			minUTICount: 1,
			wantErr:     false,
		},
		{
			name:        "jpg",
			extension:   "jpg",
			wantUTIs:    []string{"public.jpeg"},
			minUTICount: 1,
			wantErr:     false,
		},
		{
			name:        "html",
			extension:   "html",
			wantUTIs:    []string{"public.html"},
			minUTICount: 1,
			wantErr:     false,
		},
		{
			name:      "empty extension",
			extension: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utis, err := ResolveUTIsForExtension(tt.extension)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveUTIsForExtension() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveUTIsForExtension() error = %v", err)
				return
			}

			if len(utis) < tt.minUTICount {
				t.Errorf("ResolveUTIsForExtension() got %d UTIs, want at least %d", len(utis), tt.minUTICount)
				return
			}

			// Check if any of the expected UTIs are present
			found := false
			for _, wantUTI := range tt.wantUTIs {
				for _, gotUTI := range utis {
					if gotUTI == wantUTI {
						found = true
						break
					}
				}
				if found {
					break
				}
			}

			if !found {
				t.Errorf("ResolveUTIsForExtension() = %v, want to contain one of %v", utis, tt.wantUTIs)
			}

			t.Logf("UTIs for .%s: %v", tt.extension, utis)
		})
	}
}

// TestSetDefaultForUTI tests setting default app for UTI with round-trip
func TestSetDefaultForUTI(t *testing.T) {
	// This test requires TextEdit to be available
	if _, err := os.Stat(textEditPath); os.IsNotExist(err) {
		t.Skipf("TextEdit not found at %s, skipping test", textEditPath)
	}

	// Use a specific UTI for testing
	testUTI := "public.plain-text"

	// Get current default to restore later
	originalApp, err := GetDefaultAppForUTI(testUTI)
	if err != nil {
		t.Fatalf("Failed to get original default app: %v", err)
	}

	// Ensure we restore the original default
	defer func() {
		if originalApp != "" {
			_ = SetDefaultForUTI(originalApp, testUTI)
		}
	}()

	t.Logf("Original default app for %s: %s", testUTI, originalApp)

	// Set TextEdit as default
	err = SetDefaultForUTI(textEditPath, testUTI)
	if err != nil {
		t.Fatalf("SetDefaultForUTI() error = %v", err)
	}

	// Verify it was set
	currentApp, err := GetDefaultAppForUTI(testUTI)
	if err != nil {
		t.Fatalf("Failed to verify default app: %v", err)
	}

	// Normalize paths for comparison (resolve symlinks, etc.)
	if !pathsMatch(currentApp, textEditPath) {
		t.Errorf("SetDefaultForUTI() verification failed: got %s, want %s", currentApp, textEditPath)
	}

	t.Logf("Successfully set and verified TextEdit as default for %s", testUTI)
}

// TestSetDefaultForUTI_InvalidApp tests error handling for invalid app paths
func TestSetDefaultForUTI_InvalidApp(t *testing.T) {
	tests := []struct {
		name    string
		appPath string
		uti     string
	}{
		{
			name:    "non-existent app",
			appPath: "/Applications/NonExistent.app",
			uti:     "public.plain-text",
		},
		{
			name:    "empty app path",
			appPath: "",
			uti:     "public.plain-text",
		},
		{
			name:    "empty UTI",
			appPath: textEditPath,
			uti:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetDefaultForUTI(tt.appPath, tt.uti)
			if err == nil {
				t.Errorf("SetDefaultForUTI() expected error for invalid input, got nil")
			}
			t.Logf("Got expected error: %v", err)
		})
	}
}

// TestListAppsForUTI tests listing apps that can open a UTI
func TestListAppsForUTI(t *testing.T) {
	tests := []struct {
		name        string
		uti         string
		minAppCount int // Minimum number of apps expected
		wantErr     bool
	}{
		{
			name:        "plain text",
			uti:         "public.plain-text",
			minAppCount: 1, // At least TextEdit should be available
			wantErr:     false,
		},
		{
			name:        "html",
			uti:         "public.html",
			minAppCount: 1, // At least Safari or a browser
			wantErr:     false,
		},
		{
			name:    "invalid UTI",
			uti:     "com.example.nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apps, err := ListAppsForUTI(tt.uti)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ListAppsForUTI() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ListAppsForUTI() error = %v", err)
				return
			}

			if len(apps) < tt.minAppCount {
				t.Errorf("ListAppsForUTI() got %d apps, want at least %d", len(apps), tt.minAppCount)
			}

			t.Logf("Apps that can open %s: %v", tt.uti, apps)
		})
	}
}

// TestListAppsForScheme tests listing apps that can handle a URL scheme
func TestListAppsForScheme(t *testing.T) {
	tests := []struct {
		name        string
		scheme      string
		minAppCount int
		wantErr     bool
	}{
		{
			name:        "http",
			scheme:      "http",
			minAppCount: 1, // At least one browser
			wantErr:     false,
		},
		{
			name:        "mailto",
			scheme:      "mailto",
			minAppCount: 1, // At least Mail.app
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apps, err := ListAppsForScheme(tt.scheme)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ListAppsForScheme() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ListAppsForScheme() error = %v", err)
				return
			}

			if len(apps) < tt.minAppCount {
				t.Errorf("ListAppsForScheme() got %d apps, want at least %d", len(apps), tt.minAppCount)
			}

			t.Logf("Apps that can handle %s: %v", tt.scheme, apps)
		})
	}
}

// TestListAllApplications tests listing all installed applications
func TestListAllApplications(t *testing.T) {
	apps, err := ListAllApplications()
	if err != nil {
		t.Fatalf("ListAllApplications() error = %v", err)
	}

	if len(apps) == 0 {
		t.Errorf("ListAllApplications() returned zero applications")
	}

	t.Logf("Total applications found: %d", len(apps))

	// Check for presence of field in the first app as a sample
	firstApp := apps[0]
	if firstApp.Name == "" || firstApp.Path == "" || firstApp.BundleID == "" {
		t.Errorf("ListAllApplications() returned app with missing fields: %+v", firstApp)
	} else {
		t.Logf("Sample application: Name=%s, Path=%s, BundleID=%s", firstApp.Name, firstApp.Path, firstApp.BundleID)
	}
}

// TestListSupportedDocumentTypes tests listing supported document types for an application
func TestListSupportedDocumentTypes(t *testing.T) {
	// Test with TextEdit - should exist on all macOS systems
	if _, err := os.Stat(textEditPath); os.IsNotExist(err) {
		t.Skipf("TextEdit not found at %s, skipping test", textEditPath)
	}

	docTypes, err := ListSupportedDocumentTypes(textEditPath)
	if err != nil {
		t.Fatalf("ListSupportedDocumentTypes() error = %v", err)
	}

	// TextEdit should support at least some document types
	if len(docTypes) == 0 {
		t.Errorf("ListSupportedDocumentTypes() returned zero document types for TextEdit")
	}

	// Verify structure is valid
	for i, docType := range docTypes {
		// Role should not be empty
		if docType.Role == "" {
			t.Errorf("ListSupportedDocumentTypes() returned empty role at index %d", i)
		}

		// Should have at least one UTI
		if len(docType.UTIs) == 0 {
			t.Errorf("ListSupportedDocumentTypes() returned zero UTIs at index %d", i)
		}

		// Check UTIs are non-empty
		for j, uti := range docType.UTIs {
			if uti == "" {
				t.Errorf("ListSupportedDocumentTypes() returned empty UTI at index %d,%d", i, j)
			}
		}

		// Extensions array can be empty (some UTIs don't have extensions)
		// but if present, should be non-empty strings
		for j, ext := range docType.Extensions {
			if ext == "" {
				t.Errorf("ListSupportedDocumentTypes() returned empty extension at index %d,%d", i, j)
			}
		}

		t.Logf("Document type %d: Name=%q, Role=%q, HandlerRank=%q, UTIs=%v, Extensions=%v, IsPackage=%v",
			i, docType.TypeName, docType.Role, docType.HandlerRank, docType.UTIs, docType.Extensions, docType.IsPackage)
	}

	t.Logf("TextEdit supports %d document types", len(docTypes))
}

// TestListSupportedDocumentTypes_ErrorCases tests error handling
func TestListSupportedDocumentTypes_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		appPath string
	}{
		{
			name:    "non-existent app",
			appPath: "/Applications/NonExistent.app",
		},
		{
			name:    "empty path",
			appPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docTypes, err := ListSupportedDocumentTypes(tt.appPath)
			if err == nil {
				t.Errorf("ListSupportedDocumentTypes() expected error, got nil with docTypes: %v", docTypes)
			} else {
				t.Logf("Got expected error: %v", err)
			}
		})
	}
}

// TestListDefaultDocumentTypes tests listing document types where an app is the system default
func TestListDefaultDocumentTypes(t *testing.T) {
	// Test with TextEdit - should exist on all macOS systems
	if _, err := os.Stat(textEditPath); os.IsNotExist(err) {
		t.Skipf("TextEdit not found at %s, skipping test", textEditPath)
	}

	defaultDocTypes, err := ListDefaultDocumentTypes(textEditPath)
	if err != nil {
		t.Fatalf("ListDefaultDocumentTypes() error = %v", err)
	}

	// TextEdit may or may not be the default for any types (depends on system configuration)
	// So we can't assert a minimum count, but we can validate structure if any are returned
	t.Logf("TextEdit is the default for %d document types", len(defaultDocTypes))

	for i, docType := range defaultDocTypes {
		// Verify structure is valid
		if docType.Role == "" {
			t.Errorf("ListDefaultDocumentTypes() returned empty role at index %d", i)
		}

		if len(docType.UTIs) == 0 {
			t.Errorf("ListDefaultDocumentTypes() returned zero UTIs at index %d", i)
		}

		// Verify that this app IS actually the default for ALL returned UTIs (not just some)
		for _, uti := range docType.UTIs {
			defaultApp, err := GetDefaultAppForUTI(uti)
			if err != nil {
				t.Errorf("ListDefaultDocumentTypes() returned UTI %s at index %d, but got error when checking default: %v", uti, i, err)
				continue
			}

			if !pathsMatch(defaultApp, textEditPath) {
				t.Errorf("ListDefaultDocumentTypes() returned UTI %s at index %d, but TextEdit is not the default (default is %s)",
					uti, i, defaultApp)
			}
		}

		t.Logf("Default document type %d: Name=%q, Role=%q, UTIs=%v, Extensions=%v",
			i, docType.TypeName, docType.Role, docType.UTIs, docType.Extensions)
	}

	// Also verify that we're filtering correctly - supported types should be >= default types
	supportedDocTypes, err := ListSupportedDocumentTypes(textEditPath)
	if err != nil {
		t.Fatalf("ListSupportedDocumentTypes() error = %v", err)
	}

	if len(defaultDocTypes) > len(supportedDocTypes) {
		t.Errorf("ListDefaultDocumentTypes() returned %d types, but ListSupportedDocumentTypes() returned %d - default should be a subset",
			len(defaultDocTypes), len(supportedDocTypes))
	}

	t.Logf("TextEdit supports %d types total, is default for %d", len(supportedDocTypes), len(defaultDocTypes))
}

// Helper function to compare app paths (handles symlinks and normalization)
func pathsMatch(path1, path2 string) bool {
	// Clean both paths
	clean1 := filepath.Clean(path1)
	clean2 := filepath.Clean(path2)

	// Direct match
	if clean1 == clean2 {
		return true
	}

	// Try resolving symlinks
	real1, err1 := filepath.EvalSymlinks(clean1)
	real2, err2 := filepath.EvalSymlinks(clean2)

	if err1 == nil && err2 == nil {
		return real1 == real2
	}

	// Case-insensitive match (macOS filesystems are often case-insensitive)
	return strings.EqualFold(clean1, clean2)
}
