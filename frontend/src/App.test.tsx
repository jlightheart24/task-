import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { App } from "./App";

type GoMock = {
  app: {
    App: {
      ListTasks: () => Promise<any[]>;
      UpdateTaskDetails: (
        id: string,
        description: string,
        dueDate: string,
        priority: string
      ) => Promise<any>;
    };
  };
};

const setGoMock = (mock: GoMock) => {
  (window as unknown as { go?: GoMock }).go = mock;
};

describe("App task details modal", () => {
  it("opens modal and saves task details", async () => {
    const updateTaskDetails = vi.fn(async (_id, description, dueDate, priority) => ({
      id: "t1",
      title: "first",
      description,
      status: "open",
      priority,
      due_date: dueDate,
      created_at: "2026-02-04T10:00:00Z",
    }));

    setGoMock({
      app: {
        App: {
          ListTasks: async () => [
            {
              id: "t1",
              title: "first",
              description: "",
              status: "open",
              priority: "normal",
              due_date: "2026-02-04T00:00:00Z",
              created_at: "2026-02-04T10:00:00Z",
            },
          ],
          UpdateTaskDetails: updateTaskDetails,
        },
      },
    });

    const user = userEvent.setup();
    render(<App />);

    const task = await screen.findByText("first");
    await user.click(task);

    expect(await screen.findByText("Task details")).toBeInTheDocument();

    await user.clear(screen.getByLabelText("Details"));
    await user.type(screen.getByLabelText("Details"), "New notes");
    await user.clear(screen.getByLabelText("Due date"));
    await user.type(screen.getByLabelText("Due date"), "2026-02-10");
    await user.selectOptions(screen.getByLabelText("Priority"), "high");

    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => {
      expect(updateTaskDetails).toHaveBeenCalledWith(
        "t1",
        "New notes",
        "2026-02-10",
        "high"
      );
    });
  });

  it("shows weekday and date headers when configured", async () => {
    window.localStorage.setItem("taskpp.dateHeaderMode", "both");
    setGoMock({
      app: {
        App: {
          ListTasks: async () => [
            {
              id: "t2",
              title: "second",
              description: "",
              status: "open",
              priority: "normal",
              due_date: "2026-02-04T00:00:00Z",
              created_at: "2026-02-04T10:00:00Z",
            },
          ],
          UpdateTaskDetails: async () => ({
            id: "t2",
          }),
        },
      },
    });

    render(<App />);

    const header = await screen.findByText(/·/);
    expect(header).toBeInTheDocument();
  });
});
