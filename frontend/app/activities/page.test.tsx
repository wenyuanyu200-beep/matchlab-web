import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import ActivitiesPage from "./page";

const { request } = vi.hoisted(() => ({ request: vi.fn() }));
vi.mock("@/lib/api", () => ({ request, ApiError: class ApiError extends Error {} }));

describe("ActivitiesPage", () => {
  it("renders activities returned by the API", async () => {
    request.mockResolvedValue({ activities: [{ id: "a1", creator_id: "u1", title: "机器人项目", type: "project", description: "一起做机器人", required_count: 3, joined_count: 1, tags: ["控制"], time_text: "周末", location_text: "实验室", status: "recruiting" }] });
    render(<ActivitiesPage />);
    expect(await screen.findByText("机器人项目")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "发布活动" })).toHaveAttribute("href", "/activities/create");
  });
});
