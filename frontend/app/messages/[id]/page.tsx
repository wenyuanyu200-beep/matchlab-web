"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import MessageComposer from "@/components/chat/MessageComposer";
import MessageList from "@/components/chat/MessageList";
import { usePollingMessages } from "@/hooks/usePollingMessages";
import { getToken, request } from "@/lib/api";
import type { ChatMessage, Conversation, User } from "@/lib/types";

export default function ConversationPage() {
  const { id } = useParams<{ id: string }>(); const router = useRouter(); const [conversation, setConversation] = useState<Conversation | null>(null); const [me, setMe] = useState<User | null>(null); const [error, setError] = useState(""); const polling = usePollingMessages(`/conversations/${id}/messages`);
  useEffect(() => { if (!getToken()) { router.replace("/login"); return; } Promise.all([request<Conversation | { conversation: Conversation }>(`/conversations/${id}`), request<User | { user: User }>("/me"), request(`/conversations/${id}/read`, { method: "POST" })]).then(([data, user]) => { setConversation("conversation" in data ? data.conversation : data); setMe("user" in user ? user.user : user); }).catch(() => setError("会话加载失败")); }, [id, router]);
  async function send(content: string) { const data = await request<ChatMessage | { message: ChatMessage }>(`/conversations/${id}/messages`, { method: "POST", body: JSON.stringify({ content }) }); polling.append("message" in data ? data.message : data); }
  return <main className="page-shell page-section"><div className="mx-auto max-w-3xl"><Link className="text-sm font-bold text-indigo-700" href="/messages">← 返回私信</Link><section className="card mt-5"><div className="border-b border-slate-200 pb-4"><h1 className="text-2xl font-black">{conversation?.peer?.nickname || conversation?.title || "私信会话"}</h1><p className="text-sm text-slate-500">{conversation?.peer?.school || "校园伙伴"}</p></div>{error && <p className="notice-error mt-4">{error}</p>}<MessageList messages={polling.messages} currentUserId={me?.id} />{polling.error && <p className="notice-soft">{polling.error}</p>}<MessageComposer onSend={send} /></section></div></main>;
}
