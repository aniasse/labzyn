import { useQuery } from '@tanstack/react-query';
import { NavLink } from 'react-router-dom';
import { LogOut, Search, Sparkles } from 'lucide-react';
import { api } from '@/lib/api';
import { useAuth } from '@/lib/authContext';
import { getIcon } from '@/lib/icons';
import { cn } from '@/lib/utils';
import { WorkspaceSwitcher } from './WorkspaceSwitcher';
import { Avatar } from '@/ui/Avatar';

// Matches Twenty's NavigationDrawerItem: 28px tall, md font, 4px left padding,
// sm radius, secondary text, transparent-light hover.
const navItemBase =
  'flex h-7 items-center gap-2 rounded-sm px-1 text-md text-secondary transition-colors duration-fast hover:bg-accent';

export function Sidebar({ onOpenSearch }: { onOpenSearch: () => void }) {
  const { user, logout } = useAuth();
  const { data: objects, isLoading, isError } = useQuery({
    queryKey: ['objects'],
    queryFn: api.listObjects,
  });

  return (
    <aside className="flex h-full w-56 shrink-0 flex-col px-2">
      {/* Workspace header / switcher */}
      <div className="py-2">
        <WorkspaceSwitcher />
      </div>

      <nav className="flex-1 overflow-y-auto pb-3">
        {/* Search (Twenty puts it in the sidebar, not a top bar) */}
        <button type="button" onClick={onOpenSearch} className={cn(navItemBase, 'w-full')}>
          <Search size={16} strokeWidth={2} className="shrink-0" />
          <span>Search</span>
          <kbd className="ml-auto rounded-xs border border-border bg-background px-1 text-xxs text-tertiary">
            ⌘K
          </kbd>
        </button>

        <NavLink
          to="/agent"
          className={({ isActive }) => cn(navItemBase, isActive && 'bg-accent text-foreground')}
        >
          <Sparkles size={16} strokeWidth={2} className="shrink-0 text-primary" />
          <span className="font-medium text-foreground">Ask Labzyn</span>
        </NavLink>

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
          <Avatar placeholder={user.firstName || user.email} type="rounded" size="sm" />
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
    <div className="flex h-7 items-center px-1 pt-2 text-xs font-medium text-tertiary">
      {children}
    </div>
  );
}

function Hint({ children }: { children: React.ReactNode }) {
  return <div className="px-2 py-1.5 text-xs text-tertiary">{children}</div>;
}
