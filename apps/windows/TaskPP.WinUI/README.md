# TaskPP.WinUI

WinUI 3 (.NET) shell for the Go core.

## Build (Windows)
1. Build the Go DLL:
   - Run `scripts/build-core-windows.ps1` to produce `build/windows/taskppcore.dll` and `taskppcore.h`.
2. Ensure the DLL is available to the WinUI app (copy to output or add to PATH).
3. Open this project in Visual Studio and run.

## Notes
- P/Invoke bindings are in `CoreNative.cs`.
- All Go core calls return JSON strings; free them with `Core_FreeString`.
