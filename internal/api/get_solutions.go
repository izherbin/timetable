package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (tt *TimetableAPI) getSolutions(c *gin.Context) {
	sol, att := tt.searcher.GetSolutions()
	if len(sol) > 5000 {
		sol = sol[:5000]
	}

	c.JSON(http.StatusOK, gin.H{
		"solutions": sol,
		"attempts":  att,
	})
}
