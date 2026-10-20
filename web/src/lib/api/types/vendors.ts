/** One vendor record. Mirrors `vendors.VendorResponse`. */
export interface Vendor {
	id: string;
	name: string;
	type: 'brokerage' | 'bank';
	active: boolean;
	import_disabled: boolean;
	created_at: string;
	updated_at: string;
}

/** `GET /vendors` returns a bare array of vendors. */
export type VendorsResponse = Vendor[];
