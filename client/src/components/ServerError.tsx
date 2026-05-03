"use client";

import { AlertCircle, RefreshCw } from "lucide-react";

interface ServerErrorProps {
  onRetry?: () => void;
}

export default function ServerError({ onRetry }: ServerErrorProps) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[80vh] p-6 text-center">
      <div className="w-20 h-20 bg-red-50 rounded-full flex items-center justify-center mb-6">
        <AlertCircle size={40} className="text-red-500" />
      </div>
      <h2 className="text-2xl font-bold text-gray-900 mb-2">サーバーに接続できません</h2>
      <p className="text-gray-500 mb-8 max-w-xs">
        バックエンドサーバーが停止しているか、ネットワーク接続に問題があります。
      </p>
      <button
        onClick={() => onRetry ? onRetry() : window.location.reload()}
        className="flex items-center gap-2 bg-blue-600 text-white px-6 py-3 rounded-xl font-bold shadow-lg shadow-blue-200 active:scale-95 transition-transform"
      >
        <RefreshCw size={20} />
        再試行する
      </button>
    </div>
  );
}
