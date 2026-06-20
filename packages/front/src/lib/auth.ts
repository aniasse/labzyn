// Auth token persistence (localStorage).

const KEY = 'labzyn_token';

export const tokenStore = {
  get: () => localStorage.getItem(KEY),
  set: (t: string) => localStorage.setItem(KEY, t),
  clear: () => localStorage.removeItem(KEY),
};

export type AuthUser = {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
};

export type AuthResponse = {
  token: string;
  user: AuthUser;
  workspaces: { id: string; displayName: string; subdomain: string }[];
  currentWorkspaceId: string;
};
