import Link from "next/link";
import { activityTypeLabel } from "@/lib/activityTypes";
import { friendlyStatus } from "@/lib/forms";
import type { Activity } from "@/lib/types";
import TagList from "./TagList";

export default function ActivityCard({ activity }: { activity: Activity }) {
  const creator = activity.creator
    ? `${activity.creator.nickname} · ${activity.creator.school}`
    : "校园伙伴";

  return (
    <article className="card group flex h-full flex-col">
      <div className="mb-4 flex items-start justify-between gap-3">
        <div>
          <span className="status-badge bg-indigo-50 text-indigo-700">{activityTypeLabel(activity.type)}</span>
          <h2 className="mt-3 text-xl font-black text-slate-950 group-hover:text-indigo-700">{activity.title}</h2>
        </div>
        <span className="status-badge">{friendlyStatus(activity.status)}</span>
      </div>
      <p className="line-clamp-3 flex-1 text-sm leading-6 text-slate-650">{activity.description}</p>
      <div className="mt-4">
        <TagList tags={activity.tags} />
      </div>
      <p className="mt-4 rounded-xl bg-slate-50 px-3 py-2 text-sm leading-6 text-slate-600">
        适合关注{activity.tags.slice(0, 2).join("、") || activityTypeLabel(activity.type)}，并希望参与校园协作的同学。
      </p>
      <dl className="mt-5 grid grid-cols-2 gap-3 border-t border-slate-100 pt-4 text-sm text-slate-600">
        <div>
          <dt className="text-xs text-slate-500">当前人数</dt>
          <dd className="mt-1 font-semibold text-slate-800">
            {activity.joined_count} / {activity.required_count} 人
          </dd>
        </div>
        <div>
          <dt className="text-xs text-slate-500">发起人</dt>
          <dd className="mt-1 font-semibold text-slate-800">{creator}</dd>
        </div>
        <div>
          <dt className="text-xs text-slate-500">时间</dt>
          <dd className="mt-1 font-semibold text-slate-800">{activity.time_text || "待沟通"}</dd>
        </div>
        <div>
          <dt className="text-xs text-slate-500">地点</dt>
          <dd className="mt-1 font-semibold text-slate-800">{activity.location_text || "待沟通"}</dd>
        </div>
      </dl>
      <Link className="button-secondary mt-5 w-full" href={`/activities/${activity.id}`}>
        查看推荐理由
      </Link>
    </article>
  );
}
