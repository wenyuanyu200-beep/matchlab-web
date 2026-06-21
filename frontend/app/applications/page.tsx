"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import EmptyState from "@/components/EmptyState";
import { ApiError, getToken, request } from "@/lib/api";
import { friendlyStatus } from "@/lib/forms";
import type { Application } from "@/lib/types";

export default function ApplicationsPage() {
  const router = useRouter(); const [applications, setApplications] = useState<Application[] | null>(null); const [error, setError] = useState("");
  useEffect(() => { if (!getToken()) { router.replace("/login"); return; } request<{ applications: Application[] }>("/me/applications").then((data) => setApplications(data.applications || [])).catch((cause) => setError(cause instanceof ApiError ? cause.message : "无法加载报名记录")); }, [router]);
  return <section className="page-shell page-section"><p className="eyebrow">My applications</p><h1 className="page-heading mt-3">我的报名</h1><p className="page-subtitle">查看已经报名的活动与当前审核状态。</p>{error && <p className="notice-error mt-8">{error}</p>}{!applications && !error && <p className="mt-8 text-slate-600">正在加载报名记录…</p>}{applications && <div className="mt-9 grid gap-4">{applications.length ? applications.map((application) => <article className="card flex flex-wrap items-center justify-between gap-5" key={application.id}><div><span className="status-badge">{friendlyStatus(application.status)}</span><h2 className="mt-3 text-xl font-black text-slate-950">{application.activity_title || "校园活动"}</h2><p className="mt-2 text-sm text-slate-600">报名理由：{application.reason || "未填写"}</p></div><Link className="button-secondary" href={`/activities/${application.activity_id}`}>查看活动</Link></article>) : <EmptyState title="还没有报名记录" description="去活动广场寻找感兴趣的协作机会吧。" />}</div>}</section>;
}
