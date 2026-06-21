"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useSyncExternalStore } from "react";
import { clearToken, getToken, subscribeAuth } from "@/lib/api";

export default function Navbar() {
  const router = useRouter();
  const loggedIn = useSyncExternalStore(subscribeAuth, () => Boolean(getToken()), () => false);

  function logout() {
    clearToken();
    router.push("/");
  }

  return (
    <header className="sticky top-0 z-50 border-b border-slate-200/80 bg-white/90 backdrop-blur-xl">
      <nav className="page-shell flex min-h-16 flex-wrap items-center justify-between gap-3 py-3" aria-label="主导航">
        <Link href="/" className="flex items-center gap-2 text-lg font-black tracking-tight text-indigo-950">
          <span className="grid size-8 place-items-center rounded-xl bg-indigo-600 text-sm text-white shadow-lg shadow-indigo-200">M</span>
          MatchLab
        </Link>
        <div className="flex flex-wrap items-center justify-end gap-1 text-sm font-semibold text-slate-700">
          <Link className="nav-link" href="/activities">活动广场</Link>
          <Link className="nav-link" href="/match">智能推荐</Link>
          {loggedIn ? (
            <>
              <Link className="nav-link" href="/dashboard">工作台</Link>
              <button className="nav-link cursor-pointer" onClick={logout}>退出</button>
            </>
          ) : (
            <Link className="button-primary !px-4 !py-2" href="/login">登录</Link>
          )}
        </div>
      </nav>
    </header>
  );
}
