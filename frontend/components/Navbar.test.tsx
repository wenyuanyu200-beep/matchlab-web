import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Navbar from "./Navbar";

const { push, request, clearToken } = vi.hoisted(() => ({ push: vi.fn(), request: vi.fn(), clearToken: vi.fn() }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ push }) }));
vi.mock("@/lib/api", () => ({ getToken: () => localStorage.getItem("matchlab_token"), subscribeAuth: () => () => undefined, request, clearToken }));

describe("Navbar", () => {
  beforeEach(() => { localStorage.clear(); push.mockClear(); request.mockReset(); clearToken.mockClear(); });
  it("未登录时展示登录入口", () => { render(<Navbar />); expect(screen.getByRole("link", { name: "登录" })).toHaveAttribute("href", "/login"); });
  it("登录后展示消息未读数并可退出", async () => {
    localStorage.setItem("matchlab_token", "token"); request.mockResolvedValue({ unread_count: 3 }); render(<Navbar />);
    expect(await screen.findByText("3")).toBeInTheDocument(); fireEvent.click(screen.getByRole("button", { name: "退出" }));
    expect(clearToken).toHaveBeenCalled(); expect(push).toHaveBeenCalledWith("/");
  });
});
