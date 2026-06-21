"use client";

import { FormEvent, useEffect, useState, useSyncExternalStore } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import TagList from "@/components/TagList";
import { ApiError, getToken, postJSON, request, subscribeAuth } from "@/lib/api";
import { friendlyStatus } from "@/lib/forms";
import type { Activity } from "@/lib/types";

export default function ActivityDetailPage() {
  const { id } = useParams<{ id: string }>(); const [activity, setActivity] = useState<Activity | null>(null); const [reason, setReason] = useState(""); const [error, setError] = useState(""); const [success, setSuccess] = useState(false); const [pending, setPending] = useState(false); const loggedIn = useSyncExternalStore(subscribeAuth, () => Boolean(getToken()), () => false);
  useEffect(() => { request<{ activity: Activity }>(`/activities/${id}`).then((data) => setActivity(data.activity)).catch((cause) => setError(cause instanceof ApiError ? cause.message : "无法加载活动详情")); }, [id]);
  async function apply(event: FormEvent) { event.preventDefault(); setPending(true); setError(""); try { await postJSON(`/activities/${id}/apply`, { reason }); setSuccess(true); } catch (cause) { setError(cause instanceof ApiError ? cause.message : "报名提交失败"); } finally { setPending(false); } }
  return <section className="page-shell page-section">{!activity && !error && <p>正在加载活动详情…</p>}{error && <p className="notice-error">{error}</p>}{activity && <div className="grid gap-7 lg:grid-cols-[1fr_360px]">
    <article className="card p-6 sm:p-9"><div className="flex flex-wrap justify-between gap-3"><span className="eyebrow">{activity.type}</span><span className="status-badge">{friendlyStatus(activity.status)}</span></div><h1 className="mt-5 text-3xl font-black text-slate-950 sm:text-5xl">{activity.title}</h1><p className="mt-6 whitespace-pre-wrap text-base leading-8 text-slate-650">{activity.description}</p><div className="mt-7"><TagList tags={activity.tags} /></div><dl className="mt-8 grid gap-5 border-t border-slate-200 pt-7 sm:grid-cols-2"><div><dt className="text-sm text-slate-500">时间</dt><dd className="mt-1 font-bold text-slate-900">{activity.time_text || "待沟通"}</dd></div><div><dt className="text-sm text-slate-500">地点</dt><dd className="mt-1 font-bold text-slate-900">{activity.location_text || "待沟通"}</dd></div><div><dt className="text-sm text-slate-500">参与人数</dt><dd className="mt-1 font-bold text-slate-900">{activity.joined_count} / {activity.required_count}</dd></div><div><dt className="text-sm text-slate-500">发起人</dt><dd className="mt-1 font-bold text-slate-900">{activity.creator ? `${activity.creator.nickname} · ${activity.creator.school}` : "校园伙伴"}</dd></div></dl>{activity.preferred_tags?.length ? <div className="mt-8"><h2 className="mb-3 font-black text-slate-950">期待的队友标签</h2><TagList tags={activity.preferred_tags} /></div> : null}</article>
    <aside className="card h-fit lg:sticky lg:top-24"><h2 className="text-xl font-black text-slate-950">加入这个活动</h2>{loggedIn ? <form className="mt-5 grid gap-4" onSubmit={apply}><label className="field-label">报名理由<textarea className="field min-h-32" required value={reason} onChange={(e) => setReason(e.target.value)} placeholder="介绍你的兴趣、技能或参与目标" /></label>{success && <p className="notice-success">报名已提交，请等待活动发起人审核。</p>}<button className="button-primary" disabled={pending || success}>{pending ? "提交中…" : success ? "已提交" : "提交报名"}</button></form> : <div className="mt-5"><p className="text-sm leading-6 text-slate-600">登录后即可提交报名。</p><Link className="button-primary mt-4 w-full" href="/login">前往登录</Link></div>}</aside>
  </div>}</section>;
}
