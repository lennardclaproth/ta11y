// button.types.ts

export const buttonIntents = [
  'primary',
  'secondary',
  'warning',
  'error',
  'success',
  'info'
] as const;

// `ruled` is the editorial supporting action: it hangs on a rule instead of a filled or
// outlined box, so a region keeps exactly one visually dominant button.
export const buttonVariants = [
  'solid',
  'outline',
  'ghost',
  'ruled'
] as const;

export const buttonSizes = [
  'sm',
  'md',
  'lg'
] as const;

export const buttonShapes = [
  'default',
  'rounded',
  'pill'
] as const;

export type ButtonIntent = typeof buttonIntents[number];
export type ButtonVariant = typeof buttonVariants[number];
export type ButtonSize = typeof buttonSizes[number];
export type ButtonShape = typeof buttonShapes[number];

export type ButtonType = 'button' | 'submit' | 'reset';