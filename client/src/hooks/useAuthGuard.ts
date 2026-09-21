"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

// 未ログイン時にログイン画面へリダイレクトする。各ページのuseEffectに直書きされていた重複処理を集約。
export function useAuthGuard() {
  const router = useRouter();

  useEffect(() => {
    if (!localStorage.getItem("token")) {
      router.push("/login");
    }
  }, [router]);
}
