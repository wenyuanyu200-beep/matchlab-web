"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import ActivityCard from "@/components/ActivityCard";
import EmptyState from "@/components/EmptyState";
import { ApiError, request } from "@/lib/api";
import type { Activity } from "@/lib/types";

export default function ActivitiesPage() {
  const [activities, setActivities] = useState<Activity[] | null>(null);
  const [error, setError] = useState("");
  useEffect(() => { request<{ activities: Activity[] }>("/activities").then((data) => setActivities(data.activities || [])).catch((cause) => setError(cause instanceof ApiError ? cause.message : "无法加载活动")); }, []);
  return <section className="page-shell page-section">
    <div className="flex flex-wrap items-end justify-between gap-5"><div><p className="eyebrow">Explore campus</p><h1 className="page-heading mt-3">活动广场</h1><p className="page-subtitle">发现值得一起完成的比赛、项目与校园活动。</p></div><Link className="button-primary" href="/activities/create">发布活动</Link></div>
    {error && <p className="notice-error mt-8">{error}</p>}
    {!activities && !error && <p className="mt-10 text-slate-600">正在加载活动…</p>}
    {activities && <div className="mt-10 grid gap-5 md:grid-cols-2 xl:grid-cols-3">{activities.length ? activities.map((activity) => <ActivityCard key={activity.id} activity={activity} />) : <EmptyState title="暂时没有招募中的活动" description="可以成为第一个发布活动的人。" />}</div>}
  </section>;
}
