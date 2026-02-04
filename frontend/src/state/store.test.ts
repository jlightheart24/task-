import { describe, it, expect } from "vitest";
import { useTaskStore } from "./store";

describe("useTaskStore", () => {
  it("adds a task", () => {
    const { addTask } = useTaskStore.getState();
    addTask({ id: "t1", title: "first" });

    const { tasks } = useTaskStore.getState();
    expect(tasks).toHaveLength(1);
    expect(tasks[0].title).toBe("first");
  });
});
