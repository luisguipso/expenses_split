import { useHousehold } from '../lib/household';
import Layout from '../components/Layout';
import { useNavigate } from 'react-router-dom';

export default function Dashboard() {
  const { activeHousehold, isLoading } = useHousehold();
  const navigate = useNavigate();

  if (isLoading) {
    return (
      <Layout>
        <p className="text-gray-500">Carregando...</p>
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

  return (
    <Layout>
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
          {activeHousehold.name}
        </h2>
        <p className="mt-2 text-gray-500">
          Os dados do painel serão preenchidos nas próximas fases. Use a
          navegação acima para gerenciar moradores.
        </p>
      </div>
    </Layout>
  );
}
