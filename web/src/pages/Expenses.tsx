import { useState, useEffect } from 'react';
import { useHousehold } from '../lib/household';
import { expenseApi, CreateExpenseInput, UpdateExpenseInput } from '../lib/expense-api';
import { categoryApi } from '../lib/category-api';
import { householdApi } from '../lib/household-api';
import { Expense, Category, Member } from '../lib/types';
import Layout from '../components/Layout';
import Spinner from '../components/Spinner';
import ErrorAlert from '../components/ErrorAlert';

function formatCurrency(cents: number): string {
  return (cents / 100).toLocaleString('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  });
}

function todayISO(): string {
  return new Date().toISOString().slice(0, 10);
}

const emptyForm = {
  category_id: '',
  description: '',
  amount: '',
  expense_date: todayISO(),
  is_shared: true,
  assigned_to: '',
};

type FormState = typeof emptyForm;

export default function Expenses() {
  const { activeHousehold } = useHousehold();
  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Filters
  const now = new Date();
  const [filterMonth, setFilterMonth] = useState(now.getMonth() + 1);
  const [filterYear, setFilterYear] = useState(now.getFullYear());
  const [filterCategory, setFilterCategory] = useState('');
  const [filterUser, setFilterUser] = useState('');

  // Form
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<FormState>(emptyForm);

  const fetchExpenses = async () => {
    if (!activeHousehold) return;
    try {
      const data = await expenseApi.list(activeHousehold.id, {
        month: filterMonth,
        year: filterYear,
        category_id: filterCategory || undefined,
        user_id: filterUser || undefined,
      });
      setExpenses(data);
      setError('');
    } catch {
      setError('Erro ao carregar despesas');
    } finally {
      setLoading(false);
    }
  };

  const fetchMeta = async () => {
    if (!activeHousehold) return;
    try {
      const [catsData, membersData] = await Promise.all([
        categoryApi.list(activeHousehold.id),
        householdApi.listMembers(activeHousehold.id),
      ]);
      setCategories(catsData);
      setMembers(membersData);
    } catch {
      // non-critical
    }
  };

  useEffect(() => {
    setLoading(true);
    fetchMeta();
  }, [activeHousehold?.id]);

  useEffect(() => {
    setLoading(true);
    fetchExpenses();
  }, [activeHousehold?.id, filterMonth, filterYear, filterCategory, filterUser]);

  const openCreate = () => {
    setForm({ ...emptyForm, expense_date: todayISO() });
    setEditingId(null);
    setShowForm(true);
    setError('');
  };

  const openEdit = (exp: Expense) => {
    setForm({
      category_id: exp.category_id || '',
      description: exp.description,
      amount: (exp.amount_cents / 100).toFixed(2),
      expense_date: exp.expense_date,
      is_shared: exp.is_shared,
      assigned_to: exp.assigned_to || '',
    });
    setEditingId(exp.id);
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
    if (isNaN(amountCents) || amountCents <= 0) {
      setError('Valor inválido');
      return;
    }

    try {
      if (editingId) {
        const data: UpdateExpenseInput = {
          category_id: form.category_id,
          description: form.description.trim(),
          amount_cents: amountCents,
          expense_date: form.expense_date,
          is_shared: form.is_shared,
          assigned_to: form.is_shared ? '' : form.assigned_to,
        };
        await expenseApi.update(activeHousehold.id, editingId, data);
      } else {
        const data: CreateExpenseInput = {
          category_id: form.category_id,
          description: form.description.trim(),
          amount_cents: amountCents,
          expense_date: form.expense_date,
          is_shared: form.is_shared,
          assigned_to: form.is_shared ? '' : form.assigned_to,
        };
        await expenseApi.create(activeHousehold.id, data);
      }
      closeForm();
      await fetchExpenses();
    } catch {
      setError(editingId ? 'Erro ao atualizar despesa.' : 'Erro ao criar despesa.');
    }
  };

  const handleDelete = async (id: string, desc: string) => {
    if (!activeHousehold) return;
    if (!confirm(`Excluir "${desc}"?`)) return;
    try {
      await expenseApi.delete(activeHousehold.id, id);
      setError('');
      await fetchExpenses();
    } catch {
      setError('Erro ao excluir despesa.');
    }
  };

  const totalCents = expenses.reduce((sum, e) => sum + e.amount_cents, 0);

  const monthNames = [
    'Janeiro', 'Fevereiro', 'Março', 'Abril', 'Maio', 'Junho',
    'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro',
  ];

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
        <h2 className="text-xl font-bold text-gray-900">Despesas</h2>
        <button
          onClick={openCreate}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          + Nova Despesa
        </button>
      </div>

      {error && (
        <ErrorAlert message={error} onDismiss={() => setError('')} />
      )}

      {/* Filters */}
      <div className="mb-6 flex flex-wrap items-end gap-3 rounded-lg bg-white p-4 shadow">
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-500">Mês</label>
          <select
            value={filterMonth}
            onChange={(e) => setFilterMonth(Number(e.target.value))}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm"
          >
            {monthNames.map((m, i) => (
              <option key={i} value={i + 1}>{m}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-500">Ano</label>
          <select
            value={filterYear}
            onChange={(e) => setFilterYear(Number(e.target.value))}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm"
          >
            {Array.from({ length: 5 }, (_, i) => now.getFullYear() - 2 + i).map((y) => (
              <option key={y} value={y}>{y}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-500">Categoria</label>
          <select
            value={filterCategory}
            onChange={(e) => setFilterCategory(e.target.value)}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm"
          >
            <option value="">Todas</option>
            {categories.map((cat) => (
              <option key={cat.id} value={cat.id}>{cat.icon} {cat.name}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-500">Pessoa</label>
          <select
            value={filterUser}
            onChange={(e) => setFilterUser(e.target.value)}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm"
          >
            <option value="">Todos</option>
            {members.map((m) => (
              <option key={m.user_id} value={m.user_id}>{m.user_name}</option>
            ))}
          </select>
        </div>
        <div className="ml-auto text-right">
          <span className="block text-xs text-gray-500">Total do período</span>
          <span className="text-lg font-bold text-green-700">{formatCurrency(totalCents)}</span>
        </div>
      </div>

      {/* Form */}
      {showForm && (
        <div className="mb-6 rounded-lg bg-white p-6 shadow">
          <h3 className="mb-4 text-lg font-semibold text-gray-800">
            {editingId ? 'Editar Despesa' : 'Nova Despesa'}
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
                placeholder="Ex: Compras do mercado"
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
                Data *
              </label>
              <input
                type="date"
                value={form.expense_date}
                onChange={(e) => setForm({ ...form, expense_date: e.target.value })}
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
      ) : expenses.length === 0 ? (
        <p className="text-center text-gray-400">
          Nenhuma despesa em {monthNames[filterMonth - 1]} {filterYear}.
        </p>
      ) : (
        <div className="overflow-hidden rounded-lg bg-white shadow">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">
                  Data
                </th>
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
                  Tipo
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">
                  Pago por
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500">
                  Ações
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {expenses.map((exp) => {
                const dateParts = exp.expense_date.split('-');
                const dateFormatted =
                  dateParts.length === 3
                    ? `${dateParts[2]}/${dateParts[1]}`
                    : exp.expense_date;
                const assignedMember = members.find((m) => m.user_id === exp.assigned_to);

                return (
                  <tr key={exp.id}>
                    <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-500">
                      {dateFormatted}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-sm font-medium text-gray-900">
                      {exp.description}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-500">
                      {exp.category_name || '—'}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-right text-sm font-semibold text-gray-900">
                      {formatCurrency(exp.amount_cents)}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-center text-sm">
                      {exp.is_shared ? (
                        <span className="rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">
                          Compartilhada
                        </span>
                      ) : (
                        <span className="rounded-full bg-yellow-100 px-2 py-0.5 text-xs font-medium text-yellow-700">
                          {assignedMember ? assignedMember.user_name : 'Pessoal'}
                        </span>
                      )}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-500">
                      {exp.paid_by_name || '—'}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-right text-sm">
                      <span className="flex justify-end gap-3">
                        <button
                          onClick={() => openEdit(exp)}
                          className="text-blue-600 hover:text-blue-800"
                        >
                          Editar
                        </button>
                        <button
                          onClick={() => handleDelete(exp.id, exp.description)}
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
