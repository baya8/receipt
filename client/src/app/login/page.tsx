"use client";

import { useState } from "react";
import { apiRequest } from "@/lib/api";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useApi } from "@/lib/ApiContext";
import { toast } from "sonner";

export default function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const router = useRouter();
  const { checkAuth } = useApi();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const data = await apiRequest("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });
      localStorage.setItem("token", data.token);
      localStorage.setItem("user", JSON.stringify(data.user));
      checkAuth();
      toast.success("ログインしました");
      router.push("/");
    } catch (err: any) {
      toast.error(err.message || "ログインに失敗しました");
    }
  };

  return (
    <div className="p-8 flex flex-col justify-center min-h-screen space-y-8 bg-white">
      <div className="text-center">
        <h1 className="text-3xl font-black text-blue-600">Receipt Share</h1>
        <p className="text-gray-500">夫婦で家計をシェア</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-1">
          <label className="text-sm font-semibold text-gray-800">メールアドレス</label>
          <input 
            type="email" 
            className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900" 
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        <div className="space-y-1">
          <label className="text-sm font-semibold text-gray-800">パスワード</label>
          <input 
            type="password" 
            className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900" 
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>
        <button className="w-full py-4 bg-blue-600 text-white rounded-2xl font-bold shadow-lg shadow-blue-200 active:scale-[0.98] transition-all">
          ログイン
        </button>
      </form>

      <div className="text-center">
        <p className="text-sm text-gray-500">
          アカウントをお持ちでないですか？{" "}
          <Link href="/signup" className="text-blue-600 font-bold hover:underline">
            新規登録
          </Link>
        </p>
      </div>
    </div>
  );
}
