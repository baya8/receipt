"use client";

import { useEffect, useState, Suspense } from "react";
import { ChevronRight, ChevronLeft } from "lucide-react";
import Link from "next/link";
import { apiRequest } from "@/lib/api";
import { handleApiError } from "@/lib/errors";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuthGuard } from "@/hooks/useAuthGuard";
import { useGroups } from "@/hooks/useGroups";
import LoadingScreen from "@/components/ui/LoadingScreen";
import EmptyState from "@/components/ui/EmptyState";
import type { Receipt } from "@/types";

function HomeContent() {
  useAuthGuard();
  const { primaryGroup, loading: groupsLoading } = useGroups();
  const [receipts, setReceipts] = useState<Receipt[]>([]);
  const [receiptsLoading, setReceiptsLoading] = useState(true);
  const router = useRouter();
  const searchParams = useSearchParams();

  // URLパラメータから年月を取得、なければ現在の年月を使用
  const queryYear = searchParams.get("year");
  const queryMonth = searchParams.get("month");

  const now = new Date();
  const year = queryYear ? parseInt(queryYear) : now.getFullYear();
  const month = queryMonth ? parseInt(queryMonth) : now.getMonth() + 1;

  useEffect(() => {
    async function fetchReceipts() {
      if (groupsLoading) return;
      if (!primaryGroup) {
        setReceiptsLoading(false);
        return;
      }

      setReceiptsLoading(true);
      try {
        const data = await apiRequest(`/api/receipts?group_id=${primaryGroup.id}&year=${year}&month=${month}`);
        setReceipts(data);
      } catch (err) {
        handleApiError(err, "レシートの取得に失敗しました");
      } finally {
        setReceiptsLoading(false);
      }
    }
    fetchReceipts();
  }, [groupsLoading, primaryGroup, year, month]);

  const loading = groupsLoading || receiptsLoading;

  const changeMonth = (offset: number) => {
    const date = new Date(year, month - 1);
    date.setMonth(date.getMonth() + offset);
    const newYear = date.getFullYear();
    const newMonth = date.getMonth() + 1;
    
    // URLを更新
    router.push(`/?year=${newYear}&month=${newMonth}`);
  };

  const totalAmount = receipts.reduce((sum, r) => sum + r.amount, 0);

  const getPayerColor = (userId: string) => {
    const colors = [
      "bg-blue-50 text-blue-600 border-blue-100",
      "bg-purple-50 text-purple-600 border-purple-100",
      "bg-pink-50 text-pink-600 border-pink-100",
      "bg-indigo-50 text-indigo-600 border-indigo-100",
      "bg-cyan-50 text-cyan-600 border-cyan-100",
    ];
    let hash = 0;
    for (let i = 0; i < userId.length; i++) {
      hash = userId.charCodeAt(i) + ((hash << 5) - hash);
    }
    return colors[Math.abs(hash) % colors.length];
  };

  const getPaymentMethodLabel = (method: string) => {
    switch (method) {
      case "half": return "折半";
      case "self": return "自分が10割負担";
      case "other": return "全額相手負担";
      default: return method;
    }
  };

  if (loading) return <LoadingScreen />;

  if (!primaryGroup) {
    return (
      <EmptyState description="レシートを記録するには、まず設定画面からグループを作成するか、招待を受けてください。" />
    );
  }

  return (
    <div>
      {/* Header */}
      <header className="sticky top-0 bg-white border-b border-gray-100 p-4 flex justify-between items-center z-10 shadow-sm">
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

      {/* Summary Bar */}
      <div className="px-4 py-2 bg-blue-50/50 border-b border-blue-50 flex justify-between items-center">
        <span className="text-xs font-semibold text-blue-600">支出合計</span>
        <span className="text-sm font-black text-blue-700">¥{totalAmount.toLocaleString()}</span>
      </div>

      {/* List */}
      <div className="divide-y divide-gray-100">
        {receipts.length === 0 ? (
          <div className="p-10 text-center text-gray-400">レシートがありません</div>
        ) : (
          receipts.map((receipt) => (
            <Link
              key={receipt.id}
              href={`/receipt/${receipt.id}`}
              className="p-4 flex items-center justify-between hover:bg-gray-50 active:bg-gray-100 cursor-pointer"
            >
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs text-gray-500 font-medium">{new Date(receipt.date).toLocaleDateString()}</span>
                  <span className={`text-[10px] px-2 py-0.5 rounded-full ${
                    receipt.payment_method === "half" ? "bg-orange-100 text-orange-600" :
                    receipt.payment_method === "other" ? "bg-green-100 text-green-600" :
                    "bg-gray-100 text-gray-600"
                  }`}>
                    {getPaymentMethodLabel(receipt.payment_method)}
                  </span>
                  {receipt.payer && (
                    <span className={`text-[10px] px-2 py-0.5 rounded-full font-bold border ${getPayerColor(receipt.payer_id)}`}>
                      {receipt.payer.nickname}
                    </span>
                  )}
                </div>
                <h2 className="text-base font-semibold text-gray-900">{receipt.shop || "店名なし"}</h2>
                <p className="text-sm text-gray-500">{receipt.item}</p>
              </div>
              <div className="text-right flex items-center gap-3">
                <div>
                  <span className="text-lg font-bold text-gray-900">¥{receipt.amount.toLocaleString()}</span>
                </div>
                <ChevronRight size={18} className="text-gray-300" />
              </div>
            </Link>
          ))
        )}
      </div>
    </div>
  );
}

export default function Home() {
  return (
    <Suspense fallback={<LoadingScreen />}>
      <HomeContent />
    </Suspense>
  );
}
