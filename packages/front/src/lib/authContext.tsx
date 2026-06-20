import { createContext, useContext } from 'react';
import type { MeResponse } from './api';

type AuthContextValue = {
  user: MeResponse['user'];
  workspaces: MeResponse['workspaces'];
  logout: () => void;
};

export const AuthContext = createContext<AuthContextValue | null>(null);

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthContext');
  return ctx;
}
