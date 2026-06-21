import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { setToken } from "@/lib/api";
import Navbar from "./Navbar";

const push = vi.fn();
vi.mock("next/navigation", () => ({ useRouter: () => ({ push }) }));

describe("Navbar", () => {
  beforeEach(() => {
    localStorage.clear();
    push.mockClear();
  });

  it("shows the login action when logged out", () => {
    render(<Navbar />);
    expect(screen.getByRole("link", { name: "登录" })).toHaveAttribute("href", "/login");
  });

  it("clears the token when logging out", async () => {
    setToken("token");
    render(<Navbar />);
    fireEvent.click(await screen.findByRole("button", { name: "退出" }));
    expect(localStorage.getItem("matchlab_token")).toBeNull();
    expect(push).toHaveBeenCalledWith("/");
  });
});
