"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import StatCard from "@/components/StatCard";
import { ApiError, getToken, request } from "@/lib/api";
import { friendlyStatus } from "@/lib/forms";
import type { AdminActivity, AdminStats, Application, Feedback, User } from "@/lib/types";

interface AdminData { stats: AdminStats; users: User[]; activities: AdminActivity[]; applications: Application[]; feedbacks: Feedback[]; }

export default function AdminPage() {
  const router = useRouter(); const [denied, setDenied] = useState(false); const [data, setData] = useState<AdminData | null>(null); const [error, setError] = useState("");
  useEffect(() => {
    if (!getToken()) { router.replace("/login"); return; }
    async function load() {
      try {
        const current = await request<{ user: User }>("/me");
        if (current.user.role !== "admin") { setDenied(true); return; }
        const [stats, users, activities, applications, feedbacks] = await Promise.all([
          request<{ stats: AdminStats }>("/admin/stats"), request<{ users: User[] }>("/admin/users?limit=100&offset=0"), request<{ activities: AdminActivity[] }>("/admin/activities?limit=100&offset=0"), request<{ applications: Application[] }>("/admin/applications?limit=100&offset=0"), request<{ feedbacks: Feedback[] }>("/admin/feedbacks?limit=100&offset=0"),
        ]);
        setData({ stats: stats.stats, users: users.users || [], activities: activities.activities || [], applications: applications.applications || [], feedbacks: feedbacks.feedbacks || [] });
      } catch (cause) { setError(cause instanceof ApiError ? cause.message : "管理员数据加载失败"); }
    }
    load();
  }, [router]);
  if (denied) return <section className="page-shell page-section"><div className="card mx-auto max-w-xl py-16 text-center"><div className="text-4xl">⊘</div><h1 className="mt-5 text-2xl font-black text-slate-950">无权限访问</h1><p className="mt-3 text-slate-600">当前账号不是管理员，无法查看后台数据。</p></div></section>;
  return <section className="page-shell page-section"><p className="eyebrow">Administration</p><h1 className="page-heading mt-3">平台数据总览</h1><p className="page-subtitle">查看用户、活动、报名、推荐、问卷与反馈的运营概况。</p>{error && <p className="notice-error mt-8">{error}</p>}{!data && !error && <p className="mt-8 text-slate-600">正在加载管理员数据…</p>}{data && <>
    <div className="mt-9 grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6"><StatCard label="用户数" value={data.stats.users_count} /><StatCard label="活动数" value={data.stats.activities_count} accent="cyan" /><StatCard label="报名数" value={data.stats.applications_count} /><StatCard label="推荐数" value={data.stats.matches_count} accent="orange" /><StatCard label="问卷数" value={data.stats.questionnaires_count} /><StatCard label="反馈数" value={data.stats.feedbacks_count} accent="cyan" /></div>
    <AdminSection title="用户"><div className="table-wrap"><table className="data-table"><thead><tr><th>昵称</th><th>邮箱</th><th>学校</th><th>角色</th><th>注册时间</th></tr></thead><tbody>{data.users.map((user) => <tr key={user.id}><td>{user.nickname || "-"}</td><td>{user.email}</td><td>{user.school || "-"}</td><td>{user.role}</td><td>{formatDate(user.created_at)}</td></tr>)}</tbody></table></div></AdminSection>
    <AdminSection title="活动"><div className="table-wrap"><table className="data-table"><thead><tr><th>标题</th><th>类型</th><th>状态</th><th>发起人</th><th>人数</th></tr></thead><tbody>{data.activities.map((activity) => <tr key={activity.id}><td>{activity.title}</td><td>{activity.type}</td><td>{friendlyStatus(activity.status)}</td><td>{activity.creator?.nickname || "-"}</td><td>{activity.joined_count}/{activity.required_count}</td></tr>)}</tbody></table></div>{!data.activities.length && <p className="card text-slate-600">暂无活动数据</p>}</AdminSection>
    <AdminSection title="报名"><div className="table-wrap"><table className="data-table"><thead><tr><th>活动</th><th>申请人</th><th>理由</th><th>状态</th><th>时间</th></tr></thead><tbody>{data.applications.map((item) => <tr key={item.id}><td>{item.activity_title || item.activity_id}</td><td>{item.applicant?.nickname || item.applicant_id || "-"}</td><td>{item.reason || "-"}</td><td>{friendlyStatus(item.status)}</td><td>{formatDate(item.created_at)}</td></tr>)}</tbody></table></div>{!data.applications.length && <p className="card text-slate-600">暂无报名数据</p>}</AdminSection>
    <AdminSection title="反馈">{data.feedbacks.length ? <div className="table-wrap"><table className="data-table"><thead><tr><th>用户</th><th>评分</th><th>内容</th><th>时间</th></tr></thead><tbody>{data.feedbacks.map((item) => <tr key={item.id}><td>{item.user_id}</td><td>{item.rating}/5</td><td>{item.comment || "-"}</td><td>{formatDate(item.created_at)}</td></tr>)}</tbody></table></div> : <p className="card text-slate-600">暂无反馈数据</p>}</AdminSection>
  </>}</section>;
}

function AdminSection({ title, children }: { title: string; children: React.ReactNode }) { return <section className="mt-12"><h2 className="mb-5 text-2xl font-black text-slate-950">{title}</h2>{children}</section>; }
function formatDate(value?: string) { return value ? new Date(value).toLocaleDateString("zh-CN") : "-"; }
