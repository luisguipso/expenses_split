import { useState, useEffect } from 'react';
import { useHousehold } from '../lib/household';
import { householdApi } from '../lib/household-api';
import { Member } from '../lib/types';
import { useAuth } from '../lib/auth';
import Layout from '../components/Layout';
import Spinner from '../components/Spinner';
import ErrorAlert from '../components/ErrorAlert';

function formatCurrency(cents: number): string {
  return (cents / 100).toLocaleString('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  });
}

export default function Members() {
  const { user } = useAuth();
  const { activeHousehold } = useHousehold();
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [salaryInput, setSalaryInput] = useState('');
  const [copied, setCopied] = useState(false);

  const fetchMembers = async () => {
    if (!activeHousehold) return;
    try {
      const data = await householdApi.listMembers(activeHousehold.id);
      setMembers(data);
    } catch {
      setError('Erro ao carregar moradores');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMembers();
  }, [activeHousehold?.id]);

  const currentUserMember = members.find((m) => m.user_id === user?.id);
  const isAdmin = currentUserMember?.role === 'admin';

  const handleSalaryEdit = (member: Member) => {
    setEditingId(member.user_id);
    setSalaryInput((member.salary_cents / 100).toFixed(2));
    setError('');
  };

  const handleSalarySave = async (memberId: string) => {
    if (!activeHousehold) return;
    const cents = Math.round(parseFloat(salaryInput) * 100);
    if (isNaN(cents) || cents < 0) {
      setError('Valor inválido');
      return;
    }
    try {
      await householdApi.updateSalary(activeHousehold.id, memberId, cents);
      setEditingId(null);
      await fetchMembers();
    } catch {
      setError('Erro ao atualizar salário. Sem permissão.');
    }
  };

  const handleRemove = async (memberId: string, memberName: string) => {
    if (!activeHousehold) return;
    if (!confirm(`Remover ${memberName} da residência?`)) return;
    try {
      await householdApi.removeMember(activeHousehold.id, memberId);
      await fetchMembers();
    } catch {
      setError('Erro ao remover morador. Sem permissão.');
    }
  };

  const copyInviteCode = () => {
    if (!activeHousehold?.invite_code) return;
    navigator.clipboard.writeText(activeHousehold.invite_code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
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
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h2 className="text-xl font-bold text-gray-900">Moradores</h2>
        {activeHousehold.invite_code && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-500">Código:</span>
            <code className="rounded bg-gray-100 px-2 py-1 text-sm font-mono font-semibold text-gray-800">
              {activeHousehold.invite_code}
            </code>
            <button
              onClick={copyInviteCode}
              className="rounded-md bg-gray-200 px-2 py-1 text-xs text-gray-700 hover:bg-gray-300"
            >
              {copied ? '✓ Copiado' : 'Copiar'}
            </button>
          </div>
        )}
      </div>

      {error && (
        <ErrorAlert message={error} onDismiss={() => setError('')} />
      )}

      {loading ? (
        <Spinner />
      ) : (
        <div className="rounded-lg bg-white shadow">
          {/* Mobile cards */}
          <div className="divide-y divide-gray-200 sm:hidden">
            {members.map((m) => (
              <div key={m.user_id} className="px-4 py-4 space-y-2">
                <div className="flex items-center justify-between">
                  <div>
                    <span className="font-medium text-gray-900">
                      {m.user_name}
                      {m.user_id === user?.id && (
                        <span className="ml-1 text-xs text-gray-400">(você)</span>
                      )}
                    </span>
                    <p className="text-xs text-gray-500">{m.user_email}</p>
                  </div>
                  <span
                    className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                      m.role === 'admin'
                        ? 'bg-blue-100 text-blue-700'
                        : 'bg-gray-100 text-gray-600'
                    }`}
                  >
                    {m.role === 'admin' ? 'Admin' : 'Membro'}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  {editingId === m.user_id ? (
                    <div className="flex items-center gap-2">
                      <span className="text-gray-500">R$</span>
                      <input
                        type="number"
                        step="0.01"
                        min="0"
                        value={salaryInput}
                        onChange={(e) => setSalaryInput(e.target.value)}
                        className="w-28 rounded border border-gray-300 px-2 py-1 text-sm"
                        autoFocus
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') handleSalarySave(m.user_id);
                          if (e.key === 'Escape') setEditingId(null);
                        }}
                      />
                      <button onClick={() => handleSalarySave(m.user_id)} className="text-green-600">✓</button>
                      <button onClick={() => setEditingId(null)} className="text-gray-400">✕</button>
                    </div>
                  ) : (
                    <span
                      className={`text-sm cursor-pointer hover:text-blue-600 ${
                        m.salary_cents === 0 ? 'text-gray-400 italic' : 'text-gray-900'
                      }`}
                      onClick={() =>
                        (isAdmin || m.user_id === user?.id) && handleSalaryEdit(m)
                      }
                    >
                      {m.salary_cents === 0 ? 'Salário não definido' : formatCurrency(m.salary_cents)}
                    </span>
                  )}
                  <div>
                    {isAdmin && m.user_id !== user?.id && (
                      <button onClick={() => handleRemove(m.user_id, m.user_name)} className="text-sm text-red-600 hover:text-red-800">
                        Remover
                      </button>
                    )}
                    {!isAdmin && m.user_id === user?.id && (
                      <button onClick={() => handleRemove(m.user_id, m.user_name)} className="text-sm text-red-600 hover:text-red-800">
                        Sair
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
          {/* Desktop table */}
          <table className="hidden min-w-full divide-y divide-gray-200 sm:table">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">
                  Nome
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">
                  Email
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">
                  Papel
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">
                  Salário
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500">
                  Ações
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {members.map((m) => (
                <tr key={m.user_id}>
                  <td className="whitespace-nowrap px-6 py-4 text-sm font-medium text-gray-900">
                    {m.user_name}
                    {m.user_id === user?.id && (
                      <span className="ml-2 text-xs text-gray-400">(você)</span>
                    )}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-500">
                    {m.user_email}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm">
                    <span
                      className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                        m.role === 'admin'
                          ? 'bg-blue-100 text-blue-700'
                          : 'bg-gray-100 text-gray-600'
                      }`}
                    >
                      {m.role === 'admin' ? 'Admin' : 'Membro'}
                    </span>
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-900">
                    {editingId === m.user_id ? (
                      <div className="flex items-center gap-2">
                        <span className="text-gray-500">R$</span>
                        <input
                          type="number"
                          step="0.01"
                          min="0"
                          value={salaryInput}
                          onChange={(e) => setSalaryInput(e.target.value)}
                          className="w-28 rounded border border-gray-300 px-2 py-1 text-sm"
                          autoFocus
                          onKeyDown={(e) => {
                            if (e.key === 'Enter')
                              handleSalarySave(m.user_id);
                            if (e.key === 'Escape') setEditingId(null);
                          }}
                        />
                        <button
                          onClick={() => handleSalarySave(m.user_id)}
                          className="text-green-600 hover:text-green-800"
                        >
                          ✓
                        </button>
                        <button
                          onClick={() => setEditingId(null)}
                          className="text-gray-400 hover:text-gray-600"
                        >
                          ✕
                        </button>
                      </div>
                    ) : (
                      <span
                        className={`cursor-pointer hover:text-blue-600 ${
                          m.salary_cents === 0 ? 'text-gray-400 italic' : ''
                        }`}
                        onClick={() =>
                          (isAdmin || m.user_id === user?.id) &&
                          handleSalaryEdit(m)
                        }
                        title={
                          isAdmin || m.user_id === user?.id
                            ? 'Clique para editar'
                            : ''
                        }
                      >
                        {m.salary_cents === 0
                          ? 'Não definido'
                          : formatCurrency(m.salary_cents)}
                      </span>
                    )}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-right text-sm">
                    {isAdmin && m.user_id !== user?.id && (
                      <button
                        onClick={() => handleRemove(m.user_id, m.user_name)}
                        className="text-red-600 hover:text-red-800"
                      >
                        Remover
                      </button>
                    )}
                    {!isAdmin && m.user_id === user?.id && (
                      <button
                        onClick={() =>
                          handleRemove(m.user_id, m.user_name)
                        }
                        className="text-red-600 hover:text-red-800"
                      >
                        Sair
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Layout>
  );
}
