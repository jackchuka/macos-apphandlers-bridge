#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>
#import "bridge.h"
#import <string.h>

// Require macOS 12.0 or later
#if !defined(__MAC_12_0) || MAC_OS_X_VERSION_MIN_REQUIRED < __MAC_12_0
#warning "This code requires macOS 12.0 (Monterey) or later"
#endif

// Helper function to create a C string from NSString
static char* NSStringToCString(NSString* str) {
    if (!str) return NULL;
    const char* utf8 = [str UTF8String];
    if (!utf8) return NULL;
    return strdup(utf8);
}

// Helper function to set error message
static void SetError(char** outError, NSString* message) {
    if (outError) {
        *outError = NSStringToCString(message);
    }
}

// Helper function to get absolute path from URL
static char* URLToPath(NSURL* url) {
    if (!url) return NULL;
    return NSStringToCString([url path]);
}

// Get the default application for a UTI
int GetDefaultAppForUTI(const char* uti, char** outAppPath, char** outError) {
    @autoreleasepool {
        if (!uti || !outAppPath) {
            SetError(outError, @"Invalid parameters");
            return BRIDGE_ERROR_INVALID_UTI;
        }

        *outAppPath = NULL;

        NSString* utiString = [NSString stringWithUTF8String:uti];
        if (!utiString) {
            SetError(outError, @"Invalid UTF-8 in UTI string");
            return BRIDGE_ERROR_INVALID_UTI;
        }

        // Create UTType from identifier
        UTType* utType = [UTType typeWithIdentifier:utiString];
        if (!utType) {
            SetError(outError, [NSString stringWithFormat:@"Invalid or unknown UTI: %s", uti]);
            return BRIDGE_ERROR_INVALID_UTI;
        }

        // Get default application URL for this UTType
        NSError* error = nil;
        NSWorkspace* workspace = [NSWorkspace sharedWorkspace];
        NSURL* appURL = [workspace URLForApplicationToOpenContentType:utType];

        if (error) {
            SetError(outError, [error localizedDescription]);
            return BRIDGE_ERROR_SYSTEM;
        }

        if (!appURL) {
            SetError(outError, [NSString stringWithFormat:@"No default application found for UTI: %s", uti]);
            return BRIDGE_ERROR_NOT_FOUND;
        }

        *outAppPath = URLToPath(appURL);
        return BRIDGE_OK;
    }
}

// Get the default application for a URL scheme
int GetDefaultAppForScheme(const char* scheme, char** outAppPath, char** outError) {
    @autoreleasepool {
        if (!scheme || !outAppPath) {
            SetError(outError, @"Invalid parameters");
            return BRIDGE_ERROR_INVALID_SCHEME;
        }

        *outAppPath = NULL;

        NSString* schemeString = [NSString stringWithUTF8String:scheme];
        if (!schemeString) {
            SetError(outError, @"Invalid UTF-8 in scheme string");
            return BRIDGE_ERROR_INVALID_SCHEME;
        }

        // Create a sample URL with this scheme
        NSString* urlString = [NSString stringWithFormat:@"%@://", schemeString];
        NSURL* url = [NSURL URLWithString:urlString];

        if (!url) {
            SetError(outError, [NSString stringWithFormat:@"Invalid URL scheme: %s", scheme]);
            return BRIDGE_ERROR_INVALID_SCHEME;
        }

        // Get default application URL for this scheme
        NSWorkspace* workspace = [NSWorkspace sharedWorkspace];
        NSURL* appURL = [workspace URLForApplicationToOpenURL:url];

        if (!appURL) {
            SetError(outError, [NSString stringWithFormat:@"No default application found for scheme: %s", scheme]);
            return BRIDGE_ERROR_NOT_FOUND;
        }

        *outAppPath = URLToPath(appURL);
        return BRIDGE_OK;
    }
}

