import { purposeLabel } from '$lib/api/wealthgoal';

/**
 * The Purpose header filter. `none` is what the wire calls the empty purpose, so the option
 * value is the query value and nothing has to be translated in the page.
 */
export const purposeFilterOptions = [
	{ value: 'income', label: purposeLabel.income },
	{ value: 'wealth', label: purposeLabel.wealth },
	{ value: 'none', label: purposeLabel[''] }
];
