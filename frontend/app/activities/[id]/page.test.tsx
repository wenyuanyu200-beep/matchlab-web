import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import ActivityDetailPage from "./page";

const { request, postJSON } = vi.hoisted(() => ({ request: vi.fn(), postJSON: vi.fn() }));
vi.mock("@/lib/api", () => ({ getToken: () => "token", request, postJSON, ApiError: class ApiError extends Error {} }));
vi.mock("next/navigation", () => ({ useParams: () => ({ id: "activity-1" }) }));

describe("ActivityDetailPage", () => {
  it("loads the detail and submits an application", async () => {
    request.mockResolvedValue({ activity: { id: "activity-1", creator_id: "u1", title: "创客项目", type: "project", description: "完成校园装置", required_count: 4, joined_count: 1, tags: ["硬件"], preferred_tags: ["嵌入式"], time_text: "周六", location_text: "创新中心", status: "recruiting" } });
    postJSON.mockResolvedValue({ application: { id: "ap1" } });
    render(<ActivityDetailPage />);
    expect(await screen.findByText("创客项目")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("报名理由"), { target: { value: "我有嵌入式经验" } });
    fireEvent.click(screen.getByRole("button", { name: "提交报名" }));
    await waitFor(() => expect(postJSON).toHaveBeenCalledWith("/activities/activity-1/apply", { reason: "我有嵌入式经验" }));
    expect(await screen.findByText("报名已提交，请等待活动发起人审核。")).toBeInTheDocument();
  });
});
