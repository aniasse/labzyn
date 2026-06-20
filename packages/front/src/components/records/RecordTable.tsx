import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowDown, ArrowUp, ChevronRight } from 'lucide-react';
import { api, type MetadataField } from '@/lib/api';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { EditableCell } from './EditableCell';

const PAGE = 25;

export function RecordTable({ plural, fields }: { plural: string; fields: MetadataField[] }) {
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [sort, setSort] = useState<{ by: string; desc: boolean } | null>(null);
  const [offset, setOffset] = useState(0);

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ['records', plural, sort, offset],
    queryFn: () =>
      api.listRecords(plural, {
        limit: PAGE,
        offset,
        orderBy: sort?.by,
        order: sort ? (sort.desc ? 'desc' : 'asc') : undefined,
      }),
  });

  const update = useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: Record<string, unknown> }) =>
      api.updateRecord(plural, id, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['records', plural] }),
  });

  const toggleSort = (name: string) =>
    setSort((s) => (s?.by === name ? { by: name, desc: !s.desc } : { by: name, desc: false }));

  if (isError) {
    return <p className="p-4 text-sm text-danger">{(error as Error).message}</p>;
  }

  const rows = data?.data ?? [];
  const total = data?.total ?? 0;

  return (
    <div className="flex h-full flex-col">
      <div className="min-h-0 flex-1 overflow-auto">
        <table className="w-full border-collapse text-sm">
          <thead className="sticky top-0 z-10 bg-muted">
            <tr className="text-left text-xs text-tertiary">
              <th className="w-9 border-b border-border" />
              {fields.map((f) => (
                <th
                  key={f.id}
                  className="cursor-pointer select-none border-b border-border px-3 py-2 font-medium hover:text-secondary"
                  onClick={() => toggleSort(f.name)}
                >
                  <span className="inline-flex items-center gap-1">
                    {f.label}
                    {sort?.by === f.name &&
                      (sort.desc ? <ArrowDown size={12} /> : <ArrowUp size={12} />)}
                  </span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id} className="group border-b border-border hover:bg-muted/50">
                <td className="px-1 text-center align-middle">
                  <button
                    type="button"
                    aria-label="Open record"
                    onClick={() => navigate(`/objects/${plural}/${row.id}`)}
                    className="rounded-sm p-1 text-tertiary opacity-0 transition-opacity duration-fast hover:bg-accent group-hover:opacity-100"
                  >
                    <ChevronRight size={14} />
                  </button>
                </td>
                {fields.map((f) => (
                  <td key={f.id} className="px-2 py-1 align-middle">
                    <EditableCell
                      field={f}
                      value={row.data[f.name]}
                      onSave={(value) => update.mutate({ id: row.id, patch: { [f.name]: value } })}
                    />
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>

        {isLoading && <p className="p-4 text-sm text-tertiary">Loading records…</p>}
        {!isLoading && rows.length === 0 && (
          <div className="flex flex-col items-center gap-1 p-10 text-center">
            <p className="text-sm font-medium text-foreground">No records yet</p>
            <p className="text-xs text-tertiary">Create the first one with “New”.</p>
          </div>
        )}
      </div>

      <Pagination
        offset={offset}
        count={rows.length}
        total={total}
        onPrev={() => setOffset((o) => Math.max(0, o - PAGE))}
        onNext={() => setOffset((o) => o + PAGE)}
      />
    </div>
  );
}

function Pagination({
  offset,
  count,
  total,
  onPrev,
  onNext,
}: {
  offset: number;
  count: number;
  total: number;
  onPrev: () => void;
  onNext: () => void;
}) {
  const from = total === 0 ? 0 : offset + 1;
  const to = offset + count;
  return (
    <div className="flex items-center justify-between border-t border-border px-4 py-2 text-xs text-tertiary">
      <span>
        {from}–{to} of {total}
      </span>
      <div className="flex gap-1">
        <Button size="sm" variant="secondary" onClick={onPrev} disabled={offset === 0}>
          Previous
        </Button>
        <Button
          size="sm"
          variant="secondary"
          onClick={onNext}
          disabled={to >= total}
          className={cn(to >= total && 'opacity-50')}
        >
          Next
        </Button>
      </div>
    </div>
  );
}
