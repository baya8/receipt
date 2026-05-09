import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { Receipt, PlusCircle, User, List, PieChart } from "lucide-react";
import Link from "next/link";
import { ApiProvider } from "@/lib/ApiContext";
import BottomNav from "@/components/BottomNav";
import { Toaster } from "sonner";

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
        <ApiProvider>
          <main className="max-w-md mx-auto min-h-screen bg-white pb-20 shadow-lg relative">
            {children}
          </main>
          <BottomNav />
          <Toaster 
            position="top-center" 
            richColors 
            toastOptions={{
              style: {
                marginTop: '80vh',
              },
            }}
          />
        </ApiProvider>
      </body>
    </html>
  );
}
