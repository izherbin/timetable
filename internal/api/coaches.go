package api

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sergrom/timetable/internal/api/req"
	"github.com/sergrom/timetable/internal/repository"
	"github.com/xuri/excelize/v2"
)

var (
	coachesTmpl, _ = template.New(`coachesTemplate`).Parse(`
	<div class="bttns-top-panel">
		<div class="pull-right">
			<button id="AddEntity" data-tag="coach" data-id="-1" type="button" class="btn btn-sm btn-success"><i class="fa fa-plus" aria-hidden="true"></i> Добавить</button>
		</div>
		<div class="clearfix"></div>
	</div>
	<table class="data-table-powered table table-sm">
	<thead>
	  <tr>
		<th scope="col">#ID</th>
		<th scope="col">Имя</th>
		<th scope="col"></th>
	  </tr>
	</thead>
	<tbody>
	  {{range $key, $coach := .coaches }}
	  <tr>
		<td scope="row">{{ $coach.ID }}</td>
		<td>{{ $coach.Name }}</td>
		<td style="text-align:right">
			<button data-tag="coach" data-id="{{ $coach.ID }}" type="button" class="edit-btn btn btn-sm btn-info"><i class="fa fa-pencil" aria-hidden="true"></i></button>
			<button data-tag="coach" data-id="{{ $coach.ID }}" type="button" class="del-btn btn btn-sm btn-danger"><i class="fa fa-times" aria-hidden="true"></i></button>
		</td>
	  </tr>
	  {{end}}
	</tbody>
  </table>
`)
)

// coaches ...
func (tt *TimetableAPI) coaches(c *gin.Context) {
	errs := make([]string, 0)
	body := ""

	coaches, err := tt.repo.GetCoaches()
	if err != nil {
		log.Println(err.Error())
		errs = append(errs, err.Error())
	} else {
		body = tt.renderTemplate(coachesTmpl, map[string]interface{}{
			"coaches": coaches,
		})
	}

	c.HTML(http.StatusOK, "tmpl.html", gin.H{
		"title":    "Конструктор турниров",
		"subtitle": "Тренеры",
		"body":     template.HTML(body),
		"page":     "coaches",
	})
}

func (tt *TimetableAPI) delCoach(id int) error {
	coaches, err := tt.repo.GetCoaches()
	if err != nil {
		return err
	}
	teams, err := tt.repo.GetTeams()
	if err != nil {
		return err
	}

	coachTeams := make([]string, 0, 10)
	for _, t := range teams {
		if id == t.CoachID {
			coachTeams = append(coachTeams, t.Name)
		}
	}
	if len(coachTeams) > 0 {
		return fmt.Errorf("Невозможно удалить тренера ID:%d, т.к. есть команды с таким тренером: %s", id, strings.Join(coachTeams, ", "))
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
	f.SetCellStr("Sheet1", "B1", "Имя")

	for i, c := range coaches {
		if c.ID == id {
			continue
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), c.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), c.Name)
	}

	f.SetActiveSheet(index)

	return f.SaveAs(filepath.Join(repository.DataDir, repository.CoachesFile))
}

func (tt *TimetableAPI) saveCoach(msg req.SaveCoachRequest) error {
	if err := tt.validateCoach(msg); err != nil {
		return err
	}

	coachID, err := strconv.Atoi(msg.ID)
	if err != nil {
		return err
	}

	coaches, err := tt.repo.GetCoaches()
	if err != nil {
		return err
	}

	for _, c := range coaches {
		if c.ID != coachID && strings.ToLower(msg.Name) == strings.ToLower(c.Name) {
			return fmt.Errorf("Тренер с именем %s уже существует", c.Name)
		}
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
	f.SetCellStr("Sheet1", "B1", "Имя")

	maxID := 0
	for i, coach := range coaches {
		if maxID < coach.ID {
			maxID = coach.ID
		}
		if coach.ID == coachID {
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), msg.ID)
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), msg.Name)
			continue
		}

		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), coach.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), coach.Name)
	}

	if coachID == -1 {
		idx := len(coaches) + 2
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", idx), maxID+1)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", idx), msg.Name)
	}

	f.SetActiveSheet(index)

	return f.SaveAs(filepath.Join(repository.DataDir, repository.CoachesFile))
}

func (tt *TimetableAPI) validateCoach(msg req.SaveCoachRequest) error {
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

	return nil
}

// func (tt *TimetableAPI) coachesDownload(c *gin.Context) {
// 	coaches, err := tt.repo.GetCoaches()
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

// 	f.SetCellStr("Sheet1", "A1", "Имя")

// 	for i, coach := range coaches {
// 		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), coach.Name)
// 	}

// 	// Set active sheet of the workbook.
// 	f.SetActiveSheet(index)

// 	c.Header("Content-Disposition", "attachment; filename=Тренеры.xlsx")
// 	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
// 	c.Header("Pragma", "public")
// 	c.Header("Content-Transfer-Encoding", "binary")
// 	c.Header("Cache-Control", "must-revalidate")

// 	f.WriteTo(c.Writer)
// }

// func (tt *TimetableAPI) uploadCoaches() error {
// 	return tt.repo.UplodCoaches()
// }
