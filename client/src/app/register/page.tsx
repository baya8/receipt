"use client";

import { useState, useRef } from "react";
import { Camera, ImagePlus, Save, Loader2 } from "lucide-react";
import { apiRequest } from "@/lib/api";
import { handleApiError } from "@/lib/errors";
import { compressImage } from "@/lib/image";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { useAuthGuard } from "@/hooks/useAuthGuard";
import { useCurrentUser } from "@/hooks/useCurrentUser";
import { useGroups } from "@/hooks/useGroups";
import LoadingScreen from "@/components/ui/LoadingScreen";
import EmptyState from "@/components/ui/EmptyState";
import Input from "@/components/ui/Input";
import Select from "@/components/ui/Select";
import Button from "@/components/ui/Button";

export default function Register() {
  useAuthGuard();
  const router = useRouter();
  const currentUser = useCurrentUser();
  const { primaryGroup, loading: fetchingGroups } = useGroups();
  const cameraInputRef = useRef<HTMLInputElement>(null);
  const galleryInputRef = useRef<HTMLInputElement>(null);
  const [formData, setFormData] = useState({
    date: new Date().toISOString().split('T')[0],
    settlement_month: new Date().toISOString().slice(0, 7), // YYYY-MM
    shop: "",
    item: "",
    amount: 0,
    payer_id: currentUser?.id || "",
    payment_method: "half",
  });
  const [loading, setLoading] = useState(false);
  const [analyzing, setAnalyzing] = useState(false);

  const handleCameraClick = () => {
    cameraInputRef.current?.click();
  };

  const handleGalleryClick = () => {
    galleryInputRef.current?.click();
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setAnalyzing(true);
    try {
      const uploadFile = await compressImage(file);
      const formDataBody = new FormData();
      formDataBody.append("image", uploadFile);

      const data = await apiRequest("/api/receipts/analyze", {
        method: "POST",
        body: formDataBody,
      });

      setFormData((prev) => {
        const newDate = data.date || prev.date;
        return {
          ...prev,
          date: newDate,
          settlement_month: newDate.slice(0, 7),
          shop: data.shop || prev.shop,
          item: data.item || prev.item,
          amount: data.amount || prev.amount,
        };
      });
    } catch (err) {
      // AnalyzeReceiptはサーバー内部のエラーをそのまま返すことがあるため、詳細は表示しない
      handleApiError(err, "解析に失敗しました。手動で入力してください。");
    } finally {
      setAnalyzing(false);
      e.target.value = "";
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!primaryGroup) return;

    if (formData.amount <= 0) {
      toast.error("金額は1円以上にしてください");
      return;
    }

    setLoading(true);
    try {
      const [sYear, sMonth] = formData.settlement_month.split('-').map(Number);
      await apiRequest("/api/receipts", {
        method: "POST",
        body: JSON.stringify({
          ...formData,
          amount: Number(formData.amount),
          group_id: primaryGroup.id,
          date: new Date(formData.date).toISOString(),
          settlement_year: sYear,
          settlement_month: sMonth,
        }),
      });
      toast.success("レシートを登録しました");
      router.push("/");
    } catch (err) {
      handleApiError(err, "登録に失敗しました");
    } finally {
      setLoading(false);
    }
  };

  if (fetchingGroups) return <LoadingScreen />;

  if (!primaryGroup) {
    return <EmptyState description="レシートを登録するには、まずグループを作成してください。" />;
  }

  return (
    <div className="pb-10">
      <header className="p-4 border-b border-gray-100 flex justify-between items-center bg-white sticky top-0 z-10">
        <h1 className="text-xl font-bold text-gray-800">レシート登録</h1>
        <div className="text-[10px] font-bold bg-blue-50 text-blue-600 px-2 py-1 rounded">
          {primaryGroup.name}
        </div>
      </header>

      <div className="p-6 space-y-8">
        <section>
          <input
            type="file"
            accept="image/*"
            capture="environment"
            className="hidden"
            ref={cameraInputRef}
            onChange={handleFileChange}
          />
          <input
            type="file"
            accept="image/*"
            className="hidden"
            ref={galleryInputRef}
            onChange={handleFileChange}
          />

          {analyzing ? (
            <div className="w-full aspect-video border-2 border-dashed border-blue-200 rounded-2xl bg-blue-50 flex flex-col items-center justify-center gap-2 text-blue-600">
              <Loader2 size={48} className="animate-spin text-blue-400" />
              <span className="font-semibold text-blue-400">解析中...</span>
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-3">
              <button
                type="button"
                onClick={handleCameraClick}
                className="aspect-square border-2 border-dashed border-blue-200 rounded-2xl bg-blue-50 flex flex-col items-center justify-center gap-2 text-blue-600 active:bg-blue-100 transition-colors"
              >
                <Camera size={40} strokeWidth={1.5} />
                <span className="font-semibold text-sm">撮影する</span>
              </button>
              <button
                type="button"
                onClick={handleGalleryClick}
                className="aspect-square border-2 border-dashed border-blue-200 rounded-2xl bg-blue-50 flex flex-col items-center justify-center gap-2 text-blue-600 active:bg-blue-100 transition-colors"
              >
                <ImagePlus size={40} strokeWidth={1.5} />
                <span className="font-semibold text-sm">アルバムから選択</span>
              </button>
            </div>
          )}
          <p className="text-xs text-blue-400 text-center mt-2">Gemini AI が内容を読み取ります</p>
        </section>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-semibold text-gray-800">購入日</label>
              <Input
                type="date"
                value={formData.date}
                onChange={(e) => {
                  const newDate = e.target.value;
                  setFormData({...formData, date: newDate, settlement_month: newDate.slice(0, 7)});
                }}
                required
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-semibold text-gray-800">精算対象月</label>
              <Input
                type="month"
                value={formData.settlement_month}
                onChange={(e) => setFormData({...formData, settlement_month: e.target.value})}
                required
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">お店</label>
            <Input
              type="text"
              placeholder="お店の名前を入力"
              value={formData.shop}
              onChange={(e) => setFormData({...formData, shop: e.target.value})}
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">品名</label>
            <Input
              type="text"
              placeholder="例：夕食の買い物"
              value={formData.item}
              onChange={(e) => setFormData({...formData, item: e.target.value})}
            />
          </div>

          <div className="space-y-1">
            <label className="text-sm font-semibold text-gray-800">金額</label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 font-bold">¥</span>
              <Input
                type="number"
                placeholder="0"
                className="pl-8 font-bold text-lg"
                value={formData.amount || ""}
                onChange={(e) => setFormData({...formData, amount: Number(e.target.value)})}
                required
                min="1"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-semibold text-gray-800">支払者</label>
              <Select
                value={formData.payer_id}
                onChange={(e) => setFormData({...formData, payer_id: e.target.value})}
              >
                {primaryGroup.members.map((member) => (
                  <option key={member.id} value={member.id}>{member.nickname}</option>
                ))}
              </Select>
            </div>
            <div className="space-y-1">
              <label className="text-sm font-semibold text-gray-800">精算方法</label>
              <Select
                value={formData.payment_method}
                onChange={(e) => setFormData({...formData, payment_method: e.target.value})}
              >
                <option value="half">折半</option>
                <option value="self">自分が10割負担</option>
                <option value="other">全額相手負担</option>
              </Select>
            </div>
          </div>

          <Button type="submit" disabled={loading || analyzing}>
            <Save size={20} />
            {loading ? "保存中..." : "保存する"}
          </Button>
        </form>
      </div>
    </div>
  );
}
