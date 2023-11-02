package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sergrom/timetable/internal/api/req"
	"github.com/sergrom/timetable/internal/ds"
	"github.com/sergrom/timetable/internal/pkg"
	"github.com/sergrom/timetable/internal/services/searcher"
)

func (tt *TimetableAPI) searchStart(c *gin.Context) {
	var msg req.SearchStartRequest
	if err := c.BindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var stad *ds.Stadium
	stads, err := tt.repo.GetStadiums()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, s := range stads {
		if s.ID == msg.StaduiumID {
			stad = &s
			break
		}
	}
	if stad == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not find stadium by id " + strconv.Itoa(msg.StaduiumID)})
		return
	}

	fields := make([]ds.Field, 0, len(msg.Fields))
	for i, f := range msg.Fields {
		fields = append(fields, ds.NewField(i+1, f.Format, time.Duration(f.Dur)*time.Minute, f.From, f.To))
	}

	divisions, err := tt.repo.GetDivisions()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	divsMap := make(map[int]string, len(divisions))
	for _, d := range divisions {
		divsMap[d.ID] = d.Name
	}

	coaches, err := tt.repo.GetCoaches()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allTeams, err := tt.repo.GetTeams()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	teamsMap := msg.GetTemsMap()
	teams := make([]ds.Team, 0, len(msg.Teams))
	for _, t := range allTeams {
		if teamsMap[t.ID] {
			teams = append(teams, t)
		}
	}

	wishes := make([]ds.Wish, 0, len(msg.Wishes))
	for _, w := range msg.Wishes {
		from, err1 := pkg.ParseHM(w.From)
		to, err2 := pkg.ParseHM(w.To)
		if err1 != nil || err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err1.Error() + ", " + err2.Error()})
			return
		}
		wishes = append(wishes, ds.Wish{
			TeamID:   w.TeamID,
			TimeFrom: from,
			TimeTo:   to,
		})
	}

	games := make([]ds.Game, 0, len(msg.Games))
	for _, g := range msg.Games {
		games = append(games, ds.Game{
			Tour:       "",
			TeamID1:    g.TeamID1,
			TeamID2:    g.TeamID2,
			CanRematch: 0,
		})
	}

	cond := searcher.NewCondition(msg.TourName, fields, divisions, coaches, teams, wishes, games)
	solutions, att, err := tt.searcher.Search(cond)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tour_name": cond.TourName,
		"solutions": solutions,
		"attempts":  att,
		"teams":     cond.TeamsPrettyMap,
		"day_start": cond.DayStart,
		"day_end":   cond.DayEnd,
	})
}