// Set the default application for a UTI
int SetDefaultForUTI(const char* appPath, const char* uti, char** outError) {
    @autoreleasepool {
        if (!appPath || !uti) {
            SetError(outError, @"Invalid parameters");
            return BRIDGE_ERROR_INVALID_APP;
        }

        NSString* appPathString = [NSString stringWithUTF8String:appPath];
        NSString* utiString = [NSString stringWithUTF8String:uti];

        if (!appPathString || !utiString) {
            SetError(outError, @"Invalid UTF-8 in parameters");
            return BRIDGE_ERROR_INVALID_APP;
        }

        // Create app URL
        NSURL* appURL = [NSURL fileURLWithPath:appPathString];
        if (!appURL) {
            SetError(outError, [NSString stringWithFormat:@"Invalid application path: %s", appPath]);
            return BRIDGE_ERROR_INVALID_APP;
        }

        // Verify app exists
        if (![[NSFileManager defaultManager] fileExistsAtPath:appPathString]) {
            SetError(outError, [NSString stringWithFormat:@"Application not found: %s", appPath]);
            return BRIDGE_ERROR_INVALID_APP;
        }

        // Create UTType
        UTType* utType = [UTType typeWithIdentifier:utiString];
        if (!utType) {
            SetError(outError, [NSString stringWithFormat:@"Invalid or unknown UTI: %s", uti]);
            return BRIDGE_ERROR_INVALID_UTI;
        }

        // Use semaphore to wait for async completion
        dispatch_semaphore_t semaphore = dispatch_semaphore_create(0);
        __block int resultCode = BRIDGE_OK;
        __block NSError* resultError = nil;

        NSWorkspace* workspace = [NSWorkspace sharedWorkspace];
        [workspace setDefaultApplicationAtURL:appURL
                         toOpenContentType:utType
                         completionHandler:^(NSError * _Nullable error) {
            if (error) {
                resultError = [error retain];
                if ([error.domain isEqualToString:NSCocoaErrorDomain] && error.code == NSUserCancelledError) {
                    resultCode = BRIDGE_ERROR_USER_DECLINED;
                } else {
                    resultCode = BRIDGE_ERROR_SYSTEM;
                }
            }
            dispatch_semaphore_signal(semaphore);
        }];

        // Wait for completion (with timeout of 10 seconds)
        long timeoutResult = dispatch_semaphore_wait(semaphore, dispatch_time(DISPATCH_TIME_NOW, 10 * NSEC_PER_SEC));

        if (timeoutResult != 0) {
            SetError(outError, @"Operation timed out");
            return BRIDGE_ERROR_SYSTEM;
        }

        if (resultError) {
            SetError(outError, [resultError localizedDescription]);
            [resultError release];
            return resultCode;
        }

        return BRIDGE_OK;
    }
}

// Set the default application for a URL scheme
int SetDefaultForScheme(const char* appPath, const char* scheme, char** outError) {
    @autoreleasepool {
        if (!appPath || !scheme) {
            SetError(outError, @"Invalid parameters");
            return BRIDGE_ERROR_INVALID_APP;
        }

        NSString* appPathString = [NSString stringWithUTF8String:appPath];
        NSString* schemeString = [NSString stringWithUTF8String:scheme];

        if (!appPathString || !schemeString) {
            SetError(outError, @"Invalid UTF-8 in parameters");
            return BRIDGE_ERROR_INVALID_APP;
        }

        // Create app URL
        NSURL* appURL = [NSURL fileURLWithPath:appPathString];
        if (!appURL) {
            SetError(outError, [NSString stringWithFormat:@"Invalid application path: %s", appPath]);
            return BRIDGE_ERROR_INVALID_APP;
        }

        // Verify app exists
        if (![[NSFileManager defaultManager] fileExistsAtPath:appPathString]) {
            SetError(outError, [NSString stringWithFormat:@"Application not found: %s", appPath]);
            return BRIDGE_ERROR_INVALID_APP;
        }

        // Use semaphore to wait for async completion
        dispatch_semaphore_t semaphore = dispatch_semaphore_create(0);
        __block int resultCode = BRIDGE_OK;
        __block NSError* resultError = nil;

        NSWorkspace* workspace = [NSWorkspace sharedWorkspace];
        [workspace setDefaultApplicationAtURL:appURL
                         toOpenURLsWithScheme:schemeString
                         completionHandler:^(NSError * _Nullable error) {
            if (error) {
                resultError = [error retain];
                if ([error.domain isEqualToString:NSCocoaErrorDomain] && error.code == NSUserCancelledError) {
                    resultCode = BRIDGE_ERROR_USER_DECLINED;
                } else {
                    resultCode = BRIDGE_ERROR_SYSTEM;
                }
            }
            dispatch_semaphore_signal(semaphore);
        }];

        // Wait for completion (with timeout of 10 seconds)
        long timeoutResult = dispatch_semaphore_wait(semaphore, dispatch_time(DISPATCH_TIME_NOW, 10 * NSEC_PER_SEC));

        if (timeoutResult != 0) {
            SetError(outError, @"Operation timed out");
            return BRIDGE_ERROR_SYSTEM;
        }

        if (resultError) {
            SetError(outError, [resultError localizedDescription]);
            [resultError release];
            return resultCode;
        }

        return BRIDGE_OK;
    }
}

