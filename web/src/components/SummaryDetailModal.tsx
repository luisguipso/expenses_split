import { useEffect, useState } from 'react';
import { summaryApi, SummaryDetailResponse } from '../lib/summary-api';
import Spinner from './Spinner';

function formatCurrency(cents: number): string {
  return ((cents ?? 0) / 100).toLocaleString('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  });
}

interface Props {
  householdId: string;
  year: number;
  month: number;
  userId: string;
  userName: string;
  onClose: () => void;
}

export default function SummaryDetailModal({
  householdId,
  year,
  month,
  userId,
  userName,
  onClose,
}: Props) {
  const [detail, setDetail] = useState<SummaryDetailResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    setError('');
    summaryApi
      .getUserDetail(householdId, year, month, userId)
      .then(setDetail)
      .catch(() => setError('Erro ao carregar detalhamento.'))
      .finally(() => setLoading(false));
  }, [householdId, year, month, userId]);

  const monthNames = [
    'Janeiro', 'Fevereiro', 'Março', 'Abril', 'Maio', 'Junho',
    'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro',
  ];

  const sharedItems = detail?.items.filter((i) => i.is_shared) ?? [];
  const personalItems = detail?.items.filter((i) => !i.is_shared) ?? [];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={onClose}>
      <div
        className="relative max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-lg bg-white shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="sticky top-0 z-10 flex items-center justify-between border-b bg-white px-6 py-4">
          <div>
            <h3 className="text-lg font-bold text-gray-900">{userName}</h3>
            <p className="text-sm text-gray-500">
              {monthNames[month - 1]} {year}
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

        {/* Body */}
        <div className="px-6 py-4">
          {loading ? (
            <Spinner text="Carregando..." />
          ) : error ? (
            <p className="text-center text-red-500">{error}</p>
          ) : detail ? (
            <>
              {/* Shared items */}
              {sharedItems.length > 0 && (
                <div className="mb-6">
                  <h4 className="mb-2 text-sm font-semibold uppercase text-gray-500">
                    Compartilhado
                  </h4>
                  {/* Mobile */}
                  <div className="space-y-2 sm:hidden">
                    {sharedItems.map((item, i) => (
                      <div key={i} className="rounded border border-gray-100 p-3 space-y-1">
                        <div className="flex items-center justify-between">
                          <span className="font-medium text-gray-900 text-sm">{item.description}</span>
                          <span className="text-xs text-gray-400">
                            {item.type === 'fixed_bill' ? 'Conta fixa' : 'Despesa'}
                          </span>
                        </div>
                        {item.category_name && (
                          <p className="text-xs text-gray-400">{item.category_name}</p>
                        )}
                        <div className="flex justify-between text-sm">
                          <span className="text-gray-500">Total</span>
                          <span className="text-gray-600">{formatCurrency(item.total_cents)}</span>
                        </div>
                        <div className="flex justify-between text-sm font-semibold">
                          <span className="text-gray-700">Sua parte ({(item.proportion * 100).toFixed(1)}%)</span>
                          <span className="text-gray-900">{formatCurrency(item.user_share_cents)}</span>
                        </div>
                        {item.paid_by_name && (
                          <p className="text-xs text-gray-400">Pago por: {item.paid_by_name}</p>
                        )}
                      </div>
                    ))}
                  </div>
                  {/* Desktop */}
                  <table className="hidden w-full text-sm sm:table">
                    <thead>
                      <tr className="border-b text-left text-xs uppercase text-gray-400">
                        <th className="pb-2 font-medium">Descrição</th>
                        <th className="pb-2 font-medium">Tipo</th>
                        <th className="pb-2 font-medium">Categoria</th>
                        <th className="pb-2 text-right font-medium">Total</th>
                        <th className="pb-2 text-right font-medium">Sua Parte</th>
                        <th className="pb-2 text-right font-medium">%</th>
                        <th className="pb-2 font-medium">Pago por</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      {sharedItems.map((item, i) => (
                        <tr key={i}>
                          <td className="py-2 text-gray-900">{item.description}</td>
                          <td className="py-2 text-gray-500">
                            {item.type === 'fixed_bill' ? 'Conta fixa' : 'Despesa'}
                          </td>
                          <td className="py-2 text-gray-500">{item.category_name || '—'}</td>
                          <td className="py-2 text-right text-gray-600">{formatCurrency(item.total_cents)}</td>
                          <td className="py-2 text-right font-semibold text-gray-900">
                            {formatCurrency(item.user_share_cents)}
                          </td>
                          <td className="py-2 text-right text-gray-500">
                            {(item.proportion * 100).toFixed(1)}%
                          </td>
                          <td className="py-2 text-gray-500">{item.paid_by_name || '—'}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {/* Personal items */}
              {personalItems.length > 0 && (
                <div className="mb-6">
                  <h4 className="mb-2 text-sm font-semibold uppercase text-gray-500">
                    Pessoal
                  </h4>
                  {/* Mobile */}
                  <div className="space-y-2 sm:hidden">
                    {personalItems.map((item, i) => (
                      <div key={i} className="rounded border border-gray-100 p-3 space-y-1">
                        <div className="flex items-center justify-between">
                          <span className="font-medium text-gray-900 text-sm">{item.description}</span>
                          <span className="text-xs text-gray-400">
                            {item.type === 'fixed_bill' ? 'Conta fixa' : 'Despesa'}
                          </span>
                        </div>
                        {item.category_name && (
                          <p className="text-xs text-gray-400">{item.category_name}</p>
                        )}
                        <div className="flex justify-between text-sm font-semibold">
                          <span className="text-gray-700">Valor</span>
                          <span className="text-gray-900">{formatCurrency(item.user_share_cents)}</span>
                        </div>
                        {item.paid_by_name && (
                          <p className="text-xs text-gray-400">Pago por: {item.paid_by_name}</p>
                        )}
                      </div>
                    ))}
                  </div>
                  {/* Desktop */}
                  <table className="hidden w-full text-sm sm:table">
                    <thead>
                      <tr className="border-b text-left text-xs uppercase text-gray-400">
                        <th className="pb-2 font-medium">Descrição</th>
                        <th className="pb-2 font-medium">Tipo</th>
                        <th className="pb-2 font-medium">Categoria</th>
                        <th className="pb-2 text-right font-medium">Valor</th>
                        <th className="pb-2 font-medium">Pago por</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      {personalItems.map((item, i) => (
                        <tr key={i}>
                          <td className="py-2 text-gray-900">{item.description}</td>
                          <td className="py-2 text-gray-500">
                            {item.type === 'fixed_bill' ? 'Conta fixa' : 'Despesa'}
                          </td>
                          <td className="py-2 text-gray-500">{item.category_name || '—'}</td>
                          <td className="py-2 text-right font-semibold text-gray-900">
                            {formatCurrency(item.user_share_cents)}
                          </td>
                          <td className="py-2 text-gray-500">{item.paid_by_name || '—'}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {detail.items.length === 0 && (
                <p className="text-center text-gray-400 py-6">
                  Nenhum item encontrado para este mês.
                </p>
              )}
            </>
          ) : null}
        </div>

        {/* Footer totals */}
        {detail && !loading && !error && (
          <div className="sticky bottom-0 border-t bg-gray-50 px-6 py-4 space-y-1">
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">Compartilhado</span>
              <span className="text-gray-700">{formatCurrency(detail.total_shared_cents)}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">Pessoal</span>
              <span className="text-gray-700">{formatCurrency(detail.total_personal_cents)}</span>
            </div>
            <div className="flex justify-between text-sm font-bold border-t border-gray-200 pt-1">
              <span className="text-gray-700">Total a Pagar</span>
              <span className="text-gray-900">{formatCurrency(detail.amount_due_cents)}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">Total Pago</span>
              <span className="text-gray-700">{formatCurrency(detail.total_paid_cents)}</span>
            </div>
            <div className="flex justify-between text-sm font-bold">
              <span className="text-gray-700">Saldo</span>
              <span
                className={
                  (detail.balance_cents ?? 0) > 0
                    ? 'text-green-600'
                    : (detail.balance_cents ?? 0) < 0
                      ? 'text-red-600'
                      : 'text-gray-500'
                }
              >
                {(detail.balance_cents ?? 0) > 0 ? '+' : ''}
                {formatCurrency(detail.balance_cents)}
              </span>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
