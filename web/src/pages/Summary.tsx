import { useState, useEffect } from 'react';
import { useHousehold } from '../lib/household';
import { summaryApi, SummaryResponse } from '../lib/summary-api';
import Layout from '../components/Layout';
import Spinner from '../components/Spinner';
import ErrorAlert from '../components/ErrorAlert';
import SummaryDetailModal from '../components/SummaryDetailModal';

function formatCurrency(cents: number): string {
  return ((cents ?? 0) / 100).toLocaleString('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  });
}

const monthNames = [
  'Janeiro', 'Fevereiro', 'Março', 'Abril', 'Maio', 'Junho',
  'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro',
];

export default function Summary() {
  const { activeHousehold } = useHousehold();
  const now = new Date();
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [year, setYear] = useState(now.getFullYear());
  const [summary, setSummary] = useState<SummaryResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedUser, setSelectedUser] = useState<{ id: string; name: string } | null>(null);

  const fetchSummary = async () => {
    if (!activeHousehold) return;
    setLoading(true);
    setError('');
    try {
      const data = await summaryApi.getSummary(activeHousehold.id, year, month);
      setSummary(data);
    } catch (err: unknown) {
      setSummary(null);
      const msg =
        err && typeof err === 'object' && 'response' in err
          ? (err as { response?: { data?: { error?: string } } }).response?.data?.error
          : undefined;
      if (msg === 'no members have salary configured') {
        setError('Nenhum morador tem salário configurado. Defina os salários na página de Moradores.');
      } else {
        setError('Erro ao gerar resumo mensal.');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSummary();
  }, [activeHousehold?.id, month, year]);

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
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h2 className="text-xl font-bold text-gray-900">Resumo Mensal</h2>
        <div className="flex items-center gap-3">
          <select
            value={month}
            onChange={(e) => setMonth(Number(e.target.value))}
            className="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm sm:flex-none"
          >
            {monthNames.map((m, i) => (
              <option key={i} value={i + 1}>{m}</option>
            ))}
          </select>
          <select
            value={year}
            onChange={(e) => setYear(Number(e.target.value))}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm"
          >
            {Array.from({ length: 5 }, (_, i) => now.getFullYear() - 2 + i).map((y) => (
              <option key={y} value={y}>{y}</option>
            ))}
          </select>
        </div>
      </div>

      {error && (
        <ErrorAlert message={error} onDismiss={() => setError('')} />
      )}

      {loading ? (
        <Spinner text="Calculando..." />
      ) : summary ? (
        <>
          {/* Totals */}
          <div className="mb-6 grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="rounded-lg bg-white p-5 shadow">
              <span className="text-sm text-gray-500">Total Compartilhado</span>
              <p className="mt-1 text-2xl font-bold text-green-600">
                {formatCurrency(summary.total_shared_cents)}
              </p>
            </div>
            <div className="rounded-lg bg-white p-5 shadow">
              <span className="text-sm text-gray-500">Total Geral</span>
              <p className="mt-1 text-2xl font-bold text-gray-900">
                {formatCurrency(summary.total_all_cents)}
              </p>
            </div>
          </div>

          {/* Per-member breakdown */}
          <div className="rounded-lg bg-white shadow">
            {/* Mobile cards */}
            <div className="divide-y divide-gray-200 sm:hidden">
              {summary.items.map((item) => (
                <div
                  key={item.user_id}
                  className="px-4 py-4 space-y-1 cursor-pointer hover:bg-gray-50 transition-colors"
                  onClick={() => setSelectedUser({ id: item.user_id, name: item.user_name })}
                >
                  <div className="flex items-center justify-between">
                    <span className="font-medium text-gray-900">{item.user_name}</span>
                    <span className="text-sm text-gray-500">{(item.proportion * 100).toFixed(1)}%</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Salário</span>
                    <span className="text-gray-700">{formatCurrency(item.salary_cents)}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Compartilhado</span>
                    <span className="text-gray-700">{formatCurrency(item.total_shared_cents)}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Pessoal</span>
                    <span className="text-gray-700">{formatCurrency(item.total_personal_cents)}</span>
                  </div>
                  <div className="flex justify-between text-sm font-bold border-t border-gray-100 pt-1">
                    <span className="text-gray-700">Total a Pagar</span>
                    <span className="text-gray-900">{formatCurrency(item.amount_due_cents)}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Pago</span>
                    <span className="text-gray-700">{formatCurrency(item.total_paid_cents)}</span>
                  </div>
                  <div className="flex justify-between text-sm font-semibold">
                    <span className="text-gray-500">Saldo</span>
                    <span className={(item.balance_cents ?? 0) > 0 ? 'text-green-600' : (item.balance_cents ?? 0) < 0 ? 'text-red-600' : 'text-gray-500'}>
                      {(item.balance_cents ?? 0) > 0 ? '+' : ''}{formatCurrency(item.balance_cents)}
                    </span>
                  </div>
                </div>
              ))}
            </div>
            {/* Desktop table */}
            <table className="hidden min-w-full divide-y divide-gray-200 sm:table">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">
                    Morador
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500">
                    Salário
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500">
                    Proporção
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500">
                    Compartilhado
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500">
                    Pessoal
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500">
                    Total a Pagar
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500">
                    Pago
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500">
                    Saldo
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {summary.items.map((item) => (
                  <tr
                    key={item.user_id}
                    className="cursor-pointer hover:bg-gray-50 transition-colors"
                    onClick={() => setSelectedUser({ id: item.user_id, name: item.user_name })}
                  >
                    <td className="whitespace-nowrap px-6 py-4 text-sm font-medium text-gray-900">
                      {item.user_name}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-right text-sm text-gray-500">
                      {formatCurrency(item.salary_cents)}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-right text-sm text-gray-500">
                      {(item.proportion * 100).toFixed(1)}%
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-right text-sm text-gray-700">
                      {formatCurrency(item.total_shared_cents)}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-right text-sm text-gray-700">
                      {formatCurrency(item.total_personal_cents)}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-right text-sm font-bold text-gray-900">
                      {formatCurrency(item.amount_due_cents)}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-right text-sm text-gray-700">
                      {formatCurrency(item.total_paid_cents)}
                    </td>
                    <td className={`whitespace-nowrap px-6 py-4 text-right text-sm font-semibold ${(item.balance_cents ?? 0) > 0 ? 'text-green-600' : (item.balance_cents ?? 0) < 0 ? 'text-red-600' : 'text-gray-500'}`}>
                      {(item.balance_cents ?? 0) > 0 ? '+' : ''}{formatCurrency(item.balance_cents)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Settlements */}
          {summary.settlements && summary.settlements.length > 0 && (
            <div className="mt-6 rounded-lg bg-white shadow">
              <div className="border-b border-gray-200 px-4 py-4 sm:px-6">
                <h3 className="text-lg font-semibold text-gray-800">
                  Acertos
                </h3>
              </div>
              <div className="divide-y divide-gray-200">
                {summary.settlements.map((s, i) => (
                  <div key={i} className="flex items-center justify-between px-4 py-3 sm:px-6">
                    <div className="flex items-center gap-2 text-sm">
                      <span className="font-medium text-red-600">{s.from_user_name}</span>
                      <span className="text-gray-400">→</span>
                      <span className="font-medium text-green-600">{s.to_user_name}</span>
                    </div>
                    <span className="text-sm font-bold text-gray-900">
                      {formatCurrency(s.amount_cents)}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          <p className="mt-4 text-right text-xs text-gray-400">
            Gerado em: {new Date(summary.generated_at).toLocaleString('pt-BR')}
          </p>
        </>
      ) : null}

      {selectedUser && activeHousehold && (
        <SummaryDetailModal
          householdId={activeHousehold.id}
          year={year}
          month={month}
          userId={selectedUser.id}
          userName={selectedUser.name}
          onClose={() => setSelectedUser(null)}
        />
      )}
    </Layout>
  );
}
