import { describe, expect, it } from "vitest";
import { asArray, splitList } from "./forms";

describe("form helpers", () => {
  it("splits comma-separated tags and removes empty values", () => {
    expect(splitList(" 电赛, STM32，硬件, ")).toEqual(["电赛", "STM32", "硬件"]);
  });

  it("normalizes missing lists to empty arrays", () => {
    expect(asArray(undefined)).toEqual([]);
    expect(asArray(["project"])).toEqual(["project"]);
  });
});
