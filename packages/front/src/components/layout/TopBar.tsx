import { Search } from 'lucide-react';

export function TopBar({ onOpenSearch }: { onOpenSearch: () => void }) {
  return (
    <header className="flex h-12 shrink-0 items-center gap-3 border-b border-border bg-background px-4">
      <button
        type="button"
        onClick={onOpenSearch}
        className="flex h-8 w-64 items-center gap-2 rounded-sm border border-border bg-muted px-2.5 text-sm text-tertiary transition-colors duration-fast hover:bg-accent"
      >
        <Search size={15} strokeWidth={2} />
        <span>Search</span>
        <kbd className="ml-auto rounded-xs border border-border bg-background px-1.5 py-0.5 text-xxs font-medium text-tertiary">
          ⌘K
        </kbd>
      </button>
    </header>
  );
}
