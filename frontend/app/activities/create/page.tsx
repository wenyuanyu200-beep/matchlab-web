"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiError, getToken, postJSON } from "@/lib/api";
import { splitList } from "@/lib/forms";

const initial = { title: "", type: "project", description: "", required_count: 2, tags: "", preferred_tags: "", time_text: "", location_text: "" };

export default function CreateActivityPage() {
  const router = useRouter(); const [form, setForm] = useState(initial); const [error, setError] = useState(""); const [pending, setPending] = useState(false);
  useEffect(() => { if (!getToken()) router.replace("/login"); }, [router]);
  async function submit(event: FormEvent) { event.preventDefault(); setPending(true); setError(""); try { await postJSON("/activities", { ...form, required_count: Number(form.required_count), tags: splitList(form.tags), preferred_tags: splitList(form.preferred_tags) }); router.push("/activities"); } catch (cause) { setError(cause instanceof ApiError ? cause.message : "活动发布失败"); } finally { setPending(false); } }
  return <section className="page-shell page-section"><div className="mx-auto max-w-3xl"><p className="eyebrow">Create activity</p><h1 className="page-heading mt-3">发布活动</h1><p className="page-subtitle">把目标和期待说清楚，更容易找到合适的项目队友。</p>
    <form className="card mt-9 grid gap-5 md:grid-cols-2" onSubmit={submit}>
      <label className="field-label md:col-span-2">活动标题<input className="field" required value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} /></label>
      <label className="field-label">活动类型<select className="field" value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })}><option value="project">项目</option><option value="competition">竞赛</option><option value="workshop">工作坊</option><option value="volunteer">志愿活动</option></select></label>
      <label className="field-label">招募人数<input className="field" type="number" min="1" required value={form.required_count} onChange={(e) => setForm({ ...form, required_count: Number(e.target.value) })} /></label>
      <label className="field-label md:col-span-2">活动介绍<textarea className="field min-h-32" required value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} /></label>
      <label className="field-label">活动标签<input className="field" placeholder="电赛, STM32, 硬件" value={form.tags} onChange={(e) => setForm({ ...form, tags: e.target.value })} /></label>
      <label className="field-label">偏好标签<input className="field" placeholder="嵌入式, 焊接" value={form.preferred_tags} onChange={(e) => setForm({ ...form, preferred_tags: e.target.value })} /></label>
      <label className="field-label">活动时间<input className="field" placeholder="周末下午" value={form.time_text} onChange={(e) => setForm({ ...form, time_text: e.target.value })} /></label>
      <label className="field-label">活动地点<input className="field" placeholder="创新实验室" value={form.location_text} onChange={(e) => setForm({ ...form, location_text: e.target.value })} /></label>
      {error && <p className="notice-error md:col-span-2">{error}</p>}<button className="button-primary md:col-span-2" disabled={pending}>{pending ? "发布中…" : "发布活动"}</button>
    </form></div></section>;
}
