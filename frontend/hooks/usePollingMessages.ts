"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { request } from "@/lib/api";
import { listFrom, mergeMessages, messageCursor } from "@/lib/circles";
import type { ChatMessage } from "@/lib/types";

export function usePollingMessages(path: string | null, enabled = true) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [error, setError] = useState("");
  const messagesRef = useRef<ChatMessage[]>([]);

  const poll = useCallback(async () => {
    if (!path || !enabled || document.hidden) return;
    const cursor = messageCursor(messagesRef.current);
    const query = new URLSearchParams(cursor).toString();
    try {
      const payload = await request<unknown>(`${path}${query ? `?${query}` : ""}`);
      const incoming = listFrom<ChatMessage>(payload, "messages");
      const merged = mergeMessages(messagesRef.current, incoming);
      messagesRef.current = merged; setMessages(merged); setError("");
    } catch { setError("消息刷新失败，将自动重试"); }
  }, [enabled, path]);

  useEffect(() => {
    messagesRef.current = [];
    let cancelled = false;
    queueMicrotask(() => { if (!cancelled) { setMessages([]); setError(""); } });
    if (!path || !enabled) return;
    void poll();
    const timer = window.setInterval(poll, 3000);
    const visibility = () => { if (!document.hidden) void poll(); };
    document.addEventListener("visibilitychange", visibility);
    return () => { cancelled = true; window.clearInterval(timer); document.removeEventListener("visibilitychange", visibility); };
  }, [enabled, path, poll]);

  const append = useCallback((message: ChatMessage) => {
    const merged = mergeMessages(messagesRef.current, [message]); messagesRef.current = merged; setMessages(merged);
  }, []);
  return { messages, error, poll, append };
}

