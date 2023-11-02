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
	"github.com/sergrom/timetable/internal/pkg"
	"github.com/sergrom/timetable/internal/repository"
	"github.com/xuri/excelize/v2"
)

var (
	stadiumsTmpl, _ = template.New(`stadiumsTemplate`).Parse(`
	<div class="bttns-top-panel">
		<div class="pull-right">
			<button id="AddEntity" data-tag="stadium" data-id="-1" type="button" class="btn btn-sm btn-success"><i class="fa fa-plus" aria-hidden="true"></i> Добавить</button>
		</div>
		<div class="clearfix"></div>
	</div>
	<table class="data-table-powered table table-sm">
	<thead>
	  <tr>
		<th scope="col">#ID</th>
		<th scope="col">Название</th>
		<th scope="col">Полей</th>
		<th scope="col">Формат</th>
		<th scope="col">Работает с</th>
		<th scope="col">Работает по</th>
		<th scope="col">Игра (минут)</th>
		<th scope="col"></th>
	  </tr>
	</thead>
	<tbody>
	  {{range $key, $stad := .stadiums }}
	  <tr>
		<td scope="row">{{ $stad.ID }}</td>
		<td>{{ $stad.Name }}</td>
		<td>{{ $stad.Fields }}</td>
		<td>{{ $stad.Format }}</td>
		<td>{{ $stad.TimeFrom.Format "15:04" }}</td>
		<td>{{ $stad.TimeTo.Format "15:04" }}</td>
		<td>{{ $stad.GameDur.Minutes }} мин.</td>
		<td style="text-align:right">
			<button data-tag="stadium" data-id="{{ $stad.ID }}" type="button" class="edit-btn btn btn-sm btn-info"><i class="fa fa-pencil" aria-hidden="true"></i></button>
			<button data-tag="stadium" data-id="{{ $stad.ID }}" type="button" class="del-btn btn btn-sm btn-danger"><i class="fa fa-times" aria-hidden="true"></i></button>
		</td>
	  </tr>
	  {{end}}
	</tbody>
  </table>
`)
)

// stadiums ...
func (tt *TimetableAPI) stadiums(c *gin.Context) {
	errs := make([]string, 0)
	body := ""

	stads, err := tt.repo.GetStadiums()
	if err != nil {
		log.Println(err.Error())
		errs = append(errs, err.Error())
	} else {
		body = tt.renderTemplate(stadiumsTmpl, map[string]interface{}{
			"stadiums": stads,
		})
	}

	c.HTML(http.StatusOK, "tmpl.html", gin.H{
		"title":    "Конструктор турниров",
		"subtitle": "Стадионы",
		"errors":   errs,
		"body":     template.HTML(body),
		"page":     "stadiums",
	})
}

func (tt *TimetableAPI) delStadium(id int) error {
	stads, err := tt.repo.GetStadiums()
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
	f.SetCellStr("Sheet1", "C1", "Полей")
	f.SetCellStr("Sheet1", "D1", "Формат")
	f.SetCellStr("Sheet1", "E1", "Работает с")
	f.SetCellStr("Sheet1", "F1", "Работает по")
	f.SetCellStr("Sheet1", "G1", "Игра (минут)")

	for i, stad := range stads {
		if stad.ID == id {
			continue
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), stad.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), stad.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), stad.Fields)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), stad.Format)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+2), stad.TimeFrom.Format("15:04"))
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", i+2), stad.TimeTo.Format("15:04"))
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", i+2), stad.GameDur.Minutes())
	}

	f.SetActiveSheet(index)

	return f.SaveAs(filepath.Join(repository.DataDir, repository.StadiumsFile))
}

