/**
 * The manual end-of-day CSV the importer accepts, described for the upload form.
 *
 * Mirrors `internal/importer/eod/parsers/brandnewday.go`: the parser requires these
 * headers by name and reads dates in `dd/mm/yyyy`. NAV becomes the row's open, high,
 * low and close — a fund publishes one price a day, not a candle — and volume is zero.
 */
export const eodCsvColumns = ['date', 'nav', 'ask', 'bid', 'dividend'] as const;

/** Date format the parser accepts. Anything else fails the row. */
export const eodCsvDateFormat = 'dd/mm/yyyy';

/** Accepted upload types, as the file picker's `accept` list. */
export const eodCsvAccept = '.csv,text/csv';
