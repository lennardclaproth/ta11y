import { apiGet, apiSend } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import type {
	ApplyIgnoreRuleResponse,
	CashflowTransaction,
	IgnoreRule,
	IgnoreRulePreview,
	IgnoreRuleRequest,
	IgnoreRulesResponse,
	IgnoredRuleGroup,
	ImportIgnoredResponse
} from '$lib/api/types';
import { ignoreRules } from '$lib/data/fixtures/ignoreRules';
import { mockCashflowRows, mutateMockIgnored, mutateMockImportId } from './cashflow';
import { clone, contains, delay, mockId } from './_mock';

/**
 * Rules written in fixture mode live here for the session. Nothing is persisted, but a
 * rule made on the rules page has to show up in its own index, or the page contradicts
 * itself the moment you use it.
 */
let mockRules: IgnoreRule[] = clone(ignoreRules);

/** Which transactions a rule matches, using the same fields the ledger filters on. */
function mockMatches(body: IgnoreRuleRequest): CashflowTransaction[] {
	return mockCashflowRows().filter((tx) => matchesMockRule(body, tx));
}

function matchesMockRule(body: IgnoreRuleRequest, tx: CashflowTransaction): boolean {
	if (body.direction && tx.direction !== body.direction) return false;
	// The bank is chosen from a list of names, so it narrows on the whole name — the API
	// matches it exactly, and a preview that matched a substring would promise more than
	// the rule delivers.
	if (body.source && tx.source.toLowerCase() !== body.source.toLowerCase()) return false;
	const field = body.match_field === 'note' ? tx.note : tx.description;
	return contains(field, body.contains);
}

/** The request body that reproduces a stored rule, for matching it against the ledger. */
function requestFor(rule: IgnoreRule): IgnoreRuleRequest {
	return {
		name: rule.name,
		match_field: rule.match_field,
		contains: rule.contains,
		direction: rule.direction,
		source: rule.source,
		enabled: rule.enabled
	};
}

function mockRule(body: IgnoreRuleRequest, base?: IgnoreRule): IgnoreRule {
	const now = new Date().toISOString();
	return {
		id: base?.id ?? mockId(),
		name: body.name,
		match_field: body.match_field,
		contains: body.contains,
		direction: body.direction ?? '',
		source: body.source ?? '',
		enabled: body.enabled ?? true,
		ignored_total: base?.ignored_total ?? 0,
		last_applied_at: base?.last_applied_at ?? null,
		created_at: base?.created_at ?? now,
		updated_at: now
	};
}

/**
 * How many of an upload's rows the session's enabled rules would recognise. Fixture mode
 * only: the import mock has no database to apply rules against, and an import that always
 * reported zero would hide the very thing this feature is about.
 */
export function countMockAutoIgnored(rows: { description: string; note: string }[]): number {
	return rows.filter((row) =>
		mockRules.some((rule) => {
			if (!rule.enabled) return false;
			const field = rule.match_field === 'note' ? row.note : row.description;
			return contains(field, rule.contains);
		})
	).length;
}

/** `GET /cashflow/ignore-rules` */
export async function listIgnoreRules(): Promise<IgnoreRule[]> {
	if (useMocks) {
		await delay();
		return clone(mockRules);
	}
	const response = await apiGet<IgnoreRulesResponse>('/cashflow/ignore-rules');
	return response.data;
}

/** `POST /cashflow/ignore-rules` */
export async function createIgnoreRule(body: IgnoreRuleRequest): Promise<IgnoreRule> {
	if (useMocks) {
		await delay();
		const created = mockRule(body);
		mockRules = [created, ...mockRules];
		return clone(created);
	}
	return apiSend<IgnoreRule>('POST', '/cashflow/ignore-rules', body);
}

/** `PUT /cashflow/ignore-rules/{rule_id}` */
export async function updateIgnoreRule(id: string, body: IgnoreRuleRequest): Promise<IgnoreRule> {
	if (useMocks) {
		await delay();
		const updated = mockRule(body, mockRules.find((rule) => rule.id === id));
		mockRules = mockRules.map((rule) => (rule.id === id ? updated : rule));
		return clone(updated);
	}
	return apiSend<IgnoreRule>('PUT', `/cashflow/ignore-rules/${id}`, body);
}

/** `DELETE /cashflow/ignore-rules/{rule_id}` */
export async function deleteIgnoreRule(id: string): Promise<void> {
	if (useMocks) {
		await delay();
		mockRules = mockRules.filter((rule) => rule.id !== id);
		return;
	}
	await apiSend<unknown>('DELETE', `/cashflow/ignore-rules/${id}`);
}

/** `POST /cashflow/ignore-rules/preview` */
export async function previewIgnoreRule(body: IgnoreRuleRequest): Promise<IgnoreRulePreview> {
	if (useMocks) {
		await delay();
		const matches = mockMatches(body);
		const sample = matches
			.slice()
			.sort((a, b) => b.date.localeCompare(a.date))
			.slice(0, 4);
		return clone({
			matching: matches.length,
			not_yet_ignored: matches.filter((tx) => !tx.ignored && !tx.ignore_overridden).length,
			scanned: mockCashflowRows().length,
			sample
		});
	}
	return apiSend<IgnoreRulePreview>('POST', '/cashflow/ignore-rules/preview', body);
}

/** `POST /cashflow/ignore-rules/{rule_id}/apply` */
export async function applyIgnoreRule(id: string): Promise<ApplyIgnoreRuleResponse> {
	if (useMocks) {
		await delay();
		const rule = mockRules.find((entry) => entry.id === id);
		if (!rule) return { ignored_count: 0, status: 'ok' };
		// The same rows the server would touch: matching, not ignored yet, and not
		// decided by hand.
		const ignored = mutateMockIgnored(
			(tx) => !tx.ignored && !tx.ignore_overridden && matchesMockRule(requestFor(rule), tx),
			true,
			rule.id
		);
		rule.ignored_total += ignored;
		rule.last_applied_at = new Date().toISOString();
		return { ignored_count: ignored, status: 'ok' };
	}
	return apiSend<ApplyIgnoreRuleResponse>('POST', `/cashflow/ignore-rules/${id}/apply`, undefined);
}

/** `GET /cashflow/imports/{import_id}/ignored` */
export async function getImportIgnored(importId: string): Promise<IgnoredRuleGroup[]> {
	if (useMocks) {
		await delay();
		// Fixture rows carry no import of their own, so the first read of a review stands
		// in for the import: each enabled rule claims what it matches, stamped with this
		// import. From then on the page works against real session state, so restoring a
		// row and showing the rest of a group behave the way they do against the API.
		for (const rule of mockRules) {
			if (!rule.enabled) continue;
			mutateMockIgnored(
				(tx) =>
					!tx.ignored &&
					!tx.ignore_overridden &&
					!tx.ignored_by_rule_id &&
					matchesMockRule(requestFor(rule), tx),
				true,
				rule.id
			);
		}
		mutateMockImportId(importId);

		const rows = mockCashflowRows().filter((tx) => tx.import_id === importId);
		return clone(
			mockRules
				.map((rule) => {
					const caught = rows.filter((tx) => tx.ignored_by_rule_id === rule.id);
					return { rule, total: caught.length, transactions: caught.slice(0, 5) };
				})
				.filter((group) => group.total > 0)
		);
	}
	const response = await apiGet<ImportIgnoredResponse>(`/cashflow/imports/${importId}/ignored`);
	return response.data;
}
