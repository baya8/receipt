import { toast } from "sonner";
import { ConnectionError } from "@/lib/api";

interface HandleApiErrorOptions {
  // サーバーが返した具体的なエラー内容（err.message）をトーストに含めるかどうか。
  // Gemini解析のようにサーバー内部のエラー文字列がそのまま返ってくるエンドポイントでは
  // 意図せず内部情報を利用者に見せてしまうため、呼び出し側で明示的にtrueを指定した場合のみ含める。
  includeDetail?: boolean;
}

export function handleApiError(err: unknown, fallbackMessage: string, options: HandleApiErrorOptions = {}) {
  console.error(fallbackMessage, err);

  // ConnectionErrorはApiContextのServerError画面が案内するため、ここでは二重に通知しない
  if (err instanceof ConnectionError) return;

  if (options.includeDetail && err instanceof Error && err.message) {
    toast.error(`${fallbackMessage}: ${err.message}`);
    return;
  }
  toast.error(fallbackMessage);
}
