// Obviously fake data for the #11 heading-weight prototype. Round numbers, invented names,
// no real balances or instruments. Only here to give the headings a realistic neighbourhood.

export const netCashflowSeries = [40, 55, 48, 62, 58, 70, 66, 80, 74, 90];
export const netCashflowTotal = 1240;

export const spendingSeries = [90, 82, 86, 70, 74, 66, 60, 58, 52, 48];
export const spendingTotal = 860;

export const ledgerRows = [
	{ description: 'Monthly rent', date: '12 Mar', category: 'Housing', amount: -1200 },
	{ description: 'Salary', date: '01 Mar', category: 'Income', amount: 3000 },
	{ description: 'Groceries', date: '28 Feb', category: 'Food', amount: -160 }
];

export const drawerAssets = [
	{ name: 'Alpha Index Fund', worth: 4000 },
	{ name: 'Beta World Equity', worth: 2500 }
];

export const drawerMutations = [
	{ changeType: 'Revaluation', date: '10 Mar', worth: 4000 },
	{ changeType: 'Purchase', date: '02 Mar', worth: 3800 }
];

export const drawerListings = [
	{ symbol: 'AAA', name: 'Alpha Index Fund', exchange: 'Example Exchange' },
	{ symbol: 'BBB', name: 'Beta World Equity', exchange: 'Example Exchange' },
	{ symbol: 'CCC', name: 'Gamma Bond Fund', exchange: 'Sample Exchange' }
];
