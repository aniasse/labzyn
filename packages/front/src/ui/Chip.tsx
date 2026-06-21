// Adapted from twenty-ui (components/chip/Chip.tsx): uses Twenty's real
// Chip.module.scss; truncating text instead of the tooltip dependency.
import { clsx } from 'clsx';
import type { ReactNode } from 'react';
import styles from './Chip.module.scss';

export type ChipSize = 'large' | 'small';
export type ChipVariant = 'highlighted' | 'regular' | 'transparent' | 'static';
export type ChipAccent = 'text-primary' | 'text-secondary';

type ChipProps = {
  label: string;
  size?: ChipSize;
  variant?: ChipVariant;
  accent?: ChipAccent;
  disabled?: boolean;
  clickable?: boolean;
  leftComponent?: ReactNode;
  rightComponent?: ReactNode;
  maxWidth?: number;
  className?: string;
  onClick?: () => void;
};

export const Chip = ({
  label,
  size = 'small',
  variant = 'regular',
  accent = 'text-primary',
  disabled = false,
  clickable = true,
  leftComponent,
  rightComponent,
  maxWidth,
  className,
  onClick,
}: ChipProps) => {
  const backgroundClass =
    variant === 'highlighted'
      ? styles.backgroundHighlighted
      : variant === 'static'
        ? styles.backgroundStatic
        : variant === 'regular' && !disabled && clickable
          ? styles.interactiveRegular
          : undefined;

  return (
    <div
      onClick={onClick}
      className={clsx(
        styles.chip,
        size === 'large' ? styles.sizeLarge : styles.sizeSmall,
        accent === 'text-primary' ? styles.accentTextPrimary : styles.accentTextSecondary,
        disabled && styles.disabled,
        accent === 'text-secondary' && styles.fontMedium,
        variant !== 'transparent' && clickable && styles.cursorPointer,
        backgroundClass,
        maxWidth ? styles.hasMaxWidth : undefined,
        variant === 'transparent' && styles.paddingLeftNone,
        className,
      )}
      style={maxWidth ? ({ '--chip-max-width': `${maxWidth}px` } as React.CSSProperties) : undefined}
    >
      {leftComponent}
      <span className={styles.content ?? ''} style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>
        {label}
      </span>
      {rightComponent}
    </div>
  );
};
