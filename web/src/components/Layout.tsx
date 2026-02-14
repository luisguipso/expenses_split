import { ReactNode, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../lib/auth';
import { useHousehold } from '../lib/household';

export default function Layout({ children }: { children: ReactNode }) {
  const { user, logout } = useAuth();
  const { activeHousehold, households, selectHousehold } = useHousehold();
  const location = useLocation();
  const [menuOpen, setMenuOpen] = useState(false);

  const navItems = [
    { path: '/', label: 'Painel' },
    { path: '/despesas', label: 'Despesas' },
    { path: '/contas-fixas', label: 'Contas Fixas' },
    { path: '/resumo', label: 'Resumo' },
    { path: '/categorias', label: 'Categorias' },
    { path: '/membros', label: 'Moradores' },
    { path: '/residencias', label: 'Residências' },
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow">
        <div className="mx-auto max-w-7xl px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link to="/" className="text-xl font-bold text-gray-900">
                Contas
              </Link>
              {/* Desktop nav */}
              {activeHousehold && (
                <nav className="hidden gap-1 lg:flex">
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
                  className="hidden rounded-md border border-gray-300 px-2 py-1 text-sm sm:block"
                >
                  {households.map((h) => (
                    <option key={h.id} value={h.id}>
                      {h.name}
                    </option>
                  ))}
                </select>
              )}
              {activeHousehold && households.length === 1 && (
                <span className="hidden text-sm font-medium text-gray-700 sm:inline">
                  {activeHousehold.name}
                </span>
              )}
              <span className="hidden text-sm text-gray-500 sm:inline">{user?.email}</span>
              <button
                onClick={logout}
                className="rounded-md bg-gray-200 px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-300"
              >
                Sair
              </button>
              {/* Mobile hamburger */}
              {activeHousehold && (
                <button
                  onClick={() => setMenuOpen(!menuOpen)}
                  className="rounded-md p-1.5 text-gray-600 hover:bg-gray-100 lg:hidden"
                  aria-label="Menu"
                >
                  <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    {menuOpen ? (
                      <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                    ) : (
                      <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />
                    )}
                  </svg>
                </button>
              )}
            </div>
          </div>
        </div>

        {/* Mobile nav drawer */}
        {menuOpen && activeHousehold && (
          <div className="border-t border-gray-200 lg:hidden">
            <nav className="mx-auto max-w-7xl space-y-1 px-4 py-3">
              {navItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  onClick={() => setMenuOpen(false)}
                  className={`block rounded-md px-3 py-2 text-sm font-medium ${
                    location.pathname === item.path
                      ? 'bg-blue-100 text-blue-700'
                      : 'text-gray-600 hover:bg-gray-100'
                  }`}
                >
                  {item.label}
                </Link>
              ))}
              {/* Mobile-only: household selector & user info */}
              <div className="border-t border-gray-100 pt-2 mt-2 space-y-2">
                {households.length > 1 && (
                  <select
                    value={activeHousehold?.id ?? ''}
                    onChange={(e) => {
                      const h = households.find((hh) => hh.id === e.target.value);
                      if (h) selectHousehold(h);
                    }}
                    className="w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm sm:hidden"
                  >
                    {households.map((h) => (
                      <option key={h.id} value={h.id}>
                        {h.name}
                      </option>
                    ))}
                  </select>
                )}
                <p className="px-3 text-xs text-gray-400 sm:hidden">{user?.email}</p>
              </div>
            </nav>
          </div>
        )}
      </header>
      <main className="mx-auto max-w-7xl px-4 py-6 sm:py-8">{children}</main>
    </div>
  );
}
