"use client";

import { useEffect, useState } from "react";
import { ChevronRight, PlusCircle } from "lucide-react";
import Link from "next/link";
import { apiRequest } from "@/lib/api";
import { useRouter } from "next/navigation";

interface Receipt {
  id: number;
  date: string;
  shop: string;
  item: string;
  amount: number;
  payment_method: string;
  payer_id: number;
  payer?: {
    nickname: string;
  };
}

interface Group {
  id: number;
  name: string;
}

export default function Home() {
  const [receipts, setReceipts] = useState<Receipt[]>([]);
  const [loading, setLoading] = useState(true);
  const [groups, setGroups] = useState<Group[]>([]);
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) {
      router.push("/login");
      return;
    }

    async function fetchData() {
      try {
        // 1. 自分が所属するグループを取得
        const myGroups = await apiRequest("/api/groups");
        setGroups(myGroups);

        if (myGroups.length > 0) {
          // 2. 最初のグループのレシートを取得
          const data = await apiRequest(`/api/receipts?group_id=${myGroups[0].id}`);
          setReceipts(data);
        }
      } catch (err) {
        console.error("Failed to fetch data:", err);
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, [router]);

  const totalAmount = receipts.reduce((sum, r) => sum + r.amount, 0);

  if (loading) return <div className="p-8 text-center text-gray-400">読み込み中...</div>;

  if (groups.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[80vh] p-8 text-center">
        <div className="w-20 h-20 bg-blue-50 rounded-full flex items-center justify-center mb-6">
          <PlusCircle size={40} className="text-blue-500" />
        </div>
        <h2 className="text-xl font-bold text-gray-900 mb-2">グループがありません</h2>
        <p className="text-gray-500 mb-8">
          レシートを記録するには、まず設定画面からグループを作成するか、招待を受けてください。
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

  return (
    <div>
      {/* Header */}
      <header className="sticky top-0 bg-white border-b border-gray-100 p-4 flex justify-between items-center z-10">
        <div>
          <h1 className="text-xl font-bold text-gray-800">レシート一覧</h1>
          <p className="text-[10px] text-gray-400 font-bold uppercase tracking-wider">{groups[0].name}</p>
        </div>
        <div className="text-sm font-medium text-blue-600 bg-blue-50 px-3 py-1 rounded-full">
          合計: ¥{totalAmount.toLocaleString()}
        </div>
      </header>

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
                    receipt.payment_method === "折半" ? "bg-orange-100 text-orange-600" :
                    receipt.payment_method === "全額相手負担" ? "bg-green-100 text-green-600" :
                    "bg-gray-100 text-gray-600"
                  }`}>
                    {receipt.payment_method}
                  </span>
                  {receipt.payer && (
                    <span className="text-[10px] bg-blue-50 text-blue-600 px-2 py-0.5 rounded-full font-bold">
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
