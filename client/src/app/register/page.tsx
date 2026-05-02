"use client";

import { useState, useEffect, useRef } from "react";
import { Camera, Save, Loader2 } from "lucide-react";
import { apiRequest } from "@/lib/api";
import { useRouter } from "next/navigation";

export default function Register() {
  const router = useRouter();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [formData, setFormData] = useState({
    date: new Date().toISOString().split('T')[0],
    shop: "",
    item: "",
    amount: 0,
    payer_id: 0, // 後でログインユーザーIDにセット
    payment_method: "折半",
  });
  const [loading, setLoading] = useState(false);
  const [analyzing, setAnalyzing] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) {
      router.push("/login");
    }
  }, [router]);

  const handleCameraClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setAnalyzing(true);
    const formData = new FormData();
    formData.append("image", file);

    try {
      const token = localStorage.getItem("token");
      const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      
      const response = await fetch(`${API_URL}/api/receipts/analyze`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || "Analysis failed");
      }

      const data = await response.json();
      setFormData((prev) => ({
        ...prev,
        date: data.date || prev.date,
        shop: data.shop || prev.shop,
        item: data.item || prev.item,
        amount: data.amount || prev.amount,
      }));
    } catch (err) {
      console.error("Failed to analyze receipt:", err);
      alert("解析に失敗しました。手動で入力してください。");
    } finally {
      setAnalyzing(false);
      // 同じファイルを再度選択できるようにリセット
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const user = JSON.parse(localStorage.getItem("user") || "{}");
      await apiRequest("/api/receipts", {
        method: "POST",
        body: JSON.stringify({
          ...formData,
          amount: Number(formData.amount),
          group_id: 1, // 固定
          payer_id: user.id || 1,
          date: new Date(formData.date).toISOString(),
        }),
      });
      router.push("/");
    } catch (err) {
      console.error("Failed to register receipt:", err);
      alert("登録に失敗しました");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="pb-10">
      <header className="p-4 border-b border-gray-100">
        <h1 className="text-xl font-bold text-gray-800">レシート登録</h1>
      </header>

      <div className="p-6 space-y-8">
        <section>
          <input 
            type="file" 
            accept="image/*" 
            capture="environment" 
            className="hidden" 
            ref={fileInputRef}
            onChange={handleFileChange}
          />
          <button 
            type="button"
            onClick={handleCameraClick}
            disabled={analyzing}
            className="w-full aspect-video border-2 border-dashed border-blue-200 rounded-2xl bg-blue-50 flex flex-col items-center justify-center gap-2 text-blue-600 active:bg-blue-100 transition-colors disabled:opacity-50"
          >
            {analyzing ? (
              <>
                <Loader2 size={48} className="animate-spin text-blue-400" />
                <span className="font-semibold text-blue-400">解析中...</span>
              </>
            ) : (
              <>
                <Camera size={48} strokeWidth={1.5} />
                <span className="font-semibold">レシートを撮影して自動入力</span>
                <span className="text-xs text-blue-400">Gemini AI が内容を読み取ります</span>
              </>
            )}
          </button>
        </section>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">購入日</label>
            <input 
              type="date" 
              className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900" 
              value={formData.date}
              onChange={(e) => setFormData({...formData, date: e.target.value})}
              required
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">お店</label>
            <input 
              type="text" 
              placeholder="お店の名前を入力" 
              className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900" 
              value={formData.shop}
              onChange={(e) => setFormData({...formData, shop: e.target.value})}
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">品名</label>
            <input 
              type="text" 
              placeholder="例：夕食の買い物" 
              className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900" 
              value={formData.item}
              onChange={(e) => setFormData({...formData, item: e.target.value})}
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">金額</label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 font-bold">¥</span>
              <input 
                type="number" 
                placeholder="0" 
                className="w-full p-3 pl-8 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 font-bold text-lg text-gray-900" 
                value={formData.amount || ""}
                onChange={(e) => setFormData({...formData, amount: Number(e.target.value)})}
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-semibold text-gray-800">精算方法</label>
              <select 
                className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 appearance-none text-gray-900"
                value={formData.payment_method}
                onChange={(e) => setFormData({...formData, payment_method: e.target.value})}
              >
                <option>折半</option>
                <option>自分が10割負担</option>
                <option>全額相手負担</option>
              </select>
            </div>
          </div>

          <button 
            type="submit" 
            disabled={loading || analyzing}
            className="w-full py-4 bg-blue-600 text-white rounded-2xl font-bold flex items-center justify-center gap-2 shadow-lg shadow-blue-200 active:scale-[0.98] transition-all disabled:opacity-50"
          >
            <Save size={20} />
            {loading ? "保存中..." : "保存する"}
          </button>
        </form>
      </div>
    </div>
  );
}
