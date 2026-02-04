import { useEffect, useRef, useState } from "react";

type Task = {
  id: string;
  title: string;
  status: string;
  created_at?: string;
  due_date?: string;
};

export function App() {
  const [message, setMessage] = useState<string>("backend not connected");
  const [tasks, setTasks] = useState<Task[]>([]);
  const [draft, setDraft] = useState<string>("");
  const editorRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const wails = (window as unknown as { go?: any }).go;
    const greet = wails?.app?.App?.Greet;
    const list = wails?.app?.App?.ListTasks;
    if (typeof greet === "function") {
      greet("jonny")
        .then((result: string) => setMessage(result))
        .catch(() => setMessage("backend error"));
    }
    if (typeof list === "function") {
      list()
        .then((result: Task[]) => setTasks(Array.isArray(result) ? result : []))
        .catch(() => setMessage("backend error"));
    }
  }, []);

  const createTask = () => {
    const wails = (window as unknown as { go?: any }).go;
    const create = wails?.app?.App?.CreateTask;
    const trimmed = draft.trim();
    if (typeof create !== "function" || trimmed === "") {
      return;
    }
    create(trimmed, "")
      .then((task: Task) => {
        setTasks((prev) => [...prev, task]);
        setDraft("");
        if (editorRef.current) {
          editorRef.current.innerText = "";
          editorRef.current.focus();
        }
      })
      .catch(() => setMessage("backend error"));
  };

  const toggleTask = (id: string) => {
    const wails = (window as unknown as { go?: any }).go;
    const toggle = wails?.app?.App?.ToggleTaskComplete;
    if (typeof toggle !== "function") {
      return;
    }
    toggle(id)
      .then((updated: Task) => {
        setTasks((prev) =>
          prev.map((task) => (task.id === updated.id ? updated : task))
        );
      })
      .catch(() => setMessage("backend error"));
  };

  const deleteTask = (id: string) => {
    const wails = (window as unknown as { go?: any }).go;
    const del = wails?.app?.App?.DeleteTask;
    if (typeof del !== "function") {
      return;
    }
    del(id)
      .then(() => {
        setTasks((prev) => prev.filter((task) => task.id !== id));
      })
      .catch(() => setMessage("backend error"));
  };

  const setDueDate = (id: string, dueDate: string) => {
    const wails = (window as unknown as { go?: any }).go;
    const update = wails?.app?.App?.UpdateTaskDueDate;
    if (typeof update !== "function") {
      return;
    }
    update(id, dueDate)
      .then((updated: Task) => {
        setTasks((prev) =>
          prev.map((task) => (task.id === updated.id ? updated : task))
        );
      })
      .catch(() => setMessage("backend error"));
  };

  const formatDate = (value?: string) => {
    if (!value) {
      return "";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return "";
    }
    return date.toLocaleDateString();
  };

  const formatInputDate = (value?: string) => {
    if (!value) {
      return "";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return "";
    }
    const year = date.getUTCFullYear();
    const month = String(date.getUTCMonth() + 1).padStart(2, "0");
    const day = String(date.getUTCDate()).padStart(2, "0");
    return `${year}-${month}-${day}`;
  };

  return (
    <div>
      <h1>task-</h1>
      <p>{message}</p>
      <div
        onClick={() => editorRef.current?.focus()}
        style={{ cursor: "text", padding: "8px 0" }}
      >
        <ul>
        {tasks.map((task) => (
          <li key={task.id}>
            <div>
              <label style={{ textDecoration: task.status === "done" ? "line-through" : "none" }}>
                <input
                  type="checkbox"
                  checked={task.status === "done"}
                  onChange={() => toggleTask(task.id)}
                />
                {task.title} ({task.status})
              </label>
              <button type="button" onClick={() => deleteTask(task.id)}>
                delete
              </button>
            </div>
            <div style={{ fontSize: "0.85rem", color: "#666", marginTop: "4px" }}>
              <span>Created: {formatDate(task.created_at) || "unknown"}</span>
              <label style={{ marginLeft: "12px" }}>
                Due:
                <input
                  type="date"
                  value={formatInputDate(task.due_date)}
                  onChange={(event) => setDueDate(task.id, event.target.value)}
                  style={{ marginLeft: "6px" }}
                />
              </label>
            </div>
          </li>
        ))}
        </ul>
        <div style={{ position: "relative", minHeight: "28px" }}>
          {draft.trim() === "" ? (
            <div
              style={{
                position: "absolute",
                inset: 0,
                color: "#888",
                pointerEvents: "none",
                padding: "2px 0",
              }}
            >
              Click here and type a task...
            </div>
          ) : null}
          <div
            ref={editorRef}
            contentEditable
            suppressContentEditableWarning
            style={{ outline: "none", minHeight: "28px", padding: "2px 0" }}
            onInput={(event) => {
              const target = event.currentTarget;
              setDraft(target.innerText);
            }}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                event.preventDefault();
                createTask();
              }
            }}
          />
        </div>
      </div>
    </div>
  );
}
