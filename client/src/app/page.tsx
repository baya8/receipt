"use client";

import { useEffect, useState } from "react";
import { ChevronRight } from "lucide-react";
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
}

export default function Home() {
  const [receipts, setReceipts] = useState<Receipt[]>([]);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) {
      router.push("/login");
      return;
    }

    async function fetchReceipts() {
      try {
        // 現在はgroup_id=1を固定で使用
        const data = await apiRequest("/api/receipts?group_id=1");
        setReceipts(data);
      } catch (err) {
        console.error("Failed to fetch receipts:", err);
      } finally {
        setLoading(false);
      }
    }
    fetchReceipts();
  }, [router]);

  const totalAmount = receipts.reduce((sum, r) => sum + r.amount, 0);

  if (loading) return <div className="p-8 text-center text-gray-400">読み込み中...</div>;

  return (
    <div>
      {/* Header */}
      <header className="sticky top-0 bg-white border-b border-gray-100 p-4 flex justify-between items-center z-10">
        <h1 className="text-xl font-bold text-gray-800">レシート一覧</h1>
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
