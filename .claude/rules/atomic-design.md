---
paths:
  - "web/src/lib/components/**"
---

# Atomic Design (strict)

Components live in `web/src/lib/components/<tier>/` where tier is one of
**atoms → molecules → organisms → templates → pages**. The taxonomy is mandatory: never place a
component elsewhere, and only import from the same or a **lower** tier. Compose upward.

Populated today: `atoms/`, `molecules/`, `organisms/`, `templates/` (`app-shell`, `page-content`).
Create `pages/` when you add its first component, keeping the anatomy below.

## The boundary rule

An atom may render native HTML elements, inline SVG, and *external*-library primitives (the `Icon`
atom wraps `@iconify/svelte`). An atom must **never** import another `$lib/components` component.
Anything composing an internal atom is a molecule or higher.

- Atom-level controls take icons through `children`: `<Button><Icon icon="heroicons:check" /> Save</Button>`.
- Icon-only / icon-decorated controls are molecules (`molecules/icon-button`, `molecules/icon-input`).
- Reusing an atom's `*.types.ts` / `*.variants.ts` tokens from a molecule is fine — that's data, not
  composition.
- `@iconify/svelte` is imported in exactly one place: the `Icon` atom.

## Component anatomy (co-location)

```
button/
  Button.svelte          # markup + logic
  button.types.ts        # option vocabularies + derived unions
  button.variants.ts     # Tailwind class maps keyed by variant
  Button.stories.svelte  # Storybook CSF (catalog + tests)
```

Folders are lowercase kebab-case; `.svelte` and `*.stories.svelte` are PascalCase; `*.types.ts` /
`*.variants.ts` are lowercase. One folder may hold several related components sharing types and
variants (`Text.svelte` + `Heading.svelte` in `typography/`).

Simple atoms may inline types/variants (see `icon`, `checkbox`). Rule of thumb: extract once there
is a real variant matrix; inline for one trivial scale. Stories are expected for anything reusable.

### `*.types.ts`

Each option axis is an `as const` array with the union derived from it — the single source of truth
shared by the component and the story's `argTypes`:

```ts
export const buttonIntents = ['primary', 'secondary', 'warning', 'error', 'success', 'info'] as const;
export type ButtonIntent = (typeof buttonIntents)[number];
```

### `*.variants.ts`

Class strings stored as data, keyed by the union and guarded with `satisfies Record<…, string>`.
**Tailwind 4 rule:** every utility must appear as a complete literal so the content scanner and
`prettier-plugin-tailwindcss` can see it — never assemble one utility from fragments
(no `` `bg-${color}-500` ``). Joining several whole utilities with `.join(' ')` is fine.

### `Component.svelte`

- **Runes only** (runes mode is forced): `$props()`, `$derived()`, `$bindable()`.
- Local `type Props = { … }` (or `interface Props extends Omit<HTML…Attributes, …>` when wrapping a
  native element, as `Checkbox` does), with defaults in the destructure.
- `class` passthrough: accept `class?: string`, rename to `class: className = ''`, append **last**.
- Build the class string from an array → `.filter(Boolean).join(' ')`, wrapped in `$derived` when it
  depends on props, pulling pieces from the variants maps.
- `children?: Snippet` rendered via `{@render children?.()}`.
- Callback props for events (`onclick?: (e: MouseEvent) => void`), guarded — a disabled or loading
  control short-circuits before calling back.
- **Accessibility is not optional**: `aria-busy`/`aria-pressed` on Button, `aria-invalid`/
  `aria-describedby` on Input, the visually-hidden `peer sr-only` input behind a styled box on
  Checkbox, a **required** `ariaLabel` on icon-only controls. The a11y addon is wired into Storybook.
- `<svelte:element this={…}>` for polymorphic tags (`Heading` `level`, `Text` `as`).

## Shared prop vocabulary

Reuse these names rather than inventing new ones:

- `intent` — `primary | secondary | warning | error | success | info` (Badge adds `neutral`)
- `variant` — `solid | outline | ghost` (Badge: `soft | solid | outline`)
- `size` — `sm | md | lg` (typography has its own scales)
- `shape` — `default | rounded | pill`

**Read [docs/DESIGN.md](../../docs/DESIGN.md) before UI work** — it is the design authority for the
ta11y portal: visual principles, tokens, component usage, financial formatting, responsive layouts,
and accessibility acceptance criteria. Reuse what it documents and update it when a shared design
decision changes. (`docs/DESIGN.html` is the generated illustrated version, regenerated with
`node scripts/render-design.cjs`.)

Palette: Tailwind `slate / emerald / amber / red / sky` plus the custom **`taupe`** family for
backgrounds. Headings are EB Garamond. Keep the calm, precise, newspaper-inspired hierarchy —
ruled sections, aligned columns, square main panels, rounded controls and dialogs. Do not recolor
the brand or introduce heavy amber as a drive-by.

## Stories

```svelte
<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import Button from './Button.svelte';
  import { buttonIntents } from './button.types';

  const { Story } = defineMeta({
    title: 'Atoms/Button',
    component: Button,
    tags: ['autodocs'],
    argTypes: { intent: { control: 'select', options: buttonIntents } }
  });
</script>
```

`title` follows the tier (`Atoms/…`, `Molecules/…`, …). Feed `argTypes.options` from the
`*.types.ts` arrays so controls and types never drift. Provide a `Playground` plus focused stories
(variants, sizes, states) — these **are** the test corpus under Vitest browser mode.

## Adding a component — checklist

1. Pick the tier folder (create it if it's the tier's first component). Import only same/lower tier.
2. Write `Component.svelte` per the conventions above. If it composes another internal component,
   it is not an atom.
3. Add `component.types.ts` + `component.variants.ts` for a real variant matrix; inline for a
   trivial atom.
4. Add `Component.stories.svelte` with `defineMeta`, the tier-correct `title`, `tags: ['autodocs']`,
   and `argTypes` fed from the option arrays. Cover the meaningful states.
5. Run `npm run check` and preview in `npm run storybook`. Match the file's existing indentation.
