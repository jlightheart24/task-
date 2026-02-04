import { create } from "zustand";

type Task = {
  id: string;
  title: string;
};

type TaskState = {
  tasks: Task[];
  addTask: (task: Task) => void;
};

export const useTaskStore = create<TaskState>((set) => ({
  tasks: [],
  addTask: (task) => set((state) => ({ tasks: [...state.tasks, task] }))
}));
