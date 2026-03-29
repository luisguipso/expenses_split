import api from './api';
import type { ImportPreviewResponse, ImportConfirmItem, Expense } from './types';

export const importApi = {
  upload: (householdId: string, file: File): Promise<ImportPreviewResponse> => {
    const formData = new FormData();
    formData.append('file', file);
    return api
      .post<ImportPreviewResponse>(
        `/households/${householdId}/import/upload`,
        formData,
        { headers: { 'Content-Type': 'multipart/form-data' } }
      )
      .then((r) => r.data);
  },

  confirm: (householdId: string, items: ImportConfirmItem[]): Promise<Expense[]> => {
    return api
      .post<Expense[]>(`/households/${householdId}/import/confirm`, { items })
      .then((r) => r.data);
  },
};
