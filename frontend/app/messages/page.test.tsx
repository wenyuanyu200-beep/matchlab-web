import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import MessagesPage from "./page";

const { request } = vi.hoisted(() => ({ request: vi.fn() }));
vi.mock("@/lib/api", () => ({ request, getToken: () => "token" }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ replace: vi.fn() }) }));

describe("MessagesPage", () => {
  it("展示会话和未读数", async () => {
    request.mockResolvedValue({ conversations: [{ id: "c1", peer: { id: "u2", nickname: "小林" }, unread_count: 2, last_message: { content: "明天见" } }] });
    render(<MessagesPage />);
    expect(await screen.findByText("小林")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("明天见")).toBeInTheDocument();
  });
});
