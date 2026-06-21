import { useState } from 'react';
import { Check, ChevronsUpDown } from 'lucide-react';
import { useAuth } from '@/lib/authContext';
import { cn } from '@/lib/utils';
import { Avatar } from '@/ui/Avatar';

export function WorkspaceSwitcher() {
  const { workspaces, currentWorkspaceId, switchWorkspace } = useAuth();
  const [open, setOpen] = useState(false);
  const current = workspaces.find((w) => w.id === currentWorkspaceId) ?? workspaces[0];

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center gap-2 rounded-sm px-1.5 py-1.5 text-left transition-colors duration-fast hover:bg-accent"
      >
        <Avatar placeholder={current?.displayName ?? 'Labzyn'} type="squared" size="sm" />
        <span className="min-w-0 flex-1 truncate text-sm font-semibold text-foreground">
          {current?.displayName ?? 'Labzyn'}
        </span>
        <ChevronsUpDown size={14} className="shrink-0 text-tertiary" />
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-10" onClick={() => setOpen(false)} />
          <div className="absolute left-0 right-0 top-full z-20 mt-1 rounded-md border border-border bg-background p-1 shadow-lg">
            {workspaces.map((w) => (
              <button
                key={w.id}
                type="button"
                onClick={() => {
                  switchWorkspace(w.id);
                  setOpen(false);
                }}
                className={cn(
                  'flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-sm text-secondary hover:bg-accent',
                  w.id === current?.id && 'text-foreground',
                )}
              >
                <span className="min-w-0 flex-1 truncate">{w.displayName}</span>
                {w.id === current?.id && <Check size={14} className="text-primary" />}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
