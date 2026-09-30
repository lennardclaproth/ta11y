# Design #11 — Minder zware koppen (definitief ontwerp)

Prototype voor [issue #11](https://github.com/lennardclaproth/ta11y/issues/11). Alles in deze map is
**ontwerpmateriaal**, geen implementatie: buiten `web/src/designs/11-lighter-heading-font-weight/`
is niets aangeraakt.

Storybook: `npm run storybook` in `web/`, open **Designs/#11 Lighter heading weight → Final**.

## De beslissing

Alle EB Garamond-koppen gaan naar **regular 400** — variant A uit ronde 1. Het is het enige gewicht
waarvoor vandaag al een echt fontbestand in de repo staat, dus er blijft nergens door de browser
namaakt vet over en er komt geen nieuw fontbestand bij.

Lennards antwoorden op de twee open vragen zijn verwerkt:

1. **Centraal oplossen**: ook koppen die het `Heading`-atom niet gebruiken lopen mee.
2. **De twee kleine labels** in de asset-class-drawer krijgen dezelfde dikte.

## Wat er te zien is

Bovenaan een vergelijkingsband met drie paren — voor elk van de drie manieren waarop een kop
vandaag te zwaar is, links het huidige beeld en rechts het voorstel. Daaronder alle
Garamond-koppen van de app op één pagina, naast de bodytekst en de knoppen waaraan de dikte
volgens de pitch beoordeeld wordt.

| # | Context | Bron | Kop vandaag |
| --- | --- | --- | --- |
| 1 | Paginakop | `organisms/top-navbar/TopNavbar.svelte:130` | `Heading level="h1" size="2xl"` — semibold (600) |
| 2 | Kaarttitels | `molecules/analytics-card/AnalyticsCard.svelte:20` | `Heading size="md" weight="medium"` — rendert al als 400 |
| 3 | Ledgerkop | `organisms/ledger-toolbar/LedgerToolbar.svelte:29` | kale `<h2 class="text-2xl">` — browser-default bold (700) |
| 4 | Modaltitel | `molecules/dialog/Dialog.svelte:102` | `Heading size="lg"` — semibold |
| 5 | Drawertitel | `organisms/drawer/Drawer.svelte:85` | `Heading size="lg"` — semibold |
| 6 | Drawerlabels | `organisms/asset-class-drawer/AssetClassDrawer.svelte:60,80` | `<h3 class="text-sm font-semibold">` — semibold op 14px |
| 7 | Admin-kop | `routes/admin/listings/+page.svelte:124` (en 3 andere) | kale `<h2 class="text-2xl">` — bold |
| 8 | Servererror-kop | `routes/+layout.svelte:36` | kale `<h1 class="text-2xl">` — bold |
| 9 | Inlogkop | `routes/login/+page.svelte:35` | `Heading size="xl"` — semibold |

Modal en drawer zijn als **headermarkup** nagebouwd (dezelfde atoms, dezelfde spacing) in plaats van
als echte overlays, zodat alle koppen op één screenshot vergelijkbaar zijn.

`shared/heading-weight.css` zet 400 op elke kop in de subtree — de prototype-vervanger van de
echte wijziging in `app.css` en het `Heading`-atom.

## Voor de build-agent

1. **`web/src/app.css`, `@layer base`** — de kopregel (`h1…h6`) zet nu alleen `font-family` en
   `text-slate-900`. Voeg daar `font-weight: 400;` aan toe. Dit is de centrale oplossing: ook kale
   koppen buiten het atom lopen dan mee. Verbreed **nooit** de `font-weight: 400`-descriptor van de
   bestaande `@font-face`; de comment in `app.css` legt uit waarom dat eerder alle koppen platsloeg.
   Er komt geen tweede `@font-face` bij.
2. **`typography.types.ts`** — voeg `'normal'` toe aan `headingWeights` (staat nu op
   `['medium', 'semibold', 'bold']`).
3. **`typography.variants.ts`** — voeg `normal: 'font-normal'` toe aan `headingWeightClasses`.
4. **`Heading.svelte`** — zet de default `weight` van `'semibold'` naar `'normal'`. Elke aanroep
   zonder expliciet `weight` staat dan meteen goed; de bestaande `medium`/`semibold`/`bold`-waarden
   blijven bestaan maar worden nergens meer gebruikt voor Garamond-koppen.
5. **`AnalyticsCard.svelte:20`** — haal `weight="medium"` weg; de default klopt dan.
6. **`AssetClassDrawer.svelte:60,80`** — haal `font-semibold` weg uit de twee
   `<h3 class="mb-2 text-sm font-semibold text-slate-900">` labels.
7. **Controleer** dat er nergens nog een expliciete `weight="semibold"`/`"bold"` of een
   `font-semibold`/`font-bold`-utility op een Garamond-kop staat:
   `rg 'font-(semi)?bold' web/src --glob '*.svelte'` en `rg 'weight="(semibold|bold|medium)"' web/src`.
   Sans-serif tekst (`Text`, `Money`, knoppen, navigatie) blijft ongemoeid.
8. **Niet aanraken:** het `ta11y`-woordmerk in `TopNavbar.svelte:108` (een `<a class="font-heading">`,
   geen kop, dus de `@layer base`-regel raakt het niet), bodytekst, knoppen, kopmaten en
   `leading`/`tracking`.
9. **`docs/DESIGN.md`** — de typografie-sectie (regel ~175) beschrijft dat zwaardere kopgewichten
   door de browser gesynthetiseerd worden. Werk die alinea bij: koppen staan voortaan op 400 en er
   wordt geen gewicht meer gerenderd waarvoor geen bestand bestaat. De regel "Heading weights are
   medium, semibold, and bold" moet `normal` erbij krijgen.
10. **`CHANGELOG.md`** — `Unreleased → Changed`, met een nieuw feature-ID.

## Mappen

```
shared/
  HeadingContexts.svelte      alle koppen in context, opgebouwd uit bestaande componenten
  WeightBand.svelte           vergelijkingsband: vandaag vs. dit ontwerp, drie paren
  heading-weight.css          één kopgewicht per subtree (alleen prototype)
  heading-weight.fixtures.ts  nepdata: ronde bedragen, verzonnen namen
final/
  FinalRegularHeadings.svelte + stories (Default, Mobile, Loading, Empty, Error, No matches)
```
