"use client";

import Image from "next/image";
import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiError, postJSON, setToken } from "@/lib/api";
import type { User } from "@/lib/types";

export default function LoginPage() {
  const router = useRouter();
  const [mode, setMode] = useState<"login" | "register">("login");
  const [form, setForm] = useState({ email: "", password: "", nickname: "", school: "" });
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault(); setPending(true); setError("");
    try {
      if (mode === "register") {
        await postJSON<{ user: User }>("/auth/register", form);
      }
      const result = await postJSON<{ token: string; user: User }>("/auth/login", { email: form.email, password: form.password });
      setToken(result.token);
      router.push("/dashboard");
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : "操作失败，请稍后重试");
    } finally { setPending(false); }
  }

  return (
    <section className="page-shell page-section grid items-stretch gap-8 lg:grid-cols-[1.05fr_.95fr]">
      <div className="relative hidden min-h-[620px] overflow-hidden rounded-[2rem] lg:block">
        <Image src="/images/matchlab-hero.png" alt="校园项目协作插画" fill className="object-cover object-[72%_center]" sizes="50vw" />
        <div className="absolute inset-0 bg-gradient-to-t from-indigo-950/80 via-transparent to-transparent" />
        <p className="absolute bottom-8 left-8 max-w-md text-2xl font-black text-white">让每次校园协作，都从清晰的目标开始。</p>
      </div>
      <div className="card flex flex-col justify-center px-6 py-10 sm:px-12">
        <p className="eyebrow">Welcome to MatchLab</p><h1 className="mt-3 text-3xl font-black text-slate-950">{mode === "login" ? "欢迎回来" : "创建校园协作账号"}</h1>
        <div className="mt-7 grid grid-cols-2 rounded-xl bg-slate-100 p-1">
          <button className={`rounded-lg py-2 font-bold ${mode === "login" ? "bg-white text-indigo-700 shadow" : "text-slate-600"}`} onClick={() => setMode("login")}>登录</button>
          <button className={`rounded-lg py-2 font-bold ${mode === "register" ? "bg-white text-indigo-700 shadow" : "text-slate-600"}`} onClick={() => setMode("register")}>注册账号</button>
        </div>
        <form className="mt-7 grid gap-4" onSubmit={submit}>
          {mode === "register" && <><label className="field-label">昵称<input className="field" required value={form.nickname} onChange={(e) => setForm({ ...form, nickname: e.target.value })} /></label><label className="field-label">学校<input className="field" required value={form.school} onChange={(e) => setForm({ ...form, school: e.target.value })} /></label></>}
          <label className="field-label">邮箱<input className="field" type="email" required value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} /></label>
          <label className="field-label">密码<input className="field" type="password" minLength={8} required value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} /></label>
          {error && <p className="notice-error">{error}</p>}
          <button className="button-primary mt-2" disabled={pending}>{pending ? "提交中…" : mode === "login" ? "登录" : "注册并登录"}</button>
        </form>
      </div>
    </section>
  );
}
