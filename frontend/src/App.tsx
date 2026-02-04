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
  const [activeTab, setActiveTab] = useState<"tasks" | "settings">("tasks");
  const [dateHeaderMode, setDateHeaderMode] = useState<"date" | "weekday" | "both">(() => {
    const stored = window.localStorage.getItem("taskpp.dateHeaderMode");
    if (stored === "weekday" || stored === "date" || stored === "both") {
      return stored;
    }
    return "date";
  });
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

  useEffect(() => {
    window.localStorage.setItem("taskpp.dateHeaderMode", dateHeaderMode);
  }, [dateHeaderMode]);

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

  const isZeroDateString = (value?: string) => {
    if (!value) {
      return true;
    }
    return value.startsWith("0001-01-01");
  };

  const formatDate = (value?: string) => {
    if (!value) {
      return "";
    }
    if (isZeroDateString(value)) {
      return "";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return "";
    }
    if (date.getUTCFullYear() <= 1) {
      return "";
    }
    return date.toLocaleDateString();
  };

  const formatInputDate = (value?: string) => {
    if (!value) {
      return "";
    }
    if (isZeroDateString(value)) {
      return "";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return "";
    }
    if (date.getFullYear() <= 1) {
      return "";
    }
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    return `${year}-${month}-${day}`;
  };

  const toDateKey = (value?: string) => {
    if (!value) {
      return "";
    }
    if (isZeroDateString(value)) {
      return "";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return "";
    }
    if (date.getFullYear() <= 1) {
      return "";
    }
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    return `${year}-${month}-${day}`;
  };

  const formatDateKey = (value: string) => {
    if (!value) {
      return "No due date";
    }
    const date = new Date(`${value}T00:00:00`);
    if (Number.isNaN(date.getTime())) {
      return "No due date";
    }
    if (dateHeaderMode === "weekday") {
      return date.toLocaleDateString(undefined, { weekday: "long" });
    }
    if (dateHeaderMode === "both") {
      const weekday = date.toLocaleDateString(undefined, { weekday: "long" });
      const fullDate = date.toLocaleDateString();
      return `${weekday} · ${fullDate}`;
    }
    return date.toLocaleDateString();
  };

  const tasksByDueDate = tasks.reduce<Record<string, Task[]>>((acc, task) => {
    const key = toDateKey(task.due_date) || "no-due";
    if (!acc[key]) {
      acc[key] = [];
    }
    acc[key].push(task);
    return acc;
  }, {});

  const sortedDueDateKeys = Object.keys(tasksByDueDate).sort((a, b) => {
    if (a === "no-due") return 1;
    if (b === "no-due") return -1;
    return a.localeCompare(b);
  });

  return (
    <div>
      <style>{`
        .task-row:hover .task-delete {
          opacity: 1;
        }
      `}</style>
      <div style={{ display: "flex", gap: "8px", marginBottom: "12px" }}>
        <button
          type="button"
          onClick={() => setActiveTab("tasks")}
          style={{ fontWeight: activeTab === "tasks" ? 600 : 400 }}
        >
          Tasks
        </button>
        <button
          type="button"
          onClick={() => setActiveTab("settings")}
          style={{ fontWeight: activeTab === "settings" ? 600 : 400 }}
        >
          Settings
        </button>
      </div>
      {activeTab === "settings" ? (
        <div>
          <div style={{ marginBottom: "8px", fontWeight: 600 }}>
            Date Display
          </div>
          <label style={{ marginRight: "12px" }}>
            <input
              type="radio"
              name="dateHeaderMode"
              value="date"
              checked={dateHeaderMode === "date"}
              onChange={() => setDateHeaderMode("date")}
            />
            <span style={{ marginLeft: "6px" }}>Show date</span>
          </label>
          <label>
            <input
              type="radio"
              name="dateHeaderMode"
              value="weekday"
              checked={dateHeaderMode === "weekday"}
              onChange={() => setDateHeaderMode("weekday")}
            />
            <span style={{ marginLeft: "6px" }}>Show weekday</span>
          </label>
          <label style={{ marginLeft: "12px" }}>
            <input
              type="radio"
              name="dateHeaderMode"
              value="both"
              checked={dateHeaderMode === "both"}
              onChange={() => setDateHeaderMode("both")}
            />
            <span style={{ marginLeft: "6px" }}>Show both</span>
          </label>
        </div>
      ) : null}
      {activeTab === "tasks" ? (
      <div
        onClick={() => editorRef.current?.focus()}
        style={{ cursor: "text", padding: "8px 0" }}
      >
        <ul>
        {sortedDueDateKeys.map((dueKey) => (
          <li key={dueKey} style={{ listStyle: "none", marginBottom: "12px" }}>
            <div style={{ fontWeight: 600, marginBottom: "6px" }}>
              {formatDateKey(dueKey === "no-due" ? "" : dueKey)}
            </div>
            <ul>
              {tasksByDueDate[dueKey].map((task) => (
                <li key={task.id}>
                  <div
                    className="task-row"
                    style={{ display: "flex", alignItems: "center", gap: "8px" }}
                  >
                    <label style={{ textDecoration: task.status === "done" ? "line-through" : "none" }}>
                      <input
                        type="checkbox"
                        checked={task.status === "done"}
                        onChange={() => toggleTask(task.id)}
                      />
                      {task.title}
                    </label>
                    <button
                      type="button"
                      onClick={() => deleteTask(task.id)}
                      style={{
                        border: "none",
                        background: "transparent",
                        cursor: "pointer",
                        opacity: 0,
                        transition: "opacity 0.12s ease",
                      }}
                      aria-label="Delete task"
                      title="Delete task"
                      className="task-delete"
                    >
                      ×
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
      ) : null}
    </div>
  );
}
