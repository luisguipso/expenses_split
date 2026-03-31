import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import api from '../lib/api';
import type { AuthResponse, AuthUser, MeResponse } from './types';

type AuthErrorResponse = {
  error?: string;
};

type AuthRequestError = {
  response?: {
    status?: number;
    data?: AuthErrorResponse;
  };
};

export class EmailNotVerifiedError extends Error {
  constructor() {
    super('email_not_verified');
    this.name = 'EmailNotVerifiedError';
  }
}

interface AuthContextType {
  user: AuthUser | null;
  token: string | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (name: string, email: string, password: string) => Promise<string>;
  verifyEmail: (email: string, code: string) => Promise<void>;
  resendCode: (email: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setToken] = useState<string | null>(
    localStorage.getItem('token')
  );
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (token) {
      api
        .get<MeResponse>('/auth/me')
        .then((res) => {
          setUser({
            id: res.data.user_id,
            name: '',
            email: res.data.email,
          });
        })
        .catch(() => {
          localStorage.removeItem('token');
          localStorage.removeItem('refresh_token');
          setToken(null);
          setUser(null);
        })
        .finally(() => setIsLoading(false));
    } else {
      setIsLoading(false);
    }
  }, [token]);

  const login = async (email: string, password: string) => {
    try {
      const res = await api.post<AuthResponse>('/auth/login', { email, password });
      const { tokens, user: userData } = res.data;
      localStorage.setItem('token', tokens.access_token);
      localStorage.setItem('refresh_token', tokens.refresh_token);
      setToken(tokens.access_token);
      setUser(userData);
    } catch (error: unknown) {
      const authError = error as AuthRequestError;
      if (
        authError.response?.status === 403 &&
        authError.response?.data?.error === 'email_not_verified'
      ) {
        throw new EmailNotVerifiedError();
      }
      throw error;
    }
  };

  const register = async (
    name: string,
    email: string,
    password: string
  ): Promise<string> => {
    await api.post('/auth/register', { name, email, password });
    return email;
  };

  const verifyEmail = async (email: string, code: string) => {
    const res = await api.post<AuthResponse>('/auth/verify-email', { email, code });
    const { tokens, user: userData } = res.data;
    localStorage.setItem('token', tokens.access_token);
    localStorage.setItem('refresh_token', tokens.refresh_token);
    setToken(tokens.access_token);
    setUser(userData);
  };

  const resendCode = async (email: string) => {
    await api.post('/auth/resend-code', { email });
  };

  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('refresh_token');
    setToken(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider
      value={{ user, token, isLoading, login, register, verifyEmail, resendCode, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
