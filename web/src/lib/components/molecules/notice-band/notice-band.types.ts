export const noticeIntents = ['info', 'success', 'warning', 'error'] as const;

export type NoticeIntent = (typeof noticeIntents)[number];

/**
 * One outcome a screen holds on to: what happened and in which intent it is printed. Shared, so the
 * overlays that report their own outcomes in a band rather than through the toast store all speak
 * the same shape.
 */
export type NoticeMessage = { intent: NoticeIntent; message: string };

/**
 * Surface the band is printed on. `paper` is a fresh white sheet on the taupe canvas and inside a
 * taupe-50 panel; `inset` is the warm sheet used inside a white dialog or drawer, where white on
 * white would vanish.
 */
export const noticeBandSurfaces = ['paper', 'inset'] as const;

export type NoticeBandSurface = (typeof noticeBandSurfaces)[number];

/**
 * Horizontal gutter. The band runs edge to edge, but its text lines up with the text of whatever
 * contains it: the page gutter under the masthead, the panel gutter above the records, the dialog
 * gutter under a modal header. `none` is for a container that already pads its own content.
 */
export const noticeBandGutters = ['page', 'panel', 'dialog', 'none'] as const;

export type NoticeBandGutter = (typeof noticeBandGutters)[number];
