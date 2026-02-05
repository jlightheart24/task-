# Desktop UI Stack

Chosen direction:
- macOS: SwiftUI (native)
- iOS: SwiftUI (native)
- Windows: WinUI 3 (.NET)
- Shared core: Go (bind-safe API)

## Integration Plan
- macOS + iOS: use `gomobile bind` to generate an XCFramework from `taskpp/core/bind`.
- Windows: expose Go core as a C-ABI DLL and consume via P/Invoke in WinUI 3.

## Next Steps
- Decide API surface to expose for gomobile + C-ABI.
- Implement build scripts for generating XCFramework and DLL.
- Scaffold native apps to call `Core` and show a minimal task list.
