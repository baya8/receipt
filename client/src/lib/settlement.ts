import type { MonthlySummary } from "@/types";

// その月が精算済み（残高0円）かどうかを判定する。
// 2人グループの精算は常にゼロサム（片方の残高ともう片方の残高が符号反転の関係）なので、
// どのメンバーを基準にしても「残高が0かどうか」の判定結果自体は変わらない。
export function isMonthFullySettled(summary: MonthlySummary): boolean {
  const [anchor] = summary.members;
  if (!anchor || summary.settlements.length === 0) return false;

  const initialBalance = anchor.share - anchor.paid;
  const totalSettledByAnchor = summary.settlements
    .filter((s) => s.settled_by === anchor.user_id)
    .reduce((sum, s) => sum + s.amount, 0);
  const totalSettledByOthers = summary.settlements
    .filter((s) => s.settled_by !== anchor.user_id)
    .reduce((sum, s) => sum + s.amount, 0);
  const currentBalance = initialBalance - totalSettledByAnchor + totalSettledByOthers;

  return currentBalance === 0;
}
