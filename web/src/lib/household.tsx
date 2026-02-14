import {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
  useCallback,
} from 'react';
import { householdApi } from './household-api';
import { Household } from './types';
import { useAuth } from './auth';

interface HouseholdContextType {
  households: Household[];
  activeHousehold: Household | null;
  isLoading: boolean;
  selectHousehold: (h: Household) => void;
  refresh: () => Promise<void>;
}

const HouseholdContext = createContext<HouseholdContextType | null>(null);

export function HouseholdProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth();
  const [households, setHouseholds] = useState<Household[]>([]);
  const [activeHousehold, setActiveHousehold] = useState<Household | null>(
    null
  );
  const [isLoading, setIsLoading] = useState(true);

  const refresh = useCallback(async () => {
    if (!user) {
      setHouseholds([]);
      setActiveHousehold(null);
      setIsLoading(false);
      return;
    }
    try {
      const list = await householdApi.list();
      setHouseholds(list);

      const savedId = localStorage.getItem('active_household_id');
      const saved = list.find((h) => h.id === savedId);
      if (saved) {
        setActiveHousehold(saved);
      } else if (list.length > 0) {
        setActiveHousehold(list[0]);
        localStorage.setItem('active_household_id', list[0].id);
      } else {
        setActiveHousehold(null);
        localStorage.removeItem('active_household_id');
      }
    } catch {
      setHouseholds([]);
      setActiveHousehold(null);
    } finally {
      setIsLoading(false);
    }
  }, [user]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const selectHousehold = (h: Household) => {
    setActiveHousehold(h);
    localStorage.setItem('active_household_id', h.id);
  };

  return (
    <HouseholdContext.Provider
      value={{ households, activeHousehold, isLoading, selectHousehold, refresh }}
    >
      {children}
    </HouseholdContext.Provider>
  );
}

export function useHousehold() {
  const context = useContext(HouseholdContext);
  if (!context) {
    throw new Error('useHousehold must be used within a HouseholdProvider');
  }
  return context;
}
