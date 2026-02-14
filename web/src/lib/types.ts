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
  assigned_to?: string;
  is_active: boolean;
}
