import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { tokenStore } from '@/lib/auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export function LoginPage({ onAuthed }: { onAuthed: () => void }) {
  const [mode, setMode] = useState<'login' | 'signup'>('login');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [firstName, setFirstName] = useState('');
  const [workspaceName, setWorkspaceName] = useState('');
  const [error, setError] = useState<string | null>(null);

  const submit = useMutation({
    mutationFn: () =>
      mode === 'login'
        ? api.login({ email, password })
        : api.signup({ email, password, firstName, workspaceName }),
    onSuccess: (res) => {
      tokenStore.set(res.token);
      onAuthed();
    },
    onError: (e: Error) => setError(e.message),
  });

  return (
    <div className="flex h-full w-full items-center justify-center bg-muted">
      <form
        onSubmit={(e) => {
          e.preventDefault();
          setError(null);
          submit.mutate();
        }}
        className="w-full max-w-sm rounded-md border border-border bg-background p-6 shadow-sm"
      >
        <div className="mb-5 flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary text-md font-semibold text-primary-foreground">
            L
          </div>
          <span className="text-lg font-semibold text-foreground">Labzyn</span>
        </div>

        <h1 className="mb-1 text-md font-semibold text-foreground">
          {mode === 'login' ? 'Sign in' : 'Create your account'}
        </h1>
        <p className="mb-4 text-xs text-tertiary">
          {mode === 'login'
            ? 'Welcome back. Enter your credentials.'
            : 'Set up your account and first workspace.'}
        </p>

        <div className="flex flex-col gap-3">
          {mode === 'signup' && (
            <Field label="First name">
              <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} />
            </Field>
          )}
          <Field label="Email">
            <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
          </Field>
          <Field label="Password">
            <Input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </Field>
          {mode === 'signup' && (
            <Field label="Workspace name">
              <Input
                value={workspaceName}
                onChange={(e) => setWorkspaceName(e.target.value)}
                placeholder="My Lab"
              />
            </Field>
          )}

          {error && <p className="text-xs text-danger">{error}</p>}

          <Button type="submit" variant="primary" disabled={submit.isPending} className="mt-1 h-9">
            {submit.isPending
              ? 'Please wait…'
              : mode === 'login'
                ? 'Sign in'
                : 'Create account'}
          </Button>
        </div>

        <button
          type="button"
          className="mt-4 w-full text-center text-xs text-secondary hover:text-foreground"
          onClick={() => {
            setMode((m) => (m === 'login' ? 'signup' : 'login'));
            setError(null);
          }}
        >
          {mode === 'login'
            ? "Don't have an account? Sign up"
            : 'Already have an account? Sign in'}
        </button>
      </form>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="flex flex-col gap-1">
      <span className="text-xs font-medium text-secondary">{label}</span>
      {children}
    </label>
  );
}
