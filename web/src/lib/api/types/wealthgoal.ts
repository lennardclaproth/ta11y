/**
 * Monthly wealth-growth goal [034]. The goal is one share of income per account, applying
 * from the calendar month it was set in; the standing scores each month against the goal
 * that applied in it. Amounts are 1e6-scaled, like everywhere else (see api/money.ts).
 */
import type { CashflowTransactionFilters } from './cashflow';

/** What a transaction counts as towards the goal. `''` is "not assigned". */
export type TransactionPurpose = '' | 'income' | 'wealth';

/** How a calendar month ended up. `in_progress` is the running month, never scored. */
export type MonthResult = 'met' | 'missed' | 'in_progress' | 'not_scored';

/** The monthly goal in force. Mirrors `wealthgoal.Goal`. */
export interface WealthGoal {
	share_percent: number;
	/** "YYYY-MM-DD", the first of the month the goal started applying in. */
	effective_from: string;
}

/** `GET /wealth-goal` / `PUT /wealth-goal` — mirrors `wealthgoal.GoalResponse`. */
export interface WealthGoalResponse {
	/** Null when the account has never set a goal. */
	goal: WealthGoal | null;
}

/** `PUT /wealth-goal` request — mirrors `wealthgoal.SetGoalRequest`. */
export interface SetWealthGoalRequest {
	share_percent: number;
}

/** One scored calendar month. Mirrors `wealthgoal.MonthStandingResponse`. */
export interface MonthStanding {
	/** "YYYY-MM-DD" (first of month). */
	month: string;
	income_cents: number;
	contributed_cents: number;
	/** The goal in force in this month, not today's goal. */
	goal_percent: number;
	result: MonthResult;
	/** Transactions in the month that carry no purpose yet. */
	unassigned_count: number;
}

/** `GET /wealth-goal/standing` — mirrors `wealthgoal.StandingResponse`. */
export interface WealthGoalStandingResponse {
	goal: WealthGoal | null;
	/** Newest month first. */
	months: MonthStanding[];
	current_streak: number;
	best_streak: number;
}

/** `POST /cashflow/transactions/purpose/selection` request. */
export interface MarkPurposeBySelectionRequest {
	purpose: TransactionPurpose | 'none';
	ids: string[];
}

/** `POST /cashflow/transactions/purpose/filter` request. */
export interface MarkPurposeByFilterRequest {
	purpose: TransactionPurpose | 'none';
	filters: CashflowTransactionFilters;
}

/**
 * Result of a purpose mutation. `matched_count` is what you pointed at, `updated_count`
 * what the purpose could apply to — income only sticks to incoming money.
 */
export interface MarkPurposeResponse {
	updated_count: number;
	matched_count: number;
	status: string;
}
