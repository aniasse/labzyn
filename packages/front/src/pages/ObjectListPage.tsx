import { useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { api } from '@/lib/api';
import { getIcon } from '@/lib/icons';
import { Button } from '@/components/ui/button';
import { RecordTable } from '@/components/records/RecordTable';
import { RecordFormDialog } from '@/components/records/RecordFormDialog';

export function ObjectListPage() {
  const { objectPlural = '' } = useParams();
  const [creating, setCreating] = useState(false);

  const { data: objects } = useQuery({ queryKey: ['objects'], queryFn: api.listObjects });
  const summary = useMemo(
    () => objects?.find((o) => o.namePlural === objectPlural),
    [objects, objectPlural],
  );

  const { data: detail } = useQuery({
    queryKey: ['object', summary?.nameSingular],
    queryFn: () => api.getObject(summary!.nameSingular),
    enabled: !!summary,
  });

  if (objects && !summary) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-1 text-center">
        <h2 className="text-md font-semibold text-foreground">Object not found</h2>
        <p className="text-sm text-tertiary">No object named “{objectPlural}”.</p>
      </div>
    );
  }

  const Icon = getIcon(summary?.icon);
  const fields = detail?.fields ?? [];

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-border px-5 py-3">
        <Icon size={18} strokeWidth={2} className="text-secondary" />
        <h1 className="text-md font-semibold text-foreground">{summary?.labelPlural ?? '…'}</h1>
        <Button
          variant="primary"
          size="sm"
          className="ml-auto"
          onClick={() => setCreating(true)}
          disabled={!detail}
        >
          <Plus size={14} strokeWidth={2.5} />
          New
        </Button>
      </div>

      <div className="min-h-0 flex-1">
        {detail && <RecordTable plural={objectPlural} fields={fields} />}
      </div>

      {detail && (
        <RecordFormDialog
          plural={objectPlural}
          label={summary?.labelSingular ?? 'record'}
          fields={fields}
          open={creating}
          onClose={() => setCreating(false)}
        />
      )}
    </div>
  );
}
