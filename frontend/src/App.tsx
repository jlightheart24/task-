import { useEffect, useRef, useState } from "react";

type Task = {
  id: string;
  title: string;
  status: string;
  created_at?: string;
  due_date?: string;
  order?: number;
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
  const dragStateRef = useRef<{ taskId: string; dueKey: string } | null>(null);
  const [draggingId, setDraggingId] = useState<string | null>(null);
  const [dragOverId, setDragOverId] = useState<string | null>(null);
  const dragUserSelectRef = useRef<string>("");

  const getTaskSortValue = (task: Task) => {
    if (typeof task.order === "number" && task.order > 0) {
      return task.order;
    }
    if (task.created_at) {
      const parsed = new Date(task.created_at);
      if (!Number.isNaN(parsed.getTime())) {
        return parsed.getTime();
      }
    }
    return 0;
  };

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
    setTasks((prev) =>
      prev.map((task) => {
        if (typeof task.order === "number" && task.order > 0) {
          return task;
        }
        return {
          ...task,
          order: getTaskSortValue(task),
        };
      })
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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

  const updateTaskOrder = (id: string, order: number) => {
    const wails = (window as unknown as { go?: any }).go;
    const update = wails?.app?.App?.UpdateTaskOrder;
    if (typeof update !== "function") {
      return;
    }
    update(id, order).catch(() => setMessage("backend error"));
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

  const reorderWithinDueDate = (dueKey: string, sourceId: string, targetId: string) => {
    if (sourceId === targetId) {
      return;
    }
    setTasks((prev) => {
      const grouped = prev.filter((task) => (toDateKey(task.due_date) || "no-due") === dueKey);
      const others = prev.filter((task) => (toDateKey(task.due_date) || "no-due") !== dueKey);

      const sourceIndex = grouped.findIndex((task) => task.id === sourceId);
      const targetIndex = grouped.findIndex((task) => task.id === targetId);
      if (sourceIndex === -1 || targetIndex === -1) {
        return prev;
      }

      const reordered = [...grouped];
      const [moved] = reordered.splice(sourceIndex, 1);
      reordered.splice(targetIndex, 0, moved);

      const updated = reordered.map((task, index) => {
        const newOrder = index + 1;
        if (task.order !== newOrder) {
          updateTaskOrder(task.id, newOrder);
          return { ...task, order: newOrder };
        }
        return task;
      });

      return [...others, ...updated];
    });
  };

  return (
    <div>
      <style>{`
        .task-delete {
          opacity: 0;
          visibility: hidden;
          transition: opacity 0.12s ease;
          color: #333;
        }
        .task-item:hover .task-delete {
          opacity: 1;
          visibility: visible;
        }
        .task-item {
          user-select: none;
          -webkit-user-select: none;
        }
        .task-item.dragging {
          opacity: 0.6;
        }
        .task-item.drag-over {
          outline: 2px dashed #bbb;
          outline-offset: 2px;
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
              {[...tasksByDueDate[dueKey]]
                .sort((a, b) => getTaskSortValue(a) - getTaskSortValue(b))
                .map((task) => (
                <li
                  key={task.id}
                  className={`task-item${draggingId === task.id ? " dragging" : ""}${dragOverId === task.id ? " drag-over" : ""}`}
                  draggable
                  onDragStart={(event) => {
                    event.dataTransfer.setData("text/plain", task.id);
                    event.dataTransfer.effectAllowed = "move";
                    dragStateRef.current = { taskId: task.id, dueKey };
                    setDraggingId(task.id);
                    dragUserSelectRef.current = document.body.style.userSelect || "";
                    document.body.style.userSelect = "none";
                  }}
                  onDragEnd={() => {
                    dragStateRef.current = null;
                    setDraggingId(null);
                    setDragOverId(null);
                    document.body.style.userSelect = dragUserSelectRef.current;
                  }}
                  onDragEnter={(event) => {
                    event.preventDefault();
                    setDragOverId(task.id);
                  }}
                  onDragOver={(event) => {
                    event.preventDefault();
                    event.dataTransfer.dropEffect = "move";
                  }}
                  onDrop={(event) => {
                    event.preventDefault();
                    event.stopPropagation();
                    const dragState = dragStateRef.current;
                    if (!dragState || dragState.dueKey !== dueKey) {
                      return;
                    }
                    reorderWithinDueDate(dueKey, dragState.taskId, task.id);
                    dragStateRef.current = null;
                    setDraggingId(null);
                    setDragOverId(null);
                  }}
                >
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
                        fontSize: "16px",
                        lineHeight: 1,
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
