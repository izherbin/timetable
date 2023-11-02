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
	divisionsTmpl, _ = template.New(`dividionsTemplate`).Parse(`
	<div class="bttns-top-panel">
		<div class="pull-right">
			<button id="AddEntity" data-tag="division" data-id="-1" type="button" class="btn btn-sm btn-success"><i class="fa fa-plus" aria-hidden="true"></i> Добавить</button>
		</div>
		<div class="clearfix"></div>
	</div>
	<table class="data-table-powered table table-sm">
	<thead>
	  <tr>
		<th scope="col">#ID</th>
		<th scope="col">Название</th>
		<th scope="col">Формат</th>
		<th scope="col"></th>
	  </tr>
	</thead>
	<tbody>
	  {{range $key, $div := .divisions }}
	  <tr>
		<td scope="row">{{ $div.ID }}</td>
		<td>{{ $div.Name }}</td>
		<td>{{ $div.Format }}</td>
		<td style="text-align:right">
			<button data-tag="division" data-id="{{ $div.ID }}" type="button" class="edit-btn btn btn-sm btn-info"><i class="fa fa-pencil" aria-hidden="true"></i></button>
			<button data-tag="division" data-id="{{ $div.ID }}" type="button" class="del-btn btn btn-sm btn-danger"><i class="fa fa-times" aria-hidden="true"></i></button>
		</td>
	  </tr>
	  {{end}}
	</tbody>
  </table>
`)
)

// divisions ...
func (tt *TimetableAPI) divisions(c *gin.Context) {
	errs := make([]string, 0)
	body := ""

	divs, err := tt.repo.GetDivisions()
	if err != nil {
		log.Println(err.Error())
		errs = append(errs, err.Error())
	} else {
		body = tt.renderTemplate(divisionsTmpl, map[string]interface{}{
			"divisions": divs,
		})
	}

	c.HTML(http.StatusOK, "tmpl.html", gin.H{
		"title":    "Конструктор турниров",
		"subtitle": "Дивизионы",
		"errors":   errs,
		"body":     template.HTML(body),
		"page":     "divisions",
	})
}

func (tt *TimetableAPI) delDivision(id int) error {
	divs, err := tt.repo.GetDivisions()
	if err != nil {
		return err
	}

	teams, err := tt.repo.GetTeams()
	if err != nil {
		return err
	}
	divTeams := make([]string, 0, 10)
	for _, t := range teams {
		if t.DivisionID == id {
			divTeams = append(divTeams, t.Name)
		}
	}
	if len(divTeams) > 0 {
		return fmt.Errorf("Невозможно удалить дивизион ID:%d, т.к. есть команды с таким дивизионом: %s", id, strings.Join(divTeams, ", "))
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
	f.SetCellStr("Sheet1", "B1", "Дивизион")
	f.SetCellStr("Sheet1", "C1", "Формат")

	for i, div := range divs {
		if div.ID == id {
			continue
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), div.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), div.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), div.Format)
	}

	f.SetActiveSheet(index)

	return f.SaveAs(filepath.Join(repository.DataDir, repository.DivisionsFile))
}

func (tt *TimetableAPI) saveDivision(msg req.SaveDivisionRequest) error {
	if err := tt.validateDivision(msg); err != nil {
		return err
	}

	divID, err := strconv.Atoi(msg.ID)
	if err != nil {
		return err
	}

	divs, err := tt.repo.GetDivisions()
	if err != nil {
		return err
	}

	for _, d := range divs {
		if d.ID != divID && strings.ToLower(msg.Name) == strings.ToLower(d.Name) {
			return fmt.Errorf("Дивизион с названием %s уже существует", d.Name)
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
	f.SetCellStr("Sheet1", "B1", "Дивизион")
	f.SetCellStr("Sheet1", "C1", "Формат")

	maxID := 0
	for i, div := range divs {
		if maxID < div.ID {
			maxID = div.ID
		}
		if div.ID == divID {
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), msg.ID)
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), msg.Name)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), msg.Format)
			continue
		}

		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), div.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), div.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), div.Format)
	}

	if divID == -1 {
		idx := len(divs) + 2
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", idx), maxID+1)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", idx), msg.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", idx), msg.Format)
	}

	f.SetActiveSheet(index)

	return f.SaveAs(filepath.Join(repository.DataDir, repository.DivisionsFile))
}

func (tt *TimetableAPI) validateDivision(msg req.SaveDivisionRequest) error {
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

	format, err := strconv.Atoi(msg.Format)
	if err != nil {
		return err
	}
	if format < 3 || format > 7 {
		return errors.New("Формат должен быть от 3 до 7")
	}

	return nil
}

// func (tt *TimetableAPI) divisionsDownload(c *gin.Context) {
// 	divs, err := tt.repo.GetDivisions()
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

// 	f.SetCellStr("Sheet1", "A1", "Дивизион")
// 	f.SetCellStr("Sheet1", "B1", "Формат")

// 	for i, stad := range divs {
// 		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), stad.Name)
// 		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), stad.Format)
// 	}

// 	// Set active sheet of the workbook.
// 	f.SetActiveSheet(index)

// 	c.Header("Content-Disposition", "attachment; filename=Дивизионы.xlsx")
// 	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
// 	c.Header("Pragma", "public")
// 	c.Header("Content-Transfer-Encoding", "binary")
// 	c.Header("Cache-Control", "must-revalidate")

// 	f.WriteTo(c.Writer)
// }

// func (tt *TimetableAPI) uploadDivisions() error {
// 	return tt.repo.UplodDivisions()
// }
