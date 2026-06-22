import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import CirclesPage from "./page";

const { request } = vi.hoisted(() => ({ request: vi.fn() }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ push: vi.fn() }) }));
vi.mock("@/lib/api", () => ({ request, getToken: () => "token" }));

describe("CirclesPage", () => {
  beforeEach(() => request.mockReset());
  it("搜索和分类筛选已通过圈子", async () => {
    request.mockResolvedValue({ circles: [
      { id: "1", name: "羽毛球搭子", description: "约球", category: "sports", status: "approved" },
      { id: "2", name: "学习组", description: "刷题", category: "study", status: "approved" },
    ] });
    render(<CirclesPage />);
    expect(await screen.findByText("羽毛球搭子")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("搜索圈子"), { target: { value: "学习" } });
    expect(screen.queryByText("羽毛球搭子")).not.toBeInTheDocument();
    expect(screen.getByText("学习组")).toBeInTheDocument();
  });
});
