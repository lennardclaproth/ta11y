import type { CashflowDirection, IgnoreRule, IgnoreRuleRequest } from './types';

/** The label a rule's "all banks" scope reads as. A rule with no source is not narrowed. */
export const allBanksLabel = 'All banks';

/** How a direction reads in a rule summary; a rule without one matches both ways. */
export function ruleDirectionLabel(direction: '' | CashflowDirection): string {
	if (direction === 'in') return 'Incoming';
	if (direction === 'out') return 'Outgoing';
	return 'In and out';
}

/** How a rule's matched field reads. */
export function ruleFieldLabel(field: IgnoreRule['match_field']): string {
	return field === 'note' ? 'Note' : 'Description';
}

/**
 * One ruled line saying what a rule matches on. It names every narrowing the rule
 * applies, so a rule that is wider than it looks cannot hide behind its own name.
 */
export function ruleSummary(rule: IgnoreRule | IgnoreRuleRequest): string {
	const field = ruleFieldLabel(rule.match_field);
	const direction = ruleDirectionLabel(rule.direction ?? '');
	const bank = rule.source ? rule.source.toUpperCase() : allBanksLabel;
	return `${field} contains “${rule.contains}” · ${direction} · ${bank}`;
}

/** The request body that recreates a rule, for editing and for previewing it unchanged. */
export function ruleToRequest(rule: IgnoreRule): IgnoreRuleRequest {
	return {
		name: rule.name,
		match_field: rule.match_field,
		contains: rule.contains,
		direction: rule.direction,
		source: rule.source,
		enabled: rule.enabled
	};
}
