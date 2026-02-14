import api from './api';
import { Category } from './types';

export const categoryApi = {
  list: (householdId: string) =>
    api.get<Category[]>(`/households/${householdId}/categories`).then((r) => r.data),

  create: (householdId: string, data: { name: string; icon: string }) =>
    api.post<Category>(`/households/${householdId}/categories`, data).then((r) => r.data),

  update: (householdId: string, id: string, data: { name: string; icon: string }) =>
    api.put<Category>(`/households/${householdId}/categories/${id}`, data).then((r) => r.data),

  delete: (householdId: string, id: string) =>
    api.delete(`/households/${householdId}/categories/${id}`),
};
