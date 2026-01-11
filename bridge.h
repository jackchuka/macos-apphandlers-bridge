#ifndef MACOS_APPHANDLERS_BRIDGE_H
#define MACOS_APPHANDLERS_BRIDGE_H

// Return codes
#define BRIDGE_OK 0
#define BRIDGE_ERROR_INVALID_APP -1
#define BRIDGE_ERROR_INVALID_UTI -2
#define BRIDGE_ERROR_INVALID_SCHEME -3
#define BRIDGE_ERROR_SYSTEM -4
#define BRIDGE_ERROR_USER_DECLINED -5
#define BRIDGE_ERROR_NOT_FOUND -6

// Application information structure
typedef struct
{
    char *name;     // Application display name
    char *path;     // Full path to application bundle
    char *bundleID; // Bundle identifier (e.g., "com.apple.Safari")
} AppInfo;

// Document type information structure
typedef struct
{
    char *typeName;      // Human-readable name (e.g., "JPEG Image", "PDF Document")
    char *role;          // Role: "Editor", "Viewer", "Shell", "None"
    char *handlerRank;   // Handler rank: "Owner", "Default", "Alternate", "None", or NULL if not specified
    char **utis;         // Array of UTI identifiers
    int utiCount;        // Number of UTIs
    char **extensions;   // Array of file extensions
    int extensionCount;  // Number of extensions
    int isPackage;       // 1 if this is a package/bundle type, 0 otherwise
} DocumentType;

// Get the default application for a UTI
//
// Parameters:
//   uti: The Uniform Type Identifier (e.g., "public.plain-text")
//   outAppPath: Pointer to receive the application path (caller must free)
//   outError: Pointer to receive error message if any (caller must free)
//
// Returns: BRIDGE_OK on success, error code otherwise
int GetDefaultAppForUTI(const char *uti, char **outAppPath, char **outError);

// Get the default application for a URL scheme
//
// Parameters:
//   scheme: The URL scheme (e.g., "http", "mailto")
//   outAppPath: Pointer to receive the application path (caller must free)
//   outError: Pointer to receive error message if any (caller must free)
//
// Returns: BRIDGE_OK on success, error code otherwise
int GetDefaultAppForScheme(const char *scheme, char **outAppPath, char **outError);

// Set the default application for a UTI
//
// Parameters:
//   appPath: Full path to the application bundle (e.g., "/Applications/TextEdit.app")
//   uti: The Uniform Type Identifier (e.g., "public.plain-text")
//   outError: Pointer to receive error message if any (caller must free)
//
// Returns: BRIDGE_OK on success, error code otherwise
int SetDefaultForUTI(const char *appPath, const char *uti, char **outError);

// Set the default application for a URL scheme
//
// Parameters:
//   appPath: Full path to the application bundle
//   scheme: The URL scheme (e.g., "http", "mailto")
//   outError: Pointer to receive error message if any (caller must free)
//
// Returns: BRIDGE_OK on success, error code otherwise
int SetDefaultForScheme(const char *appPath, const char *scheme, char **outError);

// Resolve file extension to UTI(s)
//
// Parameters:
//   extension: File extension without dot (e.g., "txt", "md")
//   outUTIs: Pointer to receive array of UTI strings (caller must free using FreeCStringArray)
//   outCount: Pointer to receive count of UTIs returned
//   outError: Pointer to receive error message if any (caller must free)
//
// Returns: BRIDGE_OK on success, error code otherwise
int ResolveUTIsForExtension(const char *extension, char ***outUTIs, int *outCount, char **outError);

// Get file extensions for a UTI
//
// Parameters:
//   uti: The UTI string (e.g., "public.plain-text", "public.html")
//   outExtensions: Pointer to receive array of extension strings (caller must free using FreeCStringArray)
//   outCount: Pointer to receive count of extensions returned
//   outError: Pointer to receive error message if any (caller must free)
//
// Returns: BRIDGE_OK on success, error code otherwise
int GetExtensionsForUTI(const char *uti, char ***outExtensions, int *outCount, char **outError);

// List all applications that can open a UTI
//
// Parameters:
//   uti: The Uniform Type Identifier
//   outAppPaths: Pointer to receive array of app path strings (caller must free using FreeCStringArray)
//   outCount: Pointer to receive count of apps returned
//   outError: Pointer to receive error message if any (caller must free)
//
// Returns: BRIDGE_OK on success, error code otherwise
int ListAppsForUTI(const char *uti, char ***outAppPaths, int *outCount, char **outError);

// List all applications that can handle a URL scheme
//
// Parameters:
//   scheme: The URL scheme
//   outAppPaths: Pointer to receive array of app path strings (caller must free using FreeCStringArray)
//   outCount: Pointer to receive count of apps returned
//   outError: Pointer to receive error message if any (caller must free)
//
// Returns: BRIDGE_OK on success, error code otherwise
int ListAppsForScheme(const char *scheme, char ***outAppPaths, int *outCount, char **outError);

// Free a single C string allocated by bridge functions
//
// Parameters:
//   str: The string to free
void FreeCString(char *str);

// Free an array of C strings allocated by bridge functions
//
// Parameters:
//   arr: The array of strings to free
//   count: The number of strings in the array
void FreeCStringArray(char **arr, int count);

// List all installed applications on the system
//
// Parameters:
//   outApps: Pointer to receive array of AppInfo structures (caller must free using FreeAppInfoArray)
//   outCount: Pointer to receive count of applications returned
//   outError: Pointer to receive error message if any (caller must free)
//
// Returns: BRIDGE_OK on success, error code otherwise
int ListAllApplications(AppInfo ***outApps, int *outCount, char **outError);

// Free an array of AppInfo structures allocated by bridge functions
//
// Parameters:
//   apps: The array of AppInfo structures to free
//   count: The number of AppInfo structures in the array
void FreeAppInfoArray(AppInfo **apps, int count);

// Get supported document types for an application
//
// Parameters:
//   appPath: Full path to the application bundle (e.g., "/Applications/TextEdit.app")
//   outDocTypes: Pointer to receive array of DocumentType structures (caller must free using FreeDocumentTypeArray)
//   outCount: Pointer to receive count of document types returned
//   outError: Pointer to receive error message if any (caller must free)
//
// Returns: BRIDGE_OK on success, error code otherwise
int GetSupportedDocumentTypesForApp(const char *appPath, DocumentType ***outDocTypes, int *outCount, char **outError);

// Free an array of DocumentType structures allocated by bridge functions
//
// Parameters:
//   docTypes: The array of DocumentType structures to free
//   count: The number of DocumentType structures in the array
void FreeDocumentTypeArray(DocumentType **docTypes, int count);

#endif // MACOS_APPHANDLERS_BRIDGE_H
