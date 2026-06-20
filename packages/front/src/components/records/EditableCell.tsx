import { useEffect, useState } from 'react';
import type { MetadataField } from '@/lib/api';
import { formatValue } from '@/lib/fields';
import { FieldInput } from './FieldInput';
import { Check } from 'lucide-react';

type Props = {
  field: MetadataField;
  value: unknown;
  onSave: (value: unknown) => void;
};

/** A table cell that turns into a typed input on click (Twenty-style). */
export function EditableCell({ field, value, onSave }: Props) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState<unknown>(value);

  useEffect(() => setDraft(value), [value]);

  if (editing) {
    const commit = () => {
      setEditing(false);
      if (draft !== value) onSave(draft);
    };
    return (
      <FieldInput
        field={field}
        value={draft}
        onChange={setDraft}
        autoFocus
        onCommit={commit}
        onCancel={() => {
          setDraft(value);
          setEditing(false);
        }}
      />
    );
  }

  return (
    <button
      type="button"
      onClick={() => setEditing(true)}
      className="flex h-7 w-full items-center rounded-sm px-1 text-left text-sm hover:bg-accent"
    >
      <CellValue field={field} value={value} />
    </button>
  );
}

function CellValue({ field, value }: { field: MetadataField; value: unknown }) {
  if (value === null || value === undefined || value === '') {
    return <span className="text-tertiary">—</span>;
  }
  if (field.type === 'BOOLEAN') {
    return value ? (
      <Check size={15} className="text-primary" />
    ) : (
      <span className="text-tertiary">—</span>
    );
  }
  if (field.type === 'SELECT') {
    return (
      <span className="rounded-pill bg-muted px-2 py-0.5 text-xs text-secondary">
        {formatValue(field, value)}
      </span>
    );
  }
  return <span className="truncate text-foreground">{formatValue(field, value)}</span>;
}
