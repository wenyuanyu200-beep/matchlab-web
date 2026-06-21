import Image from "next/image";
import Link from "next/link";

const steps = [
  ["01", "填写画像", "用兴趣、技能与时间偏好建立协作画像。"],
  ["02", "浏览活动", "发现比赛、项目与校园活动。"],
  ["03", "报名组队", "获取智能推荐，加入合适的协作团队。"],
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
            <p className="mt-7 max-w-lg text-lg leading-8 text-slate-700">从个人画像出发，发现值得参与的校园活动，用智能组队推荐找到更合适的协作机会。</p>
            <div className="mt-8 flex flex-wrap gap-3">
              <Link className="button-primary" href="/activities">进入活动广场</Link>
              <Link className="button-secondary" href="/match">智能推荐</Link>
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
    </>
  );
}
