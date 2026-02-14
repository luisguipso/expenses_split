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
