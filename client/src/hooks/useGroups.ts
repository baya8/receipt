"use client";

import { useEffect, useState } from "react";
import { apiRequest } from "@/lib/api";
import { handleApiError } from "@/lib/errors";
import type { Group } from "@/types";

interface UseGroupsResult {
  groups: Group[];
  // 現状のUIは「所属グループは1つだけ」という前提で作られている。
  // その前提をここに集約しておくことで、複数グループ対応する際はこのフックだけを直せばよいようにする。
  primaryGroup: Group | null;
  loading: boolean;
}

export function useGroups(): UseGroupsResult {
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchGroups() {
      // 未ログイン時はuseAuthGuardがリダイレクトするので、ここでは無駄なAPI呼び出しをしない
      if (!localStorage.getItem("token")) {
        setLoading(false);
        return;
      }

      setLoading(true);
      try {
        const data = await apiRequest("/api/groups");
        setGroups(data);
      } catch (err) {
        handleApiError(err, "グループの取得に失敗しました");
      } finally {
        setLoading(false);
      }
    }
    fetchGroups();
  }, []);

  return { groups, primaryGroup: groups[0] ?? null, loading };
}
