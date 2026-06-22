"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import MessageComposer from "@/components/chat/MessageComposer";
import MessageList from "@/components/chat/MessageList";
import { usePollingMessages } from "@/hooks/usePollingMessages";
import { getToken, request } from "@/lib/api";
import { categoryLabel, listFrom } from "@/lib/circles";
import type { ChatMessage, Circle, CircleMember, User } from "@/lib/types";

type Tab = "info" | "chat" | "members";

export default function CircleDetailPage() {
  const { id } = useParams<{ id: string }>(); const router = useRouter(); const loggedIn = Boolean(getToken());
  const [circle, setCircle] = useState<Circle | null>(null); const [members, setMembers] = useState<CircleMember[]>([]); const [me, setMe] = useState<User | null>(null); const [tab, setTab] = useState<Tab>("info"); const [error, setError] = useState("");
  const joined = circle?.membership_status === "joined"; const polling = usePollingMessages(joined ? `/circles/${id}/messages` : null, joined);
  useEffect(() => { Promise.all([request<Circle | { circle: Circle }>(`/circles/${id}`), loggedIn ? request<User | { user: User }>("/me") : Promise.resolve(null)]).then(([circleData, userData]) => { const value = "circle" in circleData ? circleData.circle : circleData; setCircle(value); if (userData) setMe("user" in userData ? userData.user : userData); }).catch(() => setError("圈子信息加载失败")); }, [id, loggedIn]);
  useEffect(() => { if (!joined) return; request<unknown>(`/circles/${id}/members`).then((data) => setMembers(listFrom<CircleMember>(data, "members"))).catch(() => setError("成员加载失败")); }, [id, joined]);
  async function join() { try { await request(`/circles/${id}/join`, { method: "POST" }); setCircle((value) => value ? { ...value, membership_status: "pending" } : value); } catch { setError("加入申请提交失败"); } }
  async function send(content: string) { const data = await request<ChatMessage | { message: ChatMessage }>(`/circles/${id}/messages`, { method: "POST", body: JSON.stringify({ content }) }); polling.append("message" in data ? data.message : data); }
  async function direct(userId: string) { if (userId === me?.id) return; try { const data = await request<{ conversation?: { id: string }; id?: string }>("/conversations/direct", { method: "POST", body: JSON.stringify({ user_id: userId }) }); const conversationId = data.conversation?.id || data.id; if (conversationId) router.push(`/messages/${conversationId}`); } catch { setError("发起私信失败"); } }
  if (!circle && !error) return <main className="page-shell page-section text-slate-500">正在加载圈子…</main>;
  if (!circle) return <main className="page-shell page-section"><p className="notice-error">{error}</p></main>;
  const gate = !loggedIn ? <Gate title="登录后参与聊天" action="去登录" onClick={() => router.push(`/login?next=/circles/${id}`)} /> : !joined ? <Gate title={circle.membership_status === "pending" ? "加入申请审核中" : circle.membership_status === "rejected" ? "申请未通过" : "加入圈子后查看聊天"} action={circle.membership_status === "pending" ? undefined : "申请加入"} onClick={join} /> : null;
  return <main className="page-shell page-section"><div className="mb-5 flex items-center gap-3"><span className="tag">{categoryLabel(circle.category)}</span><span className="text-sm text-slate-500">{circle.member_count ?? members.length} 位成员</span></div><h1 className="page-heading">{circle.name}</h1>
    <div className="circle-tabs mt-6" role="tablist">{(["info", "chat", "members"] as Tab[]).map((value) => <button key={value} className={tab === value ? "active" : ""} onClick={() => setTab(value)}>{value === "info" ? "圈子信息" : value === "chat" ? "频道聊天" : "成员"}</button>)}</div>
    {error && <p className="notice-error mt-4">{error}</p>}
    <div className="circle-workspace mt-6"><aside className={`circle-sidebar ${tab !== "info" && tab !== "members" ? "mobile-hidden" : ""}`}><section className={tab === "members" ? "mobile-hidden" : "card"}><h2 className="text-lg font-black">关于圈子</h2><p className="mt-3 whitespace-pre-wrap leading-7 text-slate-600">{circle.description}</p></section><section className={`card mt-4 ${tab === "info" ? "members-mobile-hide" : ""}`}><h2 className="text-lg font-black">成员</h2>{joined ? <div className="mt-3 grid gap-2">{members.map((member) => <button key={member.user_id} className="member-row" disabled={member.user_id === me?.id} onClick={() => direct(member.user_id)}><span>{member.nickname}</span><span>{member.user_id === me?.id ? "我" : "私聊"}</span></button>)}</div> : <p className="mt-3 text-sm text-slate-500">加入后查看成员</p>}</section></aside>
      <section className={`card chat-panel ${tab !== "chat" ? "mobile-hidden" : ""}`}><div className="border-b border-slate-200 pb-4"><h2 className="text-xl font-black"># general</h2><p className="text-sm text-slate-500">圈子公共频道</p></div>{gate || <><MessageList messages={polling.messages} currentUserId={me?.id} onDirect={direct} />{polling.error && <p className="notice-soft">{polling.error}</p>}<MessageComposer onSend={send} /></>}</section></div>
  </main>;
}

function Gate({ title, action, onClick }: { title: string; action?: string; onClick: () => void }) { return <div className="grid min-h-80 place-items-center text-center"><div><p className="font-bold text-slate-700">{title}</p>{action && <button className="button-primary mt-4" onClick={onClick}>{action}</button>}</div></div>; }
