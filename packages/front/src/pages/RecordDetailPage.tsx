import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Trash2 } from 'lucide-react';
import { api, type MetadataField } from '@/lib/api';
import { getIcon } from '@/lib/icons';
import { Button } from '@/components/ui/button';
import { FieldInput } from '@/components/records/FieldInput';

export function RecordDetailPage() {
  const { objectPlural = '', id = '' } = useParams();
  const navigate = useNavigate();
  const qc = useQueryClient();

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
  const { data: record, isError } = useQuery({
    queryKey: ['record', objectPlural, id],
    queryFn: () => api.getRecord(objectPlural, id),
  });

  const update = useMutation({
    mutationFn: (patch: Record<string, unknown>) => api.updateRecord(objectPlural, id, patch),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['record', objectPlural, id] });
      qc.invalidateQueries({ queryKey: ['records', objectPlural] });
    },
  });
  const remove = useMutation({
    mutationFn: () => api.deleteRecord(objectPlural, id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['records', objectPlural] });
      navigate(`/objects/${objectPlural}`);
    },
  });

  if (isError) {
    return <div className="p-6 text-sm text-danger">Record not found.</div>;
  }

  const Icon = getIcon(summary?.icon);
  const fields = detail?.fields ?? [];
  const title = primaryTitle(fields, record?.data) ?? summary?.labelSingular ?? 'Record';

  return (
    <div className="mx-auto flex h-full max-w-2xl flex-col">
      <div className="flex items-center gap-2 border-b border-border px-5 py-3">
        <button
          type="button"
          onClick={() => navigate(`/objects/${objectPlural}`)}
          className="rounded-sm p-1 text-tertiary hover:bg-accent"
          aria-label="Back"
        >
          <ArrowLeft size={16} />
        </button>
        <Icon size={18} className="text-secondary" />
        <h1 className="text-md font-semibold text-foreground">{title}</h1>
        <Button
          variant="ghost"
          size="sm"
          className="ml-auto text-danger hover:bg-danger/10"
          onClick={() => remove.mutate()}
        >
          <Trash2 size={14} />
          Delete
        </Button>
      </div>

      <div className="min-h-0 flex-1 overflow-auto p-5">
        {record && (
          <div className="flex flex-col gap-4">
            {fields.map((f) => (
              <DetailField
                key={f.id}
                field={f}
                value={record.data[f.name]}
                onSave={(value) => update.mutate({ [f.name]: value })}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function DetailField({
  field,
  value,
  onSave,
}: {
  field: MetadataField;
  value: unknown;
  onSave: (value: unknown) => void;
}) {
  const [draft, setDraft] = useState<unknown>(value);
  useEffect(() => setDraft(value), [value]);

  return (
    <div className="grid grid-cols-[160px_1fr] items-center gap-3">
      <span className="text-xs font-medium text-secondary">{field.label}</span>
      <FieldInput
        field={field}
        value={draft}
        onChange={setDraft}
        onCommit={() => draft !== value && onSave(draft)}
      />
    </div>
  );
}

/** Use the first TEXT field as the record's display title. */
function primaryTitle(fields: MetadataField[], data?: Record<string, unknown>): string | null {
  if (!data) return null;
  const text = fields.find((f) => f.type === 'TEXT');
  const v = text ? data[text.name] : undefined;
  return v ? String(v) : null;
}
