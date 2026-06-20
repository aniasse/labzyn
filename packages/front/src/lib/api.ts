// Thin client for the Labzyn API (proxied to the Go server in dev).

export type MetadataObject = {
  id: string;
  nameSingular: string;
  namePlural: string;
  labelSingular: string;
  labelPlural: string;
  icon: string;
  description: string;
  isSystem: boolean;
  isActive: boolean;
};

export type FieldType =
  | 'TEXT'
  | 'NUMBER'
  | 'NUMERIC'
  | 'BOOLEAN'
  | 'DATE'
  | 'DATE_TIME'
  | 'UUID'
  | 'SELECT'
  | 'MULTI_SELECT'
  | 'RELATION'
  | 'RICH_TEXT'
  | 'RAW_JSON'
  | 'RATING'
  | 'FILES'
  | 'ARRAY';

export type MetadataField = {
  id: string;
  name: string;
  type: FieldType;
  label: string;
  isNullable: boolean;
  isUnique: boolean;
  position: number;
};

export type MetadataObjectWithFields = MetadataObject & { fields: MetadataField[] };

const BASE = '/api/v1';

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  });
  if (!res.ok) {
    let message = `Request failed (${res.status})`;
    try {
      const body = await res.json();
      message = body?.error?.message ?? message;
    } catch {
      /* keep default */
    }
    throw new Error(message);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  listObjects: () =>
    request<{ data: MetadataObject[] }>('/metadata/objects').then((r) => r.data),
  getObject: (nameSingular: string) =>
    request<MetadataObjectWithFields>(`/metadata/objects/${nameSingular}`),
};
