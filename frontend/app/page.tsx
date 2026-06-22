import Image from "next/image";
import Link from "next/link";

const steps = [
  ["01", "填写画像", "用兴趣、技能与时间偏好建立协作画像。"],
  ["02", "浏览活动", "发现比赛、项目、学习、社团和兴趣活动。"],
  ["03", "报名参与", "获取推荐理由，加入更合适的协作团队。"],
];

const advantages = [
  "不只是找比赛队友，也能发现项目、学习、社团和兴趣活动",
  "推荐不仅看技能，也结合兴趣、时间、目标和协作风格",
  "帮你更快找到合适的人，而不只是找到“会的人”",
  "支持发布活动、报名互动、画像推荐和管理员查看，一站式完成校园协作流程",
];

export default function Home() {
  return (
    <>
      <section className="relative min-h-[680px] overflow-hidden bg-[#f7f8f2]">
        <Image src="/images/matchlab-hero.png" alt="两位大学生正在协作完成电子项目" fill priority className="object-cover object-[64%_center]" sizes="100vw" />
        <div className="absolute inset-0 bg-gradient-to-r from-[#faf9f1] via-[#faf9f1]/95 via-40% to-transparent" />
        <div className="page-shell relative flex min-h-[680px] items-center py-20">
          <div className="max-w-[590px]">
            <p className="eyebrow mb-5">校园活动与项目协作平台</p>
            <h1 className="text-[clamp(2.8rem,7vw,5.7rem)] font-black leading-[.98] tracking-[-.055em] text-indigo-950">找到活动搭子<br /><span className="text-indigo-600">遇见项目队友</span></h1>
            <p className="mt-7 max-w-lg text-lg leading-8 text-slate-700">MatchLab 帮助学生找到合适的活动搭子、项目队友、学习伙伴和校园协作对象。推荐不只看技能，也结合兴趣、目标、时间和协作风格。</p>
            <div className="mt-8 flex flex-wrap gap-3">
              <Link className="button-primary" href="/activities">进入活动广场</Link>
              <Link className="button-secondary" href="/questionnaire">填写协作画像</Link>
            </div>
          </div>
        </div>
      </section>
      <section className="page-shell page-section">
        <div className="mb-10 max-w-2xl"><p className="eyebrow">How it works</p><h2 className="mt-3 text-3xl font-black text-slate-950">三步开启校园协作</h2></div>
        <div className="grid gap-5 md:grid-cols-3">
          {steps.map(([number, title, text]) => <div className="card" key={number}><span className="text-sm font-black text-orange-500">{number}</span><h3 className="mt-8 text-xl font-black text-slate-950">{title}</h3><p className="mt-3 leading-7 text-slate-650">{text}</p></div>)}
        </div>
      </section>
      <section className="bg-white/70 py-16">
        <div className="page-shell">
          <div className="mb-8 max-w-2xl"><p className="eyebrow">Why MatchLab</p><h2 className="mt-3 text-3xl font-black text-slate-950">为什么使用 MatchLab？</h2></div>
          <div className="grid gap-4 md:grid-cols-2">
            {advantages.map((text) => <div className="rounded-2xl border border-slate-200 bg-white p-5 text-sm font-semibold leading-7 text-slate-700" key={text}>{text}</div>)}
          </div>
        </div>
      </section>
    </>
  );
}
