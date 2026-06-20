import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { Modal } from '@/components/ui/modal';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

const NAME_RE = /^[a-zA-Z][a-zA-Z0-9_]*$/;

export function CreateObjectDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const qc = useQueryClient();
  const [labelSingular, setLabelSingular] = useState('');
  const [labelPlural, setLabelPlural] = useState('');
  const [nameSingular, setNameSingular] = useState('');
  const [namePlural, setNamePlural] = useState('');
  const [error, setError] = useState<string | null>(null);

  const reset = () => {
    setLabelSingular('');
    setLabelPlural('');
    setNameSingular('');
    setNamePlural('');
    setError(null);
  };

  const create = useMutation({
    mutationFn: () =>
      api.createObject({
        nameSingular,
        namePlural,
        labelSingular: labelSingular || nameSingular,
        labelPlural: labelPlural || namePlural,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['objects'] });
      reset();
      onClose();
    },
    onError: (e: Error) => setError(e.message),
  });

  // Derive API names and the plural label from the singular label as the user types.
  const onLabelSingular = (v: string) => {
    setLabelSingular(v);
    const trimmed = v.trim();
    setLabelPlural(trimmed ? (trimmed.endsWith('s') ? trimmed : `${trimmed}s`) : '');
    const slug = trimmed.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_+|_+$/g, '');
    setNameSingular(slug);
    setNamePlural(slug ? (slug.endsWith('s') ? slug : `${slug}s`) : '');
  };

  const valid = NAME_RE.test(nameSingular) && NAME_RE.test(namePlural);

  return (
    <Modal
      open={open}
      onClose={() => {
        reset();
        onClose();
      }}
      title="New object"
      footer={
        <>
          <Button variant="secondary" onClick={() => { reset(); onClose(); }}>
            Cancel
          </Button>
          <Button
            variant="primary"
            disabled={!valid || create.isPending}
            onClick={() => create.mutate()}
          >
            {create.isPending ? 'Creating…' : 'Create'}
          </Button>
        </>
      }
    >
      <div className="flex flex-col gap-3">
        <Labeled label="Singular label">
          <Input value={labelSingular} onChange={(e) => onLabelSingular(e.target.value)} placeholder="Experiment" />
        </Labeled>
        <Labeled label="Plural label">
          <Input value={labelPlural} onChange={(e) => setLabelPlural(e.target.value)} placeholder="Experiments" />
        </Labeled>
        <div className="grid grid-cols-2 gap-3">
          <Labeled label="API name (singular)">
            <Input value={nameSingular} onChange={(e) => setNameSingular(e.target.value)} placeholder="experiment" />
          </Labeled>
          <Labeled label="API name (plural)">
            <Input value={namePlural} onChange={(e) => setNamePlural(e.target.value)} placeholder="experiments" />
          </Labeled>
        </div>
        {nameSingular && !valid && (
          <p className="text-xs text-danger">API names must match ^[a-zA-Z][a-zA-Z0-9_]*$</p>
        )}
        {error && <p className="text-xs text-danger">{error}</p>}
      </div>
    </Modal>
  );
}

function Labeled({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="flex flex-col gap-1">
      <span className="text-xs font-medium text-secondary">{label}</span>
      {children}
    </label>
  );
}
