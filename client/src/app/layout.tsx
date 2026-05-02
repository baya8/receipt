import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { Receipt, PlusCircle, User, List, PieChart } from "lucide-react";
import Link from "next/link";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Receipt Share",
  description: "夫婦で家計をシェア",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja">
      <body className={`${inter.className} bg-gray-50 text-gray-900`}>
        <main className="max-w-md mx-auto min-h-screen bg-white pb-20 shadow-lg relative">
          {children}
        </main>

        {/* Bottom Navigation */}
        <nav className="fixed bottom-0 left-0 right-0 max-w-md mx-auto bg-white border-t border-gray-200 flex justify-around py-3 z-10">
          <Link href="/" className="flex flex-col items-center text-gray-400 hover:text-blue-500">
            <List size={24} />
            <span className="text-xs mt-1">一覧</span>
          </Link>
          <Link href="/summary" className="flex flex-col items-center text-gray-400 hover:text-blue-500">
            <PieChart size={24} />
            <span className="text-xs mt-1">精算</span>
          </Link>
          <Link href="/register" className="flex flex-col items-center text-gray-400 hover:text-blue-500">
            <PlusCircle size={24} />
            <span className="text-xs mt-1">登録</span>
          </Link>
          <Link href="/profile" className="flex flex-col items-center text-gray-400 hover:text-blue-500">
            <User size={24} />
            <span className="text-xs mt-1">設定</span>
          </Link>
        </nav>
      </body>
    </html>
  );
}