// Resolve file extension to UTI(s)
int ResolveUTIsForExtension(const char* extension, char*** outUTIs, int* outCount, char** outError) {
    @autoreleasepool {
        if (!extension || !outUTIs || !outCount) {
            SetError(outError, @"Invalid parameters");
            return BRIDGE_ERROR_INVALID_UTI;
        }

        *outUTIs = NULL;
        *outCount = 0;

        NSString* extString = [NSString stringWithUTF8String:extension];
        if (!extString) {
            SetError(outError, @"Invalid UTF-8 in extension string");
            return BRIDGE_ERROR_INVALID_UTI;
        }

        // Get all UTTypes that match this extension
        NSArray<UTType*>* types = [UTType typesWithTag:extString
                                              tagClass:UTTagClassFilenameExtension
                                     conformingToType:nil];

        if (!types || [types count] == 0) {
            SetError(outError, [NSString stringWithFormat:@"No UTIs found for extension: %s", extension]);
            return BRIDGE_ERROR_NOT_FOUND;
        }

        *outCount = (int)[types count];
        *outUTIs = (char**)malloc(sizeof(char*) * (*outCount));

        if (!*outUTIs) {
            SetError(outError, @"Memory allocation failed");
            return BRIDGE_ERROR_SYSTEM;
        }

        for (int i = 0; i < *outCount; i++) {
            UTType* type = types[i];
            (*outUTIs)[i] = NSStringToCString([type identifier]);
        }

        return BRIDGE_OK;
    }
}

// List all applications that can open a UTI
int ListAppsForUTI(const char* uti, char*** outAppPaths, int* outCount, char** outError) {
    @autoreleasepool {
        if (!uti || !outAppPaths || !outCount) {
            SetError(outError, @"Invalid parameters");
            return BRIDGE_ERROR_INVALID_UTI;
        }

        *outAppPaths = NULL;
        *outCount = 0;

        NSString* utiString = [NSString stringWithUTF8String:uti];
        if (!utiString) {
            SetError(outError, @"Invalid UTF-8 in UTI string");
            return BRIDGE_ERROR_INVALID_UTI;
        }

        UTType* utType = [UTType typeWithIdentifier:utiString];
        if (!utType) {
            SetError(outError, [NSString stringWithFormat:@"Invalid or unknown UTI: %s", uti]);
            return BRIDGE_ERROR_INVALID_UTI;
        }

        NSWorkspace* workspace = [NSWorkspace sharedWorkspace];
        NSArray<NSURL*>* appURLs = [workspace URLsForApplicationsToOpenContentType:utType];

        if (!appURLs || [appURLs count] == 0) {
            // Return empty list, not an error
            return BRIDGE_OK;
        }

        *outCount = (int)[appURLs count];
        *outAppPaths = (char**)malloc(sizeof(char*) * (*outCount));

        if (!*outAppPaths) {
            SetError(outError, @"Memory allocation failed");
            return BRIDGE_ERROR_SYSTEM;
        }

        for (int i = 0; i < *outCount; i++) {
            NSURL* appURL = appURLs[i];
            (*outAppPaths)[i] = URLToPath(appURL);
        }

        return BRIDGE_OK;
    }
}

