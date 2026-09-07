---
name: ta11y
description: A calm, precise personal ledger for cashflow, assets, and investments.
colors:
  taupe-50: "#faf8f4"
  taupe-100: "#f2ede4"
  taupe-200: "#e7dfd1"
  slate-50: "oklch(98.4% 0.003 247.858)"
  slate-100: "oklch(96.8% 0.007 247.896)"
  slate-200: "oklch(92.9% 0.013 255.508)"
  slate-300: "oklch(86.9% 0.022 252.894)"
  slate-500: "oklch(55.4% 0.046 257.417)"
  slate-600: "oklch(44.6% 0.043 257.281)"
  slate-700: "oklch(37.2% 0.044 257.287)"
  slate-800: "oklch(27.9% 0.041 260.031)"
  slate-900: "oklch(20.8% 0.042 265.755)"
  slate-950: "oklch(12.9% 0.042 264.695)"
  amber-100: "oklch(96.2% 0.059 95.617)"
  amber-200: "oklch(92.4% 0.12 95.746)"
  amber-500: "oklch(76.9% 0.188 70.08)"
  emerald-400: "oklch(76.5% 0.177 163.223)"
  emerald-700: "oklch(50.8% 0.118 165.612)"
  red-700: "oklch(50.5% 0.213 27.518)"
  sky-700: "oklch(50% 0.134 242.749)"
  white: "#ffffff"
typography:
  heading:
    fontFamily: "EB Garamond, ui-serif, serif"
    fontSize: "30px"
    lineHeight: 1.25
    letterSpacing: "-0.025em"
  body:
    fontFamily: "Noto Sans, ui-sans-serif, system-ui, sans-serif"
    fontSize: "16px"
    lineHeight: 1.625
  control:
    fontFamily: "Noto Sans, ui-sans-serif, system-ui, sans-serif"
    fontSize: "14px"
rounded:
  square: "0px"
  default: "6px"
  navigation: "8px"
  rounded: "12px"
  panel-xl: "16px"
  pill: "9999px"
spacing:
  tight: "4px"
  small: "8px"
  compact: "12px"
  regular: "16px"
  roomy: "24px"
components:
  button-primary:
    backgroundColor: "{colors.slate-600}"
    textColor: "{colors.amber-200}"
    rounded: "{rounded.default}"
    height: "40px"
    padding: "0 12px"
  button-outline:
    backgroundColor: "transparent"
    textColor: "{colors.slate-700}"
    rounded: "{rounded.default}"
  button-ghost:
    backgroundColor: "transparent"
    textColor: "{colors.slate-700}"
    rounded: "{rounded.default}"
  input:
    backgroundColor: "{colors.white}"
    textColor: "{colors.slate-900}"
    rounded: "{rounded.default}"
    height: "40px"
  navigation-active:
    backgroundColor: "{colors.amber-100}"
    textColor: "{colors.slate-900}"
  badge-neutral:
    backgroundColor: "{colors.slate-100}"
    textColor: "{colors.slate-700}"
  content-panel:
    backgroundColor: "{colors.taupe-50}"
    textColor: "{colors.slate-800}"
    rounded: "{rounded.square}"
---

# Design System: ta11y

## Overview

**Creative North Star: "A calm, precise personal ledger"**

ta11y (pronounced “tally”) helps people understand cashflow, assets, and investments. Its portal pairs warm paper-like surfaces with precise financial information: serif headings establish hierarchy, sans-serif controls support scanning, and slate-and-amber actions provide a recognizable signature.

This document captures the existing light-theme portal. It is the shared design reference for new screens and changes to reusable components. The product spelling is **ta11y**; repository names and API module paths may retain `my-finances-tracker`. Existing SVG lockups contain “Tally”; preserve those supplied assets until a revised wordmark is provided rather than silently redrawing them.

**Key Characteristics:**

- Warm taupe canvas with white content surfaces.
- EB Garamond headings and Noto Sans interface text.
- Compact, aligned financial data with clear units and periods.
- Slate-and-amber primary actions; semantic color for financial direction and feedback.
- Ruled analytics, square paper content panels, rounded controls, and short state transitions.

### Editorial direction

**Approved direction for future work:** ta11y should feel like a carefully composed newspaper:
calm, precise, and editorial. Build on its existing serif headings and warm paper palette with
strong headline hierarchy, fine dividing rules, aligned columns, and restrained accent color.
Use whitespace and rules to organize reading surfaces; retain the existing rounded controls and
panels where they support operational tasks. Avoid aged paper effects, decorative news copy, and
dense columns that make financial information harder to read.

