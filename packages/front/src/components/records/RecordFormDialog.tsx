import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api, type MetadataField } from '@/lib/api';
import { Modal } from '@/components/ui/modal';
import { Button } from '@/components/ui/button';
import { FieldInput } from './FieldInput';

type Props = {
  plural: string;
  label: string;
  fields: MetadataField[];
  open: boolean;
  onClose: () => void;
};

export function RecordFormDialog({ plural, label, fields, open, onClose }: Props) {
  const qc = useQueryClient();
  const [draft, setDraft] = useState<Record<string, unknown>>({});
  const [error, setError] = useState<string | null>(null);

  const create = useMutation({
    mutationFn: () => api.createRecord(plural, draft),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['records', plural] });
      reset();
      onClose();
    },
    onError: (e: Error) => setError(e.message),
  });

  const reset = () => {
    setDraft({});
    setError(null);
  };

  const set = (name: string, value: unknown) =>
    setDraft((d) => ({ ...d, [name]: value }));

  return (
    <Modal
      open={open}
      onClose={() => {
        reset();
        onClose();
      }}
      title={`New ${label}`}
      footer={
        <>
          <Button
            variant="secondary"
            onClick={() => {
              reset();
              onClose();
            }}
          >
            Cancel
          </Button>
          <Button variant="primary" onClick={() => create.mutate()} disabled={create.isPending}>
            {create.isPending ? 'Creating…' : 'Create'}
          </Button>
        </>
      }
    >
      <div className="flex flex-col gap-3">
        {fields.map((f) => (
          <label key={f.id} className="flex flex-col gap-1">
            <span className="text-xs font-medium text-secondary">
              {f.label}
              {!f.isNullable && <span className="text-danger"> *</span>}
            </span>
            <FieldInput field={f} value={draft[f.name]} onChange={(v) => set(f.name, v)} />
          </label>
        ))}
        {error && <p className="text-xs text-danger">{error}</p>}
      </div>
    </Modal>
  );
}
