"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getToken, request } from "@/lib/api";
import { listFrom } from "@/lib/circles";
import type { Conversation } from "@/lib/types";

export default function MessagesPage() {
  const router = useRouter(); const [items, setItems] = useState<Conversation[]>([]); const [error, setError] = useState("");
  useEffect(() => { if (!getToken()) { router.replace("/login"); return; } request<unknown>("/conversations").then((data) => setItems(listFrom<Conversation>(data, "conversations"))).catch(() => setError("会话加载失败，请稍后重试")); }, [router]);
  return <main className="page-shell page-section"><p className="eyebrow">Messages</p><h1 className="page-heading mt-3">私信</h1><p className="page-subtitle">和圈子里的新伙伴继续交流。</p>{error && <p className="notice-error mt-6">{error}</p>}<div className="mt-8 grid gap-3">{items.map((item) => <Link href={`/messages/${item.id}`} className="card flex items-center gap-4 transition hover:-translate-y-0.5 hover:border-indigo-200" key={item.id}><div className="grid size-12 shrink-0 place-items-center rounded-full bg-indigo-100 font-black text-indigo-700">{(item.peer?.nickname || item.title || "聊").slice(0, 1)}</div><div className="min-w-0 flex-1"><h2 className="font-black text-slate-950">{item.peer?.nickname || item.title || "私信会话"}</h2><p className="truncate text-sm text-slate-500">{item.last_message?.content || "开始聊天"}</p></div>{Boolean(item.unread_count) && <span className="unread-badge">{item.unread_count}</span>}</Link>)}</div>{!items.length && !error && <div className="card mt-8 text-center text-slate-500">暂无私信，在圈子中向成员打个招呼吧</div>}</main>;
}
