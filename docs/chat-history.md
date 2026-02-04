# Chat History

## 2026-02-04
- You asked for a project status snapshot.
- I reported repo structure and docs, and asked what area to dive into.
- You requested inline task entry below the list (editor-like).
- I updated `frontend/src/App.tsx` to replace the top input bar with a contentEditable line below the tasks; Enter creates a task and focus stays in the editor.
- You asked to keep chat history in the repo; we agreed on `docs/chat-history.md`.
- You asked to add created and due dates to tasks.
- I updated the backend to store due dates on create and to update due dates later, and updated the frontend to display created dates and allow setting a due date via a date input.
- You asked whether any of the unexpected files should be in .gitignore.
- You asked to add recommended entries to `.gitignore`; I added `taskminus.db`, `frontend/node_modules/`, and `frontend/dist/`, and untracked `taskminus.db` and `frontend/dist`.
- Tests were run with `go test ./...`.
