"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import TagList from "@/components/TagList";
import StatCard from "@/components/StatCard";
import { activityTypeOptions } from "@/lib/activityTypes";
import { ApiError, getToken, postJSON } from "@/lib/api";
import { splitList } from "@/lib/forms";
import type { Profile } from "@/lib/types";

const initial = {
  interests: "AI, 编程, 设计",
  hobbies: "",
  explore_fields: "",
  skills: "",
  skill_level: "一般",
  experiences: "",
  mbti: "暂不确定",
  communication_style: "稳定沟通",
  team_role: "稳定执行者",
  work_rhythm: "按阶段推进",
  available_time: "工作日晚间, 周末下午",
  participation_mode: "线上线下都可以",
  duration_preference: "短期组队",
  campus_or_location: "",
  main_goal: "找项目队友",
  partner_expectation: "",
  avoid_points: "",
  preferred_activity_types: "project, study, social",
  participation_purpose: "学习, 提升技能, 丰富经历",
};

const optionGroups = {
  interests: ["AI", "编程", "动画", "设计", "摄影", "音乐", "运动", "阅读", "公益", "创业", "游戏", "社团", "电子设计", "影视创作"],
  skills: ["前端开发", "后端开发", "单片机", "嵌入式", "焊接", "视频剪辑", "平面设计", "UI", "文案", "策划", "组织", "答辩", "外联", "摄影", "运营", "数据整理"],
  experiences: ["参加过比赛", "做过课程项目", "做过社团活动", "做过志愿服务", "做过调研报告", "做过答辩展示", "暂无经验但愿意尝试"],
  communication: ["稳定沟通", "直接高效", "喜欢讨论", "慢热谨慎", "执行导向", "创意发散"],
  roles: ["组织协调者", "稳定执行者", "创意提出者", "技术支持者", "资料整理者", "展示表达者"],
  rhythm: ["提前规划", "按阶段推进", "边做边调整", "截止前冲刺"],
  time: ["工作日白天", "工作日晚间", "周末上午", "周末下午", "周末晚间"],
  goals: ["找项目队友", "找比赛队友", "找学习搭子", "参加校园活动", "丰富经历", "提升技能", "认识同方向同学", "完成课程任务"],
  expectations: ["靠谱执行型", "技术强", "好沟通", "有创意", "认真负责", "时间稳定", "愿意一起学习"],
  avoid: ["不回消息", "临时鸽子", "分工不清", "长期摸鱼", "截止前失联", "沟通不尊重"],
  purpose: ["获奖", "学习", "丰富经历", "找长期搭档", "放松体验", "完成课程任务"],
};

const mbtiOptions = ["INTJ", "INTP", "ENTJ", "ENTP", "INFJ", "INFP", "ENFJ", "ENFP", "ISTJ", "ISFJ", "ESTJ", "ESFJ", "ISTP", "ISFP", "ESTP", "ESFP", "暂不确定"];

function toggleListValue(current: string, value: string) {
  const items = splitList(current);
  const exists = items.includes(value);
  return (exists ? items.filter((item) => item !== value) : [...items, value]).join(", ");
}

function ChoiceChips({ values, selected, onToggle }: { values: string[]; selected: string; onToggle: (next: string) => void }) {
  const selectedItems = splitList(selected);
  return (
    <div className="flex flex-wrap gap-2">
      {values.map((value) => {
        const active = selectedItems.includes(value);
        return (
          <button
            className={active ? "chip chip-active" : "chip"}
            key={value}
            type="button"
            onClick={() => onToggle(toggleListValue(selected, value))}
          >
            {value}
          </button>
        );
      })}
    </div>
  );
}

function ActivityTypeChips({ selected, onToggle }: { selected: string; onToggle: (next: string) => void }) {
  const selectedItems = splitList(selected);
  return (
    <div className="flex flex-wrap gap-2">
      {activityTypeOptions.map((item) => {
        const active = selectedItems.includes(item.value);
        return (
          <button
            className={active ? "chip chip-active" : "chip"}
            key={item.value}
            type="button"
            onClick={() => onToggle(toggleListValue(selected, item.value))}
          >
            {item.label}
          </button>
        );
      })}
    </div>
  );
}

