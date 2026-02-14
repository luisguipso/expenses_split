import api from './api';
import { Expense } from './types';

export interface CreateExpenseInput {
  category_id: string;
  description: string;
  amount_cents: number;
  expense_date: string;
  is_shared: boolean;
  assigned_to: string;
}

export interface UpdateExpenseInput {
  category_id: string;
  description: string;
  amount_cents: number;
  expense_date: string;
  is_shared: boolean;
  assigned_to: string;
}

export interface ExpenseFilter {
  month?: number;
  year?: number;
  category_id?: string;
  user_id?: string;
}

export const expenseApi = {
  list: (householdId: string, filter?: ExpenseFilter) => {
    const params = new URLSearchParams();
    if (filter?.month) params.set('month', String(filter.month));
    if (filter?.year) params.set('year', String(filter.year));
    if (filter?.category_id) params.set('category_id', filter.category_id);
    if (filter?.user_id) params.set('user_id', filter.user_id);
    const qs = params.toString();
    return api
      .get<Expense[]>(`/households/${householdId}/expenses${qs ? '?' + qs : ''}`)
      .then((r) => r.data);
  },

  create: (householdId: string, data: CreateExpenseInput) =>
    api.post<Expense>(`/households/${householdId}/expenses`, data).then((r) => r.data),

  update: (householdId: string, id: string, data: UpdateExpenseInput) =>
    api.put<Expense>(`/households/${householdId}/expenses/${id}`, data).then((r) => r.data),

  delete: (householdId: string, id: string) =>
    api.delete(`/households/${householdId}/expenses/${id}`),
};
