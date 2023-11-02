package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (tt *TimetableAPI) status(c *gin.Context) {

	status := tt.searcher.StatusLabel()
	attemptsCnt := tt.searcher.AttemptsCnt()
	solutionsCnt := tt.searcher.SolutionsCnt()

	retJson := gin.H{
		"solutions_cnt": solutionsCnt,
		"attempts":      attemptsCnt,
		"status":        status,
	}

	withData := c.Request.URL.Query().Get("with-data")

	if status == "process" && withData == "1" {
		retJson["tour_name"] = tt.searcher.Condition.TourName
		retJson["teams"] = tt.searcher.Condition.TeamsPrettyMap
		retJson["day_start"] = tt.searcher.Condition.DayStart
		retJson["day_end"] = tt.searcher.Condition.DayEnd
	}

	c.JSON(http.StatusOK, retJson)
}
