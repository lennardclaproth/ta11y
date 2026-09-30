# Design #11 — Minder zware koppen (ronde 1: varianten)

Prototypes voor [issue #11](https://github.com/lennardclaproth/ta11y/issues/11). Alles in deze map
is **ontwerpmateriaal**, geen implementatie: buiten `web/src/designs/11-lighter-heading-font-weight/`
is niets aangeraakt.

Storybook: `npm run storybook` in `web/`, open **Designs/#11 Lighter heading weight**.

## Wat er te zien is

Elke variant toont dezelfde pagina met álle EB Garamond-koppen uit de app naast elkaar, plus een
band bovenaan die de huidige kop (semibold, door de browser namaakt) naast de voorgestelde dikte
zet.

| # | Context | Bron | Kop |
| --- | --- | --- | --- |
| 1 | Paginakop | `organisms/top-navbar/TopNavbar.svelte:130` | `Heading level="h1" size="2xl"` (semibold) |
| 2 | Kaarttitels | `molecules/analytics-card/AnalyticsCard.svelte:20` | `Heading size="md" weight="medium"` |
| 3 | Ledgerpaneel | `organisms/ledger-toolbar/LedgerToolbar.svelte:29` | **kale `<h2 class="text-2xl">`**, niet via het Heading-atom |
| 4 | Modaltitel | `molecules/dialog/Dialog.svelte:102` | `Heading size="lg"` (semibold) |
| 5 | Drawertitel | `organisms/drawer/Drawer.svelte:85` | `Heading size="lg"` (semibold) |
| 6 | Inlogkop | `routes/login/+page.svelte:35` | `Heading size="xl"` (semibold) |

Modal en drawer zijn als **headermarkup** nagebouwd (dezelfde atoms, dezelfde spacing) in plaats van
als echte overlays, zodat alle koppen op één screenshot vergelijkbaar zijn.

## De varianten

| Variant | Dikte | Fontbestand |
| --- | --- | --- |
| A — Regular | 400 overal | geen nieuw bestand; `eb-garamond-v32-latin-regular.woff2` dekt dit al |
| B — Medium | 500 overal | één extra EB Garamond 500-bestand, met een eigen `font-weight: 500`-descriptor |

`shared/heading-weight.css` zet per variant één dikte op elke kop in de subtree — de
prototype-vervanger van het aanpassen van de Heading-atom-default.

> **Let op:** variant B is in het prototype **benaderd** met een haarlijn (`-webkit-text-stroke`),
> omdat er vandaag alleen een regular-bestand in de repo staat. Zonder die benadering zou de browser
> stilletjes terugvallen op 400 en zou B identiek zijn aan A. De build levert een écht 500-bestand
> mee, zonder stroke en zonder synthetisch vet.

## Voor de build-agent

1. **`app.css`** — declareer elk gewicht alleen samen met zijn eigen bestand. Bij variant B komt er
   een tweede `@font-face` voor `EB Garamond` met `font-weight: 500;` en het nieuwe bestand. Verbreed
   nooit de bestaande `font-weight: 400`-descriptor; de comment in `app.css` legt uit waarom dat
   eerder alle koppen platsloeg.
2. **`typography.types.ts` / `typography.variants.ts`** — bij variant A moet `normal` aan
   `headingWeights` + `headingWeightClasses` (`font-normal`) worden toegevoegd. Bij variant B kan
   `medium` blijven bestaan.
3. **`Heading.svelte`** — zet de default `weight` op de gekozen dikte, zodat elke aanroep zonder
   expliciet `weight` meteen goed staat.
4. **Kale koppen** — `LedgerToolbar.svelte`, `routes/+layout.svelte`, `routes/admin/+layout.svelte`,
   `routes/admin/credentials/+page.svelte` en `routes/admin/listings/+page.svelte` gebruiken een
   kale `<h1>`/`<h2>` en erven dus het browser-default gewicht (bold, 700 → synthetisch vet). Los dit
   op in de `@layer base`-kopregel van `app.css` (daar staat al `font-family` + `text-slate-900`),
   zodat ook koppen buiten het atom meelopen.
5. **`AssetClassDrawer.svelte:60,80`** — twee `<h3 class="text-sm font-semibold">` labels. Ook
   Garamond, dus ook synthetisch vet. Zie "Open vragen" in de issue-comment.
6. **`AnalyticsCard.svelte:20`** — het expliciete `weight="medium"` kan weg zodra de default klopt.
7. **Niet aanraken:** het `ta11y`-woordmerk in `TopNavbar.svelte:108` (een `<a>`, geen kop),
   bodytekst, knoppen, kopmaten en `leading`/`tracking`.
8. **`docs/DESIGN.md`** — de typografie-sectie (regel ~175) beschrijft de huidige situatie met één
   regular; die tekst moet mee veranderen.

## Mappen

```
shared/
  HeadingContexts.svelte      alle zes koppen in context, opgebouwd uit bestaande componenten
  WeightBand.svelte           vergelijkingsband: vandaag vs. deze variant
  heading-weight.css          één kopgewicht per subtree (alleen prototype)
  heading-weight.fixtures.ts  nepdata: ronde bedragen, verzonnen namen
variant-a/                    Regular 400
variant-b/                    Medium 500
```
