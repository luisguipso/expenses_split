import api from './api';
import { Household, Member } from './types';

export const householdApi = {
  list: () => api.get<Household[]>('/households').then((r) => r.data),

  get: (id: string) =>
    api.get<Household>(`/households/${id}`).then((r) => r.data),

  create: (name: string) =>
    api.post<Household>('/households', { name }).then((r) => r.data),

  update: (id: string, name: string) =>
    api.put<Household>(`/households/${id}`, { name }).then((r) => r.data),

  delete: (id: string) => api.delete(`/households/${id}`),

  join: (inviteCode: string) =>
    api
      .post<Household>('/households/join', { invite_code: inviteCode })
      .then((r) => r.data),

  listMembers: (id: string) =>
    api.get<Member[]>(`/households/${id}/members`).then((r) => r.data),

  updateSalary: (householdId: string, memberId: string, salaryCents: number) =>
    api.put(`/households/${householdId}/members/${memberId}/salary`, {
      salary_cents: salaryCents,
    }),

  updateSplitMode: (householdId: string, splitMode: string) =>
    api.put(`/households/${householdId}/split-mode`, {
      split_mode: splitMode,
    }),

  updateSplitPercentage: (householdId: string, memberId: string, percentage: number) =>
    api.put(`/households/${householdId}/members/${memberId}/percentage`, {
      split_percentage: percentage,
    }),

  removeMember: (householdId: string, memberId: string) =>
    api.delete(`/households/${householdId}/members/${memberId}`),
};