The first portal rollout implements the ruled ta11y masthead, visible desktop navigation, labeled
mobile menu, serif analytics headings, flat KPI sections, and square paper content panels. More
extensive editorial treatments remain forward-looking; dialogs and operational controls retain
their existing shapes and behavior. Primary creation actions on Cashflow, Assets and Portfolio
use labeled buttons with a plus icon in a ruled ledger header, immediately above the records.
Portfolio shares this header with its view tabs. Actions remain in normal document flow and wrap on mobile.
The standalone [HTML guide](DESIGN.html) demonstrates the editorial direction; this Markdown file
remains the editable source of truth. Regenerate the HTML with `node scripts/render-design.cjs`
after changing this document or its companion metadata (requires the existing frontend dependencies).
The HTML embeds fonts and styles and opens offline without running the application.
Its masthead and reading layout use a larger editorial type hierarchy than operational portal
screens. Specimen controls have enhanced dark keyboard-focus outlines for the standalone guide;
these presentation enhancements do not change the runtime components.

### Sources and ownership

| Source | Responsibility |
| --- | --- |
| [Canonical theme](web/src/app.css) | Fonts, taupe palette, body and heading defaults |
| [Layout stylesheet](web/src/routes/layout.css) | Imports the canonical theme into the portal |
| [Component library](web/src/lib/components/) | Executable component APIs and variant styles |
| [Theme stories](web/src/lib/foundations/Theme.stories.svelte) | Foundation examples in Storybook |
| [Chart theme](web/src/lib/charts/theme.ts) | Chart colors, ramps, and geometry |
| [Overlay layers](web/src/lib/styles/z-index.ts) | Shared stacking order |
| [Brand assets](web/static/brand/) | Existing SVG marks, lockups, and illustrations |
| [Frontend conventions](web/AGENTS.md) | Atomic Design boundaries and component implementation |

The frontmatter records the extracted token values. Runtime styles remain in the sources above; this document does not inject CSS. Update implementation and documentation together when changing a shared decision. Tailwind's installed default palette supplies the slate, amber, emerald, red, and sky utilities; preserve its OKLCH values. Chart hex values are a separate existing implementation, not exact aliases of Tailwind 4 colors.

## Colors

The palette balances warm neutral backgrounds with cool, legible slate text.

### Primary

Primary solid actions use slate-600 with amber-200 text and border; hover uses slate-500. Outline and ghost primary actions use slate-700 text with pale slate interaction surfaces. Amber also identifies the active navigation item. Keep amber accents localized so the main action remains easy to find.

### Secondary and semantic roles

| Role | Existing palette | Usage |
| --- | --- | --- |
| Secondary action | Emerald-400 surface, slate-800 text | Supporting action using the secondary intent |
| Success | Emerald-700; pale emerald support surfaces | Completed operations and positive values where appropriate |
| Error | Red-700; pale red support surfaces | Validation failures and destructive actions |
| Warning | Amber-500 with slate-800 text | Conditions that need attention |
| Information | Sky-700; pale sky support surfaces | Informational feedback |

**The Meaning Rule.** Pair semantic color with a label, sign, icon, or explanation. An outgoing payment is a financial direction, not necessarily an application error; an increase in expenses is not automatically a positive trend.

### Neutrals

Use taupe-100 for the app canvas, white for normal/floating panels, and taupe-50 for muted panels. The full taupe scale lives in the canonical theme. Body text starts at slate-800; the Text atom normally uses slate-700. Muted text uses slate-500, strong text slate-950, and common control borders slate-300.

No dark theme is currently defined. A future dark theme requires explicit role mapping and contrast checks rather than automatic inversion.

## Typography

| Role | Family | Existing scale and treatment |
| --- | --- | --- |
| Headings | EB Garamond | `sm` 18px, `md` 20px, `lg` 24px, `xl` 30px, `2xl` 36px; tight leading and tracking |
| Interface/body | Noto Sans | Text `xs` 12px, `sm` 14px, `md` 16px, `lg` 18px; relaxed leading |
| Controls | Noto Sans | 14px for small/medium controls; 16px for large controls |
| Monetary values | Noto Sans | Reuse Money's tabular numeric styling and formatter |

Heading weights are medium, semibold, and bold; Text also supports normal. The `font-display` token currently resolves to Noto Sans; use `font-heading` for serif headings. Fonts are locally served from `web/static/fonts/` with swap loading.

**The Hierarchy Rule.** Choose heading level for document structure and size for visual hierarchy. Use one page h1 and meaningful subordinate levels. Reserve small text for secondary information, not the only explanation of an error or amount.

### Financial formatting and language

