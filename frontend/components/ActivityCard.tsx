import Link from "next/link";
import { friendlyStatus } from "@/lib/forms";
import type { Activity } from "@/lib/types";
import TagList from "./TagList";

export default function ActivityCard({ activity }: { activity: Activity }) {
  return (
    <article className="card group flex h-full flex-col">
      <div className="mb-4 flex items-start justify-between gap-3">
        <div>
          <span className="eyebrow">{activity.type || "project"}</span>
          <h2 className="mt-2 text-xl font-black text-slate-950 group-hover:text-indigo-700">{activity.title}</h2>
        </div>
        <span className="status-badge">{friendlyStatus(activity.status)}</span>
      </div>
      <p className="line-clamp-3 flex-1 text-sm leading-6 text-slate-650">{activity.description}</p>
      <div className="mt-4"><TagList tags={activity.tags} /></div>
      <dl className="mt-5 grid grid-cols-2 gap-2 border-t border-slate-100 pt-4 text-sm text-slate-600">
        <div><dt className="text-xs text-slate-500">发起人</dt><dd className="mt-1 font-semibold text-slate-800">{activity.creator ? `${activity.creator.nickname} · ${activity.creator.school}` : "校园伙伴"}</dd></div>
        <div><dt className="text-xs text-slate-500">当前人数</dt><dd className="mt-1 font-semibold text-slate-800">{activity.joined_count} / {activity.required_count} 人</dd></div>
      </dl>
      <Link className="button-secondary mt-5 w-full" href={`/activities/${activity.id}`}>查看详情</Link>
    </article>
  );
}
