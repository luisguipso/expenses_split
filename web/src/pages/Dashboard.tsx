import { useState, useEffect } from 'react';
import { useHousehold } from '../lib/household';
import { summaryApi, DashboardResponse } from '../lib/summary-api';
import Layout from '../components/Layout';
import Spinner from '../components/Spinner';
import ErrorAlert from '../components/ErrorAlert';
import { useNavigate } from 'react-router-dom';

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

export default function Dashboard() {
  const { activeHousehold, isLoading } = useHousehold();
  const navigate = useNavigate();
  const [dashboard, setDashboard] = useState<DashboardResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!activeHousehold) return;
    setLoading(true);
    setError('');
    summaryApi
      .getDashboard(activeHousehold.id)
      .then(setDashboard)
      .catch(() => setError('Erro ao carregar painel.'))
      .finally(() => setLoading(false));
  }, [activeHousehold?.id]);

  if (isLoading) {
    return (
      <Layout>
        <Spinner />
      </Layout>
    );
  }

  if (!activeHousehold) {
    return (
      <Layout>
        <div className="rounded-lg bg-white p-8 text-center shadow">
          <h2 className="text-xl font-semibold text-gray-900">
            Bem-vindo ao Contas!
          </h2>
          <p className="mt-2 text-gray-500">
            Crie ou entre em uma residência para começar.
          </p>
          <button
            onClick={() => navigate('/residencias')}
            className="mt-4 rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
          >
            Gerenciar Residências
          </button>
        </div>
      </Layout>
    );
  }

  if (loading) {
    return (
      <Layout>
        <Spinner text="Carregando painel..." />
      </Layout>
    );
  }

  if (error) {
    return (
      <Layout>
        <ErrorAlert message={error} />
      </Layout>
    );
  }

  if (!dashboard) return <Layout><p className="text-gray-500">Sem dados.</p></Layout>;

  const total = dashboard.total_expenses + dashboard.total_fixed_bills;

  return (
    <Layout>
      <h2 className="mb-6 text-xl font-bold text-gray-900">
        {dashboard.household_name} — {monthNames[dashboard.month - 1]} {dashboard.year}
      </h2>

      {/* Summary cards */}
      <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-lg bg-white p-5 shadow">
          <span className="text-sm text-gray-500">Total do Mês</span>
          <p className="mt-1 text-2xl font-bold text-green-600">
            {formatCurrency(total)}
          </p>
        </div>
        <div className="rounded-lg bg-white p-5 shadow">
          <span className="text-sm text-gray-500">
            Contas Fixas ({dashboard.fixed_bill_count})
          </span>
          <p className="mt-1 text-2xl font-bold text-blue-600">
            {formatCurrency(dashboard.total_fixed_bills)}
          </p>
        </div>
        <div className="rounded-lg bg-white p-5 shadow">
          <span className="text-sm text-gray-500">
            Despesas ({dashboard.expense_count})
          </span>
          <p className="mt-1 text-2xl font-bold text-orange-600">
            {formatCurrency(dashboard.total_expenses)}
          </p>
        </div>
        <div className="rounded-lg bg-white p-5 shadow">
          <span className="text-sm text-gray-500">Compartilhado / Pessoal</span>
          <p className="mt-1 text-lg font-semibold text-gray-800">
            {formatCurrency(dashboard.total_shared)}{' '}
            <span className="text-sm font-normal text-gray-400">/</span>{' '}
            {formatCurrency(dashboard.total_personal)}
          </p>
        </div>
      </div>

      {/* Member breakdown */}
      {dashboard.member_breakdown && dashboard.member_breakdown.length > 0 ? (
        <div className="rounded-lg bg-white shadow">
          <div className="border-b border-gray-200 px-4 py-4 sm:px-6">
            <h3 className="text-lg font-semibold text-gray-800">
              Divisão por Morador
            </h3>
          </div>
          {/* Mobile cards */}
          <div className="divide-y divide-gray-200 sm:hidden">
            {dashboard.member_breakdown.map((item) => (
              <div key={item.user_id} className="px-4 py-4 space-y-1">
                <div className="flex items-center justify-between">
                  <span className="font-medium text-gray-900">{item.user_name}</span>
                  <span className="text-sm text-gray-500">{(item.proportion * 100).toFixed(1)}%</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Compartilhado</span>
                  <span className="text-gray-700">{formatCurrency(item.total_shared_cents)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Pessoal</span>
                  <span className="text-gray-700">{formatCurrency(item.total_personal_cents)}</span>
                </div>
                <div className="flex justify-between text-sm font-bold">
                  <span className="text-gray-700">Total a Pagar</span>
                  <span className="text-gray-900">{formatCurrency(item.amount_due_cents)}</span>
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
              {dashboard.member_breakdown.map((item) => (
                <tr key={item.user_id}>
                  <td className="whitespace-nowrap px-6 py-4 text-sm font-medium text-gray-900">
                    {item.user_name}
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
      ) : (
        <div className="rounded-lg bg-white p-6 shadow">
          <p className="text-gray-500">
            Configure os salários dos moradores para ver a divisão proporcional.
          </p>
          <button
            onClick={() => navigate('/membros')}
            className="mt-3 rounded-md bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700"
          >
            Configurar Salários
          </button>
        </div>
      )}

      <div className="mt-6 text-right">
        <button
          onClick={() => navigate('/resumo')}
          className="text-sm text-blue-600 hover:text-blue-800"
        >
          Ver resumo mensal detalhado →
        </button>
      </div>
    </Layout>
  );
}
