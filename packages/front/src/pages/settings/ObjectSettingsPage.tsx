import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Plus, Trash2 } from 'lucide-react';
import { api } from '@/lib/api';
import { getIcon } from '@/lib/icons';
import { Button } from '@/components/ui/button';
import { CreateFieldDialog } from '@/components/settings/CreateFieldDialog';

export function ObjectSettingsPage() {
  const { nameSingular = '' } = useParams();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [creating, setCreating] = useState(false);

  const { data: object, isError } = useQuery({
    queryKey: ['object', nameSingular],
    queryFn: () => api.getObject(nameSingular),
  });

  const deleteObject = useMutation({
    mutationFn: () => api.deleteObject(nameSingular),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['objects'] });
      navigate('/settings/objects');
    },
  });
  const deleteField = useMutation({
    mutationFn: (fieldId: string) => api.deleteField(nameSingular, fieldId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['object', nameSingular] }),
  });

  if (isError) {
    return <div className="p-6 text-sm text-danger">Object not found.</div>;
  }

  const Icon = getIcon(object?.icon);
  const fields = object?.fields ?? [];

  return (
    <div className="mx-auto max-w-3xl p-6">
      <div className="mb-4 flex items-center gap-2">
        <button
          type="button"
          onClick={() => navigate('/settings/objects')}
          className="rounded-sm p-1 text-tertiary hover:bg-accent"
          aria-label="Back"
        >
          <ArrowLeft size={16} />
        </button>
        <Icon size={18} className="text-secondary" />
        <h1 className="text-lg font-semibold text-foreground">{object?.labelPlural ?? '…'}</h1>
        <Button
          variant="primary"
          size="sm"
          className="ml-auto"
          onClick={() => setCreating(true)}
          disabled={!object}
        >
          <Plus size={14} strokeWidth={2.5} />
          New field
        </Button>
        <Button
          variant="ghost"
          size="sm"
          className="text-danger hover:bg-danger/10"
          disabled={!object}
          onClick={() => {
            if (window.confirm(`Delete “${object?.labelPlural}” and all its records?`)) {
              deleteObject.mutate();
            }
          }}
        >
          <Trash2 size={14} />
          Delete
        </Button>
      </div>

      <div className="overflow-hidden rounded-md border border-border">
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="bg-muted text-left text-xs text-tertiary">
              <th className="px-3 py-2 font-medium">Field</th>
              <th className="px-3 py-2 font-medium">API name</th>
              <th className="px-3 py-2 font-medium">Type</th>
              <th className="px-3 py-2 font-medium">Required</th>
              <th className="w-10 px-3 py-2" />
            </tr>
          </thead>
          <tbody>
            {fields.map((f) => (
              <tr key={f.id} className="group border-t border-border">
                <td className="px-3 py-2 text-foreground">{f.label}</td>
                <td className="px-3 py-2 text-tertiary">{f.name}</td>
                <td className="px-3 py-2 text-secondary">{f.type}</td>
                <td className="px-3 py-2 text-secondary">{f.isNullable ? '—' : 'yes'}</td>
                <td className="px-2 py-2 text-right">
                  <button
                    type="button"
                    aria-label={`Delete field ${f.label}`}
                    onClick={() => {
                      if (window.confirm(`Delete field “${f.label}”?`)) deleteField.mutate(f.id);
                    }}
                    className="rounded-sm p-1 text-tertiary opacity-0 transition-opacity duration-fast hover:bg-danger/10 hover:text-danger group-hover:opacity-100"
                  >
                    <Trash2 size={14} />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {object && fields.length === 0 && (
          <div className="p-6 text-center text-sm text-tertiary">No fields yet.</div>
        )}
      </div>

      {object && (
        <CreateFieldDialog
          nameSingular={nameSingular}
          open={creating}
          onClose={() => setCreating(false)}
        />
      )}
    </div>
  );
}
