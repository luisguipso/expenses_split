import { useState, useEffect } from 'react';
import { summaryApi, FixedBillSnapshot, UpdateFixedBillSnapshotInput } from '../lib/summary-api';
import { categoryApi } from '../lib/category-api';
import { householdApi } from '../lib/household-api';
import { Category, Member } from '../lib/types';
import Spinner from './Spinner';

function formatCurrency(cents: number): string {
  return ((cents ?? 0) / 100).toLocaleString('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  });
}

interface Props {
  householdId: string;
  snapshot: FixedBillSnapshot;
  onClose: () => void;
  onSaved: () => void;
}

export default function SnapshotEditModal({ householdId, snapshot, onClose, onSaved }: Props) {
  const [categories, setCategories] = useState<Category[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const [form, setForm] = useState({
    category_id: snapshot.category_id || '',
    description: snapshot.description,
    amount: (snapshot.amount_cents / 100).toFixed(2),
    due_day: String(snapshot.due_day),
    is_shared: snapshot.is_shared,
    paid_by: snapshot.paid_by || '',
    assigned_to: snapshot.assigned_to || '',
  });

  useEffect(() => {
    Promise.all([
      categoryApi.list(householdId),
      householdApi.listMembers(householdId),
    ])
      .then(([cats, mems]) => {
        setCategories(cats);
        setMembers(mems);
      })
      .catch(() => setError('Erro ao carregar dados'))
      .finally(() => setLoading(false));
  }, [householdId]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    const amountCents = Math.round(parseFloat(form.amount) * 100);
    const dueDay = parseInt(form.due_day, 10);

    if (isNaN(amountCents) || amountCents <= 0) {
      setError('Valor inválido');
      return;
    }
    if (isNaN(dueDay) || dueDay < 1 || dueDay > 31) {
      setError('Dia de vencimento deve ser entre 1 e 31');
      return;
    }

    const data: UpdateFixedBillSnapshotInput = {
      category_id: form.category_id,
      description: form.description.trim(),
      amount_cents: amountCents,
      due_day: dueDay,
      is_shared: form.is_shared,
      paid_by: form.paid_by,
      assigned_to: form.is_shared ? '' : form.assigned_to,
    };

    setSaving(true);
    try {
      await summaryApi.updateSnapshot(householdId, snapshot.id, data);
      onSaved();
    } catch {
      setError('Erro ao salvar alteração');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={onClose}>
      <div
        className="relative max-h-[90vh] w-full max-w-md overflow-y-auto rounded-lg bg-white shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b px-6 py-4">
          <div>
            <h3 className="text-lg font-bold text-gray-900">Editar Conta Fixa</h3>
            <p className="text-sm text-gray-500">
              Valor congelado — {formatCurrency(snapshot.amount_cents)}
            </p>
          </div>
          <button
            onClick={onClose}
            className="rounded-full p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
            aria-label="Fechar"
          >
            ✕
          </button>
        </div>

        <div className="px-6 py-4">
          {loading ? (
            <Spinner text="Carregando..." />
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              {error && <p className="text-sm text-red-600">{error}</p>}

              <div>
                <label className="block text-sm font-medium text-gray-700">Descrição</label>
                <input
                  type="text"
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">Valor (R$)</label>
                <input
                  type="number"
                  step="0.01"
                  min="0.01"
                  value={form.amount}
                  onChange={(e) => setForm({ ...form, amount: e.target.value })}
                  className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">Dia de vencimento</label>
                <input
                  type="number"
                  min="1"
                  max="31"
                  value={form.due_day}
                  onChange={(e) => setForm({ ...form, due_day: e.target.value })}
                  className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">Categoria</label>
                <select
                  value={form.category_id}
                  onChange={(e) => setForm({ ...form, category_id: e.target.value })}
                  className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                >
                  <option value="">Sem categoria</option>
                  {categories.map((c) => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">Pago por</label>
                <select
                  value={form.paid_by}
                  onChange={(e) => setForm({ ...form, paid_by: e.target.value })}
                  className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                >
                  <option value="">Selecionar</option>
                  {members.map((m) => (
                    <option key={m.user_id} value={m.user_id}>{m.user_name}</option>
                  ))}
                </select>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="is_shared"
                  checked={form.is_shared}
                  onChange={(e) => setForm({ ...form, is_shared: e.target.checked, assigned_to: '' })}
                  className="h-4 w-4 rounded border-gray-300"
                />
                <label htmlFor="is_shared" className="text-sm text-gray-700">Compartilhada</label>
              </div>

              {!form.is_shared && (
                <div>
                  <label className="block text-sm font-medium text-gray-700">Atribuída a</label>
                  <select
                    value={form.assigned_to}
                    onChange={(e) => setForm({ ...form, assigned_to: e.target.value })}
                    className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                  >
                    <option value="">Selecionar</option>
                    {members.map((m) => (
                      <option key={m.user_id} value={m.user_id}>{m.user_name}</option>
                    ))}
                  </select>
                </div>
              )}

              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={onClose}
                  className="rounded-md border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
                >
                  {saving ? 'Salvando...' : 'Salvar'}
                </button>
              </div>
            </form>
          )}
        </div>
      </div>
    </div>
  );
}
