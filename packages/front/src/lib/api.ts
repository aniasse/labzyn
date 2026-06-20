// Thin client for the Labzyn API (proxied to the Go server in dev).
import { tokenStore, type AuthResponse } from './auth';

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
  options?: unknown;
  defaultValue?: unknown;
};

export type MetadataObjectWithFields = MetadataObject & { fields: MetadataField[] };

export type CreateObjectInput = {
  nameSingular: string;
  namePlural: string;
  labelSingular: string;
  labelPlural: string;
  icon?: string;
  description?: string;
};

export type CreateFieldInput = {
  name: string;
  type: FieldType;
  label: string;
  isNullable?: boolean;
  isUnique?: boolean;
  options?: { value: string; label: string }[];
};

const BASE = '/api/v1';

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message);
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = tokenStore.get();
  const res = await fetch(BASE + path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });
  if (!res.ok) {
    let message = `Request failed (${res.status})`;
    try {
      const body = await res.json();
      message = body?.error?.message ?? message;
    } catch {
      /* keep default */
    }
    throw new ApiError(message, res.status);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export type RecordRow = {
  id: string;
  data: Record<string, unknown>;
  position: number;
  createdAt: string;
  updatedAt: string;
};

export type ListRecordsParams = {
  limit?: number;
  offset?: number;
  orderBy?: string;
  order?: 'asc' | 'desc';
  filters?: Record<string, string>;
};

export type RecordPage = {
  data: RecordRow[];
  total: number;
  limit: number;
  offset: number;
};

function buildQuery(params: ListRecordsParams): string {
  const q = new URLSearchParams();
  if (params.limit != null) q.set('limit', String(params.limit));
  if (params.offset != null) q.set('offset', String(params.offset));
  if (params.orderBy) q.set('orderBy', params.orderBy);
  if (params.order) q.set('order', params.order);
  for (const [k, v] of Object.entries(params.filters ?? {})) {
    if (v !== '') q.set(k, v);
  }
  const s = q.toString();
  return s ? `?${s}` : '';
}

export type MeResponse = {
  user: { id: string; email: string; firstName: string; lastName: string };
  workspaces: { id: string; displayName: string; subdomain: string }[];
  currentWorkspaceId: string;
};

export const api = {
  signup: (input: {
    email: string;
    password: string;
    firstName?: string;
    workspaceName?: string;
  }) => request<AuthResponse>('/auth/signup', { method: 'POST', body: JSON.stringify(input) }),
  login: (input: { email: string; password: string }) =>
    request<AuthResponse>('/auth/login', { method: 'POST', body: JSON.stringify(input) }),
  me: () => request<MeResponse>('/auth/me'),

  listObjects: () =>
    request<{ data: MetadataObject[] }>('/metadata/objects').then((r) => r.data),
  getObject: (nameSingular: string) =>
    request<MetadataObjectWithFields>(`/metadata/objects/${nameSingular}`),
  createObject: (input: CreateObjectInput) =>
    request<MetadataObject>('/metadata/objects', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  createField: (nameSingular: string, input: CreateFieldInput) =>
    request<MetadataField>(`/metadata/objects/${nameSingular}/fields`, {
      method: 'POST',
      body: JSON.stringify(input),
    }),

  listRecords: (plural: string, params: ListRecordsParams = {}) =>
    request<RecordPage>(`/${plural}${buildQuery(params)}`),
  getRecord: (plural: string, id: string) =>
    request<RecordRow>(`/${plural}/${id}`),
  createRecord: (plural: string, data: Record<string, unknown>) =>
    request<RecordRow>(`/${plural}`, { method: 'POST', body: JSON.stringify(data) }),
  updateRecord: (plural: string, id: string, patch: Record<string, unknown>) =>
    request<RecordRow>(`/${plural}/${id}`, { method: 'PATCH', body: JSON.stringify(patch) }),
  deleteRecord: (plural: string, id: string) =>
    request<void>(`/${plural}/${id}`, { method: 'DELETE' }),
};
