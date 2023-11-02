package api

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sergrom/timetable/internal/api/req"
	"github.com/sergrom/timetable/internal/repository"
	"github.com/sergrom/timetable/internal/services/searcher"
)

// TimetableAPI ...
type TimetableAPI struct {
	repo     *repository.Repo
	searcher *searcher.Searcher
}

// NewTimetableAPI ...
func NewTimetableAPI() *TimetableAPI {
	return &TimetableAPI{
		repo:     repository.NewRepo(),
		searcher: searcher.NewSearcher(),
	}
}

// GetHandlers ...
func (tt *TimetableAPI) GetHandlers() map[string]Handler {
	return map[string]Handler{
		"/": {
			Method: http.MethodGet,
			Fn:     tt.index,
		},
		"/teams": {
			Method: http.MethodGet,
			Fn:     tt.teams,
		},
		"/coaches": {
			Method: http.MethodGet,
			Fn:     tt.coaches,
		},
		"/stadiums": {
			Method: http.MethodGet,
			Fn:     tt.stadiums,
		},
		"/divisions": {
			Method: http.MethodGet,
			Fn:     tt.divisions,
		},
		// "/stadiums-download": {
		// 	Method: http.MethodGet,
		// 	Fn:     tt.stadiumsDownload,
		// },
		// "/divisions-download": {
		// 	Method: http.MethodGet,
		// 	Fn:     tt.divisionsDownload,
		// },
		// "/coaches-download": {
		// 	Method: http.MethodGet,
		// 	Fn:     tt.coachesDownload,
		// },
		// "/teams-download": {
		// 	Method: http.MethodGet,
		// 	Fn:     tt.teamsDownload,
		// },
		"/wishes": {
			Method: http.MethodGet,
			Fn:     tt.wishes,
		},
		"/games": {
			Method: http.MethodGet,
			Fn:     tt.games,
		},
		"/status": {
			Method: http.MethodGet,
			Fn:     tt.status,
		},
		"/search-start": {
			Method: http.MethodPost,
			Fn:     tt.searchStart,
		},
		"/search-stop": {
			Method: http.MethodPost,
			Fn:     tt.searchStop,
		},
		"/get-solutions": {
			Method: http.MethodGet,
			Fn:     tt.getSolutions,
		},
		"/download-solution": {
			Method: http.MethodGet,
			Fn:     tt.downloadSolution,
		},
		"/del-entity": {
			Method: http.MethodPost,
			Fn:     tt.delEntity,
		},
		"/save-entity": {
			Method: http.MethodPost,
			Fn:     tt.saveEntity,
		},
	}
}

func (tt *TimetableAPI) renderTemplate(tmpl *template.Template, data any) string {
	var tpl bytes.Buffer
	_ = tmpl.Execute(&tpl, data)
	return tpl.String()
}

// func (tt *TimetableAPI) uploadFile(c *gin.Context) {
// 	file, err := c.FormFile("file")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	c.SaveUploadedFile(file, "data/uploaded.xlsx")

// 	fmt.Println(c.Request.Form.Get("id"))

// 	switch c.Request.Form.Get("id") {
// 	case "stadiums":
// 		err = tt.uploadStadiums()
// 	case "divisions":
// 		err = tt.uploadDivisions()
// 	case "coaches":
// 		err = tt.uploadCoaches()
// 	case "teams":
// 		err = tt.uploadTeams()
// 	}

// 	if err != nil {
// 		c.String(http.StatusBadRequest, err.Error())
// 		return
// 	}

// 	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
// }

func (tt *TimetableAPI) delEntity(c *gin.Context) {
	q := c.Request.URL.Query()
	id, _ := strconv.Atoi(q.Get("id"))
	tag := q.Get("tag")

	var err error
	switch tag {
	case "stadium":
		err = tt.delStadium(id)
	case "division":
		err = tt.delDivision(id)
	case "coach":
		err = tt.delCoach(id)
	case "team":
		err = tt.delTeam(id)
	case "wish":
		err = tt.delWish(id)
	case "game":
		err = tt.delGame(id)
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"result": false,
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": true,
	})
}

func (tt *TimetableAPI) saveEntity(c *gin.Context) {
	q := c.Request.URL.Query()
	//id := q.Get("id")
	tag := q.Get("tag")
	var err error

	switch tag {
	case "stadium":
		var msg req.SaveStadiumRequest
		if err = c.BindJSON(&msg); err != nil {
			log.Println(err)
			break
		}
		err = tt.saveStadium(msg)
	case "division":
		var msg req.SaveDivisionRequest
		if err = c.BindJSON(&msg); err != nil {
			log.Println(err)
			break
		}
		err = tt.saveDivision(msg)
	case "coach":
		var msg req.SaveCoachRequest
		if err = c.BindJSON(&msg); err != nil {
			log.Println(err)
			break
		}
		err = tt.saveCoach(msg)
	case "team":
		var msg req.SaveTeamRequest
		if err = c.BindJSON(&msg); err != nil {
			log.Println(err)
			break
		}
		err = tt.saveTeam(msg)
	case "wish":
		var msg req.SaveWishRequest
		if err = c.BindJSON(&msg); err != nil {
			log.Println(err)
			break
		}
		err = tt.saveWish(msg)
	case "game":
		var msg req.SaveGameRequest
		if err = c.BindJSON(&msg); err != nil {
			log.Println(err)
			break
		}
		err = tt.saveGame(msg)
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"result": false,
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": true,
	})
}
