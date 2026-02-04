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
- You asked to group tasks into different days based on due date; I grouped the task list by due date with a header per day and a No due date bucket.
- You reported due dates defaulting to 12/31/1; I treated zero/0001 dates as empty so they fall into No due date and don't display.
- You asked for new tasks to default due date to today; I set CreateTask to use today when no due date is provided.
- You reported tasks sorting into the previous date; I switched due-date grouping and inputs to use local date parts instead of UTC.
- You asked to remove the title and greeting text from the task view and remove status labels from task rows; I removed those UI elements.
- You asked for a Settings tab with a toggle for date headers; I added a Settings tab and a date header mode (date vs weekday) with localStorage persistence.
- You asked to add a date header mode that shows both weekday and date; I added a "Show both" option.
- You asked to change the delete button to an X that appears on hover; I updated the task row to show a hover-only "×" button.
- You reported the hover "×" delete button not appearing; I changed the hover target to the full list item and bumped the glyph size.
- You reported the hover delete "×" still not appearing; I moved its visibility/opacity into CSS and added an explicit color so it shows on hover.
