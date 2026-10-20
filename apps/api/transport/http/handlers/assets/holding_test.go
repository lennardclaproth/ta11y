package assets

import "testing"

// TestPurchaseRequestValidateRejectsUnusableQuantities covers the strings
// strconv.ParseFloat accepts but a worth cannot be derived from: NaN and Inf pass
// every comparison with zero, and an absurd magnitude overflows money.Price into a
// negative amount once it is multiplied by a unit price.
func TestPurchaseRequestValidateRejectsUnusableQuantities(t *testing.T) {
	tests := []struct {
		name     string
		quantity string
		wantOK   bool
	}{
		{name: "decimal", quantity: "0.5", wantOK: true},
		{name: "not a number", quantity: "NaN"},
		{name: "lower case not a number", quantity: "nan"},
		{name: "infinity", quantity: "Inf"},
		{name: "signed infinity", quantity: "+Inf"},
		{name: "beyond the bound", quantity: "1e300"},
		{name: "zero", quantity: "0"},
		{name: "negative", quantity: "-1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := PurchaseRequest{Date: "2026-01-01", Quantity: test.quantity, UnitPrice: "100.00"}
			problems := make(map[string]string)
			req.validate(problems, "")

			_, hasProblem := problems["quantity"]
			if hasProblem == test.wantOK {
				t.Errorf("quantity %q: problem reported = %v, want %v (problems: %v)", test.quantity, hasProblem, !test.wantOK, problems)
			}
		})
	}
}
