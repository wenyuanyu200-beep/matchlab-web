"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { ApiError, getToken, request } from "@/lib/api";
import type { User } from "@/lib/types";

const shortcuts = [
  ["活动广场", "发现校园活动和项目", "/activities", "⌁"],
  ["发布活动", "发起你的协作计划", "/activities/create", "+"],
  ["填写问卷", "完善兴趣与技能画像", "/questionnaire", "◎"],
  ["智能推荐", "生成个性化活动推荐", "/match", "✦"],
  ["我的报名", "查看报名审核状态", "/applications", "✓"],
];

export default function DashboardPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [error, setError] = useState("");
  useEffect(() => {
    if (!getToken()) { router.replace("/login"); return; }
    request<{ user: User }>("/me").then((data) => setUser(data.user)).catch((cause) => setError(cause instanceof ApiError ? cause.message : "无法加载用户信息"));
  }, [router]);
  return (
    <section className="page-shell page-section">
      {error && <p className="notice-error">{error}</p>}
      {!user && !error && <p className="text-slate-600">正在加载工作台…</p>}
      {user && <>
        <div className="rounded-[2rem] bg-gradient-to-br from-indigo-950 to-indigo-700 p-7 text-white shadow-2xl shadow-indigo-200 sm:p-10">
          <p className="text-sm font-bold text-cyan-200">MY MATCHLAB</p><h1 className="mt-3 text-3xl font-black">你好，{user.nickname || "校园伙伴"}</h1>
          <div className="mt-6 flex flex-wrap gap-x-8 gap-y-2 text-sm text-indigo-100"><span>{user.email}</span><span>{user.school || "学校未填写"}</span><span>{user.role === "admin" ? "管理员" : "普通用户"}</span></div>
        </div>
        <div className="mt-10 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {shortcuts.map(([title, text, href, icon]) => <Link className="card group" href={href} key={href}><span className="grid size-10 place-items-center rounded-xl bg-indigo-50 text-xl font-black text-indigo-600">{icon}</span><h2 className="mt-7 text-xl font-black text-slate-950 group-hover:text-indigo-700">{title}</h2><p className="mt-2 text-sm text-slate-600">{text}</p></Link>)}
          {user.role === "admin" && <Link className="card border-orange-200 bg-orange-50/60" href="/admin"><span className="eyebrow text-orange-700">ADMIN</span><h2 className="mt-7 text-xl font-black text-slate-950">管理员后台</h2><p className="mt-2 text-sm text-slate-600">查看平台数据与运营概况</p></Link>}
        </div>
      </>}
    </section>
  );
}
