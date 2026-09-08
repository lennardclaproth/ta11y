---
name: new-component
description: Create a SvelteKit component in the right Atomic Design tier with its four co-located files (component, types, variants, stories). Use when adding a component under web/src/lib/components.
---

# Add a frontend component

## 1. Pick the tier — this is the decision that matters

- **atom** — renders only native elements, inline SVG, or an external primitive. An atom must
  **never** import another `$lib/components` component.
- **molecule** — composes atoms. Anything that renders `Icon` alongside a control is a molecule.
- **organism** — a self-contained section composed of molecules and atoms.
- **template** — page-level layout without real data (`app-shell`, `page-content`).
- **page** — concrete instance; create the tier folder when you add its first component.

Import only from the same or a lower tier. This is the most common review finding: if you reach for
`Icon` inside an atom, the component belongs in `molecules/` instead.

## 2. Create the folder and files

Under `web/src/lib/components/<tier>/<kebab-case-name>/`:

```
money-field/
  MoneyField.svelte
  money-field.types.ts
  money-field.variants.ts
  MoneyField.stories.svelte
```

A trivial atom with one scale may inline its types and classes — see `icon/` and `checkbox/`.

## 3. The types file

```ts
export const moneyFieldSizes = ['sm', 'md', 'lg'] as const;
export type MoneyFieldSize = (typeof moneyFieldSizes)[number];
```

The `as const` array is the single source of truth, shared with the story's `argTypes`.

## 4. The variants file

```ts
export const moneyFieldSizeClasses = {
	sm: 'h-8 px-2 text-sm',
	md: 'h-10 px-3 text-sm',
	lg: 'h-12 px-3 text-base'
} satisfies Record<MoneyFieldSize, string>;
```

Every Tailwind utility must appear as a complete literal so the content scanner and
`prettier-plugin-tailwindcss` can see it — never assemble one utility from fragments. Joining
several whole utilities with `.join(' ')` is fine.

## 5. The component

Runes only: `$props()`, `$derived()`, `$bindable()`. Declare a local `type Props`, with defaults in
the destructure. Accept `class?: string`, rename it to `class: className = ''`, and append it
**last**. Build the class string from an array, then `.filter(Boolean).join(' ')`, wrapped in
`$derived` when it depends on props. Render `children?: Snippet` with `{@render children?.()}`.
Expose callback props for events and guard them — a disabled or loading control short-circuits
before calling back.

Accessibility is not optional: mirror the existing atoms (`aria-busy`/`aria-pressed` on Button,
`aria-invalid`/`aria-describedby` on Input, the visually-hidden `peer sr-only` input on Checkbox),
and an icon-only control takes a **required** `ariaLabel`.

Reuse the shared prop vocabulary: `intent`, `variant`, `size`, `shape`.

## 6. The stories file

`defineMeta` with the tier-correct `title` (`Atoms/...`, `Molecules/...`), `tags: ['autodocs']`, and
`argTypes.options` fed from the types file's arrays. Add a `Playground` plus focused stories for
variants, sizes and states — under Vitest browser mode these **are** the tests.

## 7. Finish

Run `npm run check` from `web/` — that is the gate — then preview in `npm run storybook`. Match the
file's existing indentation and don't run `npm run format` across the tree; `web/` has known
tabs-vs-spaces drift and `npm run lint` is not green out of the box.
