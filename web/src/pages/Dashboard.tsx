import { useAuth } from '../lib/auth';

export default function Dashboard() {
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-6">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Contas</h1>
            <p className="mt-1 text-sm text-gray-500">
              Controle de despesas domésticas
            </p>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600">{user?.email}</span>
            <button
              onClick={logout}
              className="rounded-md bg-gray-200 px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-300"
            >
              Sair
            </button>
          </div>
        </div>
      </header>
      <main className="mx-auto max-w-7xl px-4 py-8">
        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          <div className="rounded-lg bg-white p-6 shadow">
            <h2 className="text-lg font-semibold text-gray-700">
              Total do Mês
            </h2>
            <p className="mt-2 text-3xl font-bold text-green-600">R$ 0,00</p>
          </div>
          <div className="rounded-lg bg-white p-6 shadow">
            <h2 className="text-lg font-semibold text-gray-700">
              Contas Fixas
            </h2>
            <p className="mt-2 text-3xl font-bold text-blue-600">R$ 0,00</p>
          </div>
          <div className="rounded-lg bg-white p-6 shadow">
            <h2 className="text-lg font-semibold text-gray-700">
              Despesas Variáveis
            </h2>
            <p className="mt-2 text-3xl font-bold text-orange-600">R$ 0,00</p>
          </div>
        </div>
        <div className="mt-8 rounded-lg bg-white p-6 shadow">
          <h2 className="text-lg font-semibold text-gray-700">
            Bem-vindo ao Contas!
          </h2>
          <p className="mt-2 text-gray-500">
            O sistema está funcionando. As próximas fases irão adicionar
            autenticação, gerenciamento de moradores, e controle de despesas.
          </p>
        </div>
      </main>
    </div>
  );
}
