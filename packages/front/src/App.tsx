import { useEffect, useState } from 'react';
import { QueryClient, QueryClientProvider, useQuery, useQueryClient } from '@tanstack/react-query';
import { Navigate, Route, BrowserRouter, Routes } from 'react-router-dom';
import { AppShell } from '@/components/layout/AppShell';
import { api } from '@/lib/api';
import { tokenStore } from '@/lib/auth';
import { AuthContext } from '@/lib/authContext';
import { HomePage } from '@/pages/HomePage';
import { LoginPage } from '@/pages/LoginPage';
import { ObjectListPage } from '@/pages/ObjectListPage';
import { RecordDetailPage } from '@/pages/RecordDetailPage';
import { DataModelPage } from '@/pages/settings/DataModelPage';
import { ObjectSettingsPage } from '@/pages/settings/ObjectSettingsPage';

const queryClient = new QueryClient({
  defaultOptions: { queries: { refetchOnWindowFocus: false, staleTime: 30_000, retry: false } },
});

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AuthGate />
      </BrowserRouter>
    </QueryClientProvider>
  );
}

function AuthGate() {
  const [authed, setAuthed] = useState(() => !!tokenStore.get());
  if (!authed) return <LoginPage onAuthed={() => setAuthed(true)} />;
  return <AuthedApp onLogout={() => setAuthed(false)} />;
}

function AuthedApp({ onLogout }: { onLogout: () => void }) {
  const qc = useQueryClient();
  const { data: me, isLoading, isError } = useQuery({ queryKey: ['me'], queryFn: api.me });

  const logout = () => {
    tokenStore.clear();
    qc.clear();
    onLogout();
  };

  // Token invalid/expired → drop it and show login.
  useEffect(() => {
    if (isError) logout();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isError]);

  if (isLoading || !me) {
    return <div className="flex h-full items-center justify-center text-sm text-tertiary">Loading…</div>;
  }

  return (
    <AuthContext.Provider value={{ user: me.user, workspaces: me.workspaces, logout }}>
      <Routes>
        <Route element={<AppShell />}>
          <Route path="/" element={<HomePage />} />
          <Route path="/objects/:objectPlural" element={<ObjectListPage />} />
          <Route path="/objects/:objectPlural/:id" element={<RecordDetailPage />} />
          <Route path="/settings" element={<Navigate to="/settings/objects" replace />} />
          <Route path="/settings/objects" element={<DataModelPage />} />
          <Route path="/settings/objects/:nameSingular" element={<ObjectSettingsPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </AuthContext.Provider>
  );
}
