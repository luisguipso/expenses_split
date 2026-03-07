import { useState, FormEvent } from 'react';
import { useAuth, EmailNotVerifiedError } from '../lib/auth';
import { useNavigate, Link } from 'react-router-dom';
import api from '../lib/api';
import ErrorAlert from '../components/ErrorAlert';

export default function Login() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [emailNotVerified, setEmailNotVerified] = useState(false);
  const [loading, setLoading] = useState(false);
  const [resending, setResending] = useState(false);

  const handleResendCode = async () => {
    setResending(true);
    try {
      await api.post('/auth/resend-code', { email });
      navigate('/verificar-email', { state: { email } });
    } catch {
      setError('Erro ao reenviar código. Tente novamente.');
    } finally {
      setResending(false);
    }
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setEmailNotVerified(false);
    setLoading(true);
    try {
      await login(email, password);
      navigate('/');
    } catch (err) {
      if (err instanceof EmailNotVerifiedError) {
        setError('Seu email ainda não foi verificado.');
        setEmailNotVerified(true);
      } else {
        setError('Email ou senha inválidos');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="w-full max-w-md rounded-lg bg-white p-8 shadow-md">
        <h1 className="mb-6 text-center text-2xl font-bold text-gray-900">
          Entrar no Contas
        </h1>
        <ErrorAlert message={error} onDismiss={() => { setError(''); setEmailNotVerified(false); }} />
        {emailNotVerified && (
          <button
            type="button"
            onClick={handleResendCode}
            disabled={resending}
            className="mb-4 w-full rounded-md bg-yellow-500 px-4 py-2 text-white hover:bg-yellow-600 disabled:opacity-50"
          >
            {resending ? 'Reenviando...' : 'Reenviar código'}
          </button>
        )}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700">
              Email
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder="seu@email.com"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700">
              Senha
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={6}
              className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder="••••••"
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {loading ? 'Entrando...' : 'Entrar'}
          </button>
        </form>
        <p className="mt-3 text-center text-sm">
          <Link to="/password-recover" className="text-blue-600 hover:underline">
            Esqueceu a senha?
          </Link>
        </p>
        <p className="mt-2 text-center text-sm text-gray-600">
          Não tem conta?{' '}
          <Link to="/register" className="text-blue-600 hover:underline">
            Cadastre-se
          </Link>
        </p>
      </div>
    </div>
  );
}
