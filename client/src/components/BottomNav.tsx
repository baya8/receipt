"use client";

import Link from "next/link";
import { List, PieChart, PlusCircle, User } from "lucide-react";
import { useApi } from "@/lib/ApiContext";
import { usePathname } from "next/navigation";

export default function BottomNav() {
  const { isAuthenticated } = useApi();
  const pathname = usePathname();

  // ログイン、新規登録画面では表示しない
  const hidePaths = ["/login", "/signup"];
  if (!isAuthenticated || hidePaths.includes(pathname)) {
    return null;
  }

  const navItems = [
    { href: "/", icon: List, label: "一覧" },
    { href: "/summary", icon: PieChart, label: "精算" },
    { href: "/register", icon: PlusCircle, label: "登録" },
    { href: "/profile", icon: User, label: "設定" },
  ];

  return (
    <nav className="fixed bottom-0 left-0 right-0 max-w-md mx-auto bg-white border-t border-gray-200 flex justify-around py-3 z-10">
      {navItems.map((item) => {
        const Icon = item.icon;
        const isActive = pathname === item.href;
        return (
          <Link
            key={item.href}
            href={item.href}
            className={`flex flex-col items-center ${
              isActive ? "text-blue-600" : "text-gray-400"
            } hover:text-blue-500`}
          >
            <Icon size={24} />
            <span className="text-xs mt-1">{item.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}
