import { useEffect } from 'react';
import { useGameStore } from './useGameStore';

const MOBILE_QUERY = '(max-width: 768px)';

export function useIsMobile() {
  useEffect(() => {
    const mql = window.matchMedia(MOBILE_QUERY);
    const update = () => useGameStore.getState().setIsMobile(mql.matches);
    update();
    mql.addEventListener('change', update);
    return () => mql.removeEventListener('change', update);
  }, []);
}
