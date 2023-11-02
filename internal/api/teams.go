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
	teamsTmpl, _ = template.New(`teamsTemplate`).Parse(`
	<script>
		var Divisions = {
			{{range $key, $div := .divsMap }}
			{{$div.ID}}: "{{$div.Name}}",
			{{end}}
		};
		var Coaches = {
			{{range $key, $coach := .coachesMap }}
			{{$coach.ID}}: "{{$coach.Name}}",
			{{end}}
		};
	</script>
	<div class="bttns-top-panel">
		<div class="pull-right">
			<button id="AddEntity" data-tag="team" data-id="-1" type="button" class="btn btn-sm btn-success"><i class="fa fa-plus" aria-hidden="true"></i> Добавить</button>
		</div>
		<div class="clearfix"></div>
	</div>
	<table class="data-table-powered table table-sm">
	<thead>
	  <tr>
		<th scope="col">#ID</th>
		<th scope="col">Название</th>
		<th scope="col">Дивизион</th>
		<th scope="col">Тренер</th>
		<th scope="col"></th>
	  </tr>
	</thead>
	<tbody>
	  {{range $key, $team := .teams }}
	  <tr>
		<td scope="row">{{ index $team 0 }}</td>
		<td>{{ index $team 1}}</td>
		<td>{{ index $team 2}}</td>
		<td>{{ index $team 3}}</td>
		<td style="text-align:right">
			<button data-tag="team" data-id="{{ index $team 0 }}" data-div-id="{{ index $team 4 }}" data-coach-id="{{ index $team 5 }}" type="button" class="edit-btn btn btn-sm btn-info"><i class="fa fa-pencil" aria-hidden="true"></i></button>
			<button data-tag="team" data-id="{{ index $team 0 }}" type="button" class="del-btn btn btn-sm btn-danger"><i class="fa fa-times" aria-hidden="true"></i></button>
		</td>
	  </tr>
	  {{end}}
	</tbody>
  </table>
	`)
)

// teams ...
func (tt *TimetableAPI) teams(c *gin.Context) {
	errs := make([]string, 0)
	body := ""

	teams, err1 := tt.repo.GetTeams()
	coachesMap, err2 := tt.repo.GetCoachesMap()
	divsMap, err3 := tt.repo.GetDivisionsMap()

	if err1 != nil || err2 != nil || err3 != nil {
		log.Println(err1.Error(), err2.Error(), err3.Error())
		//errs = append(errs, err1.Error()) // todo
	} else {
		teamsData := make([][]string, 0, len(teams))
		for _, t := range teams {
			coach, ok := coachesMap[t.CoachID]
			if !ok {
				// todo err
				continue
			}
			div, ok := divsMap[t.DivisionID]
			if !ok {
				// todo err
				continue
			}
			teamsData = append(teamsData, []string{strconv.Itoa(t.ID), t.Name, div.Name, coach.Name, strconv.Itoa(div.ID), strconv.Itoa(coach.ID)})
		}

		body = tt.renderTemplate(teamsTmpl, map[string]interface{}{
			"teams":      teamsData,
			"divsMap":    divsMap,
			"coachesMap": coachesMap,
		})
	}

	c.HTML(http.StatusOK, "tmpl.html", gin.H{
		"title":    "Конструктор турниров",
		"subtitle": "Команды",
		"errors":   errs,
		"body":     template.HTML(body),
		"page":     "teams",
	})
}

func (tt *TimetableAPI) delTeam(id int) error {
	teams, err := tt.repo.GetTeams()
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
	f.SetCellStr("Sheet1", "B1", "Название")
	f.SetCellStr("Sheet1", "C1", "ID тренера")
	f.SetCellStr("Sheet1", "D1", "ID дивизиона")

	for i, team := range teams {
		if team.ID == id {
			continue
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), team.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), team.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), team.CoachID)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), team.DivisionID)
	}

	f.SetActiveSheet(index)

	return f.SaveAs(filepath.Join(repository.DataDir, repository.TeamsFile))
}

