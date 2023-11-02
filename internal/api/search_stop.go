package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (tt *TimetableAPI) searchStop(c *gin.Context) {
	tt.searcher.Stop()
	//tt.searcher.Reset()
	c.JSON(http.StatusOK, gin.H{})
}
