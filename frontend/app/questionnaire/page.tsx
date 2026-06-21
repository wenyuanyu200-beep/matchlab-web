"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import TagList from "@/components/TagList";
import StatCard from "@/components/StatCard";
import { ApiError, getToken, postJSON } from "@/lib/api";
import { splitList } from "@/lib/forms";
import type { Profile } from "@/lib/types";

const initial = { interests: "", skills: "", available_time: "", activity_types: "competition, project", goal: "", communication_style: "" };

export default function QuestionnairePage() {
  const router = useRouter(); const [form, setForm] = useState(initial); const [profile, setProfile] = useState<Profile | null>(null); const [pending, setPending] = useState(false); const [error, setError] = useState("");
  useEffect(() => { if (!getToken()) router.replace("/login"); }, [router]);
  async function submit(event: FormEvent) { event.preventDefault(); setPending(true); setError(""); try { const data = await postJSON<{ profile: Profile }>("/questionnaires", { mode: "activity", answers: { interests: splitList(form.interests), skills: splitList(form.skills), available_time: form.available_time, activity_types: splitList(form.activity_types), goal: form.goal, communication_style: form.communication_style } }); setProfile(data.profile); } catch (cause) { setError(cause instanceof ApiError ? cause.message : "画像生成失败"); } finally { setPending(false); } }
  return <section className="page-shell page-section"><div className="grid gap-8 lg:grid-cols-[.95fr_1.05fr]"><div><p className="eyebrow">Know your collaboration style</p><h1 className="page-heading mt-3">填写协作画像</h1><p className="page-subtitle">告诉 MatchLab 你的兴趣、技能和目标，我们会据此生成活动推荐。</p>
    <form className="card mt-8 grid gap-4" onSubmit={submit}><label className="field-label">兴趣方向<input className="field" placeholder="电赛, STM32, 硬件" value={form.interests} onChange={(e) => setForm({ ...form, interests: e.target.value })} /></label><label className="field-label">技能特长<input className="field" placeholder="嵌入式, 焊接, 控制" value={form.skills} onChange={(e) => setForm({ ...form, skills: e.target.value })} /></label><label className="field-label">可参与时间<input className="field" placeholder="周末下午" value={form.available_time} onChange={(e) => setForm({ ...form, available_time: e.target.value })} /></label><label className="field-label">活动类型<input className="field" placeholder="competition, project" value={form.activity_types} onChange={(e) => setForm({ ...form, activity_types: e.target.value })} /></label><label className="field-label">参与目标<textarea className="field min-h-24" placeholder="找项目队友一起参加比赛" value={form.goal} onChange={(e) => setForm({ ...form, goal: e.target.value })} /></label><label className="field-label">沟通风格<input className="field" placeholder="稳定沟通" value={form.communication_style} onChange={(e) => setForm({ ...form, communication_style: e.target.value })} /></label>{error && <p className="notice-error">{error}</p>}<button className="button-primary" disabled={pending}>{pending ? "生成中…" : "生成我的画像"}</button></form></div>
    <div>{profile ? <div className="card lg:sticky lg:top-24"><p className="eyebrow">Your profile</p><h2 className="mt-3 text-2xl font-black text-slate-950">我的活动画像</h2><p className="mt-5 leading-8 text-slate-650">{profile.summary}</p><div className="mt-6"><TagList tags={profile.tags} /></div><div className="mt-8 grid grid-cols-2 gap-3"><StatCard label="兴趣" value={profile.scores.interest_score} /><StatCard label="技能" value={profile.scores.skill_score} accent="cyan" /><StatCard label="时间" value={profile.scores.time_score} /><StatCard label="目标" value={profile.scores.goal_score} accent="orange" /><StatCard label="沟通" value={profile.scores.communication_score} accent="cyan" /></div></div> : <div className="card flex min-h-80 items-center justify-center text-center"><div><span className="text-5xl text-indigo-300">◎</span><h2 className="mt-5 text-xl font-black text-slate-950">画像将在这里生成</h2><p className="mt-2 text-sm text-slate-600">完成左侧问卷即可查看标签、评分与总结。</p></div></div>}</div>
  </div></section>;
}
