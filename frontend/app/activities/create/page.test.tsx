import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import CreateActivityPage from "./page";

const { postJSON, push } = vi.hoisted(() => ({ postJSON: vi.fn(), push: vi.fn() }));
vi.mock("@/lib/api", () => ({ getToken: () => "token", postJSON, ApiError: class ApiError extends Error {} }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ push, replace: vi.fn() }) }));

describe("CreateActivityPage", () => {
  it("converts tags and creates an activity", async () => {
    postJSON.mockResolvedValue({ activity: { id: "a1" } });
    render(<CreateActivityPage />);
    fireEvent.change(screen.getByLabelText("活动标题"), { target: { value: "电赛组队" } });
    fireEvent.change(screen.getByLabelText("活动介绍"), { target: { value: "寻找硬件队友" } });
    fireEvent.change(screen.getByLabelText("活动标签"), { target: { value: "电赛, STM32" } });
    fireEvent.click(screen.getByRole("button", { name: "发布活动" }));
    await waitFor(() => expect(postJSON).toHaveBeenCalledWith("/activities", expect.objectContaining({ tags: ["电赛", "STM32"] })));
    expect(push).toHaveBeenCalledWith("/activities");
  });
});
