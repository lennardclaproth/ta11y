import type { CashflowDirection, IgnoreRule, IgnoreRuleMatchField } from '$lib/api/types';

/** Which half a narrow screen shows; both are side by side from `lg`. */
export const deskPanes = ['list', 'detail'] as const;
export type DeskPane = (typeof deskPanes)[number];

/** The rule as it is being written. A draft with no id has not been saved yet. */
export interface IgnoreRuleDraft {
	id: string | null;
	name: string;
	match_field: IgnoreRuleMatchField;
	contains: string;
	direction: '' | CashflowDirection;
	source: string;
	enabled: boolean;
}

/** The field a rule reads, as the form offers it. */
export const matchFieldOptions = [
	{ value: 'description', label: 'Description' },
	{ value: 'note', label: 'Note' }
];

/** The directions a rule can narrow to. Empty is "both", not "none". */
export const ruleDirectionOptions = [
	{ value: '', label: 'Incoming and outgoing' },
	{ value: 'in', label: 'Incoming only' },
	{ value: 'out', label: 'Outgoing only' }
];

/** An empty draft, used by "New rule" and by the Cashflow page's prefill. */
export function emptyDraft(overrides: Partial<IgnoreRuleDraft> = {}): IgnoreRuleDraft {
	return {
		id: null,
		name: '',
		match_field: 'description',
		contains: '',
		direction: '',
		source: '',
		enabled: true,
		...overrides
	};
}

/** The draft that reproduces a saved rule. */
export function draftFromRule(rule: IgnoreRule): IgnoreRuleDraft {
	return {
		id: rule.id,
		name: rule.name,
		match_field: rule.match_field,
		contains: rule.contains,
		direction: rule.direction,
		source: rule.source,
		enabled: rule.enabled
	};
}
