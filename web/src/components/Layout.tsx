import { ReactNode } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../lib/auth';
import { useHousehold } from '../lib/household';

export default function Layout({ children }: { children: ReactNode }) {
  const { user, logout } = useAuth();
  const { activeHousehold, households, selectHousehold } = useHousehold();
  const location = useLocation();

  const navItems = [
    { path: '/', label: 'Painel' },
    { path: '/despesas', label: 'Despesas' },
    { path: '/contas-fixas', label: 'Contas Fixas' },
    { path: '/categorias', label: 'Categorias' },
    { path: '/membros', label: 'Moradores' },
    { path: '/residencias', label: 'Residências' },
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow">
        <div className="mx-auto max-w-7xl px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-6">
              <Link to="/" className="text-xl font-bold text-gray-900">
                Contas
              </Link>
              {activeHousehold && (
                <nav className="flex gap-1">
                  {navItems.map((item) => (
                    <Link
                      key={item.path}
                      to={item.path}
                      className={`rounded-md px-3 py-1.5 text-sm font-medium ${
                        location.pathname === item.path
                          ? 'bg-blue-100 text-blue-700'
                          : 'text-gray-600 hover:bg-gray-100'
                      }`}
                    >
                      {item.label}
                    </Link>
                  ))}
                </nav>
              )}
            </div>
            <div className="flex items-center gap-3">
              {households.length > 1 && (
                <select
                  value={activeHousehold?.id ?? ''}
                  onChange={(e) => {
                    const h = households.find((hh) => hh.id === e.target.value);
                    if (h) selectHousehold(h);
                  }}
                  className="rounded-md border border-gray-300 px-2 py-1 text-sm"
                >
                  {households.map((h) => (
                    <option key={h.id} value={h.id}>
                      {h.name}
                    </option>
                  ))}
                </select>
              )}
              {activeHousehold && households.length === 1 && (
                <span className="text-sm font-medium text-gray-700">
                  {activeHousehold.name}
                </span>
              )}
              <span className="text-sm text-gray-500">{user?.email}</span>
              <button
                onClick={logout}
                className="rounded-md bg-gray-200 px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-300"
              >
                Sair
              </button>
            </div>
          </div>
        </div>
      </header>
      <main className="mx-auto max-w-7xl px-4 py-8">{children}</main>
    </div>
  );
}
