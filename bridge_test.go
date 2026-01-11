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
