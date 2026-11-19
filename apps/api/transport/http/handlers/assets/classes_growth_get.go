package assets

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// GetClassesGrowthRequest contains the date window for the class worth histories.
type GetClassesGrowthRequest struct {
	From string `json:"from,omitempty" query:"from"`
	To   string `json:"to,omitempty" query:"to"`
}

// ClassGrowthResponse is the worth history of one asset class.
type ClassGrowthResponse struct {
	ClassID uuid.UUID                  `json:"class_id"`
	Name    string                     `json:"name"`
	Growth  []ClassGrowthPointResponse `json:"growth"`
}

// GetClassesGrowth returns the worth history of every non-archived asset class.
//
// @Summary List asset class growth
// @Description Returns the per-day worth history of every non-archived asset class, sorted by class name.
// @Tags assets
// @Accept json
// @Produce json
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} ClassGrowthResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/classes/growth [get]
func GetClassesGrowth(log logging.Logger, queries assets.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		req, err := httpx.DecodeQuery[GetClassesGrowthRequest](r)
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
			return
		}
		from, to, err := date.ParseFromTo(strings.TrimSpace(req.From), strings.TrimSpace(req.To))
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"date_range": err.Error()})
			return
		}
		series, err := queries.ClassGrowth(r.Context(), accountID, from, to)
		if err != nil {
			log.Error(r.Context(), "An error occurred while trying to get class growth", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get class growth"})
			return
		}
		res := make([]ClassGrowthResponse, 0, len(series))
		for _, class := range series {
			points := make([]ClassGrowthPointResponse, 0, len(class.Growth))
			for _, point := range class.Growth {
				points = append(points, ClassGrowthPointResponse{
					Date:       point.Date.Format("2006-01-02"),
					TotalWorth: point.TotalWorth.String(),
				})
			}
			res = append(res, ClassGrowthResponse{ClassID: class.ClassID, Name: class.Name, Growth: points})
		}
		_ = httpx.JSONEncode(w, http.StatusOK, res)
	})
}
