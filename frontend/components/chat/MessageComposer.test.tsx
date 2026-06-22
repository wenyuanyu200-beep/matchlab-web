import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import MessageComposer from "./MessageComposer";

describe("MessageComposer", () => {
  it("禁止空白消息并 trim 后发送", async () => {
    const send = vi.fn().mockResolvedValue(undefined);
    render(<MessageComposer onSend={send} />);
    const input = screen.getByLabelText("消息内容");
    fireEvent.change(input, { target: { value: "   " } });
    expect(screen.getByRole("button", { name: "发送" })).toBeDisabled();
    fireEvent.change(input, { target: { value: "  你好  " } });
    fireEvent.click(screen.getByRole("button", { name: "发送" }));
    await waitFor(() => expect(send).toHaveBeenCalledWith("你好"));
  });

  it("发送失败时保留草稿并提示错误", async () => {
    render(<MessageComposer onSend={() => Promise.reject(new Error("发送失败"))} />);
    const input = screen.getByLabelText("消息内容") as HTMLTextAreaElement;
    fireEvent.change(input, { target: { value: "保留我" } });
    fireEvent.click(screen.getByRole("button", { name: "发送" }));
    expect(await screen.findByText("发送失败，请稍后重试")).toBeInTheDocument();
    expect(input.value).toBe("保留我");
  });
});
