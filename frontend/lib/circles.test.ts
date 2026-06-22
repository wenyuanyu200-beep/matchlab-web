import { describe, expect, it } from "vitest";
import { filterCircles, mergeMessages, messageCursor } from "./circles";
import type { Circle, ChatMessage } from "./types";

const circles = [
  { id: "1", name: "羽毛球搭子", description: "每周约球", category: "sports", status: "approved" },
  { id: "2", name: "Python 学习组", description: "一起刷题", category: "study", status: "approved" },
  { id: "3", name: "待审核", description: "不可见", category: "other", status: "pending" },
] as Circle[];

describe("filterCircles", () => {
  it("只展示已通过且匹配搜索与分类的圈子", () => {
    expect(filterCircles(circles, "羽毛", "sports").map((item) => item.id)).toEqual(["1"]);
    expect(filterCircles(circles, "", "all").map((item) => item.id)).toEqual(["1", "2"]);
  });
});

describe("message utilities", () => {
  const first = { id: "m1", content: "a", created_at: "2026-01-01T00:00:00Z" } as ChatMessage;
  const second = { id: "m2", content: "b", created_at: "2026-01-01T00:00:01Z" } as ChatMessage;

  it("按 id 去重并保持时间顺序", () => {
    expect(mergeMessages([first], [first, second]).map((item) => item.id)).toEqual(["m1", "m2"]);
  });

  it("从最后一条消息生成增量游标", () => {
    expect(messageCursor([first, second])).toEqual({ after_time: second.created_at, after_id: "m2" });
  });
});