func (tt *TimetableAPI) saveTeam(msg req.SaveTeamRequest) error {
	if err := tt.validateTeam(msg); err != nil {
		return err
	}

	teamID, err := strconv.Atoi(msg.ID)
	if err != nil {
		return err
	}

	teams, err := tt.repo.GetTeams()
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
	f.SetCellStr("Sheet1", "B1", "Название")
	f.SetCellStr("Sheet1", "C1", "ID тренера")
	f.SetCellStr("Sheet1", "D1", "ID дивизиона")

	maxID := 0
	for i, team := range teams {
		if maxID < team.ID {
			maxID = team.ID
		}
		if team.ID == teamID {
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), msg.ID)
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), msg.Name)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), msg.CoachID)
			f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), msg.DivisionID)
			continue
		}

		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), team.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), team.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), team.CoachID)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), team.DivisionID)
	}

	if teamID == -1 {
		idx := len(teams) + 2
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", idx), maxID+1)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", idx), msg.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", idx), msg.CoachID)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", idx), msg.DivisionID)
	}

	f.SetActiveSheet(index)

	return f.SaveAs(filepath.Join(repository.DataDir, repository.TeamsFile))
}

func (tt *TimetableAPI) validateTeam(msg req.SaveTeamRequest) error {
	id, err := strconv.Atoi(msg.ID)
	if err != nil {
		return err
	}
	if id < 1 && id != -1 {
		return errors.New("ID incorrect")
	}

	if len(msg.Name) == 0 {
		return errors.New("empty Name")
	}

	divID, err := strconv.Atoi(msg.DivisionID)
	if err != nil {
		return err
	}
	if divID < 1 {
		return errors.New("DivisionID incorrect")
	}

	coachID, err := strconv.Atoi(msg.CoachID)
	if err != nil {
		return err
	}
	if coachID < 1 {
		return errors.New("CoachID incorrect")
	}

	return nil
}

// func (tt *TimetableAPI) teamsDownload(c *gin.Context) {
// 	teams, err := tt.repo.GetTeams()
// 	if err != nil {
// 		c.Writer.WriteHeader(http.StatusInternalServerError)
// 		c.Writer.WriteString("Произошла ошибка")
// 		return
// 	}
// 	divMap, err := tt.repo.GetDivisionsMap()
// 	if err != nil {
// 		c.Writer.WriteHeader(http.StatusInternalServerError)
// 		c.Writer.WriteString("Произошла ошибка")
// 		return
// 	}
// 	coachesMap, err := tt.repo.GetCoachesMap()
// 	if err != nil {
// 		c.Writer.WriteHeader(http.StatusInternalServerError)
// 		c.Writer.WriteString("Произошла ошибка")
// 		return
// 	}

// 	f := excelize.NewFile()
// 	defer func() {
// 		if err := f.Close(); err != nil {
// 			fmt.Println(err)
// 		}
// 	}()

// 	// Create a new sheet.
// 	index, err := f.NewSheet("Sheet1")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	f.SetCellStr("Sheet1", "A1", "Название")
// 	f.SetCellStr("Sheet1", "B1", "Дивизион")
// 	f.SetCellStr("Sheet1", "C1", "Тренер")

// 	for i, team := range teams {
// 		coach := coachesMap[team.CoachID]
// 		div := divMap[team.DivisionID]
// 		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), team.Name)
// 		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), div.Name)
// 		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), coach.Name)
// 	}

// 	// Set active sheet of the workbook.
// 	f.SetActiveSheet(index)

// 	c.Header("Content-Disposition", "attachment; filename=Команды.xlsx")
// 	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
// 	c.Header("Pragma", "public")
// 	c.Header("Content-Transfer-Encoding", "binary")
// 	c.Header("Cache-Control", "must-revalidate")

// 	f.WriteTo(c.Writer)
// }

// func (tt *TimetableAPI) uploadTeams() error {
// 	return tt.repo.UplodTeams()
// }
