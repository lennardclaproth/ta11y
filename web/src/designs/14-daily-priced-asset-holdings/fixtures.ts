/**
 * Fake prototype data for design #14 (daily-priced asset holdings).
 *
 * Nothing here is real: round invented amounts, invented quantities and invented prices, only
 * public instrument names. It exists so the Storybook prototypes have something to lay out and is
 * not a contract — the field names are a proposal for the build, see the design comment on the
 * issue.
 */

/** One recorded purchase of a daily-priced holding. Money as decimal strings, like the assets API. */
export type DesignPurchase = {
	id: string;
	/** "YYYY-MM-DD". */
	date: string;
	/** Units bought, as a plain decimal string. */
	quantity: string;
	/** Price paid per unit, in euro. */
	unit_price: string;
	/** quantity x unit_price. */
	paid: string;
};

/** A point on a holding's value timeline: what it was worth against what had been paid by then. */
export type DesignValuePoint = {
	/** "YYYY-MM-DD". */
	date: string;
	value: number;
	paid: number;
};

/** An asset item whose value follows a daily price. */
export type DesignHolding = {
	id: string;
	class_id: string;
	class_name: string;
	/** Instrument name, e.g. "Bitcoin". */
	name: string;
	/** Quote symbol, e.g. "BTC/EUR". */
	symbol: string;
	/** Total units held, summed over the purchases. */
	quantity: string;
	/** Total paid over all purchases. */
	paid: number;
	/** Average price paid per unit. */
	avg_unit_price: number;
	/** Latest known daily price per unit. */
	price: number;
	/** Day the latest known price belongs to ("YYYY-MM-DD"). */
	price_date: string;
	/** True when `price_date` is older than today, so the price is carried forward. */
	price_carried_forward: boolean;
	/** quantity x price. */
	value: number;
	/** value - paid. */
	unrealized: number;
	/** (value - paid) / paid, as a percentage. */
	unrealized_pct: number;
	purchases: DesignPurchase[];
	series: DesignValuePoint[];
};

/** An asset item whose worth is still set by hand, shown next to the holdings for contrast. */
export type DesignManualItem = {
	id: string;
	class_id: string;
	class_name: string;
	name: string;
	worth: number;
	/** "YYYY-MM-DD" the worth was last set. */
	last_set: string;
};

/** A searchable daily-priced instrument in the "pick what you own" step. */
export type DesignQuote = {
	id: string;
	symbol: string;
	name: string;
	kind: string;
	currency: string;
	price: number;
	price_date: string;
	/** False when the quote cannot be used; `reason` says why. */
	selectable: boolean;
	reason?: string;
};

export const bitcoin: DesignHolding = {
	id: 'hld-btc',
	class_id: 'cls-crypto',
	class_name: 'Crypto & metals',
	name: 'Bitcoin',
	symbol: 'BTC/EUR',
	quantity: '0.2',
	paid: 9600,
	avg_unit_price: 48000,
	price: 64000,
	price_date: '2026-06-17',
	price_carried_forward: false,
	value: 12800,
	unrealized: 3200,
	unrealized_pct: 33.33,
	purchases: [
		{ id: 'buy-btc-3', date: '2026-05-20', quantity: '0.05', unit_price: '60000', paid: '3000' },
		{ id: 'buy-btc-2', date: '2026-03-02', quantity: '0.05', unit_price: '52000', paid: '2600' },
		{ id: 'buy-btc-1', date: '2026-01-15', quantity: '0.1', unit_price: '40000', paid: '4000' }
	],
	series: [
		{ date: '2026-01-15', value: 4000, paid: 4000 },
		{ date: '2026-02-01', value: 4500, paid: 4000 },
		{ date: '2026-03-02', value: 7800, paid: 6600 },
		{ date: '2026-04-01', value: 7200, paid: 6600 },
		{ date: '2026-05-20', value: 12000, paid: 9600 },
		{ date: '2026-06-01', value: 11600, paid: 9600 },
		{ date: '2026-06-17', value: 12800, paid: 9600 }
	]
};

export const ethereum: DesignHolding = {
	id: 'hld-eth',
	class_id: 'cls-crypto',
	class_name: 'Crypto & metals',
	name: 'Ethereum',
	symbol: 'ETH/EUR',
	quantity: '3',
	paid: 7000,
	avg_unit_price: 2333.33,
	price: 2200,
	price_date: '2026-06-15',
	price_carried_forward: true,
	value: 6600,
	unrealized: -400,
	unrealized_pct: -5.71,
	purchases: [
		{ id: 'buy-eth-2', date: '2026-04-18', quantity: '1', unit_price: '2000', paid: '2000' },
		{ id: 'buy-eth-1', date: '2026-02-10', quantity: '2', unit_price: '2500', paid: '5000' }
	],
	series: [
		{ date: '2026-02-10', value: 5000, paid: 5000 },
		{ date: '2026-03-01', value: 5400, paid: 5000 },
		{ date: '2026-04-18', value: 6000, paid: 7000 },
		{ date: '2026-05-01', value: 6900, paid: 7000 },
		{ date: '2026-06-01', value: 6300, paid: 7000 },
		{ date: '2026-06-15', value: 6600, paid: 7000 }
	]
};

