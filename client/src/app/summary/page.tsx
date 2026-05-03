"use client";

import { useEffect, useState } from "react";
import { ChevronLeft, ChevronRight, CheckCircle2, Circle, PlusCircle } from "lucide-react";
import { apiRequest } from "@/lib/api";
import { useRouter } from "next/navigation";
import Link from "next/link";

interface MemberSummary {
  user_id: number;
  nickname: string;
  paid: number;
  share: number;
}

interface SummaryData {
  total_spent: number;
  members: MemberSummary[];
  is_settled: boolean;
  settled_at: string | null;
}

interface Group {
  id: number;
  name: string;
}

export default function Summary() {
  const [date, setDate] = useState(new Date());
  const [summary, setSummary] = useState<SummaryData | null>(null);
  const [loading, setLoading] = useState(true);
  const [groups, setGroups] = useState<Group[]>([]);
  const router = useRouter();

  const year = date.getFullYear();
  const month = date.getMonth() + 1;

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) {
      router.push("/login");
      return;
    }

    async function fetchData() {
      setLoading(true);
      try {
        const myGroups = await apiRequest("/api/groups");
        setGroups(myGroups);

        if (myGroups.length > 0) {
          const data = await apiRequest(`/api/summary?group_id=${myGroups[0].id}&year=${year}&month=${month}`);
          setSummary(data);
        }
      } catch (err) {
        console.error("Failed to fetch summary:", err);
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, [year, month, router]);

  const changeMonth = (offset: number) => {
    const newDate = new Date(date);
    newDate.setMonth(newDate.getMonth() + offset);
    setDate(newDate);
  };

  const handleSettle = async () => {
    if (summary?.is_settled || groups.length === 0) return;
    if (!confirm(`${year}年${month}月の精算を完了としてマークしますか？`)) return;

    try {
      await apiRequest("/api/settle", {
        method: "POST",
        body: JSON.stringify({
          group_id: groups[0].id,
          year,
          month,
        }),
      });
      // リロード
      const data = await apiRequest(`/api/summary?group_id=${groups[0].id}&year=${year}&month=${month}`);
      setSummary(data);
    } catch (err) {
      console.error("Failed to settle:", err);
      if (!(err instanceof Error && err.name === "ConnectionError")) {
        alert("精算に失敗しました。");
      }
    }
  };

  if (loading) return <div className="p-8 text-center text-gray-400">読み込み中...</div>;

  if (groups.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[80vh] p-8 text-center">
        <div className="w-20 h-20 bg-blue-50 rounded-full flex items-center justify-center mb-6">
          <PlusCircle size={40} className="text-blue-500" />
        </div>
        <h2 className="text-xl font-bold text-gray-900 mb-2">グループがありません</h2>
        <p className="text-gray-500 mb-8">
          精算機能を利用するには、まずグループを作成してください。
        </p>
        <Link 
          href="/profile"
          className="bg-blue-600 text-white px-6 py-3 rounded-xl font-bold shadow-lg shadow-blue-100"
        >
          設定画面へ
        </Link>
      </div>
    );
  }

  const user = typeof window !== "undefined" ? JSON.parse(localStorage.getItem("user") || "{}") : {};
  const mySummary = summary?.members.find(m => m.user_id === user.id);
  const otherSummary = summary?.members.find(m => m.user_id !== user.id);

  const balance = mySummary ? mySummary.share - mySummary.paid : 0;

  return (
    <div className="pb-10">
      <header className="p-4 border-b border-gray-100 bg-white sticky top-0 z-10 flex justify-between items-center">
        <button onClick={() => changeMonth(-1)} className="p-2 hover:bg-gray-100 rounded-full">
          <ChevronLeft size={20} />
        </button>
        <div className="text-center">
          <h1 className="text-lg font-bold text-gray-800">{year}年{month}月</h1>
          <p className="text-[10px] text-gray-400 font-bold uppercase tracking-wider">{groups[0].name}</p>
        </div>
        <button onClick={() => changeMonth(1)} className="p-2 hover:bg-gray-100 rounded-full">
          <ChevronRight size={20} />
        </button>
      </header>

      <div className="p-6 space-y-6">
        {/* Settlement Card */}
        <section className={`rounded-3xl p-6 text-white shadow-xl shadow-blue-100 ${
          balance > 0 ? "bg-gradient-to-br from-blue-600 to-indigo-700" : "bg-gradient-to-br from-emerald-500 to-teal-600"
        }`}>
          <p className="text-blue-100 text-sm font-medium mb-1">
            {balance > 0 ? "精算が必要な金額" : "精算でもらえる金額"}
          </p>
          <div className="flex items-end gap-2 mb-4">
            <span className="text-4xl font-black">¥{Math.abs(balance).toLocaleString()}</span>
            <span className="text-blue-100 text-sm mb-1 pb-1">
              {balance > 0 ? `${otherSummary?.nickname || "相手"}へ` : "もらえる予定"}
            </span>
          </div>
          
          <div className="grid grid-cols-2 gap-4 pt-4 border-t border-white/20">
            <div>
              <p className="text-white/70 text-[10px] uppercase tracking-wider font-bold">支払った合計</p>
              <p className="text-lg font-bold">¥{mySummary?.paid.toLocaleString() || 0}</p>
            </div>
            <div>
              <p className="text-white/70 text-[10px] uppercase tracking-wider font-bold">負担する合計</p>
              <p className="text-lg font-bold">¥{mySummary?.share.toLocaleString() || 0}</p>
            </div>
          </div>
        </section>

        {/* Status Toggle */}
        <section>
          <button 
            onClick={handleSettle}
            className={`w-full p-4 rounded-2xl flex flex-col items-start gap-1 border-2 transition-all ${
              summary?.is_settled 
              ? "bg-green-50 border-green-200 text-green-700" 
              : "bg-white border-gray-100 text-gray-400 shadow-sm hover:border-blue-200 active:scale-95"
            }`}
          >
            <div className="flex items-center justify-between w-full">
              <div className="flex items-center gap-3">
                {summary?.is_settled ? <CheckCircle2 className="text-green-500" /> : <Circle />}
                <div className="text-left">
                  <p className="font-bold">{summary?.is_settled ? "精算済み" : "未精算（完了したらタップ）"}</p>
                  {summary?.is_settled && summary.settled_at && (
                    <p className="text-[10px] text-green-600 font-medium">
                      {new Date(summary.settled_at).toLocaleString()} に精算
                    </p>
                  )}
                </div>
              </div>
              {summary?.is_settled && (
                <span className="text-[10px] font-bold bg-green-200/50 px-2 py-0.5 rounded text-green-600">
                  DONE
                </span>
              )}
            </div>
          </button>
        </section>

        {/* Member Details */}
        <section className="space-y-4">
          <h2 className="text-sm font-bold text-gray-400 uppercase tracking-widest ml-1">メンバーごとの状況</h2>
          <div className="bg-gray-50 rounded-2xl divide-y divide-gray-100">
            {summary?.members.map((member) => (
              <div key={member.user_id} className="p-4 flex justify-between items-center">
                <div>
                  <p className="font-bold text-gray-800">{member.nickname}</p>
                  <p className="text-xs text-gray-500">支払額: ¥{member.paid.toLocaleString()}</p>
                </div>
                <div className="text-right">
                  <p className="text-xs text-gray-400">負担額</p>
                  <p className="font-semibold text-gray-700">¥{member.share.toLocaleString()}</p>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Monthly Info */}
        <section className="p-4 bg-orange-50 rounded-2xl border border-orange-100">
          <div className="flex justify-between items-center text-orange-800">
            <span className="text-sm font-medium">{month}月の総支出</span>
            <span className="text-lg font-bold">¥{summary?.total_spent.toLocaleString() || 0}</span>
          </div>
        </section>
      </div>
    </div>
  );
}
