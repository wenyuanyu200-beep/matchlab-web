import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import QuestionnairePage from "./page";

const { postJSON } = vi.hoisted(() => ({ postJSON: vi.fn() }));
vi.mock("@/lib/api", () => ({ getToken: () => "token", postJSON, ApiError: class ApiError extends Error {} }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ replace: vi.fn() }) }));

describe("QuestionnairePage", () => {
  it("submits array answers and shows the generated profile", async () => {
    postJSON.mockResolvedValue({ profile: { profile_type: "activity", tags: ["电赛"], scores: { interest_score: 80, skill_score: 75, time_score: 70, goal_score: 80, communication_score: 75 }, summary: "适合参与硬件项目。" } });
    render(<QuestionnairePage />);
    fireEvent.change(screen.getByLabelText("兴趣方向"), { target: { value: "电赛, 硬件" } });
    fireEvent.change(screen.getByLabelText("技能特长"), { target: { value: "嵌入式, 焊接" } });
    fireEvent.click(screen.getByRole("button", { name: "生成我的画像" }));
    await waitFor(() => expect(postJSON).toHaveBeenCalledWith("/questionnaires", expect.objectContaining({ answers: expect.objectContaining({ interests: ["电赛", "硬件"], skills: ["嵌入式", "焊接"] }) })));
    expect(await screen.findByText("适合参与硬件项目。")).toBeInTheDocument();
  });
});