export const holdings: DesignHolding[] = [bitcoin, ethereum];

/** Stays manual: the source does not deliver this one in euro (the pitch's cut). */
export const goldCoins: DesignManualItem = {
	id: 'ast-gold',
	class_id: 'cls-crypto',
	class_name: 'Crypto & metals',
	name: 'Gold coins',
	worth: 3000,
	last_set: '2026-02-28'
};

export const manualItems: DesignManualItem[] = [goldCoins];

/** Totals over both holdings, used by the account-wide variant. */
export const holdingsTotals = {
	paid: 16600,
	value: 19400,
	unrealized: 2800,
	unrealized_pct: 16.87
};

/** Combined value-against-paid timeline over both holdings. */
export const combinedSeries: DesignValuePoint[] = [
	{ date: '2026-01-15', value: 4000, paid: 4000 },
	{ date: '2026-02-10', value: 9500, paid: 9000 },
	{ date: '2026-03-02', value: 12800, paid: 11600 },
	{ date: '2026-04-18', value: 13000, paid: 13600 },
	{ date: '2026-05-20', value: 18000, paid: 16600 },
	{ date: '2026-06-17', value: 19400, paid: 16600 }
];

/** Results for the query "bit" / "go" in the quote picker. */
export const quoteResults: DesignQuote[] = [
	{
		id: 'qte-btc',
		symbol: 'BTC/EUR',
		name: 'Bitcoin',
		kind: 'Crypto',
		currency: 'EUR',
		price: 64000,
		price_date: '2026-06-17',
		selectable: true
	},
	{
		id: 'qte-eth',
		symbol: 'ETH/EUR',
		name: 'Ethereum',
		kind: 'Crypto',
		currency: 'EUR',
		price: 2200,
		price_date: '2026-06-15',
		selectable: true
	},
	{
		id: 'qte-sol',
		symbol: 'SOL/EUR',
		name: 'Solana',
		kind: 'Crypto',
		currency: 'EUR',
		price: 120,
		price_date: '2026-06-17',
		selectable: true
	},
	{
		id: 'qte-xau',
		symbol: 'XAU',
		name: 'Gold (troy ounce)',
		kind: 'Metal',
		currency: 'USD',
		price: 2400,
		price_date: '2026-06-17',
		selectable: false,
		reason: 'Priced in US dollars. Track gold as a manual item for now.'
	}
];

/** Asset classes for the prototype's class table, shaped like the real page's rows. */
export const designClasses = [
	{ id: 'cls-cash', name: 'Cash & savings', source: 'manual', worth: 18450, growth_pct: 2.1 },
	{
		id: 'cls-crypto',
		name: 'Crypto & metals',
		source: 'mixed',
		worth: 22400,
		growth_pct: 14.2
	},
	{ id: 'cls-portfolio', name: 'Portfolio', source: 'portfolio', worth: 12230.4, growth_pct: 12.31 }
];

/** One row of the prototype's asset-class table. */
export type DesignClassRow = (typeof designClasses)[number];

/** Short month-day label for a chart tick. */
export const dayTick = (iso: string): string =>
	new Date(`${iso}T00:00:00Z`).toLocaleDateString('en', {
		month: 'short',
		day: 'numeric',
		timeZone: 'UTC'
	});

/** "17 Jun 2026" — the same shape the calendar utils produce for a transaction date. */
export const longDate = (iso: string): string =>
	new Date(`${iso}T00:00:00Z`).toLocaleDateString('en-GB', {
		day: 'numeric',
		month: 'short',
		year: 'numeric',
		timeZone: 'UTC'
	});

/** A non-monetary number (never a price: those go through the Money atom). */
export const unitAmount = (value: number): string =>
	new Intl.NumberFormat('en-US', { maximumFractionDigits: 2 }).format(value);

/**
 * A held quantity. Two decimals minimum so a quantity column lines up (0.20 BTC next to 0.05 BTC),
 * eight maximum so a small crypto amount is never rounded away to nothing.
 */
export const qty = (value: number): string =>
	new Intl.NumberFormat('en-US', {
		minimumFractionDigits: 2,
		maximumFractionDigits: 8
	}).format(value);
