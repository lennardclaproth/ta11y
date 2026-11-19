---
description: Run the frontend type gate (svelte-check) and fix what it reports
allowed-tools: Bash, Read, Edit, Grep, Glob
---

Run `npm run check` from `web/` — this is the frontend's real gate.

Fix every error it reports. While fixing:

- Match each file's existing indentation. `web/` has known tabs-vs-spaces drift and
  `npm run lint` is not green out of the box, so **do not** reformat files as a drive-by and do not
  run `npm run format` across the tree.
- Keep Svelte 5 runes (`$props`, `$derived`, `$bindable`) — runes mode is forced.
- Respect the Atomic Design tier boundaries; a fix that makes an atom import another component means
  the component is in the wrong tier.

Re-run until clean, then report the before/after error count. If an error needs a decision you
can't make (an intentional `any`, a missing backend field), stop and ask instead of widening a type
to silence it.
