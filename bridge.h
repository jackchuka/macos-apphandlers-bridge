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

#endif // MACOS_APPHANDLERS_BRIDGE_H
