import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { ChevronRight, Plus } from 'lucide-react';
import { api } from '@/lib/api';
import { getIcon } from '@/lib/icons';
import { Button } from '@/components/ui/button';
import { CreateObjectDialog } from '@/components/settings/CreateObjectDialog';

export function DataModelPage() {
  const [creating, setCreating] = useState(false);
  const { data: objects } = useQuery({ queryKey: ['objects'], queryFn: api.listObjects });

  return (
    <div className="mx-auto max-w-3xl p-6">
      <div className="mb-1 flex items-center gap-2">
        <h1 className="text-lg font-semibold text-foreground">Data model</h1>
        <Button variant="primary" size="sm" className="ml-auto" onClick={() => setCreating(true)}>
          <Plus size={14} strokeWidth={2.5} />
          New object
        </Button>
      </div>
      <p className="mb-4 text-sm text-tertiary">
        Objects and their fields. New objects appear in the sidebar instantly.
      </p>

      <div className="overflow-hidden rounded-md border border-border">
        {objects?.map((obj, i) => {
          const Icon = getIcon(obj.icon);
          return (
            <Link
              key={obj.id}
              to={`/settings/objects/${obj.nameSingular}`}
              className={`flex items-center gap-3 px-4 py-3 transition-colors duration-fast hover:bg-accent ${
                i > 0 ? 'border-t border-border' : ''
              }`}
            >
              <Icon size={16} className="text-secondary" />
              <div className="min-w-0">
                <div className="text-sm font-medium text-foreground">{obj.labelPlural}</div>
                <div className="text-xs text-tertiary">
                  {obj.nameSingular} · {obj.namePlural}
                </div>
              </div>
              <ChevronRight size={16} className="ml-auto text-tertiary" />
            </Link>
          );
        })}
        {objects?.length === 0 && (
          <div className="p-6 text-center text-sm text-tertiary">No objects yet.</div>
        )}
      </div>

      <CreateObjectDialog open={creating} onClose={() => setCreating(false)} />
    </div>
  );
}
