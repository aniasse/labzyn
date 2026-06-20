export function HomePage() {
  return (
    <div className="mx-auto flex h-full max-w-2xl flex-col items-center justify-center gap-3 px-6 text-center">
      <div className="flex h-12 w-12 items-center justify-center rounded-md bg-primary text-lg font-semibold text-primary-foreground">
        L
      </div>
      <h1 className="text-xl font-semibold text-foreground">Welcome to Labzyn</h1>
      <p className="text-sm text-secondary">
        Pick an object in the sidebar, or press{' '}
        <kbd className="rounded-xs border border-border bg-muted px-1.5 py-0.5 text-xxs">
          ⌘K
        </kbd>{' '}
        to jump anywhere.
      </p>
    </div>
  );
}
