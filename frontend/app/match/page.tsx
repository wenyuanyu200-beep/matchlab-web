"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import EmptyState from "@/components/EmptyState";
import { activityTypeLabel } from "@/lib/activityTypes";
import { ApiError, getToken, postJSON, request } from "@/lib/api";
import type { Recommendation } from "@/lib/types";

const scoreLabels: [keyof Recommendation["detail_scores"], string][] = [
  ["interest", "兴趣"],
  ["skill", "技能"],
  ["type", "类型"],
  ["time", "时间"],
  ["goal", "目标"],
];

function RecommendationCard({ item }: { item: Recommendation }) {
  return (
    <article className="card">
      <div className="flex items-start justify-between gap-4">
        <div>
          <span className="status-badge bg-indigo-50 text-indigo-700">{activityTypeLabel(item.activity.type)}</span>
          <h2 className="mt-3 text-xl font-black text-slate-950">{item.activity.title}</h2>
        </div>
        <div className="grid size-16 shrink-0 place-items-center rounded-2xl bg-indigo-600 text-2xl font-black text-white">{Math.round(item.score)}</div>
      </div>
      <p className="mt-5 leading-7 text-slate-650">{item.reason}</p>
      <div className="mt-6 grid grid-cols-5 gap-2">
        {scoreLabels.map(([key, label]) => (
          <div className="rounded-xl bg-slate-50 p-2 text-center" key={key}>
            <div className="text-lg font-black text-indigo-700">{item.detail_scores?.[key] ?? 0}</div>
            <div className="mt-1 text-xs text-slate-500">{label}</div>
          </div>
        ))}
      </div>
      {item.activity.id && (
        <Link className="button-secondary mt-6 w-full" href={`/activities/${item.activity.id}`}>
          查看推荐理由
        </Link>
      )}
    </article>
  );
}

export default function MatchPage() {
  const router = useRouter();
  const [matches, setMatches] = useState<Recommendation[] | null>(null);
  const [generated, setGenerated] = useState<Recommendation[] | null>(null);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    request<{ matches: Recommendation[] }>("/me/matches")
      .then((data) => setMatches(data.matches || []))
      .catch((cause) => setError(cause instanceof ApiError ? cause.message : "无法加载推荐"));
  }, [router]);

  async function generate() {
    setPending(true);
    setError("");
    try {
      const data = await postJSON<{ recommendations: Recommendation[] }>("/match/recommend", { target_type: "activity", limit: 10 });
      setGenerated(data.recommendations || []);
      const saved = await request<{ matches: Recommendation[] }>("/me/matches");
      setMatches(saved.matches || data.recommendations || []);
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : "推荐生成失败");
    } finally {
      setPending(false);
    }
  }

  const display = generated || matches;
  return (
    <section className="page-shell page-section">
      <div className="rounded-2xl bg-gradient-to-r from-indigo-950 via-indigo-800 to-cyan-800 p-7 text-white sm:p-10">
        <p className="text-sm font-black tracking-widest text-cyan-200">MATCH ENGINE</p>
        <div className="mt-4 flex flex-wrap items-end justify-between gap-6">
          <div>
            <h1 className="text-3xl font-black sm:text-5xl">适合我的活动</h1>
            <p className="mt-4 max-w-2xl leading-7 text-indigo-100">
              基于你的协作画像与活动需求计算匹配分。推荐会综合兴趣、技能、活动类型、时间地点、目标和协作风格，并保留 detail scores 方便查看算法依据。
            </p>
          </div>
          <button className="rounded-xl bg-orange-400 px-6 py-3 font-black text-orange-950 shadow-lg disabled:opacity-60" onClick={generate} disabled={pending}>
            {pending ? "计算中…" : "查看适合我的活动"}
          </button>
        </div>
      </div>
      {error && <p className="notice-error mt-8">{error}</p>}
      <div className="mt-10">
        <h2 className="text-2xl font-black text-slate-950">{generated ? "本次推荐" : "当前推荐结果"}</h2>
        {display ? (
          <div className="mt-5 grid gap-5 lg:grid-cols-2">
            {display.length ? display.map((item, index) => <RecommendationCard key={item.id || item.activity.id || index} item={item} />) : <EmptyState title="还没有推荐结果" description="先填写协作画像，再点击查看适合我的活动。" />}
          </div>
        ) : (
          <p className="mt-5 text-slate-600">正在读取推荐结果…</p>
        )}
      </div>
    </section>
  );
}
