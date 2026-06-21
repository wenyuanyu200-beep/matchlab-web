import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import ApplicationsPage from "./page";

const { request } = vi.hoisted(() => ({ request: vi.fn() }));
vi.mock("@/lib/api", () => ({ getToken: () => "token", request, ApiError: class ApiError extends Error {} }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ replace: vi.fn() }) }));

describe("ApplicationsPage", () => {
  it("renders the user's application status", async () => {
    request.mockResolvedValue({ applications: [{ id: "ap1", activity_id: "a1", activity_title: "电赛组队", reason: "想参加比赛", match_score: 0, status: "approved" }] });
    render(<ApplicationsPage />);
    expect(await screen.findByText("电赛组队")).toBeInTheDocument();
    expect(screen.getByText("已通过")).toBeInTheDocument();
  });
});
