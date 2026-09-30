/**
 * One press treatment for every action in this design.
 *
 * Review of the final design: the press feedback of round 2 was its own invention (deeper amber,
 * inner shadow, a 1px shift) and read as foreign. This keeps the button exactly as the Button atom
 * already behaves — amber focus ring, 2% scale — and adds only a black bar under it, so the ring
 * also shows on a mouse press, where `focus-visible` stays off.
 *
 * `shadow-[…]` draws the bar, so pressing never reflows the row.
 */
export const pressClasses =
	'active:ring-2 active:ring-amber-300 active:shadow-[0_4px_0_0_var(--color-slate-900)]';
