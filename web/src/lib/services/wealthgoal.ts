import { apiGet, apiSend } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import type {
	CashflowTransaction,
	MarkPurposeByFilterRequest,
	MarkPurposeBySelectionRequest,
	MarkPurposeResponse,
	MonthResult,
	MonthStanding,
	SetWealthGoalRequest,
	TransactionPurpose,
	WealthGoal,
	WealthGoalResponse,
	WealthGoalStandingResponse
} from '$lib/api/types';
import { cashflowTransactions } from '$lib/data/fixtures/cashflow';
import { wealthGoal } from '$lib/data/fixtures/wealthgoal';
import { clone, delay } from './_mock';

/**
 * The goals the mock branch hands out, oldest first. A whole history rather than one value,
 * because that is what the backend keeps: adjusting the goal writes a row for the month it
 * is adjusted in, and a month that has been scored keeps the goal it was judged by.
 */
let mockGoals: WealthGoal[] = [clone(wealthGoal)];

/** The goal in force in a month: the latest one that had started by then, as `goalAt` does. */
function mockGoalAt(month: string): WealthGoal | null {
	let found: WealthGoal | null = null;
	for (const goal of mockGoals) {
		if (goal.effective_from > month) break;
		found = goal;
	}
	return found;
}

/** `GET /wealth-goal` */
export async function getWealthGoal(): Promise<WealthGoalResponse> {
	if (useMocks) {
		await delay();
		return { goal: clone(mockGoalAt(currentMockMonth())) };
	}
	return apiGet<WealthGoalResponse>('/wealth-goal');
}

/** `PUT /wealth-goal` */
export async function setWealthGoal(body: SetWealthGoalRequest): Promise<WealthGoalResponse> {
	if (useMocks) {
		await delay();
		// Mirrors the backend: a new goal applies from the first of the current month, and
		// adjusting it twice in one month replaces that month's row rather than adding one.
		const current = currentMockMonth();
		const goal: WealthGoal = { share_percent: body.share_percent, effective_from: current };
		mockGoals = [...mockGoals.filter((g) => g.effective_from !== current), goal];
		return { goal: clone(goal) };
	}
	return apiSend<WealthGoalResponse>('PUT', '/wealth-goal', body);
}

/** `GET /wealth-goal/standing` */
export async function getWealthGoalStanding(
	query: { months?: number } = {}
): Promise<WealthGoalStandingResponse> {
	if (useMocks) {
		await delay();
		return mockStanding(query.months ?? 12);
	}
	return apiGet<WealthGoalStandingResponse>('/wealth-goal/standing', { ...query });
}

/** `POST /cashflow/transactions/purpose/selection` */
export async function markCashflowPurposeBySelection(
	body: MarkPurposeBySelectionRequest
): Promise<MarkPurposeResponse> {
	if (useMocks) {
		await delay();
		const updated = markMockPurpose(
			(tx) => body.ids.includes(tx.id),
			normalizePurpose(body.purpose)
		);
		return {
			updated_count: updated,
			matched_count: body.ids.length,
			status: `marked ${updated} of ${body.ids.length} transactions`
		};
	}
	return apiSend<MarkPurposeResponse>('POST', '/cashflow/transactions/purpose/selection', body);
}

/** `POST /cashflow/transactions/purpose/filter` */
export async function markCashflowPurposeByFilter(
	body: MarkPurposeByFilterRequest
): Promise<MarkPurposeResponse> {
	if (useMocks) {
		await delay();
		const matches = (tx: CashflowTransaction) => matchesMockFilters(tx, body.filters);
		const matched = cashflowTransactions.filter(matches).length;
		const updated = markMockPurpose(matches, normalizePurpose(body.purpose));
		return {
			updated_count: updated,
			matched_count: matched,
			status: `marked ${updated} of ${matched} transactions`
		};
	}
	return apiSend<MarkPurposeResponse>('POST', '/cashflow/transactions/purpose/filter', body);
}

function normalizePurpose(purpose: MarkPurposeBySelectionRequest['purpose']): TransactionPurpose {
	return purpose === 'none' ? '' : purpose;
}

/**
 * Mutates the shared fixture rows, which is the point: in mock mode the ledger, the donuts
 * and the standing all read the same array, so marking a row moves the score the way the
 * real API does. The direction rule is enforced here too — income only sticks to incoming
 * money — so a mixed selection reports the same partial count as the backend.
 */
