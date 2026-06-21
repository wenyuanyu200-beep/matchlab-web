import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "MatchLab｜校园活动与项目协作平台",
  description: "发现活动搭子与项目队友，获得智能组队推荐。",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN" className="h-full antialiased">
      <body className="flex min-h-full flex-col">{children}</body>
    </html>
  );
}
