/**
 * The message shown when the bank's parser refused the file. The API classifies the
 * refusal (`reason: "file_not_recognised"`) without naming the export it expected, so the
 * wording names the account the user picked.
 */
export function wrongBankFileMessage(bank: string): string {
	return `This is not the transaction export from ${bank}. Download your transactions from ${bank} as CSV and upload that file.`;
}
