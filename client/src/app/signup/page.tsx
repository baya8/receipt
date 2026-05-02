"use client";

import { useState } from "react";
import { apiRequest } from "@/lib/api";
import { useRouter } from "next/navigation";
import Link from "next/link";

export default function Signup() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [nickname, setNickname] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await apiRequest("/auth/register", {
        method: "POST",
        body: JSON.stringify({ email, password, nickname }),
      });
      // 登録成功したらそのままログイン画面へ
      alert("登録が完了しました。ログインしてください。");
      router.push("/login");
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-8 flex flex-col justify-center min-h-screen space-y-8 bg-white">
      <div className="text-center">
        <h1 className="text-3xl font-black text-blue-600">新規アカウント登録</h1>
        <p className="text-gray-500">家計シェアを始めましょう</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        {error && <p className="text-red-500 text-sm text-center bg-red-50 p-2 rounded">{error}</p>}
        
        <div className="space-y-1">
          <label className="text-sm font-semibold text-gray-800">ニックネーム</label>
          <input 
            type="text" 
            placeholder="例: たろう"
            className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900" 
            value={nickname}
            onChange={(e) => setNickname(e.target.value)}
            required
          />
        </div>

        <div className="space-y-1">
          <label className="text-sm font-semibold text-gray-800">メールアドレス</label>
          <input 
            type="email" 
            placeholder="example@mail.com"
            className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900" 
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>

        <div className="space-y-1">
          <label className="text-sm font-semibold text-gray-800">パスワード (8文字以上)</label>
          <input 
            type="password" 
            className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900" 
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
          />
        </div>

        <button 
          disabled={loading}
          className="w-full py-4 bg-blue-600 text-white rounded-2xl font-bold shadow-lg shadow-blue-200 active:scale-[0.98] transition-all disabled:opacity-50"
        >
          {loading ? "登録中..." : "アカウントを作成する"}
        </button>
      </form>

      <div className="text-center">
        <p className="text-sm text-gray-500">
          すでにアカウントをお持ちですか？{" "}
          <Link href="/login" className="text-blue-600 font-bold hover:underline">
            ログイン
          </Link>
        </p>
      </div>
    </div>
  );
}
