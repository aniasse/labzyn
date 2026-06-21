import { useQuery } from '@tanstack/react-query';
import { NavLink } from 'react-router-dom';
import { LogOut } from 'lucide-react';
import { api } from '@/lib/api';
import { useAuth } from '@/lib/authContext';
import { getIcon } from '@/lib/icons';
import { cn } from '@/lib/utils';
import { WorkspaceSwitcher } from './WorkspaceSwitcher';

const navItemBase =
  'flex items-center gap-2 rounded-sm px-2 py-1.5 text-sm text-secondary transition-colors duration-fast hover:bg-accent';

export function Sidebar() {
  const { user, logout } = useAuth();
  const { data: objects, isLoading, isError } = useQuery({
    queryKey: ['objects'],
    queryFn: api.listObjects,
  });

  return (
    <aside className="flex h-full w-60 shrink-0 flex-col border-r border-sidebar-border bg-sidebar">
      {/* Workspace header / switcher */}
      <div className="px-2 py-2">
        <WorkspaceSwitcher />
      </div>

      <nav className="flex-1 overflow-y-auto px-2 pb-3">
        <SectionLabel>Workspace</SectionLabel>

        {isLoading && <Hint>Loading objects…</Hint>}
        {isError && <Hint>Could not load objects</Hint>}
        {objects?.length === 0 && <Hint>No objects yet</Hint>}

        <ul className="flex flex-col gap-0.5">
          {objects?.map((obj) => {
            const Icon = getIcon(obj.icon);
            return (
              <li key={obj.id}>
                <NavLink
                  to={`/objects/${obj.namePlural}`}
                  className={({ isActive }) =>
                    cn(navItemBase, isActive && 'bg-accent text-foreground')
                  }
                >
                  <Icon size={16} strokeWidth={2} className="shrink-0" />
                  <span className="truncate">{obj.labelPlural}</span>
                </NavLink>
              </li>
            );
          })}
        </ul>
      </nav>

      <div className="border-t border-sidebar-border px-2 py-2">
        <NavLink to="/settings" className={navItemBase}>
          {(() => {
            const Icon = getIcon('IconSettings');
            return <Icon size={16} strokeWidth={2} />;
          })()}
          <span>Settings</span>
        </NavLink>

        <div className="mt-1 flex items-center gap-2 px-2 py-1.5">
          <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-pill bg-accent text-xxs font-semibold uppercase text-secondary">
            {(user.firstName || user.email).charAt(0)}
          </div>
          <span className="min-w-0 flex-1 truncate text-xs text-secondary" title={user.email}>
            {user.email}
          </span>
          <button
            type="button"
            onClick={logout}
            aria-label="Log out"
            className="rounded-sm p-1 text-tertiary transition-colors duration-fast hover:bg-accent hover:text-foreground"
          >
            <LogOut size={14} />
          </button>
        </div>
      </div>
    </aside>
  );
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <div className="px-2 pb-1 pt-2 text-xxs font-semibold uppercase tracking-wide text-tertiary">
      {children}
    </div>
  );
}

function Hint({ children }: { children: React.ReactNode }) {
  return <div className="px-2 py-1.5 text-xs text-tertiary">{children}</div>;
}
