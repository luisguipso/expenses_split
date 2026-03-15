import { useState, useEffect } from 'react';
import { useHousehold } from '../lib/household';
import { householdApi } from '../lib/household-api';
import { Member } from '../lib/types';

interface SplitWarning {
  showWarning: boolean;
  percentageSum: number;
  members: Member[];
}

export function useSplitWarning(): SplitWarning {
  const { activeHousehold } = useHousehold();
  const [members, setMembers] = useState<Member[]>([]);

  useEffect(() => {
    if (!activeHousehold || activeHousehold.split_mode !== 'percentage') {
      setMembers([]);
      return;
    }

    householdApi
      .listMembers(activeHousehold.id)
      .then(setMembers)
      .catch(() => setMembers([]));
  }, [activeHousehold?.id, activeHousehold?.split_mode]);

  if (!activeHousehold || activeHousehold.split_mode !== 'percentage') {
    return { showWarning: false, percentageSum: 0, members: [] };
  }

  const percentageSum = members.reduce((sum, m) => sum + m.split_percentage, 0);
  const showWarning = members.length > 0 && percentageSum !== 10000;

  return { showWarning, percentageSum, members };
}
