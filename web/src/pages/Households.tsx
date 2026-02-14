import { useState, FormEvent } from 'react';
import { useHousehold } from '../lib/household';
import { householdApi } from '../lib/household-api';
import { useNavigate } from 'react-router-dom';
import ErrorAlert from '../components/ErrorAlert';

export default function Households() {
  const { households, selectHousehold, refresh } = useHousehold();
  const navigate = useNavigate();

  const [showCreate, setShowCreate] = useState(false);
  const [showJoin, setShowJoin] = useState(false);
  const [name, setName] = useState('');
  const [inviteCode, setInviteCode] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const h = await householdApi.create(name);
      await refresh();
      selectHousehold(h);
      navigate('/');
    } catch {
      setError('Erro ao criar residência');
    } finally {
      setLoading(false);
    }
  };

  const handleJoin = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const h = await householdApi.join(inviteCode);
      await refresh();
      selectHousehold(h);
      navigate('/');
    } catch (err: unknown) {
      const axiosErr = err as { response?: { status?: number } };
      if (axiosErr.response?.status === 404) {
        setError('Código de convite inválido');
      } else if (axiosErr.response?.status === 409) {
        setError('Você já é membro desta residência');
      } else {
        setError('Erro ao entrar na residência');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Tem certeza que deseja excluir esta residência?')) return;
    try {
      await householdApi.delete(id);
      await refresh();
    } catch {
      setError('Erro ao excluir. Apenas admins podem excluir.');
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="mx-auto max-w-2xl px-4 py-8">
        <h1 className="mb-6 text-2xl font-bold text-gray-900">
          Suas Residências
        </h1>

        {error && (
          <ErrorAlert message={error} onDismiss={() => setError('')} />
        )}

        {/* Household list */}
        <div className="space-y-3">
          {households.map((h) => (
            <div
              key={h.id}
              className="flex items-center justify-between rounded-lg bg-white p-4 shadow"
            >
              <div>
                <h3 className="font-semibold text-gray-900">{h.name}</h3>
                <p className="text-xs text-gray-400">
                  Criado em{' '}
                  {new Date(h.created_at).toLocaleDateString('pt-BR')}
                </p>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => {
                    selectHousehold(h);
                    navigate('/');
                  }}
                  className="rounded-md bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700"
                >
                  Selecionar
                </button>
                <button
                  onClick={() => handleDelete(h.id)}
                  className="rounded-md bg-red-100 px-3 py-1.5 text-sm text-red-600 hover:bg-red-200"
                >
                  Excluir
                </button>
              </div>
            </div>
          ))}

          {households.length === 0 && (
            <p className="py-8 text-center text-gray-500">
              Você ainda não tem nenhuma residência. Crie uma ou entre com um
              código de convite.
            </p>
          )}
        </div>

        {/* Actions */}
        <div className="mt-6 flex flex-col gap-3 sm:flex-row">
          <button
            onClick={() => {
              setShowCreate(!showCreate);
              setShowJoin(false);
              setError('');
            }}
            className="rounded-md bg-green-600 px-4 py-2 text-sm text-white hover:bg-green-700"
          >
            + Criar Residência
          </button>
          <button
            onClick={() => {
              setShowJoin(!showJoin);
              setShowCreate(false);
              setError('');
            }}
            className="rounded-md bg-purple-600 px-4 py-2 text-sm text-white hover:bg-purple-700"
          >
            Entrar com Código
          </button>
        </div>

        {/* Create form */}
        {showCreate && (
          <form
            onSubmit={handleCreate}
            className="mt-4 rounded-lg bg-white p-4 shadow"
          >
            <h3 className="mb-3 font-semibold text-gray-700">
              Nova Residência
            </h3>
            <div className="flex gap-3">
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Nome da residência"
                required
                className="flex-1 rounded-md border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
              <button
                type="submit"
                disabled={loading}
                className="rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
              >
                {loading ? 'Criando...' : 'Criar'}
              </button>
            </div>
          </form>
        )}

        {/* Join form */}
        {showJoin && (
          <form
            onSubmit={handleJoin}
            className="mt-4 rounded-lg bg-white p-4 shadow"
          >
            <h3 className="mb-3 font-semibold text-gray-700">
              Entrar em uma Residência
            </h3>
            <div className="flex gap-3">
              <input
                type="text"
                value={inviteCode}
                onChange={(e) => setInviteCode(e.target.value)}
                placeholder="Código de convite"
                required
                className="flex-1 rounded-md border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
              <button
                type="submit"
                disabled={loading}
                className="rounded-md bg-purple-600 px-4 py-2 text-white hover:bg-purple-700 disabled:opacity-50"
              >
                {loading ? 'Entrando...' : 'Entrar'}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
