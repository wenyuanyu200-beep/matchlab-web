"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { getToken, request } from "@/lib/api";
import { categoryLabel, CIRCLE_CATEGORIES, filterCircles, listFrom } from "@/lib/circles";
import type { Circle } from "@/lib/types";

export default function CirclesPage() {
  const router = useRouter();
  const [circles, setCircles] = useState<Circle[]>([]); const [query, setQuery] = useState(""); const [category, setCategory] = useState("all"); const [error, setError] = useState("");
  useEffect(() => { request<unknown>("/circles").then((data) => setCircles(listFrom<Circle>(data, "circles"))).catch(() => setError("圈子加载失败，请稍后重试")); }, []);
  const visible = useMemo(() => filterCircles(circles, query, category), [circles, query, category]);
  async function join(circle: Circle) {
    if (!getToken()) { router.push("/login"); return; }
    try { await request(`/circles/${circle.id}/join`, { method: "POST" }); setCircles((items) => items.map((item) => item.id === circle.id ? { ...item, membership_status: "pending" } : item)); }
    catch { setError("加入申请提交失败，请稍后重试"); }
  }
  return <main className="page-shell page-section">
    <div className="flex flex-wrap items-end justify-between gap-5"><div><p className="eyebrow">Campus Circles</p><h1 className="page-heading mt-3">找到你的校园同好</h1><p className="page-subtitle">一起学习、运动、做项目，让兴趣变成真实连接。</p></div><Link className="button-primary" href="/circles/create">创建圈子</Link></div>
    <div className="card mt-8 flex flex-col gap-4 md:flex-row"><input className="field flex-1" aria-label="搜索圈子" placeholder="搜索名称或介绍" value={query} onChange={(e) => setQuery(e.target.value)} /><select className="field md:w-44" aria-label="圈子分类" value={category} onChange={(e) => setCategory(e.target.value)}>{CIRCLE_CATEGORIES.map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select></div>
    {error && <p className="notice-error mt-5">{error}</p>}
    <div className="mt-7 grid gap-5 md:grid-cols-2 xl:grid-cols-3">{visible.map((circle) => <article className="card flex flex-col" key={circle.id}><div className="flex items-start justify-between gap-3"><span className="tag">{categoryLabel(circle.category)}</span><span className="text-sm text-slate-500">{circle.member_count ?? 0} 位成员</span></div><h2 className="mt-4 text-xl font-black text-slate-950"><Link href={`/circles/${circle.id}`}>{circle.name}</Link></h2><p className="mt-2 flex-1 text-sm leading-6 text-slate-600">{circle.description}</p><div className="mt-5 flex gap-3"><Link className="button-secondary flex-1 text-center" href={`/circles/${circle.id}`}>查看详情</Link>{circle.membership_status === "joined" ? <span className="tag self-center">已加入</span> : <button className="button-primary flex-1" disabled={circle.membership_status === "pending"} onClick={() => join(circle)}>{circle.membership_status === "pending" ? "审核中" : "加入"}</button>}</div></article>)}</div>
    {!visible.length && !error && <div className="card mt-7 text-center text-slate-500">没有找到匹配的圈子</div>}
  </main>;
}
