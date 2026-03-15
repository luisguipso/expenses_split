export interface Household {
  id: string;
  name: string;
  invite_code?: string;
  split_mode: string;
  created_at: string;
}

export interface Member {
  user_id: string;
  user_name: string;
  user_email: string;
  salary_cents: number;
  split_percentage: number;
  role: string;
  joined_at: string;
}

export interface Category {
  id: string;
  name: string;
  icon: string;
}

export interface FixedBill {
  id: string;
  category_id?: string;
  category_name?: string;
  description: string;
  amount_cents: number;
  due_day: number;
  is_shared: boolean;
  paid_by: string;
  paid_by_name?: string;
  assigned_to?: string;
  is_active: boolean;
}

export interface Expense {
  id: string;
  category_id?: string;
  category_name?: string;
  description: string;
  amount_cents: number;
  expense_date: string;
  is_shared: boolean;
  paid_by: string;
  paid_by_name?: string;
  assigned_to?: string;
}
