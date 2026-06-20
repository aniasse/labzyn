import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api, type FieldType } from '@/lib/api';
import { FIELD_TYPE_OPTIONS, typeNeedsOptions } from '@/lib/fieldTypes';
import { Modal } from '@/components/ui/modal';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';

const NAME_RE = /^[a-zA-Z][a-zA-Z0-9_]*$/;

export function CreateFieldDialog({
  nameSingular,
  open,
  onClose,
}: {
  nameSingular: string;
  open: boolean;
  onClose: () => void;
}) {
  const qc = useQueryClient();
  const [label, setLabel] = useState('');
  const [name, setName] = useState('');
  const [type, setType] = useState<FieldType>('TEXT');
  const [required, setRequired] = useState(false);
  const [optionsText, setOptionsText] = useState('');
  const [error, setError] = useState<string | null>(null);

  const reset = () => {
    setLabel('');
    setName('');
    setType('TEXT');
    setRequired(false);
    setOptionsText('');
    setError(null);
  };

  const create = useMutation({
    mutationFn: () =>
      api.createField(nameSingular, {
        name,
        type,
        label: label || name,
        isNullable: !required,
        options: typeNeedsOptions(type)
          ? optionsText
              .split(',')
              .map((s) => s.trim())
              .filter(Boolean)
              .map((v) => ({ value: v, label: v }))
          : undefined,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['object', nameSingular] });
      reset();
      onClose();
    },
    onError: (e: Error) => setError(e.message),
  });

  const onLabel = (v: string) => {
    setLabel(v);
    setName(v.trim().toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_+|_+$/g, ''));
  };

  const valid = NAME_RE.test(name);

  return (
    <Modal
      open={open}
      onClose={() => { reset(); onClose(); }}
      title="New field"
      footer={
        <>
          <Button variant="secondary" onClick={() => { reset(); onClose(); }}>
            Cancel
          </Button>
          <Button variant="primary" disabled={!valid || create.isPending} onClick={() => create.mutate()}>
            {create.isPending ? 'Creating…' : 'Create'}
          </Button>
        </>
      }
    >
      <div className="flex flex-col gap-3">
        <label className="flex flex-col gap-1">
          <span className="text-xs font-medium text-secondary">Label</span>
          <Input value={label} onChange={(e) => onLabel(e.target.value)} placeholder="Status" />
        </label>
        <div className="grid grid-cols-2 gap-3">
          <label className="flex flex-col gap-1">
            <span className="text-xs font-medium text-secondary">API name</span>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="status" />
          </label>
          <label className="flex flex-col gap-1">
            <span className="text-xs font-medium text-secondary">Type</span>
            <Select value={type} onChange={(e) => setType(e.target.value as FieldType)}>
              {FIELD_TYPE_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </Select>
          </label>
        </div>
        {typeNeedsOptions(type) && (
          <label className="flex flex-col gap-1">
            <span className="text-xs font-medium text-secondary">Options (comma-separated)</span>
            <Input value={optionsText} onChange={(e) => setOptionsText(e.target.value)} placeholder="ACTIVE, ARCHIVED" />
          </label>
        )}
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={required}
            onChange={(e) => setRequired(e.target.checked)}
            className="h-4 w-4 accent-[var(--primary)]"
          />
          <span className="text-xs text-secondary">Required</span>
        </label>
        {name && !valid && (
          <p className="text-xs text-danger">API name must match ^[a-zA-Z][a-zA-Z0-9_]*$</p>
        )}
        {error && <p className="text-xs text-danger">{error}</p>}
      </div>
    </Modal>
  );
}
