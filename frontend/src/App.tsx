import { useEffect, useRef, useState } from "react";

type Task = {
  id: string;
  title: string;
  status: string;
};

export function App() {
  const [message, setMessage] = useState<string>("backend not connected");
  const [tasks, setTasks] = useState<Task[]>([]);
  const [title, setTitle] = useState<string>("");
  const inputRef = useRef<HTMLInputElement | null>(null);

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
    if (typeof create !== "function" || title.trim() === "") {
      return;
    }
    create(title.trim())
      .then((task: Task) => {
        setTasks((prev) => [...prev, task]);
        setTitle("");
        inputRef.current?.focus();
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

  return (
    <div>
      <h1>task-</h1>
      <p>{message}</p>
      <div
        onClick={() => inputRef.current?.focus()}
        style={{ cursor: "text" }}
      >
        <input
          ref={inputRef}
          value={title}
          placeholder="new task"
          onChange={(event) => setTitle(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Enter") {
              event.preventDefault();
              createTask();
            }
          }}
        />
      </div>
      <ul>
        {tasks.map((task) => (
          <li key={task.id}>
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
          </li>
        ))}
      </ul>
    </div>
  );
}
