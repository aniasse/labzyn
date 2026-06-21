import { useEffect, useState } from 'react';
import { Outlet } from 'react-router-dom';
import { CommandMenu } from '@/components/CommandMenu';
import { Sidebar } from './Sidebar';

export function AppShell() {
  const [cmdOpen, setCmdOpen] = useState(false);

  // Global ⌘K / Ctrl+K to open the command menu.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setCmdOpen((v) => !v);
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  // Twenty-style layout: no global top bar; search lives in the sidebar and
  // each page renders its own header.
  return (
    <div className="flex h-full w-full overflow-hidden bg-muted p-2 pl-0">
      <Sidebar onOpenSearch={() => setCmdOpen(true)} />
      <main className="min-h-0 min-w-0 flex-1 overflow-auto rounded-md border border-border bg-background">
        <Outlet />
      </main>
      <CommandMenu open={cmdOpen} onOpenChange={setCmdOpen} />
    </div>
  );
}