function markMockPurpose(
	matches: (tx: CashflowTransaction) => boolean,
	purpose: TransactionPurpose
): number {
	let updated = 0;
	for (const tx of cashflowTransactions) {
		if (!matches(tx)) continue;
		if (purpose === 'income' && tx.direction !== 'in') continue;
		if (purpose === 'wealth' && tx.direction !== 'out') continue;
		tx.purpose = purpose;
		updated += 1;
	}
	return updated;
}

function matchesMockFilters(
	tx: CashflowTransaction,
	filters: MarkPurposeByFilterRequest['filters']
): boolean {
	const tags = (filters.tags ?? '')
		.split(',')
		.map((tag) => tag.trim())
		.filter(Boolean);
	const purposes = (filters.purpose ?? '')
		.split(',')
		.map((value) => value.trim())
		.filter(Boolean)
		.map((value) => (value === 'none' ? '' : value));

	if (filters.direction && tx.direction !== filters.direction) return false;
	if (filters.hide_ignored && tx.ignored) return false;
	if (filters.untagged && tx.tag !== '') return false;
	if (tags.length > 0 && !tags.includes(tx.tag)) return false;
	if (purposes.length > 0 && !purposes.includes(tx.purpose)) return false;
	if (
		filters.description &&
		!tx.description.toLowerCase().includes(filters.description.toLowerCase())
	) {
		return false;
	}
	if (filters.from && tx.date.slice(0, 10) < filters.from) return false;
	if (filters.to && tx.date.slice(0, 10) > filters.to) return false;
	return true;
}

/**
 * Scores the fixture transactions month by month the way the backend does: marked rows count
 * whatever their ignored state, unassigned rows are counted only when they are visible in the
 * ledger, the running month is never scored, and the streak counts finished months only.
 */
function mockStanding(months: number): WealthGoalStandingResponse {
	const currentMonth = startOfMonth(new Date());
	const earliest = new Date(
		Date.UTC(currentMonth.getUTCFullYear(), currentMonth.getUTCMonth() - (months - 1), 1)
	);

	const buckets = new Map<string, MonthStanding>();
	for (const tx of cashflowTransactions) {
		const month = `${tx.date.slice(0, 7)}-01`;
		if (month < earliest.toISOString().slice(0, 10)) continue;
		const bucket = buckets.get(month) ?? {
			month,
			income_cents: 0,
			contributed_cents: 0,
			goal_percent: mockGoalAt(month)?.share_percent ?? 0,
			result: 'in_progress' as MonthResult,
			unassigned_count: 0
		};
		if (tx.purpose === 'income') bucket.income_cents += tx.amountCents;
		if (tx.purpose === 'wealth') bucket.contributed_cents += tx.amountCents;
		if (tx.purpose === '' && !tx.ignored) bucket.unassigned_count += 1;
		buckets.set(month, bucket);
	}

	const current = currentMonth.toISOString().slice(0, 10);
	const currentGoal = mockGoalAt(current);
	if (currentGoal && !buckets.has(current)) {
		buckets.set(current, {
			month: current,
			income_cents: 0,
			contributed_cents: 0,
			goal_percent: currentGoal.share_percent,
			result: 'in_progress',
			unassigned_count: 0
		});
	}

	const scored = [...buckets.values()]
		.map((month) => ({ ...month, result: mockResult(month, current) }))
		.sort((a, b) => b.month.localeCompare(a.month));

	return {
		goal: clone(currentGoal),
		months: scored,
		...mockStreaks(scored)
	};
}

function mockResult(month: MonthStanding, currentMonth: string): MonthResult {
	if (!mockGoalAt(month.month)) return 'not_scored';
	if (month.month >= currentMonth) return 'in_progress';
	if (month.income_cents <= 0) return 'missed';
	return month.contributed_cents * 100 >= month.income_cents * month.goal_percent
		? 'met'
		: 'missed';
}

function mockStreaks(months: MonthStanding[]): { current_streak: number; best_streak: number } {
	let counting = true;
	let run = 0;
	let current = 0;
	let best = 0;
	for (const month of months) {
		if (month.result === 'in_progress' || month.result === 'not_scored') continue;
		if (month.result === 'met') {
			run += 1;
			best = Math.max(best, run);
			if (counting) current = run;
			continue;
		}
		run = 0;
		counting = false;
	}
	return { current_streak: current, best_streak: best };
}

function startOfMonth(date: Date): Date {
	return new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), 1));
}

/** The running month as a "YYYY-MM-DD" first-of-month. */
function currentMockMonth(): string {
	return startOfMonth(new Date()).toISOString().slice(0, 10);
}
