/**
 * One press treatment for every action in this design.
 *
 * Review of the revised design: the black bar under a pressed control had to go everywhere, so
 * the press is now exactly the treatment the existing buttons already use — the amber ring.
 * It is applied on `active:` as well as `focus-visible:`, because `focus-visible` stays off on a
 * mouse press and the click would otherwise give no feedback at all. The Button atom adds its own
 * 2% scale on top; nothing is drawn outside the control's own box, so a press never reflows a row.
 */
export const pressClasses = 'active:ring-2 active:ring-amber-300';
