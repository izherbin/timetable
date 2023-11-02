package api

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sergrom/timetable/internal/api/req"
	"github.com/sergrom/timetable/internal/repository"
	"github.com/xuri/excelize/v2"
)

var (
	gamesTmpl, _ = template.New(`gamesTemplate`).Parse(`
	<script>
		var Teams = {
			{{range $key, $team := .teamsMap }}
			{{$key}}: "{{$team}}",
			{{end}}
		};
	</script>
	<div class="bttns-top-panel">
		<div class="pull-right">
			<button id="AddEntity" data-tag="game" data-id="-1" type="button" class="btn btn-sm btn-success"><i class="fa fa-plus" aria-hidden="true"></i> Добавить</button>
		</div>
		<div class="clearfix"></div>
	</div>
	<table class="data-table-powered table table-sm">
	<thead>
	  <tr>
		<th scope="col">Тур</th>
		<th scope="col">Команда1</th>
		<th scope="col">Команда2</th>
		<th scope="col">Переигровка (0-нет,1-да)</th>
		<th scope="col"></th>
	  </tr>
	</thead>
	<tbody>
	  {{range $key, $game := .games }}
	  <tr>
		<td scope="row">{{ index $game 1 }}</td>
		<td>{{ index $game 2}}</td>
		<td>{{ index $game 3}}</td>
		<td>{{ index $game 4}}</td>
		<td style="text-align:right">
			<button data-tag="game" data-id="{{ index $game 0 }}" data-team-id-1="{{ index $game 5}}" data-team-id-2="{{ index $game 6}}" type="button" class="edit-btn btn btn-sm btn-info"><i class="fa fa-pencil" aria-hidden="true"></i></button>
			<button data-tag="game" data-id="{{ index $game 0 }}" type="button" class="del-btn btn btn-sm btn-danger"><i class="fa fa-times" aria-hidden="true"></i></button>
		</td>
	  </tr>
	  {{end}}
	</tbody>
  </table>
	`)
)

// games ...
func (tt *TimetableAPI) games(c *gin.Context) {
	errs := make([]string, 0)
	body := ""

	games, err := tt.repo.GetGames()
	if err != nil {
		log.Println(err.Error())
		return
	}
	teamsMap, err := tt.repo.GetTeamsMap()
	if err != nil {
		log.Println(err.Error())
	}
	divsMap, err := tt.repo.GetDivisionsMap()
	if err != nil {
		log.Println(err.Error())
	}
	teamDivsMap := make(map[int]string, len(teamsMap))
	for _, t := range teamsMap {
		teamDivsMap[t.ID] = fmt.Sprintf("%s (%s)", t.Name, divsMap[t.DivisionID].Name)
	}
	fmt.Println(len(teamsMap), len(teamDivsMap))

	gamesData := make([][]string, 0, len(games))
	for _, g := range games {
		team1, ok := teamsMap[g.TeamID1]
		if !ok {
			// todo err
			continue
		}
		team2, ok := teamsMap[g.TeamID2]
		if !ok {
			// todo err
			continue
		}

		gamesData = append(gamesData, []string{strconv.Itoa(g.ID), g.Tour, fmt.Sprintf("%s (%s)", team1.Name, divsMap[team1.DivisionID].Name), fmt.Sprintf("%s (%s)", team2.Name, divsMap[team2.DivisionID].Name), strconv.Itoa(g.CanRematch), strconv.Itoa(g.TeamID1), strconv.Itoa(g.TeamID2)})
	}

	body = tt.renderTemplate(gamesTmpl, map[string]interface{}{
		"games":    gamesData,
		"teamsMap": teamDivsMap,
	})

	c.HTML(http.StatusOK, "tmpl.html", gin.H{
		"title":    "Конструктор турниров",
		"subtitle": "Предыдущие игры",
		"errors":   errs,
		"body":     template.HTML(body),
		"page":     "games",
	})
}

func (tt *TimetableAPI) delGame(id int) error {
	games, err := tt.repo.GetGames()
	if err != nil {
		return err
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Create a new sheet.
	index, err := f.NewSheet("Sheet1")
	if err != nil {
		return err
	}

	f.SetCellStr("Sheet1", "A1", "ID")
	f.SetCellStr("Sheet1", "B1", "тур (просто текст)")
	f.SetCellStr("Sheet1", "C1", "id первой команды")
	f.SetCellStr("Sheet1", "D1", "id второй команды")
	f.SetCellStr("Sheet1", "E1", "возможна переиговка (0-нет, 1-да)")

	for i, game := range games {
		if game.ID == id {
			continue
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), game.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), game.Tour)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), game.TeamID1)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), game.TeamID2)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+2), game.CanRematch)
	}

	f.SetActiveSheet(index)

	return f.SaveAs(filepath.Join(repository.DataDir, repository.GamesFile))
}

func (tt *TimetableAPI) saveGame(msg req.SaveGameRequest) error {
	if err := tt.validateGame(msg); err != nil {
		return err
	}

	gameID, err := strconv.Atoi(msg.ID)
	if err != nil {
		return err
	}

	if msg.TeamID1 == msg.TeamID2 {
		return errors.New("Команда1 не может быть равна Команде2")
	}

	games, err := tt.repo.GetGames()
	if err != nil {
		return err
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Create a new sheet.
	index, err := f.NewSheet("Sheet1")
	if err != nil {
		return err
	}

	f.SetCellStr("Sheet1", "A1", "ID")
	f.SetCellStr("Sheet1", "B1", "тур (просто текст)")
	f.SetCellStr("Sheet1", "C1", "id первой команды")
	f.SetCellStr("Sheet1", "D1", "id второй команды")
	f.SetCellStr("Sheet1", "E1", "возможна переиговка (0-нет, 1-да)")
	maxID := 0
	for i, game := range games {
		if maxID < game.ID {
			maxID = game.ID
		}
		if game.ID == gameID {
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), msg.ID)
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), msg.Tour)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), msg.TeamID1)
			f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), msg.TeamID2)
			f.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+2), msg.CanRematch)
			continue
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), game.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), game.Tour)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), game.TeamID1)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), game.TeamID2)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+2), game.CanRematch)
	}

	if gameID == -1 {
		idx := len(games) + 2
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", idx), maxID+1)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", idx), msg.Tour)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", idx), msg.TeamID1)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", idx), msg.TeamID2)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", idx), msg.CanRematch)
	}

	f.SetActiveSheet(index)

	return f.SaveAs(filepath.Join(repository.DataDir, repository.GamesFile))
}

func (tt *TimetableAPI) validateGame(msg req.SaveGameRequest) error {
	id, err := strconv.Atoi(msg.ID)
	if err != nil {
		return err
	}
	if id < 1 && id != -1 {
		return errors.New("ID incorrect")
	}

	team1, err := strconv.Atoi(msg.TeamID1)
	if err != nil {
		return err
	}
	team2, err := strconv.Atoi(msg.TeamID1)
	if err != nil {
		return err
	}
	if team1 < 1 || team2 < 1 {
		return errors.New("TeamID incorrect")
	}

	if msg.CanRematch != "0" && msg.CanRematch != "1" {
		return errors.New("CanRematch incorrect")
	}

	return nil
}
