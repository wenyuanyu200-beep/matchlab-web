"use client";

import type { ChatMessage } from "@/lib/types";

export default function MessageList({ messages, currentUserId, onDirect }: { messages: ChatMessage[]; currentUserId?: string; onDirect?: (userId: string) => void }) {
  if (!messages.length) return <div className="grid min-h-56 place-items-center text-sm text-slate-500">还没有消息，来打个招呼吧</div>;
  return <div className="chat-list" aria-live="polite">{messages.map((message) => {
    const mine = message.sender_id === currentUserId;
    return <article key={message.id} className={`chat-row ${mine ? "chat-row-mine" : ""}`}>
      <div className={`chat-bubble ${mine ? "chat-bubble-mine" : ""}`}>
        {!mine && <button className="mb-1 text-xs font-bold text-indigo-700" onClick={() => onDirect?.(message.sender_id)}>{message.sender?.nickname || "同学"}</button>}
        <p className="whitespace-pre-wrap break-words">{message.content}</p>
        <time className="mt-1 block text-[11px] opacity-60">{new Date(message.created_at).toLocaleString("zh-CN", { month: "numeric", day: "numeric", hour: "2-digit", minute: "2-digit" })}</time>
      </div>
    </article>;
  })}</div>;
}
