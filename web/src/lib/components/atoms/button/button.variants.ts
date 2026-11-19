import type { ButtonIntent, ButtonSize, ButtonVariant, ButtonShape } from './button.types';

// Shared button styling tokens. These are styling data (not components), so molecules such as
// IconButton may import them without breaking the atom/molecule boundary.

export const baseButtonClasses = [
  'inline-flex items-center justify-center gap-1',
  'select-none whitespace-nowrap',
  'transition-all duration-150 ease-out',
  // Amber is the app's focus colour; intents may override it, but the default keeps
  // keyboard focus consistent across every button.
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-amber-300',
  'disabled:pointer-events-none disabled:opacity-50',
  // A press shows the same amber ring as keyboard focus. `focus-visible` stays off for pointer
  // input, so without this a mouse click on a solid button -- whose `active:` colour equals its
  // resting colour -- gives no feedback at all. It is a colour change, not only the scale below,
  // so it survives a reduced-motion preference.
  'active:ring-2 active:ring-amber-300',
  'active:scale-[0.98]'
].join(' ');

export const buttonSizeClasses = {
  sm: 'h-8 px-2 text-sm',
  md: 'h-10 px-3 text-sm',
  lg: 'h-12 px-3 text-base'
} satisfies Record<ButtonSize, string>;

export const iconSizeClasses = {
  sm: 'size-2',
  md: 'size-3',
  lg: 'size-4'
} satisfies Record<ButtonSize, string>;

export const buttonShapeClasses = {
  default: 'rounded-md',
  rounded: 'rounded-xl',
  pill: 'rounded-full'
} satisfies Record<ButtonShape, string>;

// The `ruled` variant carries its own geometry -- it is a rule with a label on it, not a box --
// so it replaces the shape instead of being rounded by it.
export const ruledShapeClasses = 'rounded-t-sm';

export const intentVariantClasses = {
  // Primary is amber-filled with slate ink; the outline/ghost variants invert that
  // into slate-on-transparent for secondary actions. All three focus amber.
  primary: {
    solid:
      'bg-amber-400 text-slate-800 border border-slate-800 hover:bg-amber-300 focus:ring-amber-300 focus:ring-2 focus:bg-amber-400 active:bg-amber-400',
    outline:
      'border border-slate-700 text-slate-700 hover:bg-amber-100 focus:ring-amber-300 focus:ring-2 focus:bg-amber-100 active:bg-amber-200',
    ghost:
      'text-slate-700 hover:bg-amber-100 focus:ring-amber-300 focus:ring-2 focus:bg-amber-100 active:bg-amber-200',
    ruled:
      'border-b border-slate-400 text-slate-700 hover:border-slate-700 hover:bg-amber-100 hover:text-slate-950 focus:ring-amber-300 focus:ring-2 active:border-slate-900 active:bg-amber-200 active:text-slate-950'
  },

  secondary: {
    solid:
      'bg-emerald-400 text-slate-800 border-1 border-slate-800 hover:bg-emerald-500 focus:ring-emerald-200 focus:ring-2 focus:bg-emerald-400 active:bg-emerald-400',
    outline:
      'border border-slate-800 text-slate-800 hover:bg-emerald-100 focus:ring-emerald-50 focus:ring-2 focus:bg-emerald-200 active:bg-emerald-200',
    ghost:
      'text-slate-800 hover:bg-emerald-100 focus:ring-emerald-50 focus:ring-2 focus:bg-emerald-200 active:bg-emerald-200',
    ruled:
      'border-b border-slate-400 text-slate-800 hover:border-slate-700 hover:bg-emerald-100 hover:text-slate-950 focus:ring-emerald-200 focus:ring-2 active:border-slate-900 active:bg-emerald-200'
  },

  warning: {
    solid:
      'bg-amber-500 text-slate-800 hover:bg-amber-400 focus:ring-amber-300 focus:ring-2 focus:bg-amber-500 active:bg-amber-500',
    outline:
      'border border-amber-600 text-amber-800 hover:bg-amber-100 focus:ring-amber-50 focus:ring-2 focus:bg-amber-200 active:bg-amber-200',
    ghost:
      'text-amber-800 hover:bg-amber-100 focus:ring-amber-50 focus:ring-2 focus:bg-amber-200 active:bg-amber-200',
    ruled:
      'border-b border-amber-600 text-amber-800 hover:border-amber-700 hover:bg-amber-100 focus:ring-amber-300 focus:ring-2 active:border-amber-800 active:bg-amber-200'
  },

  error: {
    solid:
      'bg-red-700 text-white hover:bg-red-600 focus:ring-red-300 focus:ring-2 focus:bg-red-700 active:bg-red-700',
    outline:
      'border border-red-600 text-red-700 hover:bg-red-100 focus:ring-red-50 focus:ring-2 focus:bg-red-200 active:bg-red-200',
    ghost:
      'text-red-700 hover:bg-red-50 focus:ring-red-50 focus:ring-2 focus:bg-red-200 active:bg-red-200',
    ruled:
      'border-b border-red-600 text-red-700 hover:border-red-700 hover:bg-red-50 focus:ring-red-300 focus:ring-2 active:border-red-800 active:bg-red-100'
  },

  success: {
    solid:
      'bg-emerald-700 text-white hover:bg-emerald-500 focus:ring-emerald-400 focus:ring-2 focus:bg-emerald-700 active:bg-emerald-700',
    outline:
      'border border-emerald-700 text-emerald-800 hover:bg-emerald-100 focus:ring-emerald-50 focus:ring-2 focus:bg-emerald-200 active:bg-emerald-200',
    ghost:
      'text-emerald-800 hover:bg-emerald-100 focus:ring-emerald-50 focus:ring-2 focus:bg-emerald-200 active:bg-emerald-200',
    ruled:
      'border-b border-emerald-700 text-emerald-800 hover:border-emerald-800 hover:bg-emerald-100 focus:ring-emerald-300 focus:ring-2 active:border-emerald-900 active:bg-emerald-200'
  },

  info: {
    solid:
      'bg-sky-700 text-white hover:bg-sky-500 focus:ring-sky-400 focus:ring-2 focus:bg-sky-700 active:bg-sky-700',
    outline:
      'border border-sky-600 text-sky-700 hover:bg-sky-100 focus:ring-sky-50 focus:ring-2 focus:bg-sky-200 active:bg-sky-200',
    ghost:
      'text-sky-700 hover:bg-sky-100 focus:ring-sky-50 focus:ring-2 focus:bg-sky-200 active:bg-sky-200',
    ruled:
      'border-b border-sky-600 text-sky-700 hover:border-sky-700 hover:bg-sky-100 focus:ring-sky-300 focus:ring-2 active:border-sky-800 active:bg-sky-200'
  }
} satisfies Record<ButtonIntent, Record<ButtonVariant, string>>;