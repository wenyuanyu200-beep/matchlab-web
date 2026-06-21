import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import DashboardPage from "./page";

const { request } = vi.hoisted(() => ({ request: vi.fn() }));
vi.mock("@/lib/api", () => ({
  getToken: () => "token",
  request,
  ApiError: class ApiError extends Error {},
}));
vi.mock("next/navigation", () => ({ useRouter: () => ({ replace: vi.fn() }) }));

describe("DashboardPage", () => {
  it("shows admin entry only for administrators", async () => {
    request.mockResolvedValue({ user: { id: "1", nickname: "管理员", email: "a@b.com", school: "示例大学", role: "admin" } });
    render(<DashboardPage />);
    expect(await screen.findByText("你好，管理员")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /管理员后台/ })).toHaveAttribute("href", "/admin");
  });
});
