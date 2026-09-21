"use client";

import { useEffect, useState } from "react";
import { ChevronLeft, ChevronRight, CheckCircle2, Circle } from "lucide-react";
import { apiRequest } from "@/lib/api";
import { toast } from "sonner";
import { useAuthGuard } from "@/hooks/useAuthGuard";
import { useCurrentUser } from "@/hooks/useCurrentUser";
import { useGroups } from "@/hooks/useGroups";
import LoadingScreen from "@/components/ui/LoadingScreen";
import EmptyState from "@/components/ui/EmptyState";
import Button from "@/components/ui/Button";
import type { MonthlySummary, MemberSummary, Settlement } from "@/types";

export default function Summary() {
  useAuthGuard();
  const currentUser = useCurrentUser();
  const { primaryGroup, loading: groupsLoading } = useGroups();
  const [date, setDate] = useState(new Date());
  const [summary, setSummary] = useState<MonthlySummary | null>(null);
  const [summaryLoading, setSummaryLoading] = useState(true);
  const [settleAmount, setSettleAmount] = useState<number>(0);
  const [settling, setSettling] = useState(false);

  const year = date.getFullYear();
  const month = date.getMonth() + 1;

  useEffect(() => {
    async function fetchSummary() {
      if (groupsLoading) return;
      if (!primaryGroup) {
        setSummaryLoading(false);
        return;
      }

      setSummaryLoading(true);
      try {
        const data = await apiRequest(`/api/summary?group_id=${primaryGroup.id}&year=${year}&month=${month}`);
        setSummary(data);

        // 残高を計算してデフォルトの精算額にセット
        const mySum = data.members.find((m: MemberSummary) => m.user_id === currentUser?.id);
        const initialBalance = mySum ? Math.max(0, mySum.share - mySum.paid) : 0;
        const totalSettledByMe = data.settlements
          .filter((s: Settlement) => s.settled_by === currentUser?.id)
          .reduce((sum: number, s: Settlement) => sum + s.amount, 0);
        setSettleAmount(Math.max(0, initialBalance - totalSettledByMe));
      } catch (err) {
        console.error("Failed to fetch summary:", err);
      } finally {
        setSummaryLoading(false);
      }
    }
    fetchSummary();
  }, [groupsLoading, primaryGroup, year, month, currentUser]);

  const loading = groupsLoading || summaryLoading;

  const changeMonth = (offset: number) => {
    const newDate = new Date(date);
    newDate.setMonth(newDate.getMonth() + offset);
    setDate(newDate);
  };

  const handleSettle = async () => {
    if (!primaryGroup || settleAmount <= 0) {
      toast.error("精算金額を入力してください");
      return;
    }

    if (settleAmount > maxSettleAmount) {
      toast.error(`精算金額は ¥${maxSettleAmount.toLocaleString()} 以下にしてください`);
      return;
    }

    if (!confirm(`${year}年${month}月の精算として ¥${settleAmount.toLocaleString()} を記録しますか？`)) return;

    setSettling(true);
    try {
      await apiRequest("/api/settle", {
        method: "POST",
        body: JSON.stringify({
          group_id: primaryGroup.id,
          year,
          month,
          amount: settleAmount,
        }),
      });
      // リロード
      const data = await apiRequest(`/api/summary?group_id=${primaryGroup.id}&year=${year}&month=${month}`);
      setSummary(data);
      toast.success("精算を記録しました");
    } catch (err: any) {
      console.error("Failed to settle:", err);
      toast.error("精算に失敗しました: " + err.message);
    } finally {
      setSettling(false);
    }
  };

  if (loading) return <LoadingScreen />;

  if (!primaryGroup) {
    return <EmptyState description="精算機能を利用するには、まずグループを作成してください。" />;
  }

  const mySummary = summary?.members.find(m => m.user_id === currentUser?.id);
  const otherSummary = summary?.members.find(m => m.user_id !== currentUser?.id);

  // 初期バランス（レシートのみ）
  const initialBalance = mySummary ? mySummary.share - mySummary.paid : 0;

  // すでに精算された額の合計（自分が払った分）
  const totalSettledByMe = summary?.settlements
    .filter(s => s.settled_by === currentUser?.id)
    .reduce((sum, s) => sum + s.amount, 0) || 0;

  // 相手が精算した額の合計
  const totalSettledByOther = summary?.settlements
    .filter(s => s.settled_by !== currentUser?.id)
    .reduce((sum, s) => sum + s.amount, 0) || 0;

  // 現在の残高
  const currentBalance = initialBalance - totalSettledByMe + totalSettledByOther;
  
  const canSettle = currentBalance > 0;
  const maxSettleAmount = Math.max(0, currentBalance);

  return (
    <div className="pb-10">
      <header className="p-4 border-b border-gray-100 bg-white sticky top-0 z-10 flex justify-between items-center">
        <button onClick={() => changeMonth(-1)} className="p-2 hover:bg-gray-100 active:bg-gray-200 rounded-full transition-colors">
          <ChevronLeft size={24} className="text-gray-900" strokeWidth={2.5} />
        </button>
        <div className="text-center">
          <h1 className="text-lg font-bold text-gray-800">{year}年{month}月</h1>
          <p className="text-[10px] text-gray-400 font-bold uppercase tracking-wider">{primaryGroup.name}</p>
        </div>
        <button onClick={() => changeMonth(1)} className="p-2 hover:bg-gray-100 active:bg-gray-200 rounded-full transition-colors">
          <ChevronRight size={24} className="text-gray-900" strokeWidth={2.5} />
        </button>
      </header>

      <div className="p-6 space-y-6">
        {/* Settlement Card */}
        <section className={`rounded-3xl p-6 text-white shadow-xl shadow-blue-100 ${
          currentBalance > 0 ? "bg-gradient-to-br from-blue-600 to-indigo-700" : 
          currentBalance < 0 ? "bg-gradient-to-br from-emerald-500 to-teal-600" :
          "bg-gradient-to-br from-gray-500 to-gray-600"
        }`}>
          <p className="text-white/80 text-sm font-medium mb-1">
            {currentBalance > 0 ? "現在の未精算額" : 
             currentBalance < 0 ? "精算でもらえる額" : "精算完了"}
          </p>
          <div className="flex items-end gap-2 mb-4">
            <span className="text-4xl font-black">¥{Math.abs(currentBalance).toLocaleString()}</span>
            <span className="text-white/80 text-sm mb-1 pb-1">
              {currentBalance > 0 ? `${otherSummary?.nickname || "相手"}へ` : 
               currentBalance < 0 ? "もらえる予定" : ""}
            </span>
          </div>

          <div className="grid grid-cols-2 gap-4 pt-4 border-t border-white/20">
            <div>
              <p className="text-white/70 text-[10px] uppercase tracking-wider font-bold">初期バランス</p>
              <p className="text-base font-bold">¥{initialBalance.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-white/70 text-[10px] uppercase tracking-wider font-bold">精算済み合計</p>
              <p className="text-base font-bold">¥{(totalSettledByMe - totalSettledByOther).toLocaleString()}</p>
            </div>
          </div>
        </section>

        {/* Settlement Action */}
        <section className="bg-white border border-gray-100 rounded-3xl p-4 shadow-sm space-y-4">
          <div className="flex items-center gap-3 mb-2">
            <div className={`p-2 rounded-xl ${canSettle ? 'bg-blue-50 text-blue-600' : 'bg-gray-50 text-gray-400'}`}>
              <CheckCircle2 size={20} />
            </div>
            <div>
              <p className="text-sm font-bold text-gray-800">
                {canSettle ? "精算を記録する" : 
                 currentBalance < 0 ? "相手からの精算待ち" : "精算の必要はありません"}
              </p>
              <p className="text-[10px] text-gray-400">実際に支払った後に金額を入力して記録します</p>
            </div>
          </div>

          {canSettle && (
            <div className="space-y-3">
              <div className="relative">
                <span className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 font-bold">¥</span>
                <input 
                  type="number"
                  className="w-full p-4 pl-10 bg-gray-50 border border-gray-100 rounded-2xl focus:outline-none focus:ring-2 focus:ring-blue-500 font-black text-xl text-gray-900"
                  value={settleAmount || ""}
                  onChange={(e) => {
                    const val = Number(e.target.value);
                    if (val <= maxSettleAmount) setSettleAmount(val);
                  }}
                  max={maxSettleAmount}
                  min="1"
                  placeholder="0"
                />
                <button 
                  onClick={() => setSettleAmount(maxSettleAmount)}
                  className="absolute right-4 top-1/2 -translate-y-1/2 text-[10px] font-bold bg-blue-100 text-blue-600 px-2 py-1 rounded-lg"
                >
                  全額
                </button>
              </div>
              <Button variant="dark" onClick={handleSettle} disabled={settling || settleAmount <= 0}>
                {settling ? "処理中..." : `¥${settleAmount.toLocaleString()} を精算済みにする`}
              </Button>
            </div>
          )}
        </section>

        {/* Settlement History */}
        {summary && summary.settlements.length > 0 && (
          <section className="space-y-3">
            <h2 className="text-sm font-bold text-gray-400 uppercase tracking-widest ml-1 text-center">精算履歴</h2>
            <div className="bg-gray-50 rounded-2xl overflow-hidden divide-y divide-gray-100">
              {summary.settlements.map((s) => (
                <div key={s.id} className="p-3 flex justify-between items-center bg-white/50">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-blue-50 flex items-center justify-center text-blue-600 font-bold text-xs">
                      {s.settled_by_user.nickname[0]}
                    </div>
                    <div>
                      <p className="text-xs font-bold text-gray-800">{s.settled_by_user.nickname}が精算</p>
                      <p className="text-[10px] text-gray-400">{new Date(s.created_at).toLocaleString()}</p>
                    </div>
                  </div>
                  <p className="font-bold text-gray-700">¥{s.amount.toLocaleString()}</p>
                </div>
              ))}
            </div>
          </section>
        )}        {/* Member Details */}
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
