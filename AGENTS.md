# Agent Instructions

## macOS Deployment Notes (Important)
- macOS builds must be done on macOS. Wails cannot build mac targets on Windows.
- If deploying to real users, plan for codesigning + notarization (Apple Developer account required).
- For quick testing, unsigned builds are possible but Gatekeeper warnings will appear.
- Decide whether to build for Apple Silicon (arm64), Intel (amd64), or universal (both).
- Packaging is typically a .dmg (or .pkg).

## Recommended Decision Checklist
1. Build location: local Mac or CI (GitHub Actions macOS runner).
2. Signing: unsigned (dev only) vs signed + notarized (production).
3. Architecture: arm64, amd64, or universal.
