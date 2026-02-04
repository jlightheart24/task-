import { useEffect, useRef, useState } from "react";

type Task = {
  id: string;
  title: string;
  description?: string;
  status: string;
  priority?: string;
  created_at?: string;
  due_date?: string;
  order?: number;
};

export function App() {
  const [message, setMessage] = useState<string>("backend not connected");
  const [tasks, setTasks] = useState<Task[]>([]);
  const [draft, setDraft] = useState<string>("");
  const [activeTab, setActiveTab] = useState<"tasks" | "settings" | "calendar">("tasks");
  const [dateHeaderMode, setDateHeaderMode] = useState<"date" | "weekday" | "both">(() => {
    const stored = window.localStorage.getItem("taskpp.dateHeaderMode");
    if (stored === "weekday" || stored === "date" || stored === "both") {
      return stored;
    }
    return "date";
  });
  const [weekStartDay, setWeekStartDay] = useState<"sunday" | "monday">(() => {
    const stored = window.localStorage.getItem("taskpp.weekStartDay");
    if (stored === "sunday" || stored === "monday") {
      return stored;
    }
    return "monday";
  });
  const [dateFormat, setDateFormat] = useState<"mdy" | "dmy">(() => {
    const stored = window.localStorage.getItem("taskpp.dateFormat");
    if (stored === "mdy" || stored === "dmy") {
      return stored;
    }
    return "mdy";
  });
  const editorRef = useRef<HTMLDivElement | null>(null);
  const dragStateRef = useRef<{ taskId: string; dueKey: string } | null>(null);
  const [draggingId, setDraggingId] = useState<string | null>(null);
  const [dragOverId, setDragOverId] = useState<string | null>(null);
  const [dragOverPosition, setDragOverPosition] = useState<"above" | "below" | null>(null);
  const [dragOverGroup, setDragOverGroup] = useState<{
    dueKey: string;
    position: "above" | "below";
  } | null>(null);
  const dragUserSelectRef = useRef<string>("");
  const [activeTaskId, setActiveTaskId] = useState<string | null>(null);
  const [detailsDraft, setDetailsDraft] = useState<string>("");
  const [priorityDraft, setPriorityDraft] = useState<string>("normal");
  const [dueDateDraft, setDueDateDraft] = useState<string>("");
  const [weekOffset, setWeekOffset] = useState<number>(0);

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

  useEffect(() => {
    window.localStorage.setItem("taskpp.weekStartDay", weekStartDay);
  }, [weekStartDay]);

  useEffect(() => {
    window.localStorage.setItem("taskpp.dateFormat", dateFormat);
  }, [dateFormat]);

  useEffect(() => {
    const task = tasks.find((item) => item.id === activeTaskId);
    if (!task) {
      return;
    }
    setDetailsDraft(task.description || "");
    setPriorityDraft(task.priority || "normal");
    setDueDateDraft(formatInputDate(task.due_date));
  }, [activeTaskId, tasks]);

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

  const updateTaskDetails = (id: string, description: string, dueDate: string, priority: string) => {
    const wails = (window as unknown as { go?: any }).go;
    const update = wails?.app?.App?.UpdateTaskDetails;
    if (typeof update !== "function") {
      return;
    }
    update(id, description, dueDate, priority)
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

  const extractDateKey = (value?: string) => {
    if (!value) {
      return "";
    }
    if (isZeroDateString(value)) {
      return "";
    }
    const trimmed = value.trim();
    if (trimmed.length >= 10) {
      return trimmed.slice(0, 10);
    }
    return "";
  };

  const formatDateParts = (date: Date) => {
    const day = String(date.getDate()).padStart(2, "0");
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const year = String(date.getFullYear());
    return dateFormat === "dmy"
      ? `${day}/${month}/${year}`
      : `${month}/${day}/${year}`;
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
    return formatDateParts(date);
  };

  const formatInputDate = (value?: string) => {
    return extractDateKey(value);
  };

  const toDateKey = (value?: string) => {
    return extractDateKey(value);
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
      const fullDate = formatDateParts(date);
      return `${weekday} · ${fullDate}`;
    }
    return formatDateParts(date);
  };

  const formatWeekdayShort = (date: Date) =>
    date.toLocaleDateString(undefined, { weekday: "short" });

  const formatMonthDay = (date: Date) => {
    const day = String(date.getDate());
    const monthName = date.toLocaleDateString(undefined, { month: "short" });
    return dateFormat === "dmy" ? `${day} ${monthName}` : `${monthName} ${day}`;
  };

  const dateToKeyLocal = (date: Date) => {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    return `${year}-${month}-${day}`;
  };

  const startOfWeek = (date: Date) => {
    const day = date.getDay(); // 0 = Sun
    const diff =
      weekStartDay === "monday"
        ? (day + 6) % 7
        : day;
    const start = new Date(date);
    start.setDate(date.getDate() - diff);
    start.setHours(0, 0, 0, 0);
    return start;
  };

  const weekDates = (() => {
    const base = new Date();
    base.setDate(base.getDate() + weekOffset * 7);
    const start = startOfWeek(base);
    return Array.from({ length: 7 }, (_, index) => {
      const next = new Date(start);
      next.setDate(start.getDate() + index);
      return next;
    });
  })();

  const tasksByDateKey = tasks.reduce<Record<string, Task[]>>((acc, task) => {
    const key = toDateKey(task.due_date);
    if (!key) {
      return acc;
    }
    if (!acc[key]) {
      acc[key] = [];
    }
    acc[key].push(task);
    return acc;
  }, {});

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

  const reorderGroup = (
    grouped: Task[],
    sourceId: string,
    targetId: string,
    position: "above" | "below"
  ) => {
    const sourceIndex = grouped.findIndex((task) => task.id === sourceId);
    const targetIndex = grouped.findIndex((task) => task.id === targetId);
    if (sourceIndex === -1 || targetIndex === -1) {
      return grouped;
    }

    const reordered = [...grouped];
    const [moved] = reordered.splice(sourceIndex, 1);

    let insertIndex = targetIndex;
    if (position === "below") {
      insertIndex += 1;
    }
    if (sourceIndex < insertIndex) {
      insertIndex -= 1;
    }
    reordered.splice(insertIndex, 0, moved);
    return reordered;
  };

  const applyOrderUpdates = (grouped: Task[]) =>
    grouped.map((task, index) => {
      const newOrder = index + 1;
      if (task.order !== newOrder) {
        updateTaskOrder(task.id, newOrder);
        return { ...task, order: newOrder };
      }
      return task;
    });

  const reorderWithinDueDate = (
    dueKey: string,
    sourceId: string,
    targetId: string,
    position: "above" | "below"
  ) => {
    if (sourceId === targetId) {
      return;
    }
    setTasks((prev) => {
      const grouped = prev.filter((task) => (toDateKey(task.due_date) || "no-due") === dueKey);
      const others = prev.filter((task) => (toDateKey(task.due_date) || "no-due") !== dueKey);

      const reordered = reorderGroup(grouped, sourceId, targetId, position);
      const updated = applyOrderUpdates(reordered);
      return [...others, ...updated];
    });
  };

  const moveTaskToDueDate = (
    sourceId: string,
    targetDueKey: string,
    targetId?: string,
    position: "above" | "below" = "below"
  ) => {
    setTasks((prev) => {
      const source = prev.find((task) => task.id === sourceId);
      if (!source) {
        return prev;
      }

      const sourceDueKey = toDateKey(source.due_date) || "no-due";
      const targetDate = targetDueKey === "no-due" ? "" : targetDueKey;

      if (sourceDueKey === targetDueKey) {
        if (targetId) {
          const grouped = prev.filter(
            (task) => (toDateKey(task.due_date) || "no-due") === targetDueKey
          );
          const others = prev.filter(
            (task) => (toDateKey(task.due_date) || "no-due") !== targetDueKey
          );
          const reordered = reorderGroup(grouped, sourceId, targetId, position);
          const updated = applyOrderUpdates(reordered);
          return [...others, ...updated];
        }
        return prev;
      }

      const withUpdatedDue = prev.map((task) =>
        task.id === sourceId ? { ...task, due_date: targetDate } : task
      );

      const grouped = withUpdatedDue.filter(
        (task) => (toDateKey(task.due_date) || "no-due") === targetDueKey
      );
      const others = withUpdatedDue.filter(
        (task) => (toDateKey(task.due_date) || "no-due") !== targetDueKey
      );

      let reordered = grouped;
      if (targetId && grouped.some((task) => task.id === targetId)) {
        reordered = reorderGroup(grouped, sourceId, targetId, position);
      } else {
        reordered = [...grouped];
      }

      const updated = applyOrderUpdates(reordered);

      setDueDate(sourceId, targetDate);

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
          position: relative;
          user-select: none;
          -webkit-user-select: none;
        }
        .task-item.dragging {
          opacity: 0.6;
        }
        .task-item.drag-over-above::before,
        .task-item.drag-over-below::after {
          content: "";
          position: absolute;
          left: 0;
          right: 0;
          border-top: 2px solid #999;
        }
        .task-item.drag-over-above::before {
          top: -4px;
        }
        .task-item.drag-over-below::after {
          bottom: -4px;
        }
        .group-drop {
          position: relative;
          height: 10px;
          margin: 4px 0;
        }
        .group-drop.active::before {
          content: "";
          position: absolute;
          left: 0;
          right: 0;
          top: 4px;
          border-top: 2px solid #666;
        }
        .modal-backdrop {
          position: fixed;
          inset: 0;
          background: rgba(0, 0, 0, 0.35);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 50;
        }
        .modal {
          background: #fff;
          width: min(520px, 92vw);
          border-radius: 8px;
          padding: 16px;
          box-shadow: 0 18px 40px rgba(0, 0, 0, 0.2);
        }
        .modal-actions {
          display: flex;
          gap: 8px;
          justify-content: flex-end;
          margin-top: 12px;
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
          onClick={() => setActiveTab("calendar")}
          style={{ fontWeight: activeTab === "calendar" ? 600 : 400 }}
        >
          Calendar
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
          <div style={{ marginTop: "16px", marginBottom: "8px", fontWeight: 600 }}>
            Week Starts On
          </div>
          <label style={{ marginRight: "12px" }}>
            <input
              type="radio"
              name="weekStartDay"
              value="sunday"
              checked={weekStartDay === "sunday"}
              onChange={() => setWeekStartDay("sunday")}
            />
            <span style={{ marginLeft: "6px" }}>Sunday</span>
          </label>
          <label>
            <input
              type="radio"
              name="weekStartDay"
              value="monday"
              checked={weekStartDay === "monday"}
              onChange={() => setWeekStartDay("monday")}
            />
            <span style={{ marginLeft: "6px" }}>Monday</span>
          </label>
          <div style={{ marginTop: "16px", marginBottom: "8px", fontWeight: 600 }}>
            Date Format
          </div>
          <label style={{ marginRight: "12px" }}>
            <input
              type="radio"
              name="dateFormat"
              value="mdy"
              checked={dateFormat === "mdy"}
              onChange={() => setDateFormat("mdy")}
            />
            <span style={{ marginLeft: "6px" }}>Month/Day/Year</span>
          </label>
          <label>
            <input
              type="radio"
              name="dateFormat"
              value="dmy"
              checked={dateFormat === "dmy"}
              onChange={() => setDateFormat("dmy")}
            />
            <span style={{ marginLeft: "6px" }}>Day/Month/Year</span>
          </label>
        </div>
      ) : null}
      {activeTab === "calendar" ? (
        <div>
          <div style={{ display: "flex", alignItems: "center", gap: "8px", marginBottom: "12px" }}>
            <button type="button" onClick={() => setWeekOffset((prev) => prev - 1)}>
              Prev
            </button>
            <button type="button" onClick={() => setWeekOffset(0)}>
              Today
            </button>
            <button type="button" onClick={() => setWeekOffset((prev) => prev + 1)}>
              Next
            </button>
          </div>
          <div style={{ display: "grid", gridTemplateColumns: "repeat(7, minmax(0, 1fr))", gap: "8px" }}>
          {weekDates.map((date) => {
            const dateKey = dateToKeyLocal(date);
            const dayTasks = tasksByDateKey[dateKey] || [];
            return (
              <div key={dateKey} style={{ border: "1px solid #ddd", borderRadius: "6px", padding: "8px" }}>
                <div style={{ fontWeight: 600, marginBottom: "6px" }}>
                  {formatWeekdayShort(date)} {formatMonthDay(date)}
                </div>
                {dayTasks.length === 0 ? (
                  <div style={{ color: "#888", fontSize: "0.85rem" }}>No tasks</div>
                ) : (
                  <ul>
                    {dayTasks
                      .sort((a, b) => getTaskSortValue(a) - getTaskSortValue(b))
                      .map((task) => (
                        <li key={task.id} style={{ marginBottom: "4px" }}>
                          <button
                            type="button"
                            onClick={() => setActiveTaskId(task.id)}
                            style={{
                              border: "none",
                              background: "transparent",
                              padding: 0,
                              cursor: "pointer",
                              textAlign: "left",
                              font: "inherit",
                              textDecoration: task.status === "done" ? "line-through" : "none",
                            }}
                          >
                            {task.title}
                          </button>
                        </li>
                      ))}
                  </ul>
                )}
              </div>
            );
          })}
          </div>
        </div>
      ) : null}
      {activeTab === "tasks" ? (
      <div
        onClick={() => editorRef.current?.focus()}
        style={{ cursor: "text", padding: "8px 0" }}
      >
        <ul>
        {sortedDueDateKeys.map((dueKey, dueIndex) => {
          const prevDueKey = dueIndex > 0 ? sortedDueDateKeys[dueIndex - 1] : "";
          const currentGroup = tasksByDueDate[dueKey];
          const sortedGroup = [...currentGroup].sort(
            (a, b) => getTaskSortValue(a) - getTaskSortValue(b)
          );
          const firstTaskId = sortedGroup[0]?.id;

          return (
          <li key={dueKey} style={{ listStyle: "none", marginBottom: "12px" }}>
            {dueIndex > 0 ? (
              <div
                className={`group-drop${
                  dragOverGroup?.dueKey === dueKey && dragOverGroup.position === "above"
                    ? " active"
                    : ""
                }`}
                onDragOver={(event) => {
                  event.preventDefault();
                  event.dataTransfer.dropEffect = "move";
                  setDragOverGroup({ dueKey, position: "above" });
                }}
                onDrop={(event) => {
                  event.preventDefault();
                  event.stopPropagation();
                  const dragState = dragStateRef.current;
                  if (!dragState) {
                    return;
                  }
                  const targetDueKey = prevDueKey || dueKey;
                  if (firstTaskId) {
                    moveTaskToDueDate(dragState.taskId, targetDueKey, firstTaskId, "above");
                  } else {
                    moveTaskToDueDate(dragState.taskId, targetDueKey);
                  }
                  dragStateRef.current = null;
                  setDraggingId(null);
                  setDragOverId(null);
                  setDragOverPosition(null);
                  setDragOverGroup(null);
                }}
              />
            ) : null}
            <div style={{ fontWeight: 600, marginBottom: "6px" }}>
              {formatDateKey(dueKey === "no-due" ? "" : dueKey)}
            </div>
            <div
              className={`group-drop${
                dragOverGroup?.dueKey === dueKey && dragOverGroup.position === "below"
                  ? " active"
                  : ""
              }`}
              onDragOver={(event) => {
                event.preventDefault();
                event.dataTransfer.dropEffect = "move";
                setDragOverGroup({ dueKey, position: "below" });
              }}
              onDrop={(event) => {
                event.preventDefault();
                event.stopPropagation();
                const dragState = dragStateRef.current;
                if (!dragState) {
                  return;
                }
                if (firstTaskId) {
                  moveTaskToDueDate(dragState.taskId, dueKey, firstTaskId, "above");
                } else {
                  moveTaskToDueDate(dragState.taskId, dueKey);
                }
                dragStateRef.current = null;
                setDraggingId(null);
                setDragOverId(null);
                setDragOverPosition(null);
                setDragOverGroup(null);
              }}
            />
            <ul
              onDragOver={(event) => {
                event.preventDefault();
                event.dataTransfer.dropEffect = "move";
              }}
              onDrop={(event) => {
                event.preventDefault();
                event.stopPropagation();
                const dragState = dragStateRef.current;
                if (!dragState) {
                  return;
                }
                moveTaskToDueDate(dragState.taskId, dueKey);
                dragStateRef.current = null;
                setDraggingId(null);
                setDragOverId(null);
                setDragOverPosition(null);
                setDragOverGroup(null);
              }}
            >
              {sortedGroup.map((task) => (
                <li
                  key={task.id}
                  className={`task-item${draggingId === task.id ? " dragging" : ""}${dragOverId === task.id && dragOverPosition === "above" ? " drag-over-above" : ""}${dragOverId === task.id && dragOverPosition === "below" ? " drag-over-below" : ""}`}
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
                    setDragOverPosition(null);
                    setDragOverGroup(null);
                    document.body.style.userSelect = dragUserSelectRef.current;
                  }}
                  onDragEnter={(event) => {
                    event.preventDefault();
                    setDragOverId(task.id);
                  }}
                  onDragOver={(event) => {
                    event.preventDefault();
                    event.dataTransfer.dropEffect = "move";
                    const rect = event.currentTarget.getBoundingClientRect();
                    const isAbove = event.clientY < rect.top + rect.height / 2;
                    setDragOverId(task.id);
                    setDragOverPosition(isAbove ? "above" : "below");
                  }}
                  onDrop={(event) => {
                    event.preventDefault();
                    event.stopPropagation();
                    const dragState = dragStateRef.current;
                    if (!dragState) {
                      return;
                    }
                    const position = dragOverPosition || "below";
                    if (dragState.dueKey === dueKey) {
                      reorderWithinDueDate(dueKey, dragState.taskId, task.id, position);
                    } else {
                      moveTaskToDueDate(dragState.taskId, dueKey, task.id, position);
                    }
                    dragStateRef.current = null;
                    setDraggingId(null);
                    setDragOverId(null);
                    setDragOverPosition(null);
                    setDragOverGroup(null);
                  }}
                >
                  <div
                    className="task-row"
                    style={{ display: "flex", alignItems: "center", gap: "8px" }}
                  >
                    <label
                      style={{ textDecoration: task.status === "done" ? "line-through" : "none" }}
                      onClick={(event) => event.stopPropagation()}
                    >
                      <input
                        type="checkbox"
                        checked={task.status === "done"}
                        onChange={() => toggleTask(task.id)}
                        onClick={(event) => event.stopPropagation()}
                      />
                    </label>
                    <button
                      type="button"
                      onClick={() => setActiveTaskId(task.id)}
                      style={{
                        border: "none",
                        background: "transparent",
                        padding: 0,
                        cursor: "pointer",
                        textAlign: "left",
                        font: "inherit",
                        textDecoration: task.status === "done" ? "line-through" : "none",
                      }}
                    >
                      {task.title}
                    </button>
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
                </li>
              ))}
            </ul>
          </li>
        )})}
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
      {activeTaskId ? (
        <div className="modal-backdrop" onClick={() => setActiveTaskId(null)}>
          <div className="modal" onClick={(event) => event.stopPropagation()}>
            <div style={{ fontWeight: 600, marginBottom: "8px" }}>
              Task details
            </div>
            <label style={{ display: "block", marginBottom: "8px" }}>
              Details
              <textarea
                value={detailsDraft}
                onChange={(event) => setDetailsDraft(event.target.value)}
                rows={4}
                style={{ width: "100%", marginTop: "6px" }}
              />
            </label>
            <label style={{ display: "block", marginBottom: "8px" }}>
              Due date
              <input
                type="date"
                value={dueDateDraft}
                onChange={(event) => setDueDateDraft(event.target.value)}
                style={{ display: "block", marginTop: "6px" }}
              />
            </label>
            <div style={{ marginBottom: "8px", color: "#666" }}>
              Created on: {formatDate(tasks.find((task) => task.id === activeTaskId)?.created_at) || "unknown"}
            </div>
            <label style={{ display: "block", marginBottom: "8px" }}>
              Priority
              <select
                value={priorityDraft}
                onChange={(event) => setPriorityDraft(event.target.value)}
                style={{ display: "block", marginTop: "6px" }}
              >
                <option value="low">Low</option>
                <option value="normal">Normal</option>
                <option value="high">High</option>
              </select>
            </label>
            <div className="modal-actions">
              <button type="button" onClick={() => setActiveTaskId(null)}>
                Cancel
              </button>
              <button
                type="button"
                onClick={() => {
                  if (!activeTaskId) {
                    return;
                  }
                  updateTaskDetails(activeTaskId, detailsDraft, dueDateDraft, priorityDraft);
                  setActiveTaskId(null);
                }}
              >
                Save
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
