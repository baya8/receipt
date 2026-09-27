export type PaymentMethod = "half" | "self" | "other";

export interface User {
  id: string;
  email: string;
  nickname: string;
  created_at: string;
  updated_at: string;
}

export interface Group {
  id: string;
  name: string;
  owner_id: string;
  created_at: string;
  updated_at: string;
  members: User[];
}

export interface Receipt {
  id: string;
  group_id: string;
  user_id: string;
  date: string;
  settlement_year: number;
  settlement_month: number;
  shop: string;
  item: string;
  amount: number;
  payer_id: string;
  payment_method: PaymentMethod;
  settled_at: string | null;
  created_at: string;
  updated_at: string;
  payer?: User;
}

export interface Settlement {
  id: string;
  group_id: string;
  year: number;
  month: number;
  amount: number;
  settled_by: string;
  created_at: string;
  settled_by_user: User;
}

export interface MemberSummary {
  user_id: string;
  nickname: string;
  paid: number;
  share: number;
}

export interface MonthlySummary {
  total_spent: number;
  members: MemberSummary[];
  settlements: Settlement[];
}