// List all applications that can handle a URL scheme
int ListAppsForScheme(const char* scheme, char*** outAppPaths, int* outCount, char** outError) {
    @autoreleasepool {
        if (!scheme || !outAppPaths || !outCount) {
            SetError(outError, @"Invalid parameters");
            return BRIDGE_ERROR_INVALID_SCHEME;
        }

        *outAppPaths = NULL;
        *outCount = 0;

        NSString* schemeString = [NSString stringWithUTF8String:scheme];
        if (!schemeString) {
            SetError(outError, @"Invalid UTF-8 in scheme string");
            return BRIDGE_ERROR_INVALID_SCHEME;
        }

        NSString* urlString = [NSString stringWithFormat:@"%@://", schemeString];
        NSURL* url = [NSURL URLWithString:urlString];

        if (!url) {
            SetError(outError, [NSString stringWithFormat:@"Invalid URL scheme: %s", scheme]);
            return BRIDGE_ERROR_INVALID_SCHEME;
        }

        NSWorkspace* workspace = [NSWorkspace sharedWorkspace];
        NSArray<NSURL*>* appURLs = [workspace URLsForApplicationsToOpenURL:url];

        if (!appURLs || [appURLs count] == 0) {
            // Return empty list, not an error
            return BRIDGE_OK;
        }

        *outCount = (int)[appURLs count];
        *outAppPaths = (char**)malloc(sizeof(char*) * (*outCount));

        if (!*outAppPaths) {
            SetError(outError, @"Memory allocation failed");
            return BRIDGE_ERROR_SYSTEM;
        }

        for (int i = 0; i < *outCount; i++) {
            NSURL* appURL = appURLs[i];
            (*outAppPaths)[i] = URLToPath(appURL);
        }

        return BRIDGE_OK;
    }
}

