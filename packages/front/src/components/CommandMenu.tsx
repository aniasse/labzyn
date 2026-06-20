import { Command } from 'cmdk';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { getIcon } from '@/lib/icons';

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

/** Keyboard-first command palette (⌘K / Ctrl+K), modelled on Twenty's. */
export function CommandMenu({ open, onOpenChange }: Props) {
  const navigate = useNavigate();
  const { data: objects } = useQuery({ queryKey: ['objects'], queryFn: api.listObjects });

  const go = (path: string) => {
    onOpenChange(false);
    navigate(path);
  };

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center bg-foreground/20 pt-[12vh]"
      onClick={() => onOpenChange(false)}
    >
      <Command
        label="Command menu"
        className="w-full max-w-lg overflow-hidden rounded-md border border-border bg-background shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <Command.Input
          autoFocus
          placeholder="Search objects and actions…"
          className="w-full border-b border-border bg-transparent px-4 py-3 text-sm text-foreground outline-none placeholder:text-tertiary"
        />
        <Command.List className="max-h-80 overflow-y-auto p-1.5">
          <Command.Empty className="px-3 py-6 text-center text-sm text-tertiary">
            No results found.
          </Command.Empty>

          <Command.Group
            heading="Navigate"
            className="px-1 [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5 [&_[cmdk-group-heading]]:text-xxs [&_[cmdk-group-heading]]:font-semibold [&_[cmdk-group-heading]]:uppercase [&_[cmdk-group-heading]]:tracking-wide [&_[cmdk-group-heading]]:text-tertiary"
          >
            {objects?.map((obj) => {
              const Icon = getIcon(obj.icon);
              return (
                <Command.Item
                  key={obj.id}
                  value={obj.labelPlural}
                  onSelect={() => go(`/objects/${obj.namePlural}`)}
                  className="flex cursor-pointer items-center gap-2 rounded-sm px-2 py-2 text-sm text-foreground aria-selected:bg-accent"
                >
                  <Icon size={16} strokeWidth={2} />
                  <span>{obj.labelPlural}</span>
                </Command.Item>
              );
            })}
          </Command.Group>
        </Command.List>
      </Command>
    </div>
  );
}
