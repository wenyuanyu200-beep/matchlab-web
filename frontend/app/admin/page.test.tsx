import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import AdminPage from "./page";

const { request } = vi.hoisted(() => ({ request: vi.fn() }));
vi.mock("@/lib/api", () => ({ getToken: () => "token", request, ApiError: class ApiError extends Error {} }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ replace: vi.fn() }) }));

describe("AdminPage", () => {
  it("denies access to ordinary users", async () => {
    request.mockResolvedValue({ user: { id: "u1", role: "user" } });
    render(<AdminPage />);
    expect(await screen.findByText("无权限访问")).toBeInTheDocument();
  });

  it("renders admin statistics and tables", async () => {
    request.mockImplementation((path: string) => {
      if (path === "/me") return Promise.resolve({ user: { id: "a1", email: "admin@example.com", nickname: "管理员", school: "示例大学", role: "admin" } });
      if (path === "/admin/stats") return Promise.resolve({ stats: { users_count: 2, activities_count: 3, applications_count: 4, matches_count: 5, questionnaires_count: 6, feedbacks_count: 0 } });
      if (path.startsWith("/admin/users")) return Promise.resolve({ users: [{ id: "u1", email: "user@example.com", nickname: "同学", role: "user", school: "示例大学" }] });
      if (path.startsWith("/admin/activities")) return Promise.resolve({ activities: [] });
      if (path.startsWith("/admin/applications")) return Promise.resolve({ applications: [] });
      return Promise.resolve({ feedbacks: [] });
    });
    render(<AdminPage />);
    expect(screen.getByText("平台数据总览")).toBeInTheDocument();
    expect(await screen.findByText("user@example.com")).toBeInTheDocument();
    expect(screen.getByText("暂无反馈数据")).toBeInTheDocument();
  });
});
