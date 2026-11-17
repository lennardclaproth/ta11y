package cashflow

import (
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/date"
)

// TransactionFilters contains transaction filters for bulk tag and ignore mutations.
type TransactionFilters struct {
	Q           string `json:"q,omitempty"`
	Description string `json:"description,omitempty"`
	Note        string `json:"note,omitempty"`
	Source      string `json:"source,omitempty"`
	Direction   string `json:"direction,omitempty"`
	Tags        string `json:"tags,omitempty"`
	Untagged    *bool  `json:"untagged,omitempty"`
	HideIgnored *bool  `json:"hide_ignored,omitempty"`
	// ImportID narrows to the rows one import brought in, ImportedByRule to the rows
	// one ignore rule ignored. Together they are what "restore everything this rule
	// caught in this import" asks for.
	ImportID      string `json:"import_id,omitempty"`
	IgnoredByRule string `json:"ignored_by_rule,omitempty"`
	From          string `json:"from,omitempty"`
	To            string `json:"to,omitempty"`
}

func (tf TransactionFilters) ToAppFilters() (cashflow.TransactionFilters, map[string]string) {
	problems := make(map[string]string)
	direction, directionErr := cashflow.ParseDirection(tf.Direction)
	if directionErr != nil {
		problems["filters.direction"] = directionErr.Error()
	}
	from, to, dateErr := date.ParseFromTo(tf.From, tf.To)
	if dateErr != nil {
		problems["filters.date_range"] = dateErr.Error()
	}
	importID, importErr := optionalUUID(tf.ImportID)
	if importErr != nil {
		problems["filters.import_id"] = "import_id must be a valid UUID"
	}
	ruleID, ruleErr := optionalUUID(tf.IgnoredByRule)
	if ruleErr != nil {
		problems["filters.ignored_by_rule"] = "ignored_by_rule must be a valid UUID"
	}

	return cashflow.TransactionFilters{
		Query:           tf.Q,
		Description:     tf.Description,
		Note:            tf.Note,
		Source:          tf.Source,
		Direction:       direction,
		Tags:            cashflow.SplitTags(tf.Tags),
		Untagged:        tf.Untagged,
		HideIgnored:     tf.HideIgnored,
		ImportID:        importID,
		IgnoredByRuleID: ruleID,
		From:            from,
		To:              to,
	}, problems
}
