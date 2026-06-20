import { forwardRef } from 'react';
import { cn } from '@/lib/utils';

export const Select = forwardRef<
  HTMLSelectElement,
  React.SelectHTMLAttributes<HTMLSelectElement>
>(({ className, children, ...props }, ref) => (
  <select
    ref={ref}
    className={cn(
      'h-8 w-full rounded-sm border border-border bg-background px-2 text-sm text-foreground',
      'outline-none transition-colors duration-fast focus:border-primary',
      className,
    )}
    {...props}
  >
    {children}
  </select>
));
Select.displayName = 'Select';
