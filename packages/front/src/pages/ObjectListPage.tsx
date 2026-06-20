import { useMemo } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { getIcon } from '@/lib/icons';
import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';

// Jalon 3: object header + schema preview wired to the API.
// Jalon 4 replaces the preview with the generic record table.
export function ObjectListPage() {
  const { objectPlural = '' } = useParams();

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
    return <Empty title="Object not found" subtitle={`No object named "${objectPlural}".`} />;
  }

  const Icon = getIcon(summary?.icon);

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-border px-5 py-3">
        <Icon size={18} strokeWidth={2} className="text-secondary" />
        <h1 className="text-md font-semibold text-foreground">
          {summary?.labelPlural ?? '…'}
        </h1>
        {detail && (
          <span className="text-xs text-tertiary">{detail.fields.length} fields</span>
        )}
        <Button variant="primary" size="sm" className="ml-auto">
          <Plus size={14} strokeWidth={2.5} />
          New
        </Button>
      </div>

      <div className="min-h-0 flex-1 overflow-auto p-5">
        {/* Schema preview (placeholder for the Jalon 4 record table). */}
        <div className="overflow-hidden rounded-md border border-border">
          <table className="w-full border-collapse text-sm">
            <thead>
              <tr className="bg-muted text-left text-xs text-tertiary">
                <th className="px-3 py-2 font-medium">Field</th>
                <th className="px-3 py-2 font-medium">Type</th>
                <th className="px-3 py-2 font-medium">Required</th>
              </tr>
            </thead>
            <tbody>
              {detail?.fields.map((f) => (
                <tr key={f.id} className="border-t border-border">
                  <td className="px-3 py-2 text-foreground">{f.label}</td>
                  <td className="px-3 py-2 text-secondary">{f.type}</td>
                  <td className="px-3 py-2 text-secondary">
                    {f.isNullable ? '—' : 'yes'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <p className="mt-3 text-xs text-tertiary">
          Record table comes next (Jalon 4). This previews the object schema served
          by the API.
        </p>
      </div>
    </div>
  );
}

function Empty({ title, subtitle }: { title: string; subtitle: string }) {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-1 text-center">
      <h2 className="text-md font-semibold text-foreground">{title}</h2>
      <p className="text-sm text-tertiary">{subtitle}</p>
    </div>
  );
}