// List all installed applications on the system
int ListAllApplications(AppInfo*** outApps, int* outCount, char** outError) {
    @autoreleasepool {
        if (!outApps || !outCount) {
            SetError(outError, @"Invalid parameters");
            return BRIDGE_ERROR_SYSTEM;
        }

        *outApps = NULL;
        *outCount = 0;

        NSMutableArray<NSDictionary*>* appInfoList = [NSMutableArray array];
        NSFileManager* fileManager = [NSFileManager defaultManager];
        NSWorkspace* workspace = [NSWorkspace sharedWorkspace];

        // Standard application directories to search
        NSArray<NSString*>* appDirectories = @[
            @"/Applications",
            @"/System/Applications",
            @"/System/Applications/Utilities",
            [@"~/Applications" stringByExpandingTildeInPath]
        ];

        // Enumerate all application directories
        for (NSString* directory in appDirectories) {
            BOOL isDirectory;
            if (![fileManager fileExistsAtPath:directory isDirectory:&isDirectory] || !isDirectory) {
                continue;
            }

            NSError* error = nil;
            NSArray<NSString*>* contents = [fileManager contentsOfDirectoryAtPath:directory error:&error];

            if (error) {
                continue; // Skip directories we can't read
            }

            for (NSString* item in contents) {
                if (![item hasSuffix:@".app"]) {
                    continue;
                }

                NSString* fullPath = [directory stringByAppendingPathComponent:item];
                NSURL* appURL = [NSURL fileURLWithPath:fullPath];

                // Get bundle information
                NSBundle* bundle = [NSBundle bundleWithURL:appURL];
                if (!bundle) {
                    continue;
                }

                NSString* bundleID = [bundle bundleIdentifier];
                NSString* appName = [bundle objectForInfoDictionaryKey:@"CFBundleName"];

                // Fallback to display name if CFBundleName is not available
                if (!appName) {
                    appName = [bundle objectForInfoDictionaryKey:@"CFBundleDisplayName"];
                }

                // Fallback to filename without .app extension
                if (!appName) {
                    appName = [[item lastPathComponent] stringByDeletingPathExtension];
                }

                // Store app info
                NSDictionary* appInfo = @{
                    @"name": appName ?: @"",
                    @"path": fullPath,
                    @"bundleID": bundleID ?: @""
                };

                [appInfoList addObject:appInfo];
            }
        }

        // Also get apps from LaunchServices (catches apps in other locations)
        NSArray<NSURL*>* allApps = [workspace URLsForApplicationsToOpenContentType:[UTType typeWithIdentifier:@"public.item"]];

        NSMutableSet<NSString*>* existingPaths = [NSMutableSet set];
        for (NSDictionary* info in appInfoList) {
            [existingPaths addObject:info[@"path"]];
        }

        for (NSURL* appURL in allApps) {
            NSString* fullPath = [appURL path];

            // Skip if we already have this app
            if ([existingPaths containsObject:fullPath]) {
                continue;
            }

            NSBundle* bundle = [NSBundle bundleWithURL:appURL];
            if (!bundle) {
                continue;
            }

            NSString* bundleID = [bundle bundleIdentifier];
            NSString* appName = [bundle objectForInfoDictionaryKey:@"CFBundleName"];

            if (!appName) {
                appName = [bundle objectForInfoDictionaryKey:@"CFBundleDisplayName"];
            }

            if (!appName) {
                appName = [[[fullPath lastPathComponent] stringByDeletingPathExtension] copy];
            }

            NSDictionary* appInfo = @{
                @"name": appName ?: @"",
                @"path": fullPath,
                @"bundleID": bundleID ?: @""
            };

            [appInfoList addObject:appInfo];
            [existingPaths addObject:fullPath];
        }

        // Convert to C array
        *outCount = (int)[appInfoList count];

        if (*outCount == 0) {
            return BRIDGE_OK;
        }

        *outApps = (AppInfo**)malloc(sizeof(AppInfo*) * (*outCount));
        if (!*outApps) {
            SetError(outError, @"Memory allocation failed");
            return BRIDGE_ERROR_SYSTEM;
        }

        for (int i = 0; i < *outCount; i++) {
            NSDictionary* appInfo = appInfoList[i];

            (*outApps)[i] = (AppInfo*)malloc(sizeof(AppInfo));
            if (!(*outApps)[i]) {
                // Clean up previously allocated memory
                for (int j = 0; j < i; j++) {
                    if ((*outApps)[j]->name) free((*outApps)[j]->name);
                    if ((*outApps)[j]->path) free((*outApps)[j]->path);
                    if ((*outApps)[j]->bundleID) free((*outApps)[j]->bundleID);
                    free((*outApps)[j]);
                }
                free(*outApps);
                *outApps = NULL;
                *outCount = 0;
                SetError(outError, @"Memory allocation failed");
                return BRIDGE_ERROR_SYSTEM;
            }

            (*outApps)[i]->name = NSStringToCString(appInfo[@"name"]);
            (*outApps)[i]->path = NSStringToCString(appInfo[@"path"]);
            (*outApps)[i]->bundleID = NSStringToCString(appInfo[@"bundleID"]);
        }

        return BRIDGE_OK;
    }
}

// Free a single C string
void FreeCString(char* str) {
    if (str) {
        free(str);
    }
}

// Free an array of C strings
void FreeCStringArray(char** arr, int count) {
    if (arr) {
        for (int i = 0; i < count; i++) {
            if (arr[i]) {
                free(arr[i]);
            }
        }
        free(arr);
    }
}

// Free an array of AppInfo structures
void FreeAppInfoArray(AppInfo** apps, int count) {
    if (apps) {
        for (int i = 0; i < count; i++) {
            if (apps[i]) {
                if (apps[i]->name) free(apps[i]->name);
                if (apps[i]->path) free(apps[i]->path);
                if (apps[i]->bundleID) free(apps[i]->bundleID);
                free(apps[i]);
            }
        }
        free(apps);
    }
}
