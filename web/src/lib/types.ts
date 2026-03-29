export interface Household {
  id: string;
  name: string;
  invite_code?: string;
  created_at: string;
}

export interface Member {
  user_id: string;
  user_name: string;
  user_email: string;
  salary_cents: number;
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

export interface ParsedExpensePreview {
  description: string;
  amount_cents: number;
  date: string;
  suggested_category_id?: string;
}

export interface ImportPreviewResponse {
  provider: string;
  items: ParsedExpensePreview[];
}

export interface ImportConfirmItem {
  category_id: string;
  description: string;
  amount_cents: number;
  expense_date: string;
  is_shared: boolean;
  assigned_to: string;
}
