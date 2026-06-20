import { forwardRef } from 'react';
import { cn } from '@/lib/utils';

export const Input = forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  ({ className, ...props }, ref) => (
    <input
      ref={ref}
      className={cn(
        'h-8 w-full rounded-sm border border-border bg-background px-2 text-sm text-foreground',
        'outline-none transition-colors duration-fast placeholder:text-tertiary',
        'focus:border-primary',
        className,
      )}
      {...props}
    />
  ),
);
Input.displayName = 'Input';
