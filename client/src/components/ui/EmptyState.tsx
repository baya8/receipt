import Link from "next/link";
import { PlusCircle } from "lucide-react";
import type { ComponentType } from "react";

interface EmptyStateProps {
  description: string;
  title?: string;
  icon?: ComponentType<{ size?: number; className?: string }>;
  actionHref?: string;
  actionLabel?: string;
}

export default function EmptyState({
  description,
  title = "グループがありません",
  icon: Icon = PlusCircle,
  actionHref = "/profile",
  actionLabel = "設定画面へ",
}: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[80vh] p-8 text-center">
      <div className="w-20 h-20 bg-blue-50 rounded-full flex items-center justify-center mb-6">
        <Icon size={40} className="text-blue-500" />
      </div>
      <h2 className="text-xl font-bold text-gray-900 mb-2">{title}</h2>
      <p className="text-gray-500 mb-8">{description}</p>
      <Link href={actionHref} className="bg-blue-600 text-white px-6 py-3 rounded-xl font-bold shadow-lg shadow-blue-100">
        {actionLabel}
      </Link>
    </div>
  );
}
