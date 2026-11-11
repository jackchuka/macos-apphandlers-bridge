# Agent guide

This library is the one place where Go code talks to macOS default-handler APIs: LaunchServices, NSWorkspace, CFPreferences (`LSHandlers`), `lsregister`, and Info.plist reads. Callers such as dutix don't use cgo or run system tools themselves, so when they need something new, it's added here.

## Adding a function

1. **`bridge.h`**: declare it with a comment on parameters, ownership ("caller must free using ..."), and return codes. Return `BRIDGE_OK` or a `BRIDGE_ERROR_*` code, and report details through `outError`.
2. **`bridge.m`**: implement it inside `@autoreleasepool`. The file uses manual reference counting: balance `Copy`/`Create` with `CFRelease` and use `autorelease` for objects you allocate. Return strings with `NSStringToCString`. For arrays and structs, add a matching `Free...` function that also frees the array itself.
3. **`bridge.go`**: wrap it in a Go function that returns Go types and an `error` from `cErrorToGoError`, and frees the C memory with `defer`.
4. **`bridge_test.go`**: add a test. Tests must not change the user's defaults or registrations; read-only checks are fine, and changes must be restored.
5. **`README.md`**: document it under API Reference.

## Checks

```bash
CGO_ENABLED=1 go vet ./...
CGO_ENABLED=1 go test ./...
```
