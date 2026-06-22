import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import MatchPage from "./page";

const { request, postJSON } = vi.hoisted(() => ({ request: vi.fn(), postJSON: vi.fn() }));
vi.mock("@/lib/api", () => ({ getToken: () => "token", request, postJSON, ApiError: class ApiError extends Error {} }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ replace: vi.fn() }) }));

describe("MatchPage", () => {
  it("generates recommendations and refreshes saved matches", async () => {
    request.mockResolvedValue({ matches: [] });
    postJSON.mockResolvedValue({ recommendations: [{ activity: { id: "a1", title: "智能车竞赛" }, score: 92, detail_scores: { interest: 28, skill: 22, type: 20, time: 8, goal: 14 }, reason: "兴趣与技能高度匹配" }] });
    render(<MatchPage />);
    fireEvent.click(screen.getByRole("button", { name: "查看适合我的活动" }));
    await waitFor(() => expect(postJSON).toHaveBeenCalledWith("/match/recommend", { target_type: "activity", limit: 10 }));
    expect(await screen.findByText("智能车竞赛")).toBeInTheDocument();
    expect(screen.getByText("92")).toBeInTheDocument();
  });
});
