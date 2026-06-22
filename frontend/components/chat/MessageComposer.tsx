"use client";

import { FormEvent, useState } from "react";

export default function MessageComposer({ onSend, disabled = false }: { onSend: (content: string) => Promise<unknown>; disabled?: boolean }) {
  const [draft, setDraft] = useState("");
  const [sending, setSending] = useState(false);
  const [error, setError] = useState("");
  const content = draft.trim();

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!content || disabled || sending) return;
    setSending(true); setError("");
    try { await onSend(content); setDraft(""); }
    catch { setError("发送失败，请稍后重试"); }
    finally { setSending(false); }
  }

  return <form className="chat-composer" onSubmit={submit}>
    <textarea className="field min-h-20 resize-none" aria-label="消息内容" maxLength={1000} value={draft} disabled={disabled || sending} onChange={(event) => setDraft(event.target.value)} placeholder="说点什么…" />
    <div className="flex items-center justify-between gap-3"><span className="text-xs text-slate-500">{draft.length}/1000</span><button className="button-primary" disabled={!content || disabled || sending}>{sending ? "发送中…" : "发送"}</button></div>
    {error && <p className="text-sm text-rose-600" role="alert">{error}</p>}
  </form>;
}