- Use the Money atom for currency output and CurrencyInput for monetary entry. Money accepts major units, defaults to EUR and `en-US`, and supports explicit locale/currency overrides. Apply the same locale within a view.
- Convert API fixed-point integers using the helpers in [money.ts](web/src/lib/api/money.ts) before display: these values use a scale of 1,000,000, including fields misleadingly named `amountCents`. Decimal-string endpoints follow their own contract.
- Align comparable amounts to the right and keep precision consistent within a column. Name the currency and time period where context could be ambiguous.
- Keep zero, unavailable data, and loading visibly distinct. Never substitute a displayed zero for an unknown amount in a new flow.
- Use direct sentence-case labels: “Add transaction”, “Save changes”, “Try again”. Error copy should explain what failed and what the person can do next. Avoid blame or judgment about spending.

## Layout

The portal is an operational workspace with a top navigation region, analytics above the main content, and a primary content panel that owns scrolling.

| Pattern | Existing implementation |
| --- | --- |
| App shell | `AppShellTemplate`: dynamic viewport-height flex column; main region scrolls below `lg`, while desktop panels own scrolling; skip link targets main content |
| Content | `PageContentTemplate`: 16px horizontal padding (32px at `lg`), 20px section gaps, 24px bottom padding |
| Main panel | Square muted Panel, no shadow or internal padding; 32rem tall below `lg`, remaining available height on desktop |
| Top navigation | Ruled ta11y masthead; direct links from `md`, labeled menu below; title, search, date, actions, and account controls wrap below the masthead |
| KPI grid | One column initially, two at `sm`, three or four at `lg` depending on configuration; 12px gap |
| Panel padding | None, 12px, 16px, or 24px |
| Controls | Small 32px, medium 40px, large 48px high |

The existing breakpoints are Tailwind defaults: `sm` 640px, `md` 768px, `lg` 1024px. These are layout transitions, not promises of validated device coverage.

For new screens, preserve the shell and allow the intended content region to scroll. Stack analytics and wrap toolbars on narrow screens; contain horizontal overflow inside wide tables rather than the whole page. Keep row actions and pagination reachable. Test with long labels, empty results, many rows, and browser zoom so fixed-height regions do not conceal controls.

## Elevation & Depth

Main content uses a square taupe-50 panel without shadow. AnalyticsCard and StatCard use transparent
sections with a fine top rule rather than elevated cards. Other panels retain white/taupe surfaces,
optional slate borders, and `none`, `sm`, or `md` shadows. The floating action button uses a larger
shadow to distinguish it from content.

**The Layer Rule.** Import `zClasses` or `zLayers` from the shared overlay module rather than choosing an arbitrary z-index.

| Layer | Value |
| --- | --- |
| Chart overlay | 10 |
| Sticky table header | 20 |
| Popover / table footer | 30 |
| Floating action button / expanded scrim | 40 |
| Modal scrim / toast host | 50 |
| Portaled filter popover | 60 |
| Async search dropdown | 70 |

Verify overlays together: a dropdown inside a dialog must remain usable, and an unrelated background control must not appear interactive over a modal.

## Shapes

Buttons and inputs share `default` (6px), `rounded` (12px), and `pill` shapes. Panels support `square`
(0px), `sm` (6px), `md` (12px), and `xl` (16px). Navigation triggers also use 8px rounding. The main
content panel uses square corners; existing dialog and control rounding remains available.

Use the supplied SVG artwork at its original aspect ratio. `logo-horizontal.svg` is the existing lockup; `tally-mark.svg`, `monogram.svg`, and `favicon.svg` provide compact marks; `ledger.svg` and `growth.svg` are supporting artwork. Their embedded gold treatment belongs to those assets and does not establish a gradient requirement for ordinary controls. Provide accessible naming for informative brand artwork and hide purely decorative artwork from assistive technology.

## Components

### Architecture and reuse

Follow **atoms → molecules → organisms → templates → pages**. Atoms wrap native elements or external primitives and never import internal components. Molecules compose atoms; organisms provide complete sections; templates establish page structure; route pages wire data and actions. Compose within or below the current tier.

Keep reusable components under `web/src/lib/components/` and co-locate their Svelte implementation, type vocabulary, variant map, and stories where applicable. Reuse complete Tailwind class strings; do not construct utility names dynamically. Svelte 5 props, snippets, and typed callbacks are the established API style.

### Component selection

