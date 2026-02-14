import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import api from '../lib/api';

interface User {
  id: string;
  name: string;
  email: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (name: string, email: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(
    localStorage.getItem('token')
  );
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (token) {
      api
        .get('/auth/me')
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
    const res = await api.post('/auth/login', { email, password });
    const { tokens, user: userData } = res.data;
    localStorage.setItem('token', tokens.access_token);
    localStorage.setItem('refresh_token', tokens.refresh_token);
    setToken(tokens.access_token);
    setUser(userData);
  };

  const register = async (
    name: string,
    email: string,
    password: string
  ) => {
    const res = await api.post('/auth/register', { name, email, password });
    const { tokens, user: userData } = res.data;
    localStorage.setItem('token', tokens.access_token);
    localStorage.setItem('refresh_token', tokens.refresh_token);
    setToken(tokens.access_token);
    setUser(userData);
  };

  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('refresh_token');
    setToken(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider
      value={{ user, token, isLoading, login, register, logout }}
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
