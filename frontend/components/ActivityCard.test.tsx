import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import ActivityCard from "./ActivityCard";

describe("ActivityCard", () => {
  it("renders activity metadata and its detail link", () => {
    render(
      <ActivityCard
        activity={{
          id: "activity-1",
          creator_id: "user-1",
          title: "电赛组队",
          type: "competition",
          description: "一起完成硬件项目",
          required_count: 4,
          joined_count: 2,
          tags: ["STM32", "硬件"],
          time_text: "周末下午",
          location_text: "创新实验室",
          status: "recruiting",
          creator: { nickname: "小林", school: "示例大学" },
        }}
      />,
    );
    expect(screen.getByText("电赛组队")).toBeInTheDocument();
    expect(screen.getByText("小林 · 示例大学")).toBeInTheDocument();
    expect(screen.getByText("2 / 4 人")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看详情" })).toHaveAttribute(
      "href",
      "/activities/activity-1",
    );
  });
});
