import { useState, useEffect } from 'react';
import { useHousehold } from '../lib/household';
import { fixedBillApi, CreateFixedBillInput, UpdateFixedBillInput } from '../lib/fixed-bill-api';
import { categoryApi } from '../lib/category-api';
import { householdApi } from '../lib/household-api';
import { FixedBill, Category, Member } from '../lib/types';
import Layout from '../components/Layout';
import Spinner from '../components/Spinner';
import ErrorAlert from '../components/ErrorAlert';

function formatCurrency(cents: number): string {
  return (cents / 100).toLocaleString('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  });
}

const emptyForm = {
  category_id: '',
  description: '',
  amount: '',
  due_day: '',
  is_shared: true,
  assigned_to: '',
  is_active: true,
};

type FormState = typeof emptyForm;

export default function FixedBills() {
  const { activeHousehold } = useHousehold();
  const [bills, setBills] = useState<FixedBill[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<FormState>(emptyForm);

  const fetchData = async () => {
    if (!activeHousehold) return;
    try {
      const [billsData, catsData, membersData] = await Promise.all([
        fixedBillApi.list(activeHousehold.id),
        categoryApi.list(activeHousehold.id),
        householdApi.listMembers(activeHousehold.id),
      ]);
      setBills(billsData);
      setCategories(catsData);
      setMembers(membersData);
      setError('');
    } catch {
      setError('Erro ao carregar dados');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    setLoading(true);
    fetchData();
  }, [activeHousehold?.id]);

  const openCreate = () => {
    setForm(emptyForm);
    setEditingId(null);
    setShowForm(true);
    setError('');
  };

  const openEdit = (bill: FixedBill) => {
    setForm({
      category_id: bill.category_id || '',
      description: bill.description,
      amount: (bill.amount_cents / 100).toFixed(2),
      due_day: String(bill.due_day),
      is_shared: bill.is_shared,
      assigned_to: bill.assigned_to || '',
      is_active: bill.is_active,
    });
    setEditingId(bill.id);
    setShowForm(true);
    setError('');
  };

  const closeForm = () => {
    setShowForm(false);
    setEditingId(null);
    setForm(emptyForm);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeHousehold) return;

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

    try {
      if (editingId) {
        const data: UpdateFixedBillInput = {
          category_id: form.category_id,
          description: form.description.trim(),
          amount_cents: amountCents,
          due_day: dueDay,
          is_shared: form.is_shared,
          assigned_to: form.is_shared ? '' : form.assigned_to,
          is_active: form.is_active,
        };
        await fixedBillApi.update(activeHousehold.id, editingId, data);
      } else {
        const data: CreateFixedBillInput = {
          category_id: form.category_id,
          description: form.description.trim(),
          amount_cents: amountCents,
          due_day: dueDay,
          is_shared: form.is_shared,
          assigned_to: form.is_shared ? '' : form.assigned_to,
        };
        await fixedBillApi.create(activeHousehold.id, data);
      }
      closeForm();
      await fetchData();
    } catch {
      setError(editingId ? 'Erro ao atualizar conta fixa.' : 'Erro ao criar conta fixa.');
    }
  };

  const handleDelete = async (id: string, desc: string) => {
    if (!activeHousehold) return;
    if (!confirm(`Excluir "${desc}"?`)) return;
    try {
      await fixedBillApi.delete(activeHousehold.id, id);
      setError('');
      await fetchData();
    } catch {
      setError('Erro ao excluir conta fixa.');
    }
  };

  if (!activeHousehold) {
    return (
      <Layout>
        <p className="text-center text-gray-500">
          Selecione ou crie uma residência primeiro.
        </p>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="mb-6 flex items-center justify-between">
        <h2 className="text-xl font-bold text-gray-900">Contas Fixas</h2>
        <button
          onClick={openCreate}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          + Nova Conta Fixa
        </button>
      </div>

      {error && (
        <ErrorAlert message={error} onDismiss={() => setError('')} />
      )}

      {/* Form modal */}
      {showForm && (
        <div className="mb-6 rounded-lg bg-white p-6 shadow">
          <h3 className="mb-4 text-lg font-semibold text-gray-800">
            {editingId ? 'Editar Conta Fixa' : 'Nova Conta Fixa'}
          </h3>
          <form onSubmit={handleSubmit} className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                Descrição *
              </label>
              <input
                type="text"
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                placeholder="Ex: Aluguel"
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                maxLength={255}
                required
              />
            </div>

            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                Categoria
              </label>
              <select
                value={form.category_id}
                onChange={(e) => setForm({ ...form, category_id: e.target.value })}
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
              >
                <option value="">Sem categoria</option>
                {categories.map((cat) => (
                  <option key={cat.id} value={cat.id}>
                    {cat.icon} {cat.name}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                Valor (R$) *
              </label>
              <input
                type="number"
                step="0.01"
                min="0.01"
                value={form.amount}
                onChange={(e) => setForm({ ...form, amount: e.target.value })}
                placeholder="0,00"
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                required
              />
            </div>

            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                Dia de vencimento *
              </label>
              <input
                type="number"
                min="1"
                max="31"
                value={form.due_day}
                onChange={(e) => setForm({ ...form, due_day: e.target.value })}
                placeholder="1-31"
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                required
              />
            </div>

            <div className="flex items-center gap-4">
              <label className="flex items-center gap-2 text-sm text-gray-700">
                <input
                  type="checkbox"
                  checked={form.is_shared}
                  onChange={(e) =>
                    setForm({ ...form, is_shared: e.target.checked, assigned_to: '' })
                  }
                  className="rounded border-gray-300"
                />
                Compartilhada
              </label>
              {editingId && (
                <label className="flex items-center gap-2 text-sm text-gray-700">
                  <input
                    type="checkbox"
                    checked={form.is_active}
                    onChange={(e) => setForm({ ...form, is_active: e.target.checked })}
                    className="rounded border-gray-300"
                  />
                  Ativa
                </label>
              )}
            </div>

            {!form.is_shared && (
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Responsável
                </label>
                <select
                  value={form.assigned_to}
                  onChange={(e) => setForm({ ...form, assigned_to: e.target.value })}
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                >
                  <option value="">Selecionar...</option>
                  {members.map((m) => (
                    <option key={m.user_id} value={m.user_id}>
                      {m.user_name}
                    </option>
                  ))}
                </select>
              </div>
            )}

            <div className="flex items-end gap-2 md:col-span-2">
              <button
                type="submit"
                className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
              >
                {editingId ? 'Salvar' : 'Criar'}
              </button>
              <button
                type="button"
                onClick={closeForm}
                className="rounded-md bg-gray-200 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-300"
              >
                Cancelar
              </button>
            </div>
          </form>
        </div>
      )}

      {/* List */}
      {loading ? (
        <Spinner />
      ) : bills.length === 0 ? (
        <p className="text-center text-gray-400">Nenhuma conta fixa cadastrada.</p>
      ) : (
        <div className="overflow-hidden rounded-lg bg-white shadow">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">
                  Descrição
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">
                  Categoria
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500">
                  Valor
                </th>
                <th className="px-6 py-3 text-center text-xs font-medium uppercase text-gray-500">
                  Vencimento
                </th>
                <th className="px-6 py-3 text-center text-xs font-medium uppercase text-gray-500">
                  Tipo
                </th>
                <th className="px-6 py-3 text-center text-xs font-medium uppercase text-gray-500">
                  Status
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500">
                  Ações
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {bills.map((bill) => {
                const assignedMember = members.find((m) => m.user_id === bill.assigned_to);
                return (
                  <tr key={bill.id} className={!bill.is_active ? 'opacity-50' : ''}>
                    <td className="whitespace-nowrap px-6 py-4 text-sm font-medium text-gray-900">
                      {bill.description}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-500">
                      {bill.category_name || '—'}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-right text-sm font-semibold text-gray-900">
                      {formatCurrency(bill.amount_cents)}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-center text-sm text-gray-500">
                      Dia {bill.due_day}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-center text-sm">
                      {bill.is_shared ? (
                        <span className="rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">
                          Compartilhada
                        </span>
                      ) : (
                        <span className="rounded-full bg-yellow-100 px-2 py-0.5 text-xs font-medium text-yellow-700">
                          {assignedMember ? assignedMember.user_name : 'Pessoal'}
                        </span>
                      )}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-center text-sm">
                      <span
                        className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                          bill.is_active
                            ? 'bg-green-100 text-green-700'
                            : 'bg-gray-100 text-gray-500'
                        }`}
                      >
                        {bill.is_active ? 'Ativa' : 'Inativa'}
                      </span>
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-right text-sm">
                      <span className="flex justify-end gap-3">
                        <button
                          onClick={() => openEdit(bill)}
                          className="text-blue-600 hover:text-blue-800"
                        >
                          Editar
                        </button>
                        <button
                          onClick={() => handleDelete(bill.id, bill.description)}
                          className="text-red-600 hover:text-red-800"
                        >
                          Excluir
                        </button>
                      </span>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </Layout>
  );
}
