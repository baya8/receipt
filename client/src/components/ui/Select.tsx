import type { SelectHTMLAttributes } from "react";

export default function Select({ className = "", children, ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      className={`w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 appearance-none text-gray-900 ${className}`}
      {...props}
    >
      {children}
    </select>
  );
}
