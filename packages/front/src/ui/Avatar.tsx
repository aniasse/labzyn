// Adapted from twenty-ui (display/avatar/components/Avatar.tsx): uses Twenty's
// real Avatar.module.scss; placeholder colors come from Twenty's tag tokens
// (no jotai / theme-util dependency).
import { clsx } from 'clsx';
import type { ComponentType } from 'react';
import { themeCssVariables } from './themeCssVariables';
import { colorForString } from './themeColor';
import styles from './Avatar.module.scss';

export type AvatarSize = 'xl' | 'lg' | 'md' | 'sm' | 'xs';
export type AvatarType = 'squared' | 'rounded' | 'icon';

type Props = {
  avatarUrl?: string | null;
  placeholder: string;
  placeholderColorSeed?: string;
  size?: AvatarSize;
  type?: AvatarType;
  Icon?: ComponentType<{ size?: number; color?: string }>;
  className?: string;
  onClick?: () => void;
};

export const Avatar = ({
  avatarUrl,
  placeholder,
  placeholderColorSeed,
  size = 'md',
  type = 'squared',
  Icon,
  className,
  onClick,
}: Props) => {
  const char = placeholder?.trim()?.charAt(0)?.toUpperCase() || '-';
  const color = colorForString(placeholderColorSeed ?? placeholder ?? '');
  const showImage = !!avatarUrl;

  const style = {
    '--avatar-color': themeCssVariables.tag.text[color],
    '--avatar-background': Icon ? 'inherit' : themeCssVariables.tag.background[color],
  } as React.CSSProperties;

  return (
    <div
      className={clsx(styles.root, styles[size], className)}
      data-type={type}
      data-clickable={onClick ? true : undefined}
      onClick={onClick}
      style={style}
    >
      {Icon ? (
        <Icon size={16} color="currentColor" />
      ) : showImage ? (
        <img className={styles.image} src={avatarUrl ?? ''} alt="" />
      ) : (
        <span className={styles.placeholderChar}>{char}</span>
      )}
    </div>
  );
};
