import { useState, useEffect } from 'react';
import { useHousehold } from '../lib/household';
import { summaryApi, SummaryResponse } from '../lib/summary-api';
import Layout from '../components/Layout';

function formatCurrency(cents: number): string {
  return (cents / 100).toLocaleString('pt-BR', {
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
      <div className="mb-6 flex items-center justify-between">
        <h2 className="text-xl font-bold text-gray-900">Resumo Mensal</h2>
        <div className="flex items-center gap-3">
          <select
            value={month}
            onChange={(e) => setMonth(Number(e.target.value))}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm"
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
        <div className="mb-4 rounded bg-red-50 p-3 text-sm text-red-600">
          {error}
        </div>
      )}

      {loading ? (
        <p className="text-gray-500">Calculando...</p>
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
          <div className="overflow-hidden rounded-lg bg-white shadow">
            <table className="min-w-full divide-y divide-gray-200">
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
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {summary.items.map((item) => (
                  <tr key={item.user_id}>
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
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <p className="mt-4 text-right text-xs text-gray-400">
            Gerado em: {new Date(summary.generated_at).toLocaleString('pt-BR')}
          </p>
        </>
      ) : null}
    </Layout>
  );
}
