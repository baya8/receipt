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

interface Settlement {
  id: number;
  amount: number;
  settled_by: number;
  created_at: string;
  settled_by_user: {
    nickname: string;
  };
}

interface SummaryData {
  total_spent: number;
  members: MemberSummary[];
  settlements: Settlement[];
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
  const [settleAmount, setSettleAmount] = useState<number>(0);
  const [settling, setSettling] = useState(false);
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
          
          // 残高を計算してデフォルトの精算額にセット
          const user = JSON.parse(localStorage.getItem("user") || "{}");
          const mySum = data.members.find((m: any) => m.user_id === user.id);
          const initialBalance = mySum ? Math.max(0, mySum.share - mySum.paid) : 0;
          const totalSettledByMe = data.settlements
            .filter((s: any) => s.settled_by === user.id)
            .reduce((sum: number, s: any) => sum + s.amount, 0);
          setSettleAmount(Math.max(0, initialBalance - totalSettledByMe));
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
    if (groups.length === 0 || settleAmount <= 0) return;
    if (!confirm(`${year}年${month}月の精算として ¥${settleAmount.toLocaleString()} を記録しますか？`)) return;

    setSettling(true);
    try {
      await apiRequest("/api/settle", {
        method: "POST",
        body: JSON.stringify({
          group_id: groups[0].id,
          year,
          month,
          amount: settleAmount,
        }),
      });
      // リロード
      const data = await apiRequest(`/api/summary?group_id=${groups[0].id}&year=${year}&month=${month}`);
      setSummary(data);
      alert("精算を記録しました");
    } catch (err: any) {
      console.error("Failed to settle:", err);
      alert("精算に失敗しました: " + err.message);
    } finally {
      setSettling(false);
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

  // 初期バランス（レシートのみ）
  const initialBalance = mySummary ? mySummary.share - mySummary.paid : 0;
  
  // すでに精算された額の合計（自分が払った分）
  const totalSettledByMe = summary?.settlements
    .filter(s => s.settled_by === user.id)
    .reduce((sum, s) => sum + s.amount, 0) || 0;
    
  // 相手が精算した額の合計
  const totalSettledByOther = summary?.settlements
    .filter(s => s.settled_by !== user.id)
    .reduce((sum, s) => sum + s.amount, 0) || 0;

  // 現在の残高
  const currentBalance = initialBalance - totalSettledByMe + totalSettledByOther;
  
  const canSettle = currentBalance > 0;
  const maxSettleAmount = Math.max(0, currentBalance);

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
                  placeholder="0"
                />
                <button 
                  onClick={() => setSettleAmount(maxSettleAmount)}
                  className="absolute right-4 top-1/2 -translate-y-1/2 text-[10px] font-bold bg-blue-100 text-blue-600 px-2 py-1 rounded-lg"
                >
                  全額
                </button>
              </div>
              <button 
                onClick={handleSettle}
                disabled={settling || settleAmount <= 0}
                className="w-full py-4 bg-gray-900 text-white rounded-2xl font-bold shadow-lg active:scale-[0.98] transition-all disabled:opacity-50"
              >
                {settling ? "処理中..." : `¥${settleAmount.toLocaleString()} を精算済みにする`}
              </button>
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
