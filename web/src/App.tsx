import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './lib/auth';
import { HouseholdProvider } from './lib/household';
import ErrorBoundary from './components/ErrorBoundary';
import Spinner from './components/Spinner';
import Login from './pages/Login';
import Register from './pages/Register';
import VerifyEmail from './pages/VerifyEmail';
import PasswordRecover from './pages/PasswordRecover';
import Households from './pages/Households';
import Members from './pages/Members';
import Categories from './pages/Categories';
import FixedBills from './pages/FixedBills';
import Expenses from './pages/Expenses';
import Summary from './pages/Summary';

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { user, isLoading } = useAuth();
  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Spinner />
      </div>
    );
  }
  return user ? <>{children}</> : <Navigate to="/login" />;
}

function PublicRoute({ children }: { children: React.ReactNode }) {
  const { user, isLoading } = useAuth();
  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Spinner />
      </div>
    );
  }
  return user ? <Navigate to="/" /> : <>{children}</>;
}

function App() {
  return (
    <AuthProvider>
      <HouseholdProvider>
        <ErrorBoundary>
        <BrowserRouter>
          <Routes>
            <Route
              path="/"
              element={
                <PrivateRoute>
                  <Summary />
                </PrivateRoute>
              }
            />
            <Route
              path="/residencias"
              element={
                <PrivateRoute>
                  <Households />
                </PrivateRoute>
              }
            />
            <Route
              path="/membros"
              element={
                <PrivateRoute>
                  <Members />
                </PrivateRoute>
              }
            />
            <Route
              path="/categorias"
              element={
                <PrivateRoute>
                  <Categories />
                </PrivateRoute>
              }
            />
            <Route
              path="/contas-fixas"
              element={
                <PrivateRoute>
                  <FixedBills />
                </PrivateRoute>
              }
            />
            <Route
              path="/despesas"
              element={
                <PrivateRoute>
                  <Expenses />
                </PrivateRoute>
              }
            />
            <Route
              path="/resumo"
              element={<Navigate to="/" replace />}
            />
            <Route
              path="/verificar-email"
              element={
                <PublicRoute>
                  <VerifyEmail />
                </PublicRoute>
              }
            />
            <Route
              path="/password-recover"
              element={
                <PublicRoute>
                  <PasswordRecover />
                </PublicRoute>
              }
            />
            <Route
              path="/login"
              element={
                <PublicRoute>
                  <Login />
                </PublicRoute>
              }
            />
            <Route
              path="/register"
              element={
                <PublicRoute>
                  <Register />
                </PublicRoute>
              }
            />
          </Routes>
        </BrowserRouter>
        </ErrorBoundary>
      </HouseholdProvider>
    </AuthProvider>
  );
}

export default App;