| Need | Reuse | Usage rule |
| --- | --- | --- |
| Primary/supporting action | Button | `intent`, `variant`, `size`, `shape`; one visually dominant action per task region |
| Icon-only action | IconButton | Supply the required accessible label; use the Icon atom for icon rendering |
| Text entry | Input, Textarea, FormField | Visible label, optional help, linked error; placeholders supplement labels |
| Search or affixes | SearchInput, IconInput | Compose the input atom instead of duplicating its focus and error styles |
| Money | Money, CurrencyInput | Explicit currency/locale and correct input units |
| Classification/status | Badge | `soft`, `solid`, `outline`; use semantic text, not color alone |
| Surface | Panel | Choose background, padding, border, shape, and shadow deliberately |
| Navigation/filtering | NavMenu, Tabs, filter molecules | Show current state and preserve the query/selection context |
| Financial summary | StatCard, KpiRow, AnalyticsCard | Explain metric, period, and comparison; trend direction needs domain meaning |
| Records | DataTable, CashflowTransactionsTable | Align comparable values; preserve sort/filter/pagination and selection visibility |
| Detail/edit | Drawer, Dialog, TransactionFormModal | Keep context visible where useful; isolate blocking decisions in dialogs |
| Feedback | Alert, ToastHost, Skeleton, Spinner | Persistent issues belong near the affected content; transient feedback uses the single shell toast host |

Do not assume every component implements every intent: Button supports primary/secondary/warning/error/success/info; Badge adds neutral; Input uses default/error/success. Read each component's types before extending it.

### States and accessibility acceptance criteria

Existing controls include focus rings, disabled treatment, and ARIA hooks. The following are acceptance criteria for changes, not a claim that the entire portal has passed an accessibility audit:

- Cover default, hover, keyboard focus, pressed/selected, disabled, loading, error, and success where applicable. Prevent duplicate submissions while a request is pending.
- Associate labels, help, and errors with the relevant field. Preserve entered data after a failed save and provide a retry path.
- Make interactive elements keyboard reachable; keep focus visible. Dialogs must manage initial focus, trap focus while modal, and restore it on close. Menus need appropriate keyboard navigation and dismissal.
- Check text, control boundaries, and focus indicators for sufficient contrast in their actual background combinations. Do not assume every palette shade is interchangeable.
- Provide usable touch targets and spacing; prefer large controls on touch layouts. The existing 32px compact size is not the default for a touch-first screen.
- Distinguish initial loading, empty account, no filter matches, request failure, and stale data. Provide an appropriate next action for each.
- Existing transitions commonly use 150ms ease-out; buttons also scale to 0.98 when pressed, and navigation uses a short fly transition. New or revised motion must respect reduced-motion preferences. Do not use animation as the only state signal.

### Charts

Use the existing TimeSeriesChart and DonutChart organisms. Source colors from `chartColors` and categorical sequences from `donutRamps`, rather than creating a page-specific palette.

Income/positive series use emerald, outgoing/negative series red, net slate, and information/range selection sky. Incoming donut categories use emerald/sky; outgoing categories use red/amber. Use labels and legends to disambiguate categories when colors repeat.

Preserve the existing dashed grid `[4, 4]`, 1.75px zero line, 60% donut cutout, 3px hover offset, and subtle area opacity range (0.2 to 0.02). Tooltips use a dark slate surface. Provide a textual summary or accessible data equivalent; hover-only tooltips must not be the sole way to understand a result.

### Catalog and change process

Run `npm run storybook` from `web/` for the existing component catalog on port 6006. Storybook and the portal both load `app.css`; do not fork the theme in a story. The companion [.impeccable/design.json](.impeccable/design.json) contains representative previews for design tooling, not production components.

When adding or changing a shared pattern:

1. Inspect the existing component, variants, and stories first.
2. Reuse or extend the lowest appropriate Atomic Design tier.
3. Add stories for meaningful variants, long content, and relevant interaction/failure states.
4. Check keyboard behavior, responsive layout, and financial formatting; use Storybook's accessibility tooling without treating automated checks as complete coverage.
5. Run the repository's relevant validation commands and update this guide when a shared design decision changes.

## Do's and Don'ts

### Do

- Reuse the canonical theme, component variants, chart theme, and overlay scale.
- Keep financial values aligned, labeled, and consistent in units and precision.
- Use serif headings to organize content and sans-serif text for everyday interaction.
- Preserve state and give a clear recovery action when requests fail.
- Review new patterns in Storybook and in their real page context.

### Don't

- Create a second palette or stylesheet that drifts from the portal.
- Use brand amber, financial direction, and validation color interchangeably.
- Communicate selection, status, or financial meaning through color alone.
- Hide labels or essential actions to make a narrow layout fit.
- Treat documentation or static previews as proof of runtime accessibility compliance.
