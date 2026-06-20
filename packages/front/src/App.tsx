import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Navigate, Route, BrowserRouter, Routes } from 'react-router-dom';
import { AppShell } from '@/components/layout/AppShell';
import { HomePage } from '@/pages/HomePage';
import { ObjectListPage } from '@/pages/ObjectListPage';
import { RecordDetailPage } from '@/pages/RecordDetailPage';
import { DataModelPage } from '@/pages/settings/DataModelPage';
import { ObjectSettingsPage } from '@/pages/settings/ObjectSettingsPage';

const queryClient = new QueryClient({
  defaultOptions: { queries: { refetchOnWindowFocus: false, staleTime: 30_000 } },
});

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
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
      </BrowserRouter>
    </QueryClientProvider>
  );
}
