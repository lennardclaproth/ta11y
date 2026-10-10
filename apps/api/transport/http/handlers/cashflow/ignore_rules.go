package cashflow

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// IgnoreRuleRequest is the body of a create, update, or preview of an ignore rule.
type IgnoreRuleRequest struct {
	Name       string `json:"name"`
	MatchField string `json:"match_field"`
	Contains   string `json:"contains"`
	// Direction is empty for a rule that matches both directions.
	Direction string `json:"direction,omitempty"`
	// Source is the bank the rule is limited to, empty for every bank.
	Source string `json:"source,omitempty"`
	// Enabled defaults to true when omitted: a rule is written to be used.
	Enabled *bool `json:"enabled,omitempty"`
}

func (r IgnoreRuleRequest) draft() (cashflow.IgnoreRuleDraft, error) {
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}
	return cashflow.NewIgnoreRuleDraft(r.Name, r.MatchField, r.Contains, r.Direction, r.Source, enabled)
}

// IgnoreRuleResponse is one ignore rule.
type IgnoreRuleResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	MatchField string    `json:"match_field"`
	Contains   string    `json:"contains"`
	Direction  string    `json:"direction"`
	Source     string    `json:"source"`
	Enabled    bool      `json:"enabled"`
	// IgnoredTotal is how many transactions this rule has ignored since it was made.
	IgnoredTotal  int        `json:"ignored_total"`
	LastAppliedAt *time.Time `json:"last_applied_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// IgnoreRulesResponse returns every ignore rule of the account.
type IgnoreRulesResponse struct {
	Data []IgnoreRuleResponse `json:"data"`
}

// IgnoreRulePreviewResponse reports what a rule matches in the ledger as it is today,
// so a rule that is too wide is visible before it is allowed to ignore anything.
type IgnoreRulePreviewResponse struct {
	Matching      int                         `json:"matching"`
	NotYetIgnored int                         `json:"not_yet_ignored"`
	Scanned       int                         `json:"scanned"`
	Sample        []CreateTransactionResponse `json:"sample"`
}

// ApplyIgnoreRuleResponse reports how many transactions an apply actually ignored.
type ApplyIgnoreRuleResponse struct {
	IgnoredCount int    `json:"ignored_count"`
	Status       string `json:"status"`
}

// IgnoredRuleGroupResponse is the harvest of one rule within one import.
type IgnoredRuleGroupResponse struct {
	Rule         IgnoreRuleResponse          `json:"rule"`
	Total        int                         `json:"total"`
	Transactions []CreateTransactionResponse `json:"transactions"`
}

// ImportIgnoredResponse groups what an import's rules ignored, by rule.
type ImportIgnoredResponse struct {
	Data []IgnoredRuleGroupResponse `json:"data"`
}

// GetIgnoreRules lists the account's ignore rules.
//
// @Summary     List cashflow ignore rules
// @Description Returns every ignore rule of the signed-in account, newest first.
// @Produce     application/json
// @Success     200 {object} IgnoreRulesResponse
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/ignore-rules [get]
// @Tags        Ignore rules
func GetIgnoreRules(log logging.Logger, queries *cashflow.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		rules, err := queries.IgnoreRules(r.Context(), accountID)
		if err != nil {
			log.Error(r.Context(), "cashflow ignore rules: failed to list rules", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to list ignore rules"})
			return
		}

		items := make([]IgnoreRuleResponse, 0, len(rules))
		for _, rule := range rules {
			items = append(items, toIgnoreRuleResponse(rule))
		}
		_ = httpx.JSONEncode(w, http.StatusOK, IgnoreRulesResponse{Data: items})
	})
}

// CreateIgnoreRule stores a new ignore rule.
//
// @Summary     Create a cashflow ignore rule
// @Description Stores a rule that ignores matching transactions on arrival. It takes effect on what is imported from then on; applying it to transactions already in the ledger is a separate action.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body IgnoreRuleRequest true "Ignore rule"
// @Success     201 {object} IgnoreRuleResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/ignore-rules [post]
// @Tags        Ignore rules
func CreateIgnoreRule(log logging.Logger, commands *cashflow.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		req, ok := decodeIgnoreRuleRequest(w, r)
		if !ok {
			return
		}
		draft, err := req.draft()
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, ignoreRuleProblem(err))
			return
		}

		rule, err := commands.CreateIgnoreRule(r.Context(), accountID, draft)
		if err != nil {
			log.Error(r.Context(), "cashflow ignore rules: failed to create rule", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to create the ignore rule"})
			return
		}
		_ = httpx.JSONEncode(w, http.StatusCreated, toIgnoreRuleResponse(rule))
	})
}

// UpdateIgnoreRule rewrites what an existing rule matches on.
//
// @Summary     Update a cashflow ignore rule
// @Description Rewrites what a rule matches on, or switches it off. Transactions it already ignored keep their state — the edit says what the rule catches from now on.
// @Accept      application/json
// @Produce     application/json
// @Param       rule_id path string true "Ignore rule UUID"
// @Param       payload body IgnoreRuleRequest true "Ignore rule"
// @Success     200 {object} IgnoreRuleResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/ignore-rules/{rule_id} [put]
// @Tags        Ignore rules
func UpdateIgnoreRule(log logging.Logger, commands *cashflow.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		ruleID, ok := ignoreRuleID(w, r)
		if !ok {
			return
		}
		req, ok := decodeIgnoreRuleRequest(w, r)
		if !ok {
			return
		}
		draft, err := req.draft()
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, ignoreRuleProblem(err))
			return
		}

		rule, err := commands.UpdateIgnoreRule(r.Context(), accountID, ruleID, draft)
		if err != nil {
			if errors.Is(err, cashflow.ErrIgnoreRuleNotFound) {
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"rule_id": err.Error()})
				return
			}
			log.Error(r.Context(), "cashflow ignore rules: failed to update rule", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to update the ignore rule"})
			return
		}
		_ = httpx.JSONEncode(w, http.StatusOK, toIgnoreRuleResponse(rule))
	})
}

// DeleteIgnoreRule removes one ignore rule.
//
// @Summary     Delete a cashflow ignore rule
// @Description Removes a rule. Nothing it ignored is restored: those transactions stay as they are and are put back one by one, like any other ignored row.
// @Produce     application/json
// @Param       rule_id path string true "Ignore rule UUID"
// @Success     204 "No content"
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/ignore-rules/{rule_id} [delete]
// @Tags        Ignore rules
func DeleteIgnoreRule(log logging.Logger, commands *cashflow.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		ruleID, ok := ignoreRuleID(w, r)
		if !ok {
			return
		}

		if err := commands.DeleteIgnoreRule(r.Context(), accountID, ruleID); err != nil {
			if errors.Is(err, cashflow.ErrIgnoreRuleNotFound) {
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"rule_id": err.Error()})
				return
			}
			log.Error(r.Context(), "cashflow ignore rules: failed to delete rule", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete the ignore rule"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// PreviewIgnoreRule reports what a rule would catch, before it catches anything.
//
// @Summary     Preview a cashflow ignore rule
// @Description Reports how many transactions a rule matches, how many applying it would ignore, and a sample of the most recent matches.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body IgnoreRuleRequest true "Ignore rule"
// @Success     200 {object} IgnoreRulePreviewResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/ignore-rules/preview [post]
// @Tags        Ignore rules
func PreviewIgnoreRule(log logging.Logger, queries *cashflow.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		req, ok := decodeIgnoreRuleRequest(w, r)
		if !ok {
			return
		}
		draft, err := req.draft()
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, ignoreRuleProblem(err))
			return
		}

		preview, err := queries.PreviewIgnoreRule(r.Context(), accountID, draft)
		if err != nil {
			log.Error(r.Context(), "cashflow ignore rules: failed to preview rule", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to preview the ignore rule"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, IgnoreRulePreviewResponse{
			Matching:      preview.Matching,
			NotYetIgnored: preview.NotYetIgnored,
			Scanned:       preview.Scanned,
			Sample:        toTransactionResponses(preview.Sample),
		})
	})
}

// ApplyIgnoreRule ignores the existing transactions one rule matches.
//
// @Summary     Apply a cashflow ignore rule to existing transactions
// @Description Ignores the transactions already in the ledger that the rule matches, skipping rows that are ignored already and rows whose ignored state was decided by hand.
// @Produce     application/json
// @Param       rule_id path string true "Ignore rule UUID"
// @Success     200 {object} ApplyIgnoreRuleResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/ignore-rules/{rule_id}/apply [post]
// @Tags        Ignore rules
func ApplyIgnoreRule(log logging.Logger, commands *cashflow.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		ruleID, ok := ignoreRuleID(w, r)
		if !ok {
			return
		}

		ignored, err := commands.ApplyIgnoreRuleToExisting(r.Context(), accountID, ruleID)
		if err != nil {
			if errors.Is(err, cashflow.ErrIgnoreRuleNotFound) {
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"rule_id": err.Error()})
				return
			}
			log.Error(r.Context(), "cashflow ignore rules: failed to apply rule", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to apply the ignore rule"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, ApplyIgnoreRuleResponse{
			IgnoredCount: ignored,
			Status:       "ok",
		})
	})
}

// GetImportIgnored groups what one import's rules ignored, by rule.
//
// @Summary     What an import ignored, grouped by rule
// @Description Returns the transactions one import ignored automatically, grouped under the rule that caught them, so a rule can be judged on its own harvest.
// @Produce     application/json
// @Param       import_id path string true "Import UUID"
// @Success     200 {object} ImportIgnoredResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/imports/{import_id}/ignored [get]
// @Tags        Ignore rules
func GetImportIgnored(log logging.Logger, queries *cashflow.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		importID, err := uuid.Parse(r.PathValue("import_id"))
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"import_id": "import_id must be a valid UUID"})
			return
		}

		groups, err := queries.IgnoredByRule(r.Context(), accountID, importID)
		if err != nil {
			log.Error(r.Context(), "cashflow ignore rules: failed to read what an import ignored", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to read what the import ignored"})
			return
		}

		items := make([]IgnoredRuleGroupResponse, 0, len(groups))
		for _, group := range groups {
			items = append(items, IgnoredRuleGroupResponse{
				Rule:         toIgnoreRuleResponse(group.Rule),
				Total:        group.Total,
				Transactions: toTransactionResponses(group.Transactions),
			})
		}
		_ = httpx.JSONEncode(w, http.StatusOK, ImportIgnoredResponse{Data: items})
	})
}

func decodeIgnoreRuleRequest(w http.ResponseWriter, r *http.Request) (IgnoreRuleRequest, bool) {
	req, err := httpx.JSONDecode[IgnoreRuleRequest](r)
	if err != nil {
		if httpx.WriteDecodeError(w, err) {
			return IgnoreRuleRequest{}, false
		}
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
		return IgnoreRuleRequest{}, false
	}
	return req, true
}

func ignoreRuleID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("rule_id"))
	if err != nil {
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"rule_id": "rule_id must be a valid UUID"})
		return uuid.Nil, false
	}
	return id, true
}

// ignoreRuleProblem maps a draft's refusal to the field it is about, so the form can
// put the message next to the input that caused it.
func ignoreRuleProblem(err error) map[string]string {
	switch {
	case errors.Is(err, cashflow.ErrIgnoreRuleNameRequired), errors.Is(err, cashflow.ErrIgnoreRuleNameTooLong):
		return map[string]string{"name": err.Error()}
	case errors.Is(err, cashflow.ErrIgnoreRuleInvalidMatchField):
		return map[string]string{"match_field": err.Error()}
	case errors.Is(err, cashflow.ErrIgnoreRuleContainsTooShort), errors.Is(err, cashflow.ErrIgnoreRuleContainsTooLong):
		return map[string]string{"contains": err.Error()}
	case errors.Is(err, cashflow.ErrIgnoreRuleInvalidDirection):
		return map[string]string{"direction": err.Error()}
	default:
		return map[string]string{"error": err.Error()}
	}
}

func toIgnoreRuleResponse(rule *cashflow.IgnoreRule) IgnoreRuleResponse {
	direction := ""
	if rule.Direction != nil {
		direction = string(*rule.Direction)
	}
	return IgnoreRuleResponse{
		ID:            rule.ID,
		Name:          rule.Name,
		MatchField:    string(rule.MatchField),
		Contains:      rule.Contains,
		Direction:     direction,
		Source:        rule.Source,
		Enabled:       rule.Enabled,
		IgnoredTotal:  rule.IgnoredTotal,
		LastAppliedAt: rule.LastAppliedAt,
		CreatedAt:     rule.CreatedAt,
		UpdatedAt:     rule.UpdatedAt,
	}
}
