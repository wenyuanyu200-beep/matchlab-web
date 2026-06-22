"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { activityTypeOptions, placeholdersFor } from "@/lib/activityTypes";
import { ApiError, getToken, postJSON } from "@/lib/api";
import { splitList } from "@/lib/forms";

const initial = {
  title: "",
  type: "project",
  description: "",
  required_count: 2,
  tags: "",
  preferred_tags: "",
  time_text: "",
  location_text: "",
  roles: "",
  notes: "",
};

export default function CreateActivityPage() {
  const router = useRouter();
  const [form, setForm] = useState(initial);
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  const placeholders = useMemo(() => placeholdersFor(form.type), [form.type]);

  useEffect(() => {
    if (!getToken()) router.replace("/login");
  }, [router]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setPending(true);
    setError("");
    try {
      const descriptionParts = [
        form.description.trim(),
        form.roles.trim() ? `适合对象 / 招募角色：${form.roles.trim()}` : "",
        form.notes.trim() ? `备注说明：${form.notes.trim()}` : "",
      ].filter(Boolean);
      await postJSON("/activities", {
        title: form.title,
        type: form.type,
        description: descriptionParts.join("\n\n"),
        required_count: Number(form.required_count),
        tags: splitList(form.tags),
        preferred_tags: splitList([form.preferred_tags, form.roles].filter(Boolean).join(", ")),
        time_text: form.time_text,
        location_text: form.location_text,
      });
      router.push("/activities");
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : "活动发布失败");
    } finally {
      setPending(false);
    }
  }

  return (
    <section className="page-shell page-section">
      <div className="mx-auto max-w-4xl">
        <p className="eyebrow">Create activity</p>
        <h1 className="page-heading mt-3">发布校园活动</h1>
        <p className="page-subtitle">
          你可以发布比赛组队、项目合作、学习搭子、社团活动或兴趣活动。MatchLab 会根据参与者的画像与活动要求，帮助你找到更合适的伙伴。
        </p>

        <form className="card mt-9 grid gap-5 md:grid-cols-2" onSubmit={submit}>
          <label className="field-label md:col-span-2">
            活动标题
            <input
              className="field"
              required
              placeholder={placeholders.title}
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
            />
          </label>

          <label className="field-label">
            活动类型
            <select className="field" value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })}>
              {activityTypeOptions.map((option) => (
                <option value={option.value} key={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>

          <label className="field-label">
            招募人数
            <input
              className="field"
              type="number"
              min="1"
              required
              value={form.required_count}
              onChange={(e) => setForm({ ...form, required_count: Number(e.target.value) })}
            />
          </label>

          <label className="field-label md:col-span-2">
            活动简介
            <textarea
              className="field min-h-32"
              required
              placeholder={placeholders.description}
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
            />
          </label>

          <label className="field-label">
            活动标签
            <input
              aria-label="活动标签"
              className="field"
              placeholder={placeholders.tags}
              value={form.tags}
              onChange={(e) => setForm({ ...form, tags: e.target.value })}
            />
            <span className="text-xs font-medium leading-5 text-slate-500">用于判断兴趣匹配，多个标签用逗号分隔。</span>
          </label>

          <label className="field-label">
            希望伙伴具备的标签
            <input
              aria-label="希望伙伴具备的标签"
              className="field"
              placeholder={placeholders.preferredTags}
              value={form.preferred_tags}
              onChange={(e) => setForm({ ...form, preferred_tags: e.target.value })}
            />
            <span className="text-xs font-medium leading-5 text-slate-500">用于判断技能或协作方式匹配。</span>
          </label>

          <label className="field-label">
            时间
            <input
              className="field"
              placeholder="如：周三晚间、周末下午、每周一次"
              value={form.time_text}
              onChange={(e) => setForm({ ...form, time_text: e.target.value })}
            />
          </label>

          <label className="field-label">
            地点
            <input
              className="field"
              placeholder="如：图书馆、线上会议、工程训练中心"
              value={form.location_text}
              onChange={(e) => setForm({ ...form, location_text: e.target.value })}
            />
          </label>

          <label className="field-label md:col-span-2">
            适合对象 / 招募角色
            <input
              className="field"
              placeholder={placeholders.roles}
              value={form.roles}
              onChange={(e) => setForm({ ...form, roles: e.target.value })}
            />
          </label>

          <label className="field-label md:col-span-2">
            备注说明
            <textarea
              className="field min-h-24"
              placeholder="如：是否需要固定到场、是否欢迎新手、是否有截止时间"
              value={form.notes}
              onChange={(e) => setForm({ ...form, notes: e.target.value })}
            />
          </label>

          {error && <p className="notice-error md:col-span-2">{error}</p>}
          <button className="button-primary md:col-span-2" disabled={pending}>
            {pending ? "发布中…" : "发布校园活动"}
          </button>
        </form>
      </div>
    </section>
  );
}
