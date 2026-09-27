"use client";

import { useMemo } from "react";
import type { User } from "@/types";

// localStorageの"user"を各ページで個別にJSON.parseしていた重複処理を集約。
export function useCurrentUser(): User | null {
  return useMemo(() => {
    if (typeof window === "undefined") return null;
    const userStr = localStorage.getItem("user");
    if (!userStr) return null;
    try {
      return JSON.parse(userStr) as User;
    } catch {
      return null;
    }
  }, []);
}
