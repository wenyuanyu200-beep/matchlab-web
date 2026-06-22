"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getToken, request } from "@/lib/api";
import { CIRCLE_CATEGORIES } from "@/lib/circles";
import type { Circle } from "@/lib/types";

export default function CreateCirclePage() {
  const router = useRouter(); const [name, setName] = useState(""); const [description, setDescription] = useState(""); const [category, setCategory] = useState("study"); const [busy, setBusy] = useState(false); const [error, setError] = useState("");
  useEffect(() => { if (!getToken()) router.replace("/login"); }, [router]);
  async function submit(event: FormEvent) { event.preventDefault(); if (!name.trim() || !description.trim()) return; setBusy(true); setError(""); try { const data = await request<{ circle?: Circle }>("/circles", { method: "POST", body: JSON.stringify({ name: name.trim(), description: description.trim(), category }) }); router.push(data.circle?.id ? `/circles/${data.circle.id}` : "/circles"); } catch { setError("创建失败，请稍后重试"); } finally { setBusy(false); } }
  return <main className="page-shell page-section"><div className="mx-auto max-w-2xl"><p className="eyebrow">New Circle</p><h1 className="page-heading mt-3">创建圈子</h1><p className="page-subtitle">提交后由管理员审核，通过后会出现在圈子广场。</p><form className="card mt-8 grid gap-5" onSubmit={submit}><label className="grid gap-2 font-bold">圈子名称<input className="field" maxLength={60} value={name} onChange={(e) => setName(e.target.value)} /></label><label className="grid gap-2 font-bold">分类<select className="field" value={category} onChange={(e) => setCategory(e.target.value)}>{CIRCLE_CATEGORIES.slice(1).map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select></label><label className="grid gap-2 font-bold">圈子介绍<textarea className="field min-h-36" maxLength={500} value={description} onChange={(e) => setDescription(e.target.value)} /></label>{error && <p className="notice-error">{error}</p>}<button className="button-primary" disabled={!name.trim() || !description.trim() || busy}>{busy ? "提交中…" : "提交审核"}</button></form></div></main>;
}
