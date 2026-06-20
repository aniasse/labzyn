import type { Config } from 'tailwindcss';

// Design tokens ported from Twenty's design system (twenty-ui/theme):
// Inter, 4px spacing base (Tailwind default), Twenty's type scale and radii,
// Radix-derived accent. Values live as CSS variables in src/index.css so the
// light/dark themes can swap them.
export default {
  darkMode: ['class'],
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      // Twenty FONT_COMMON.size
      fontSize: {
        xxs: '0.625rem',
        xs: '0.85rem',
        sm: '0.92rem',
        md: '1rem',
        lg: '1.23rem',
        xl: '1.54rem',
        xxl: '1.85rem',
      },
      // Twenty BORDER_COMMON.radius
      borderRadius: {
        xs: '2px',
        sm: '4px',
        md: '8px',
        xl: '20px',
        xxl: '40px',
        pill: '999px',
      },
      colors: {
        background: 'var(--background)',
        foreground: 'var(--foreground)',
        secondary: 'var(--secondary)',
        tertiary: 'var(--tertiary)',
        border: 'var(--border)',
        muted: { DEFAULT: 'var(--muted)', foreground: 'var(--muted-foreground)' },
        accent: { DEFAULT: 'var(--accent)', foreground: 'var(--accent-foreground)' },
        primary: { DEFAULT: 'var(--primary)', foreground: 'var(--primary-foreground)' },
        sidebar: { DEFAULT: 'var(--sidebar)', border: 'var(--sidebar-border)' },
        danger: 'var(--danger)',
      },
      transitionDuration: {
        instant: '75ms',
        fast: '150ms',
        normal: '300ms',
      },
    },
  },
  plugins: [],
} satisfies Config;
