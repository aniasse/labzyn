import type { FieldType, MetadataField } from './api';

/** Options parsed from a field's `options` (["a"] or [{value,label}]). */
export type SelectOption = { value: string; label: string };

export function selectOptions(field: { options?: unknown }): SelectOption[] {
  const raw = (field as { options?: unknown }).options;
  if (!Array.isArray(raw)) return [];
  return raw.map((o) =>
    typeof o === 'string'
      ? { value: o, label: o }
      : { value: String((o as any).value), label: String((o as any).label ?? (o as any).value) },
  );
}

/** Human-readable value for table cells / detail view. */
export function formatValue(field: MetadataField, value: unknown): string {
  if (value === null || value === undefined || value === '') return '';
  switch (field.type) {
    case 'BOOLEAN':
      return value ? 'Yes' : 'No';
    case 'DATE':
      return new Date(String(value)).toLocaleDateString();
    case 'DATE_TIME':
      return new Date(String(value)).toLocaleString();
    case 'MULTI_SELECT':
    case 'ARRAY':
      return Array.isArray(value) ? value.join(', ') : String(value);
    case 'RAW_JSON':
    case 'FILES':
      return JSON.stringify(value);
    default:
      return String(value);
  }
}

/** Convert a raw string from an <input> into the typed value the API expects. */
export function coerceInput(type: FieldType, raw: string): unknown {
  if (raw === '') return null;
  switch (type) {
    case 'NUMBER':
    case 'NUMERIC':
    case 'RATING': {
      const n = Number(raw);
      return Number.isNaN(n) ? raw : n;
    }
    case 'BOOLEAN':
      return raw === 'true';
    default:
      return raw;
  }
}

/** The HTML input control kind to use for a field type. */
export function inputKind(type: FieldType): 'text' | 'number' | 'select' | 'boolean' {
  switch (type) {
    case 'NUMBER':
    case 'NUMERIC':
    case 'RATING':
      return 'number';
    case 'SELECT':
      return 'select';
    case 'BOOLEAN':
      return 'boolean';
    default:
      return 'text';
  }
}
