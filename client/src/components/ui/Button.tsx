import type { ButtonHTMLAttributes } from "react";

type ButtonVariant = "primary" | "dark";

const VARIANT_CLASSES: Record<ButtonVariant, string> = {
  primary: "bg-blue-600 shadow-blue-200",
  dark: "bg-gray-900",
};

export default function Button({
  variant = "primary",
  className = "",
  children,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: ButtonVariant }) {
  return (
    <button
      className={`w-full py-4 text-white rounded-2xl font-bold flex items-center justify-center gap-2 shadow-lg active:scale-[0.98] transition-all disabled:opacity-50 ${VARIANT_CLASSES[variant]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}
