import api from './api';

export interface SummaryItem {
  user_id: string;
  user_name: string;
  salary_cents: number;
  proportion: number;
  total_shared_cents: number;
  total_personal_cents: number;
  amount_due_cents: number;
}

export interface SummaryResponse {
  id: string;
  household_id: string;
  year: number;
  month: number;
  total_shared_cents: number;
  total_all_cents: number;
  generated_at: string;
  items: SummaryItem[];
}

export interface DashboardResponse {
  household_name: string;
  year: number;
  month: number;
  total_expenses: number;
  total_fixed_bills: number;
  total_shared: number;
  total_personal: number;
  expense_count: number;
  fixed_bill_count: number;
  member_breakdown: SummaryItem[];
}

export const summaryApi = {
  getDashboard: (householdId: string) =>
    api
      .get<DashboardResponse>(`/households/${householdId}/dashboard`)
      .then((r) => r.data),

  getSummary: (householdId: string, year: number, month: number) =>
    api
      .get<SummaryResponse>(
        `/households/${householdId}/summary?year=${year}&month=${month}`
      )
      .then((r) => r.data),
};
