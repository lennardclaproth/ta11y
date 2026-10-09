/**
 * Obviously fake ledger rows for the #22 design prototypes. Round numbers, invented
 * descriptions — never copy real account data into a design folder.
 */
export interface MockRow {
	date: string;
	description: string;
	tag: string;
	amount: number;
}

export const mockRows: MockRow[] = [
	{ date: 'Mar 2, 2026', description: 'Monthly rent', tag: 'Housing', amount: -1200 },
	{ date: 'Mar 1, 2026', description: 'Salary', tag: 'Income', amount: 3000 },
	{ date: 'Feb 27, 2026', description: 'Groceries', tag: 'Food', amount: -80 },
	{ date: 'Feb 25, 2026', description: 'Transit pass', tag: 'Transport', amount: -60 },
	{ date: 'Feb 24, 2026', description: 'Book shop', tag: 'Leisure', amount: -40 }
];
