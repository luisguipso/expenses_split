import api from './api';
import { FixedBill } from './types';

export interface CreateFixedBillInput {
  category_id: string;
  description: string;
  amount_cents: number;
  due_day: number;
  is_shared: boolean;
  paid_by: string;
  assigned_to: string;
}

export interface UpdateFixedBillInput extends CreateFixedBillInput {
  is_active: boolean;
}

export const fixedBillApi = {
  list: (householdId: string) =>
    api.get<FixedBill[]>(`/households/${householdId}/bills`).then((r) => r.data),

  create: (householdId: string, data: CreateFixedBillInput) =>
    api.post<FixedBill>(`/households/${householdId}/bills`, data).then((r) => r.data),

  update: (householdId: string, id: string, data: UpdateFixedBillInput) =>
    api.put<FixedBill>(`/households/${householdId}/bills/${id}`, data).then((r) => r.data),

  delete: (householdId: string, id: string) =>
    api.delete(`/households/${householdId}/bills/${id}`),
};
