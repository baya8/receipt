"use client";

import { useEffect, useState, use } from "react";
import { ArrowLeft, Trash2, Save } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { apiRequest } from "@/lib/api";

interface Receipt {
  id: number;
  date: string;
  shop: string;
  item: string;
  amount: number;
  payer_id: number;
  payment_method: string;
  group_id: number;
}

export default function ReceiptDetail({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();
  const [receipt, setReceipt] = useState<Receipt | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) {
      router.push("/login");
      return;
    }

    async function fetchReceipt() {
      try {
        const data = await apiRequest(`/api/receipts/${id}`);
        setReceipt(data);
      } catch (err) {
        console.error("Failed to fetch receipt:", err);
        // ConnectionErrorはApiProviderがハンドルするため、ここではそれ以外のエラーのみアラートを出す
        if (!(err instanceof Error && err.name === "ConnectionError")) {
          alert("データの取得に失敗しました");
        }
      } finally {
        setLoading(false);
      }
    }
    fetchReceipt();
  }, [id, router]);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!receipt) return;
    
    setSaving(true);
    try {
      await apiRequest(`/api/receipts/${id}`, {
        method: "PUT",
        body: JSON.stringify({
          ...receipt,
          amount: Number(receipt.amount),
          date: new Date(receipt.date).toISOString(),
        }),
      });
      router.push("/");
    } catch (err) {
      console.error("Failed to update receipt:", err);
      if (!(err instanceof Error && err.name === "ConnectionError")) {
        alert("更新に失敗しました");
      }
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm("本当に削除しますか？")) return;
    
    try {
      await apiRequest(`/api/receipts/${id}`, {
        method: "DELETE",
      });
      router.push("/");
    } catch (err) {
      console.error("Failed to delete receipt:", err);
      if (!(err instanceof Error && err.name === "ConnectionError")) {
        alert("削除に失敗しました");
      }
    }
  };

  if (loading) return <div className="p-8 text-center text-gray-400">読み込み中...</div>;
  if (!receipt) return <div className="p-8 text-center text-red-500">データが見つかりませんでした</div>;

  return (
    <div className="pb-10 bg-white">
      {/* Header */}
      <header className="p-4 border-b border-gray-100 flex justify-between items-center bg-white sticky top-0 z-10">
        <Link href="/" className="text-gray-500 hover:text-gray-800">
          <ArrowLeft size={24} />
        </Link>
        <h1 className="text-lg font-bold text-gray-800">レシート詳細</h1>
        <button 
          onClick={handleDelete}
          className="text-red-500 p-2 active:bg-red-50 rounded-full transition-colors"
        >
          <Trash2 size={20} />
        </button>
      </header>

      <div className="p-6 space-y-6">
        <div className="bg-blue-50 p-4 rounded-2xl border border-blue-100 mb-4">
          <p className="text-xs text-blue-500 font-medium mb-1">精算ステータス</p>
          <p className="text-sm text-blue-800 font-semibold">
            {receipt.payment_method}で精算予定
          </p>
        </div>

        {/* Input Form (Editable) */}
        <form onSubmit={handleSave} className="space-y-4">
          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">購入日</label>
            <input 
              type="date" 
              className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all text-gray-900" 
              value={receipt.date.split('T')[0]} 
              onChange={(e) => setReceipt({...receipt, date: e.target.value})}
              required
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">お店</label>
            <input 
              type="text" 
              className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all text-gray-900" 
              value={receipt.shop} 
              onChange={(e) => setReceipt({...receipt, shop: e.target.value})}
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">品名</label>
            <input 
              type="text" 
              className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all text-gray-900" 
              value={receipt.item} 
              onChange={(e) => setReceipt({...receipt, item: e.target.value})}
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">金額</label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 font-bold">¥</span>
              <input 
                type="number" 
                className="w-full p-3 pl-8 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all font-bold text-lg text-gray-900" 
                value={receipt.amount} 
                onChange={(e) => setReceipt({...receipt, amount: Number(e.target.value)})}
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-semibold text-gray-800">精算方法</label>
              <select 
                className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 appearance-none text-gray-900" 
                value={receipt.payment_method}
                onChange={(e) => setReceipt({...receipt, payment_method: e.target.value})}
              >
                <option>折半</option>
                <option>自分が10割負担</option>
                <option>全額相手負担</option>
              </select>
            </div>
          </div>

          <div className="pt-4">
            <button 
              type="submit" 
              disabled={saving}
              className="w-full py-4 bg-gray-900 text-white rounded-2xl font-bold flex items-center justify-center gap-2 shadow-lg active:scale-[0.98] transition-all disabled:opacity-50"
            >
              <Save size={20} />
              {saving ? "保存中..." : "変更を保存する"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
