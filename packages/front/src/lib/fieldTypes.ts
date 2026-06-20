import type { FieldType } from './api';

/** Field types offered in the data-model editor, with friendly labels. */
export const FIELD_TYPE_OPTIONS: { value: FieldType; label: string }[] = [
  { value: 'TEXT', label: 'Text' },
  { value: 'RICH_TEXT', label: 'Rich text' },
  { value: 'NUMBER', label: 'Number' },
  { value: 'NUMERIC', label: 'Decimal' },
  { value: 'BOOLEAN', label: 'Boolean' },
  { value: 'DATE', label: 'Date' },
  { value: 'DATE_TIME', label: 'Date & time' },
  { value: 'SELECT', label: 'Select' },
  { value: 'MULTI_SELECT', label: 'Multi-select' },
  { value: 'RATING', label: 'Rating' },
  { value: 'RELATION', label: 'Relation (UUID)' },
  { value: 'RAW_JSON', label: 'JSON' },
  { value: 'FILES', label: 'Files' },
  { value: 'ARRAY', label: 'Array' },
];

/** Field types that need a list of options. */
export function typeNeedsOptions(type: FieldType): boolean {
  return type === 'SELECT' || type === 'MULTI_SELECT';
}
