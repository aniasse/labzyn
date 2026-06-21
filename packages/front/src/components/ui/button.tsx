import { forwardRef } from 'react';
import { cn } from '@/lib/utils';
import styles from '@/ui/Button.module.scss';

// Uses Twenty's real Button.module.scss (data-variant / data-accent API),
// keeping a children-based API so existing call sites are unchanged.
type Variant = 'primary' | 'secondary' | 'ghost';
type Size = 'sm' | 'md';

const variantAttrs: Record<Variant, { variant: string; accent: string }> = {
  primary: { variant: 'primary', accent: 'blue' },
  secondary: { variant: 'secondary', accent: 'default' },
  ghost: { variant: 'tertiary', accent: 'default' },
};

export type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: Variant;
  size?: Size;
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'secondary', size = 'md', disabled, ...props }, ref) => {
    const attrs = variantAttrs[variant];
    return (
      <div className={styles.wrapper}>
        <button
          ref={ref}
          className={cn(styles.button, size === 'sm' ? styles.small : styles.medium, className)}
          data-variant={attrs.variant}
          data-accent={attrs.accent}
          data-position="standalone"
          data-disabled={disabled || undefined}
          disabled={disabled}
          {...props}
        />
      </div>
    );
  },
);
Button.displayName = 'Button';
