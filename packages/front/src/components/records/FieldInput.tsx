import type { MetadataField } from '@/lib/api';
import { coerceInput, inputKind, selectOptions } from '@/lib/fields';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Checkbox } from '@/ui/Checkbox';

type Props = {
  field: MetadataField;
  value: unknown;
  onChange: (value: unknown) => void;
  autoFocus?: boolean;
  onCommit?: () => void; // Enter or blur
  onCancel?: () => void; // Escape
};

/** A control whose shape follows the field's type. Shared by create and edit. */
export function FieldInput({ field, value, onChange, autoFocus, onCommit, onCancel }: Props) {
  const kind = inputKind(field.type);

  const keyHandlers = {
    onKeyDown: (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') onCommit?.();
      if (e.key === 'Escape') onCancel?.();
    },
  };

  if (kind === 'select') {
    return (
      <Select
        autoFocus={autoFocus}
        value={value == null ? '' : String(value)}
        onChange={(e) => onChange(e.target.value === '' ? null : e.target.value)}
        onBlur={onCommit}
        {...keyHandlers}
      >
        <option value="">—</option>
        {selectOptions(field).map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </Select>
    );
  }

  if (kind === 'boolean') {
    return (
      <Checkbox
        checked={value === true}
        onCheckedChange={(v) => {
          onChange(v);
          onCommit?.();
        }}
      />
    );
  }

  return (
    <Input
      autoFocus={autoFocus}
      type={kind === 'number' ? 'number' : 'text'}
      value={value == null ? '' : String(value)}
      onChange={(e) => onChange(coerceInput(field.type, e.target.value))}
      onBlur={onCommit}
      {...keyHandlers}
    />
  );
}
