# Run Desktop (Wails)

## Prereqs
- Wails CLI installed
- Node.js + npm
- Go 1.22+

## Steps
1. Install frontend deps:
   - `cd frontend`
   - `npm install`

2. Run the app (dev mode):
  - `wails dev`

3. Build production desktop app:
  - `wails build`

## Notes
- Frontend build output is copied into `cmd/desktop/assets` by Wails.
- If you want manual build, you can run `scripts/build-frontend.sh`.
