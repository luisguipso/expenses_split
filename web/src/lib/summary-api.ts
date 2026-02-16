import api from './api';

export interface SummaryItem {
  user_id: string;
  user_name: string;
  salary_cents: number;
  proportion: number;
  total_shared_cents: number;
  total_personal_cents: number;
  amount_due_cents: number;
  total_paid_cents: number;
  balance_cents: number;
}

export interface SettlementTransfer {
  from_user_id: string;
  from_user_name: string;
  to_user_id: string;
  to_user_name: string;
  amount_cents: number;
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
  settlements: SettlementTransfer[];
  fixed_bills: FixedBillSnapshot[];
}

export interface FixedBillSnapshot {
  id: string;
  fixed_bill_id: string;
  category_id?: string;
  category_name?: string;
  description: string;
  amount_cents: number;
  due_day: number;
  is_shared: boolean;
  paid_by: string;
  paid_by_name?: string;
  assigned_to?: string;
  is_frozen: boolean;
}

export interface UpdateFixedBillSnapshotInput {
  category_id: string;
  description: string;
  amount_cents: number;
  due_day: number;
  is_shared: boolean;
  paid_by: string;
  assigned_to: string;
}

export interface SummaryDetailItem {
  description: string;
  type: 'fixed_bill' | 'expense';
  category_name?: string;
  total_cents: number;
  user_share_cents: number;
  proportion: number;
  is_shared: boolean;
  paid_by_name?: string;
}

export interface SummaryDetailResponse {
  user_id: string;
  user_name: string;
  items: SummaryDetailItem[];
  total_shared_cents: number;
  total_personal_cents: number;
  amount_due_cents: number;
  total_paid_cents: number;
  balance_cents: number;
}

export const summaryApi = {
  getSummary: (householdId: string, year: number, month: number) =>
    api
      .get<SummaryResponse>(
        `/households/${householdId}/summary?year=${year}&month=${month}`
      )
      .then((r) => r.data),

  getUserDetail: (householdId: string, year: number, month: number, userId: string) =>
    api
      .get<SummaryDetailResponse>(
        `/households/${householdId}/summary/detail?year=${year}&month=${month}&user_id=${userId}`
      )
      .then((r) => r.data),

  updateSnapshot: (householdId: string, snapshotId: string, data: UpdateFixedBillSnapshotInput) =>
    api
      .put<FixedBillSnapshot>(
        `/households/${householdId}/bills/snapshots/${snapshotId}`,
        data
      )
      .then((r) => r.data),
};