func (tt *TimetableAPI) saveStadium(msg req.SaveStadiumRequest) error {
	if err := tt.validateStad(msg); err != nil {
		return err
	}

	stadID, err := strconv.Atoi(msg.ID)
	if err != nil {
		return err
	}

	stads, err := tt.repo.GetStadiums()
	if err != nil {
		return err
	}

	for _, s := range stads {
		if s.ID != stadID && strings.ToLower(msg.Name) == strings.ToLower(s.Name) {
			return fmt.Errorf("Стадион с названием %s уже существует", s.Name)
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
	f.SetCellStr("Sheet1", "B1", "Название")
	f.SetCellStr("Sheet1", "C1", "Полей")
	f.SetCellStr("Sheet1", "D1", "Формат")
	f.SetCellStr("Sheet1", "E1", "Работает с")
	f.SetCellStr("Sheet1", "F1", "Работает по")
	f.SetCellStr("Sheet1", "G1", "Игра (минут)")

	maxID := 0
	for i, stad := range stads {
		if maxID < stad.ID {
			maxID = stad.ID
		}
		if stad.ID == stadID {
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), msg.ID)
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), msg.Name)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), msg.Fields)
			f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), msg.Format)
			f.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+2), msg.TimeFrom)
			f.SetCellValue("Sheet1", fmt.Sprintf("F%d", i+2), msg.TimeTo)
			f.SetCellValue("Sheet1", fmt.Sprintf("G%d", i+2), msg.GameDur)
			continue
		}

		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), stad.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), stad.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), stad.Fields)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), stad.Format)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+2), stad.TimeFrom.Format("15:04"))
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", i+2), stad.TimeTo.Format("15:04"))
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", i+2), stad.GameDur.Minutes())
	}

	if stadID == -1 {
		idx := len(stads) + 2
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", idx), maxID+1)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", idx), msg.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", idx), msg.Fields)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", idx), msg.Format)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", idx), msg.TimeFrom)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", idx), msg.TimeTo)
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", idx), msg.GameDur)
	}

	f.SetActiveSheet(index)

	return f.SaveAs(filepath.Join(repository.DataDir, repository.StadiumsFile))
}

func (tt *TimetableAPI) validateStad(msg req.SaveStadiumRequest) error {
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

	fields, err := strconv.Atoi(msg.Fields)
	if err != nil {
		return err
	}
	if fields < 1 || fields > 50 {
		return errors.New("The number of fields must be from 1 to 50")
	}

	format, err := strconv.Atoi(msg.Format)
	if err != nil {
		return err
	}
	if format < 3 || format > 7 {
		return errors.New("Формат должен быть от 3 до 7")
	}

	if !pkg.ValidateTime(msg.TimeFrom) {
		return errors.New("TimeFrom incorrect")
	}
	if !pkg.ValidateTime(msg.TimeTo) {
		return errors.New("TimeTo incorrect")
	}

	gameDur, err := strconv.Atoi(msg.GameDur)
	if err != nil {
		return err
	}
	if gameDur < 10 || gameDur > 150 {
		return errors.New("gameDur must be from 10 to 150 min")
	}

	return nil
}

// func (tt *TimetableAPI) stadiumsDownload(c *gin.Context) {
// 	stads, err := tt.repo.GetStadiums()
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
// 	f.SetCellStr("Sheet1", "B1", "Полей")
// 	f.SetCellStr("Sheet1", "C1", "Формат")
// 	f.SetCellStr("Sheet1", "D1", "Работает с")
// 	f.SetCellStr("Sheet1", "E1", "Работает по")
// 	f.SetCellStr("Sheet1", "F1", "Игра (минут)")

// 	for i, stad := range stads {
// 		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), stad.Name)
// 		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), stad.Fields)
// 		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), stad.Format)
// 		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), stad.TimeFrom.Format("15:04"))
// 		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+2), stad.TimeTo.Format("15:04"))
// 		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", i+2), stad.GameDur.Minutes())
// 	}

// 	// Set active sheet of the workbook.
// 	f.SetActiveSheet(index)

// 	c.Header("Content-Disposition", "attachment; filename=Стадионы.xlsx")
// 	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
// 	c.Header("Pragma", "public")
// 	c.Header("Content-Transfer-Encoding", "binary")
// 	c.Header("Cache-Control", "must-revalidate")

// 	f.WriteTo(c.Writer)
// }

// func (tt *TimetableAPI) uploadStadiums() error {
// 	return tt.repo.UplodStadiums()
// }
