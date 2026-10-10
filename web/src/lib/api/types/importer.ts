/**
 * Import types. The import endpoints accept `multipart/form-data` uploads (a file plus a few
 * form fields) and return `202 Accepted` with the durable import record's id + initial status.
 */

/** `POST /imports/*` — mirrors `importer.ImportAcceptedResponse`. */
export interface ImportAcceptedResponse {
	import_id: string;
	status: string;
}

/** Form fields for `POST /imports/cashflow`. */
export interface CashflowImportInput {
	file: File;
	vendor_id: string;
}

/** Form fields for `POST /imports/portfolio`. */
export interface PortfolioImportInput {
	file: File;
	vendor_id: string;
}

/** Form fields for `POST /imports/eod`. */
export interface EODImportInput {
	file: File;
	listing_id: string;
}

/** Lifecycle of one import record; `pending` and `processing` are not final. */
export type ImportStatus = 'pending' | 'processing' | 'in_progress' | 'completed' | 'failed';

/** Which destination an import filled. */
export type ImportType = 'cashflow' | 'portfolio' | 'eod';

/**
 * A product an import brought in that no listing matches. It is stored and visible in
 * the ledger, but contributes no market value until its listing exists.
 */
export interface UnlinkedProduct {
	name: string;
	isin: string | null;
	symbol: string | null;
	transactions: number;
}

/** `GET /imports/{import_id}` — mirrors `importer.ImportResultResponse`. */
export interface ImportResult {
	import_id: string;
	type: ImportType;
	status: ImportStatus;
	/**
	 * `file_not_recognised` when the parser refused the file, otherwise empty. This is
	 * all a failed import says: the server's own error message stays on the server.
	 */
	reason: string;
	total_rows: number;
	imported: number;
	duplicates: number;
	failed: number;
	created_at: string;
	updated_at: string;
	/** Empty until a portfolio import has completed. */
	unlinked_products: UnlinkedProduct[];
}
