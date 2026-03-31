import { useState, useEffect, FormEvent } from 'react';
import { useAuth } from '../lib/auth';
import { useNavigate, useLocation } from 'react-router-dom';
import ErrorAlert from '../components/ErrorAlert';

interface VerifyEmailLocationState {
  email?: string;
}

interface VerifyEmailRequestError {
  response?: {
    status?: number;
    data?: {
      error?: string;
    };
  };
}

export default function VerifyEmail() {
  const { verifyEmail, resendCode } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const email = (location.state as VerifyEmailLocationState | null)?.email;

  const [code, setCode] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [resending, setResending] = useState(false);
  const [cooldown, setCooldown] = useState(0);

  useEffect(() => {
    if (!email) {
      navigate('/register', { replace: true });
    }
  }, [email, navigate]);

  useEffect(() => {
    if (cooldown <= 0) return;
    const timer = setTimeout(() => setCooldown(cooldown - 1), 1000);
    return () => clearTimeout(timer);
  }, [cooldown]);

  const handleResend = async () => {
    if (!email || cooldown > 0) return;
    setResending(true);
    setError('');
    try {
      await resendCode(email);
      setCooldown(60);
    } catch {
      setError('Erro ao reenviar código. Tente novamente.');
    } finally {
      setResending(false);
    }
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!email) return;
    setError('');
    setLoading(true);
    try {
      await verifyEmail(email, code);
      navigate('/');
    } catch (err: unknown) {
      const error = err as VerifyEmailRequestError;
      const status = error.response?.status;
      const serverError = error.response?.data?.error;
      if (status === 400 && serverError === 'invalid_code') {
        setError('Código inválido. Verifique e tente novamente.');
      } else if (status === 400 && serverError === 'code_expired') {
        setError('Código expirado. Solicite um novo código.');
      } else {
        setError('Erro ao verificar email. Tente novamente.');
      }
    } finally {
      setLoading(false);
    }
  };

  if (!email) return null;

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="w-full max-w-md rounded-lg bg-white p-8 shadow-md">
        <h1 className="mb-2 text-center text-2xl font-bold text-gray-900">
          Verificar Email
        </h1>
        <p className="mb-6 text-center text-sm text-gray-600">
          Enviamos um código de 6 dígitos para{' '}
          <span className="font-medium text-gray-900">{email}</span>
        </p>
        <ErrorAlert message={error} onDismiss={() => setError('')} />
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700">
              Código de Verificação
            </label>
            <input
              type="text"
              inputMode="numeric"
              pattern="[0-9]{6}"
              maxLength={6}
              value={code}
              onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
              required
              className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-center text-2xl tracking-widest focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder="000000"
            />
          </div>
          <button
            type="submit"
            disabled={loading || code.length !== 6}
            className="w-full rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {loading ? 'Verificando...' : 'Verificar'}
          </button>
        </form>
        <div className="mt-4 text-center">
          <button
            type="button"
            onClick={handleResend}
            disabled={resending || cooldown > 0}
            className="text-sm text-blue-600 hover:underline disabled:text-gray-400 disabled:no-underline"
          >
            {cooldown > 0
              ? `Reenviar código (${cooldown}s)`
              : resending
                ? 'Reenviando...'
                : 'Reenviar código'}
          </button>
        </div>
      </div>
    </div>
  );
}