function Section({ title, description, children }: { title: string; description: string; children: ReactNode }) {
  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-5">
      <h2 className="text-lg font-black text-slate-950">{title}</h2>
      <p className="mt-2 text-sm leading-6 text-slate-600">{description}</p>
      <div className="mt-5 grid gap-4">{children}</div>
    </section>
  );
}

export default function QuestionnairePage() {
  const router = useRouter();
  const [form, setForm] = useState(initial);
  const [profile, setProfile] = useState<Profile | null>(null);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");
  const profileTags = useMemo(() => {
    if (!profile) return ["协作画像", "活动推荐", "校园协作"];
    return profile.tags.length ? profile.tags.slice(0, 8) : ["协作画像", "校园协作"];
  }, [profile]);

  useEffect(() => {
    if (!getToken()) router.replace("/login");
  }, [router]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setPending(true);
    setError("");
    try {
      const data = await postJSON<{ profile: Profile }>("/questionnaires", {
        mode: "activity",
        answers: {
          interests: splitList(form.interests),
          hobbies: form.hobbies,
          explore_fields: form.explore_fields,
          skills: splitList(form.skills),
          skill_level: form.skill_level,
          experiences: splitList(form.experiences),
          mbti: form.mbti,
          communication_style: form.communication_style,
          team_role: form.team_role,
          work_rhythm: form.work_rhythm,
          available_time: form.available_time,
          participation_mode: form.participation_mode,
          duration_preference: form.duration_preference,
          campus_or_location: form.campus_or_location,
          activity_types: splitList(form.preferred_activity_types),
          preferred_activity_types: splitList(form.preferred_activity_types),
          goal: form.main_goal,
          main_goal: form.main_goal,
          partner_expectation: splitList(form.partner_expectation),
          avoid_points: splitList(form.avoid_points),
          participation_purpose: splitList(form.participation_purpose),
        },
      });
      setProfile(data.profile);
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : "画像生成失败");
    } finally {
      setPending(false);
    }
  }

  return (
    <section className="page-shell page-section">
      <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_390px]">
        <div>
          <p className="eyebrow">Collaboration profile</p>
          <h1 className="page-heading mt-3">填写你的协作画像</h1>
          <p className="page-subtitle">
            告诉 MatchLab 你的兴趣、技能、时间与协作风格，我们会据此生成画像并推荐更适合你的活动与伙伴。
          </p>
          <div className="mt-6 grid gap-3 sm:grid-cols-2">
            {["不只看技能，也结合兴趣、目标和时间", "用于推荐活动搭子、项目队友和学习伙伴"].map((text) => (
              <div className="rounded-2xl border border-indigo-100 bg-indigo-50 px-4 py-3 text-sm font-semibold text-indigo-900" key={text}>
                {text}
              </div>
            ))}
          </div>

          <form className="mt-8 grid gap-5" onSubmit={submit}>
            <Section title="1. 兴趣与爱好" description="兴趣方向会帮助我们判断你更愿意参与哪些类型的校园活动。">
              <div className="field-label">
                <span>兴趣方向</span>
                <ChoiceChips values={optionGroups.interests} selected={form.interests} onToggle={(interests) => setForm({ ...form, interests })} />
                <input aria-label="兴趣方向" className="field" value={form.interests} onChange={(e) => setForm({ ...form, interests: e.target.value })} placeholder="AI, 摄影, 公益, 电子设计" />
              </div>
              <label className="field-label">
                日常爱好
                <input className="field" value={form.hobbies} onChange={(e) => setForm({ ...form, hobbies: e.target.value })} placeholder="如：羽毛球、拍照、看电影、逛展、做手工、旅行" />
              </label>
              <label className="field-label">
                想尝试的新领域
                <input className="field" value={form.explore_fields} onChange={(e) => setForm({ ...form, explore_fields: e.target.value })} placeholder="如：短视频拍摄、前端开发、公益活动、创新创业项目" />
              </label>
            </Section>

            <Section title="2. 技能与经验" description="技能不是为了筛掉新手，而是为了让系统更清楚你适合承担什么角色。">
              <div className="field-label">
                <span>技能标签</span>
                <ChoiceChips values={optionGroups.skills} selected={form.skills} onToggle={(skills) => setForm({ ...form, skills })} />
                <input aria-label="技能标签" className="field" value={form.skills} onChange={(e) => setForm({ ...form, skills: e.target.value })} placeholder="前端开发, 视频剪辑, 文案, 数据整理" />
              </div>
              <label className="field-label">
                熟练程度
                <select className="field" value={form.skill_level} onChange={(e) => setForm({ ...form, skill_level: e.target.value })}>
                  {["入门", "一般", "熟练", "较强"].map((item) => <option key={item}>{item}</option>)}
                </select>
              </label>
              <label className="field-label">
                过往经历
                <ChoiceChips values={optionGroups.experiences} selected={form.experiences} onToggle={(experiences) => setForm({ ...form, experiences })} />
              </label>
            </Section>

            <Section title="3. 性格与协作风格" description="MBTI 只作为辅助参考，系统更关注你的协作方式和团队偏好。">
              <div className="grid gap-4 md:grid-cols-2">
                <label className="field-label">
                  MBTI
                  <select className="field" value={form.mbti} onChange={(e) => setForm({ ...form, mbti: e.target.value })}>
                    {mbtiOptions.map((item) => <option key={item}>{item}</option>)}
                  </select>
                </label>
                <label className="field-label">
                  沟通风格
                  <select className="field" value={form.communication_style} onChange={(e) => setForm({ ...form, communication_style: e.target.value })}>
                    {optionGroups.communication.map((item) => <option key={item}>{item}</option>)}
                  </select>
                </label>
                <label className="field-label">
                  团队偏好
                  <select className="field" value={form.team_role} onChange={(e) => setForm({ ...form, team_role: e.target.value })}>
                    {optionGroups.roles.map((item) => <option key={item}>{item}</option>)}
                  </select>
                </label>
                <label className="field-label">
                  节奏偏好
                  <select className="field" value={form.work_rhythm} onChange={(e) => setForm({ ...form, work_rhythm: e.target.value })}>
                    {optionGroups.rhythm.map((item) => <option key={item}>{item}</option>)}
                  </select>
                </label>
              </div>
            </Section>

            <Section title="4. 时间与参与方式" description="时间和地点会影响实际协作体验，比单纯兴趣匹配更重要。">
              <label className="field-label">
                可参与时间
                <ChoiceChips values={optionGroups.time} selected={form.available_time} onToggle={(available_time) => setForm({ ...form, available_time })} />
              </label>
              <div className="grid gap-4 md:grid-cols-3">
                <label className="field-label">
                  参与方式
                  <select className="field" value={form.participation_mode} onChange={(e) => setForm({ ...form, participation_mode: e.target.value })}>
                    {["线下", "线上", "线上线下都可以"].map((item) => <option key={item}>{item}</option>)}
                  </select>
                </label>
                <label className="field-label">
                  合作周期
                  <select className="field" value={form.duration_preference} onChange={(e) => setForm({ ...form, duration_preference: e.target.value })}>
                    {["一次性活动", "短期组队", "长期合作"].map((item) => <option key={item}>{item}</option>)}
                  </select>
                </label>
                <label className="field-label">
                  校区或常用地点
                  <input className="field" value={form.campus_or_location} onChange={(e) => setForm({ ...form, campus_or_location: e.target.value })} placeholder="如：图书馆、线上会议、工程训练中心" />
                </label>
              </div>
            </Section>

            <Section title="5. 目标与期待" description="目标越清楚，推荐结果越接近你的真实需求。">
              <label className="field-label">
                主要目标
                <select className="field" value={form.main_goal} onChange={(e) => setForm({ ...form, main_goal: e.target.value })}>
                  {optionGroups.goals.map((item) => <option key={item}>{item}</option>)}
                </select>
              </label>
              <label className="field-label">
                希望遇到的伙伴
                <ChoiceChips values={optionGroups.expectations} selected={form.partner_expectation} onToggle={(partner_expectation) => setForm({ ...form, partner_expectation })} />
              </label>
              <label className="field-label">
                最不能接受的问题
                <ChoiceChips values={optionGroups.avoid} selected={form.avoid_points} onToggle={(avoid_points) => setForm({ ...form, avoid_points })} />
              </label>
            </Section>

            <Section title="6. 活动偏好" description="MatchLab 不只推荐比赛，也会推荐适合你的项目、学习和校园活动。">
              <label className="field-label">
                感兴趣的活动类型
                <ActivityTypeChips selected={form.preferred_activity_types} onToggle={(preferred_activity_types) => setForm({ ...form, preferred_activity_types })} />
                <select className="field" value="" onChange={(e) => {
                  const value = e.target.value;
                  if (value) setForm({ ...form, preferred_activity_types: toggleListValue(form.preferred_activity_types, value) });
                }}>
                  <option value="">添加活动类型</option>
                  {activityTypeOptions.map((item) => <option value={item.value} key={item.value}>{item.label}</option>)}
                </select>
              </label>
              <label className="field-label">
                参与目的
                <ChoiceChips values={optionGroups.purpose} selected={form.participation_purpose} onToggle={(participation_purpose) => setForm({ ...form, participation_purpose })} />
              </label>
            </Section>

            {error && <p className="notice-error">{error}</p>}
            <button className="button-primary" disabled={pending}>
              {pending ? "生成中…" : "生成我的协作画像"}
            </button>
          </form>
        </div>

        <aside className="space-y-5 lg:sticky lg:top-24 lg:self-start">
          <div className="card">
            <p className="eyebrow">Profile card</p>
            <h2 className="mt-3 text-2xl font-black text-slate-950">协作画像卡</h2>
            <p className="mt-4 leading-7 text-slate-650">
              {profile?.summary || "提交后会生成你的协作画像摘要，展示标签、维度评分、推荐方向和推荐依据。"}
            </p>
            <div className="mt-5">
              <TagList tags={profileTags} />
            </div>
            <div className="mt-6 grid grid-cols-2 gap-3">
              <StatCard label="兴趣广度" value={profile?.scores.interest_score ?? 0} />
              <StatCard label="技能匹配" value={profile?.scores.skill_score ?? 0} accent="cyan" />
              <StatCard label="协作稳定" value={profile?.scores.communication_score ?? 0} />
              <StatCard label="活动活跃" value={profile?.scores.time_score ?? 0} accent="orange" />
              <StatCard label="目标清晰" value={profile?.scores.goal_score ?? 0} accent="cyan" />
            </div>
            <div className="mt-6 rounded-2xl bg-slate-50 p-4 text-sm leading-7 text-slate-650">
              <p className="font-black text-slate-900">推荐方向</p>
              <p className="mt-2">优先关注项目合作、比赛组队、学习搭子和与你时间地点相近的校园活动。</p>
              <p className="mt-3 font-black text-slate-900">为什么这样推荐</p>
              <p className="mt-2">系统会把兴趣、技能、目标、时间和协作风格一起纳入匹配，而不是只看你会什么。</p>
            </div>
            {profile && (
              <Link className="button-secondary mt-5 w-full" href="/match">
                查看适合我的活动
              </Link>
            )}
          </div>

          <div className="card">
            <h2 className="text-xl font-black text-slate-950">为什么使用 MatchLab？</h2>
            <ul className="mt-4 grid gap-3 text-sm leading-6 text-slate-650">
              <li>不只是找比赛队友，也能发现项目、学习、社团和兴趣活动。</li>
              <li>推荐不仅看技能，也结合兴趣、时间、目标和协作风格。</li>
              <li>帮你更快找到合适的人，而不只是找到“会的人”。</li>
              <li>支持发布活动、报名互动、画像推荐和管理员查看。</li>
            </ul>
          </div>
        </aside>
      </div>
    </section>
  );
}
