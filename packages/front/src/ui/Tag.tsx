// Copied from twenty-ui (components/tag/Tag.tsx), adapted: relative imports,
// numeric icon sizes (no ThemeContext), plain truncating text (no tooltip dep).
import { clsx } from 'clsx';
import type { ComponentType } from 'react';
import { themeCssVariables } from './themeCssVariables';
import type { ThemeColor } from './themeColor';
import styles from './Tag.module.scss';

type IconComponent = ComponentType<{ size?: number; stroke?: number }>;

type TagWeight = 'regular' | 'medium';
type TagVariant = 'solid' | 'outline' | 'border';
export type TagColor = ThemeColor | 'transparent';

type TagProps = {
  className?: string;
  color: TagColor;
  text: string;
  Icon?: IconComponent;
  onClick?: () => void;
  weight?: TagWeight;
  variant?: TagVariant;
  preventShrink?: boolean;
  preventPadding?: boolean;
};

export const Tag = ({
  className,
  color,
  text,
  Icon,
  onClick,
  weight = 'regular',
  variant = 'solid',
  preventShrink,
  preventPadding,
}: TagProps) => {
  const tagBackground =
    color === 'transparent'
      ? 'transparent'
      : (themeCssVariables.tag.background[color] ?? themeCssVariables.tag.background.gray);

  const tagText =
    color === 'transparent'
      ? themeCssVariables.font.color.secondary
      : (themeCssVariables.tag.text[color] ?? themeCssVariables.font.color.secondary);

  return (
    <span
      className={clsx(
        styles.tag,
        weight === 'medium' && styles.weightMedium,
        variant === 'outline' && styles.variantOutline,
        variant === 'border' && styles.variantBorder,
        preventShrink && styles.preventShrink,
        preventPadding && styles.preventPadding,
        className,
      )}
      onClick={onClick}
      style={
        {
          '--tag-background': tagBackground,
          '--tag-text': tagText,
        } as React.CSSProperties
      }
    >
      {Icon ? (
        <div className={styles.iconContainer}>
          <Icon size={14} stroke={2} />
        </div>
      ) : null}
      <span className={preventShrink ? styles.nonShrinkableText : styles.content}>{text}</span>
    </span>
  );
};
