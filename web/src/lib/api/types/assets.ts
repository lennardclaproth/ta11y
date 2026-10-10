/**
 * Asset-class summary. Mirrors `assets.ClassResponse`.
 * Money (`current_worth`) is a decimal string. `growth_pct` is a fraction/percent number or null.
 */
export interface AssetClass {
	id: string;
	name: string;
	source: string;
	archived: boolean;
	current_worth: string;
	/** RFC3339 timestamp or null. */
	last_change_at?: string | null;
	growth_pct?: number | null;
	updated_at: string;
}

/** `GET /assets/classes` returns a bare array of asset classes. */
export type AssetClassesResponse = AssetClass[];

/** One asset within a class. Mirrors `assets.AssetResponse`. */
export interface Asset {
	id: string;
	name: string;
	current_worth: string;
	archived: boolean;
	updated_at: string;
}

/** A point on a class's growth timeline. Mirrors `assets.ClassGrowthPointResponse`. */
export interface ClassGrowthPoint {
	/** "YYYY-MM-DD". */
	date: string;
	total_worth: string;
}

/** A worth mutation record. Mirrors `assets.MutationResponse`. */
export interface AssetMutation {
	id: string;
	item_id: string;
	change_type: string;
	direction?: string | null;
	amount: string;
	previous_worth: string;
	new_worth: string;
	class_total_worth: string;
	/** "YYYY-MM-DD". */
	effective_date: string;
	note?: string | null;
	created_at: string;
}

/** One recorded purchase of a daily-priced item. Mirrors `assets.PurchaseResponse`. */
export interface HoldingPurchase {
	id: string;
	/** "YYYY-MM-DD". */
	date: string;
	/** Units bought, as a plain decimal string. */
	quantity: string;
	unit_price: string;
	/** quantity × unit_price. */
	paid: string;
}

/**
 * One day of a holding's history: what it was worth against what had been paid for
 * it by then. Mirrors `assets.HoldingValuePointResponse`.
 */
export interface HoldingValuePoint {
	/** "YYYY-MM-DD". */
	date: string;
	value: string;
	paid: string;
}

/**
 * An asset item whose worth follows a listing's daily price rather than a worth the
 * user sets. Mirrors `assets.HoldingResponse`.
 */
export interface AssetHolding {
	id: string;
	class_id: string;
	class_name?: string;
	name: string;
	/** Quote symbol, e.g. "BTC/EUR". Empty when the listing behind it has gone missing. */
	symbol: string;
	instrument_name?: string;
	/** Total units held, as a plain decimal string. */
	quantity: string;
	paid: string;
	avg_unit_price: string;
	/** Most recent known daily price per unit. */
	price: string;
	/** Day the known price belongs to; absent while no price is known yet. */
	price_date?: string | null;
	/** True when `price_date` is older than today, so the price is carried forward. */
	price_carried_forward: boolean;
	value: string;
	unrealized: string;
	/** Absent when nothing has been paid yet, where a percentage has no meaning. */
	unrealized_pct?: number | null;
	series: HoldingValuePoint[];
	/** Only present on `GET /assets/{asset_id}/holding`. */
	purchases?: HoldingPurchase[];
}

/** `GET /assets/classes/{class_id}` — mirrors `assets.ClassDetailsResponse`. */
export interface AssetClassDetails {
	class: AssetClass;
	assets: Asset[];
	/** The class's daily-priced items; they carry a quantity and a price that `assets` has no equivalent of. */
	holdings: AssetHolding[];
	growth: ClassGrowthPoint[];
	mutations: AssetMutation[];
}

/** Account-level total worth snapshot point. Mirrors `assets.SnapshotResponse`. */
export interface AssetSnapshotPoint {
	/** "YYYY-MM-DD". */
	date: string;
	total_worth: string;
}

/** `GET /assets/snapshots` returns a bare array of snapshot points. */
export type AssetSnapshotsResponse = AssetSnapshotPoint[];

/** `POST /assets/classes` request — mirrors `assets.CreateAssetClassRequest`. */
export interface CreateAssetClassRequest {
	name: string;
}

/** `POST /assets/classes` — mirrors `assets.CreateAssetClassResponse`. */
export interface CreateAssetClassResponse {
	id: string;
}

/** `PATCH /assets/classes` request — mirrors `assets.UpdateClassRequest`. */
export interface UpdateAssetClassRequest {
	id: string;
	name?: string;
	archived?: boolean;
}

/** `DELETE /assets/classes/{class_id}` body — mirrors `assets.DeleteClassRequest`. */
export interface DeleteAssetClassRequest {
}

/** `POST /assets` request — mirrors `assets.CreateAssetRequest`. */
export interface CreateAssetRequest {
	class_id: string;
	name: string;
	/** Non-negative decimal string. */
	initial_worth: string;
	/** "YYYY-MM-DD". */
	effective_date: string;
	note?: string;
}

/** `POST /assets` — mirrors `assets.CreateAssetResponse`. */
export interface CreateAssetResponse {
	id: string;
	name: string;
}

/** `PUT /assets/{asset_id}/worth` request — mirrors `assets.SetAssetWorthRequest`. */
export interface SetAssetWorthRequest {
	/** Non-negative decimal string. */
	worth: string;
	/** "YYYY-MM-DD". */
	effective_date: string;
	note?: string;
}

/** One recorded acquisition as the API receives it — mirrors `assets.PurchaseRequest`. */
export interface PurchaseRequest {
	/** "YYYY-MM-DD", not in the future. */
	date: string;
	/** Positive decimal string. */
	quantity: string;
	/** Non-negative decimal string. */
	unit_price: string;
}

/** `POST /assets/holdings` request — mirrors `assets.CreateHoldingRequest`. */
export interface CreateHoldingRequest {
	class_id: string;
	name: string;
	/** Quote symbol from `GET /marketdata/quotes`, e.g. "BTC/EUR". */
	symbol: string;
	purchase: PurchaseRequest;
}

/** `POST /assets/holdings` — mirrors `assets.CreateHoldingResponse`. */
export interface CreateHoldingResponse {
	id: string;
	name: string;
}

/** Direction for an asset-worth adjustment. */
export type AssetWorthDirection = 'INCREASE' | 'DECREASE';

/** `PUT /assets/{asset_id}/adjust` request — mirrors `assets.AdjustAssetWorthRequest`. */
export interface AdjustAssetWorthRequest {
	direction: AssetWorthDirection;
	/** Non-negative decimal string. */
	amount: string;
	/** "YYYY-MM-DD". */
	effective_date: string;
	note?: string;
}

/** Query filters for `GET /assets/classes`. */
export interface AssetClassesQuery {
	include_archived?: boolean;
}

/** Query filters for `GET /assets/snapshots`. */
export interface AssetSnapshotsQuery {
	from?: string;
	to?: string;
}
