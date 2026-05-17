"use client";

import { useEffect, useState, use } from "react";
import { ArrowLeft, Trash2, Save } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { apiRequest } from "@/lib/api";
import { toast } from "sonner";

interface Receipt {
  id: string;
  user_id: string;
  date: string;
  settlement_year: number;
  settlement_month: number;
  shop: string;
  item: string;
  amount: number;
  payer_id: string;
  payment_method: string;
  group_id: string;
  settled_at: string | null;
  payer?: {
    nickname: string;
  };
}

export default function ReceiptDetail({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();
  const [receipt, setReceipt] = useState<Receipt | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);

  useEffect(() => {
    const token = localStorage.getItem("token");
    const userStr = localStorage.getItem("user");
    if (!token || !userStr) {
      router.push("/login");
      return;
    }

    try {
      const user = JSON.parse(userStr);
      setCurrentUserId(user.id);
    } catch (e) {
      console.error("Failed to parse user from localStorage", e);
    }

    async function fetchReceipt() {
      try {
        const data = await apiRequest(`/api/receipts/${id}`);
        setReceipt(data);
      } catch (err) {
        console.error("Failed to fetch receipt:", err);
        // ConnectionErrorはApiProviderがハンドルするため、ここではそれ以外のエラーのみ通知を出す
        if (!(err instanceof Error && err.name === "ConnectionError")) {
          toast.error("データの取得に失敗しました");
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

    if (receipt.amount <= 0) {
      toast.error("金額は1円以上にしてください");
      return;
    }
    
    setSaving(true);
    try {
      const sYear = receipt.settlement_year || new Date(receipt.date).getFullYear();
      const sMonth = receipt.settlement_month || (new Date(receipt.date).getMonth() + 1);

      await apiRequest(`/api/receipts/${id}`, {
        method: "PUT",
        body: JSON.stringify({
          ...receipt,
          amount: Number(receipt.amount),
          date: new Date(receipt.date).toISOString(),
          settlement_year: sYear,
          settlement_month: sMonth,
        }),
      });
      toast.success("レシートを更新しました");
      router.push("/");
    } catch (err) {
      console.error("Failed to update receipt:", err);
      if (!(err instanceof Error && err.name === "ConnectionError")) {
        toast.error("更新に失敗しました");
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
      toast.success("レシートを削除しました");
      router.push("/");
    } catch (err) {
      console.error("Failed to delete receipt:", err);
      if (!(err instanceof Error && err.name === "ConnectionError")) {
        toast.error("削除に失敗しました");
      }
    }
  };

  if (loading) return <div className="p-8 text-center text-gray-400">読み込み中...</div>;
  if (!receipt) return <div className="p-8 text-center text-red-500">データが見つかりませんでした</div>;

  const isCreator = currentUserId === receipt.user_id;
  const isSettled = receipt.settled_at !== null;
  const canEdit = isCreator && !isSettled;

  const getPaymentMethodLabel = (method: string) => {
    switch (method) {
      case "half": return "折半";
      case "self": return "自分が10割負担";
      case "other": return "全額相手負担";
      default: return method;
    }
  };

  return (
    <div className="pb-10 bg-white">
      {/* Header */}
      <header className="p-4 border-b border-gray-100 flex justify-between items-center bg-white sticky top-0 z-10">
        <Link href="/" className="text-gray-500 hover:text-gray-800">
          <ArrowLeft size={24} />
        </Link>
        <h1 className="text-lg font-bold text-gray-800">レシート詳細</h1>
        {canEdit ? (
          <button 
            onClick={handleDelete}
            className="text-red-500 p-2 active:bg-red-50 rounded-full transition-colors"
          >
            <Trash2 size={20} />
          </button>
        ) : (
          <div className="w-10" />
        )}
      </header>

      <div className="p-6 space-y-6">
        <div className={`p-4 rounded-2xl border mb-4 flex justify-between items-center ${
          isSettled ? "bg-gray-50 border-gray-200" : "bg-blue-50 border-blue-100"
        }`}>
          <div>
            <p className={`text-xs font-medium mb-1 ${isSettled ? "text-gray-400" : "text-blue-500"}`}>
              精算ステータス
            </p>
            <p className={`text-sm font-semibold ${isSettled ? "text-gray-500" : "text-blue-800"}`}>
              {isSettled ? "精算済み（編集不可）" : `${getPaymentMethodLabel(receipt.payment_method)}で精算予定`}
            </p>
          </div>
          {receipt.payer && (
            <div className="text-right">
              <p className={`text-xs font-medium mb-1 ${isSettled ? "text-gray-400" : "text-blue-500"}`}>支払者</p>
              <p className={`text-sm font-bold bg-white px-3 py-1 rounded-full shadow-sm border ${
                isSettled ? "text-gray-500 border-gray-100" : "text-blue-800 border-blue-100"
              }`}>
                {receipt.payer.nickname}
              </p>
            </div>
          )}
        </div>

        {/* Input Form (Editable only for creator and not settled) */}
        <form onSubmit={handleSave} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-semibold text-gray-800">購入日</label>
              <input 
                type="date" 
                className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all text-gray-900 disabled:opacity-70" 
                value={receipt.date.split('T')[0]} 
                onChange={(e) => {
                  const newDate = e.target.value;
                  const [y, m] = newDate.split('-').map(Number);
                  setReceipt({...receipt, date: newDate, settlement_year: y, settlement_month: m});
                }}
                required
                disabled={!canEdit}
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-semibold text-gray-800">精算対象月</label>
              <input 
                type="month" 
                className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all text-gray-900 disabled:opacity-70" 
                value={receipt.settlement_year ? `${receipt.settlement_year}-${String(receipt.settlement_month).padStart(2, '0')}` : ""} 
                onChange={(e) => {
                  const [y, m] = e.target.value.split('-').map(Number);
                  setReceipt({...receipt, settlement_year: y, settlement_month: m});
                }}
                required
                disabled={!canEdit}
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">お店</label>
            <input 
              type="text" 
              className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all text-gray-900 disabled:opacity-70" 
              value={receipt.shop} 
              onChange={(e) => setReceipt({...receipt, shop: e.target.value})}
              disabled={!canEdit}
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">品名</label>
            <input 
              type="text" 
              className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all text-gray-900 disabled:opacity-70" 
              value={receipt.item} 
              onChange={(e) => setReceipt({...receipt, item: e.target.value})}
              disabled={!canEdit}
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">金額</label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 font-bold">¥</span>
              <input 
                type="number" 
                className="w-full p-3 pl-8 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all font-bold text-lg text-gray-900 disabled:opacity-70" 
                value={receipt.amount} 
                onChange={(e) => setReceipt({...receipt, amount: Number(e.target.value)})}
                required
                disabled={!canEdit}
                min="1"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-semibold text-gray-800">精算方法</label>
              <select 
                className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 appearance-none text-gray-900 disabled:opacity-70" 
                value={receipt.payment_method}
                onChange={(e) => setReceipt({...receipt, payment_method: e.target.value})}
                disabled={!canEdit}
              >
                <option value="half">折半</option>
                <option value="self">自分が10割負担</option>
                <option value="other">全額相手負担</option>
              </select>
            </div>
          </div>

          {canEdit ? (
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
          ) : (
            <div className="pt-4 p-4 bg-gray-50 rounded-2xl border border-dashed border-gray-200">
              <p className="text-xs text-gray-400 text-center">
                {isSettled ? "このレシートは精算済みのため編集・削除できません" : "このレシートは登録者本人のみ編集・削除できます"}
              </p>
            </div>
          )}
        </form>
      </div>
    </div>
  );
}
