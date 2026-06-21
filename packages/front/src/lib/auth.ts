// Auth token persistence (localStorage).

const KEY = 'labzyn_token';
const WS_KEY = 'labzyn_workspace';

export const tokenStore = {
  get: () => localStorage.getItem(KEY),
  set: (t: string) => localStorage.setItem(KEY, t),
  clear: () => localStorage.removeItem(KEY),
};

export const workspaceStore = {
  get: () => localStorage.getItem(WS_KEY),
  set: (id: string) => localStorage.setItem(WS_KEY, id),
  clear: () => localStorage.removeItem(WS_KEY),
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
