import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import Home from "./page";

describe("Home", () => {
  it("presents the product and three-step flow", () => {
    render(<Home />);
    expect(screen.getByText("校园活动与项目协作平台")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "进入活动广场" })).toHaveAttribute("href", "/activities");
    expect(screen.getByRole("link", { name: "智能推荐" })).toHaveAttribute("href", "/match");
    expect(screen.getByText("填写画像")).toBeInTheDocument();
    expect(screen.getByText("浏览活动")).toBeInTheDocument();
    expect(screen.getByText("报名组队")).toBeInTheDocument();
  });
});
