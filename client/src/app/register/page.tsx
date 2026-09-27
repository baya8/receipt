"use client";

import { useState, useRef } from "react";
import { Camera, ImagePlus, Save, Loader2 } from "lucide-react";
import { apiRequest, ApiError } from "@/lib/api";
import { handleApiError } from "@/lib/errors";
import { compressImage } from "@/lib/image";
import { isMonthFullySettled } from "@/lib/settlement";
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

  // 指定した精算対象月（YYYY-MM）が精算済み（残高0円）かどうかを確認する。
  // 精算済みの月に後から登録すると、精算後の残高計算と実際の未清算額がずれてしまうため。
  // 確認に失敗した場合（ネットワークエラー等）は、登録自体をブロックしないようfalse（未精算扱い）を返す。
  const isSettlementMonthClosed = async (yearMonth: string): Promise<boolean> => {
    if (!primaryGroup) return false;
    const [year, month] = yearMonth.split('-').map(Number);
    try {
      const summary = await apiRequest(`/api/summary?group_id=${primaryGroup.id}&year=${year}&month=${month}`);
      return isMonthFullySettled(summary);
    } catch (err) {
      console.error("Failed to check settlement status:", err);
      return false;
    }
  };

  // ユーザーが直接選択した月が精算済みなら、選択を差し戻してエラーを表示する
  const checkMonthAvailable = async (yearMonth: string): Promise<boolean> => {
    if (await isSettlementMonthClosed(yearMonth)) {
      const [year, month] = yearMonth.split('-').map(Number);
      toast.error(`${year}年${month}月は精算済みのため選択できません`);
      return false;
    }
    return true;
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

      const newDate = data.date || formData.date;
      let newMonth = newDate.slice(0, 7);
      if (await isSettlementMonthClosed(newMonth)) {
        const [closedYear, closedMonth] = newMonth.split('-').map(Number);
        newMonth = new Date().toISOString().slice(0, 7);
        toast.info(`${closedYear}年${closedMonth}月は精算済みのため、精算対象月を今月に設定しました`);
      }

      setFormData((prev) => ({
        ...prev,
        date: newDate,
        settlement_month: newMonth,
        shop: data.shop || prev.shop,
        item: data.item || prev.item,
        amount: data.amount || prev.amount,
      }));
    } catch (err) {
      if (err instanceof ApiError && err.status === 429) {
        // Gemini APIの利用上限に達した場合は、サーバー側で用意した案内メッセージをそのまま表示する
        handleApiError(err, err.message);
      } else {
        // それ以外は詳細を表示しない（想定外のエラー内容が表示されるのを避けるため）
        handleApiError(err, "解析に失敗しました。手動で入力してください。");
      }
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

    if (!(await checkMonthAvailable(formData.settlement_month))) return;

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
                onChange={async (e) => {
                  const newDate = e.target.value;
                  const newMonth = newDate.slice(0, 7);
                  if (!(await checkMonthAvailable(newMonth))) return;
                  setFormData({...formData, date: newDate, settlement_month: newMonth});
                }}
                required
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-semibold text-gray-800">精算対象月</label>
              <Input
                type="month"
                value={formData.settlement_month}
                onChange={async (e) => {
                  const newMonth = e.target.value;
                  if (!(await checkMonthAvailable(newMonth))) return;
                  setFormData({...formData, settlement_month: newMonth});
                }}
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
