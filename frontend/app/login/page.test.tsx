import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import LoginPage from "./page";

const { push, postJSON, setToken } = vi.hoisted(() => ({ push: vi.fn(), postJSON: vi.fn(), setToken: vi.fn() }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ push }) }));
vi.mock("@/lib/api", () => ({ postJSON, setToken, ApiError: class ApiError extends Error {} }));

describe("LoginPage", () => {
  it("logs in, stores the token, and opens the dashboard", async () => {
    postJSON.mockResolvedValueOnce({ token: "jwt-token", user: { role: "user" } });
    render(<LoginPage />);
    fireEvent.change(screen.getByLabelText("邮箱"), { target: { value: "test@example.com" } });
    fireEvent.change(screen.getByLabelText("密码"), { target: { value: "password123" } });
    fireEvent.click(screen.getAllByRole("button", { name: "登录" }).at(-1)!);
    await waitFor(() => expect(setToken).toHaveBeenCalledWith("jwt-token"));
    expect(push).toHaveBeenCalledWith("/dashboard");
  });

  it("switches to registration fields", () => {
    render(<LoginPage />);
    fireEvent.click(screen.getByRole("button", { name: "注册账号" }));
    expect(screen.getByLabelText("昵称")).toBeInTheDocument();
    expect(screen.getByLabelText("学校")).toBeInTheDocument();
  });
});
