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
	"github.com/sergrom/timetable/internal/pkg"
	"github.com/sergrom/timetable/internal/repository"
	"github.com/xuri/excelize/v2"
)

var (
	wishesTmpl, _ = template.New(`wishesTemplate`).Parse(`
	<script>
		var Teams = {
			{{range $key, $team := .teamsMap }}
			{{$key}}: "{{$team}}",
			{{end}}
		};
	</script>
	<div class="bttns-top-panel">
		<div class="pull-right">
			<button id="AddEntity" data-tag="wish" data-id="-1" type="button" class="btn btn-sm btn-success"><i class="fa fa-plus" aria-hidden="true"></i> Добавить</button>
		</div>
		<div class="clearfix"></div>
	</div>
	<table class="data-table-powered table table-sm">
	<thead>
	  <tr>
		<th scope="col">Команда</th>
		<th scope="col">с</th>
		<th scope="col">по</th>
		<th scope="col"></th>
	  </tr>
	</thead>
	<tbody>
	  {{range $key, $wish := .wishes }}
	  <tr>
		<td scope="row">{{ index $wish 1 }}</td>
		<td>{{ index $wish 2}}</td>
		<td>{{ index $wish 3}}</td>
		<td style="text-align:right">
			<button data-tag="wish" data-id="{{ index $wish 0 }}" data-team-id="{{ index $wish 4}}" type="button" class="edit-btn btn btn-sm btn-info"><i class="fa fa-pencil" aria-hidden="true"></i></button>
			<button data-tag="wish" data-id="{{ index $wish 0 }}" type="button" class="del-btn btn btn-sm btn-danger"><i class="fa fa-times" aria-hidden="true"></i></button>
		</td>
	  </tr>
	  {{end}}
	</tbody>
  </table>
`)
)

// wishes ...
func (tt *TimetableAPI) wishes(c *gin.Context) {
	errs := make([]string, 0)
	body := ""

	wishes, err := tt.repo.GetWishes()
	if err != nil {
		log.Println(err.Error())
		return
		//errs = append(errs, err1.Error()) // todo
	}
	teamsMap, err := tt.repo.GetTeamsMap()
	if err != nil {
		log.Println(err.Error())
		return
		//errs = append(errs, err1.Error()) // todo
	}
	divsMap, err := tt.repo.GetDivisionsMap()
	if err != nil {
		log.Println(err.Error())
		return
	}
	teamDivsMap := make(map[int]string, len(teamsMap))
	for _, t := range teamsMap {
		teamDivsMap[t.ID] = fmt.Sprintf("%s (%s)", t.Name, divsMap[t.DivisionID].Name)
	}

	wishesData := make([][]string, 0, len(wishes))
	for _, w := range wishes {
		team, ok := teamsMap[w.TeamID]
		if !ok {
			continue
		}
		from, to := "", ""
		if !w.TimeFrom.IsZero() {
			from = w.TimeFrom.Format("15:04")
		}
		if !w.TimeTo.IsZero() {
			to = w.TimeTo.Format("15:04")
		}
		wishesData = append(wishesData, []string{strconv.Itoa(w.ID), fmt.Sprintf("%s (%s)", team.Name, divsMap[team.DivisionID].Name), from, to, strconv.Itoa(team.ID)})
	}

	body = tt.renderTemplate(wishesTmpl, map[string]interface{}{
		"wishes":   wishesData,
		"teamsMap": teamDivsMap,
	})

	c.HTML(http.StatusOK, "tmpl.html", gin.H{
		"title":    "Конструктор турниров",
		"subtitle": "Пожелания",
		"errors":   errs,
		"body":     template.HTML(body),
		"page":     "wishes",
	})
}

func (tt *TimetableAPI) delWish(id int) error {
	wishes, err := tt.repo.GetWishes()
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
	f.SetCellStr("Sheet1", "B1", "ID команды")
	f.SetCellStr("Sheet1", "C1", "с")
	f.SetCellStr("Sheet1", "D1", "по")

	for i, wish := range wishes {
		if wish.ID == id {
			continue
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), wish.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), wish.TeamID)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), wish.TimeFrom.Format("15:04"))
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), wish.TimeTo.Format("15:04"))
	}

	f.SetActiveSheet(index)

	return f.SaveAs(filepath.Join(repository.DataDir, repository.WishesFile))
}

func (tt *TimetableAPI) saveWish(msg req.SaveWishRequest) error {
	if err := tt.validateWish(msg); err != nil {
		return err
	}

	wishID, err := strconv.Atoi(msg.ID)
	if err != nil {
		return err
	}
	teamID, err := strconv.Atoi(msg.TeamID)
	if err != nil {
		return err
	}

	wishes, err := tt.repo.GetWishes()
	if err != nil {
		return err
	}

	for _, w := range wishes {
		if w.ID != wishID && teamID == w.TeamID {
			return fmt.Errorf("Пожелание для этой команды уже есть")
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
	f.SetCellStr("Sheet1", "B1", "ID команды")
	f.SetCellStr("Sheet1", "C1", "с")
	f.SetCellStr("Sheet1", "D1", "по")

	maxID := 0
	for i, wish := range wishes {
		if maxID < wish.ID {
			maxID = wish.ID
		}
		if wish.ID == wishID {
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), msg.ID)
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), msg.TeamID)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), msg.TimeFrom)
			f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), msg.TimeTo)
			continue
		}

		from, to := "", ""
		if !wish.TimeFrom.IsZero() {
			from = wish.TimeFrom.Format("15:04")
		}
		if !wish.TimeTo.IsZero() {
			to = wish.TimeTo.Format("15:04")
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), wish.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), wish.TeamID)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), from)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), to)
	}

	if wishID == -1 {
		idx := len(wishes) + 2
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", idx), maxID+1)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", idx), msg.TeamID)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", idx), msg.TimeFrom)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", idx), msg.TimeTo)
	}

	f.SetActiveSheet(index)

	return f.SaveAs(filepath.Join(repository.DataDir, repository.WishesFile))
}

func (tt *TimetableAPI) validateWish(msg req.SaveWishRequest) error {
	id, err := strconv.Atoi(msg.ID)
	if err != nil {
		return err
	}
	if id < 1 && id != -1 {
		return errors.New("ID incorrect")
	}

	teamID, err := strconv.Atoi(msg.TeamID)
	if err != nil {
		return err
	}
	if teamID < 1 {
		return errors.New("TeamID incorrect")
	}

	if msg.TimeFrom == "" && msg.TimeTo == "" {
		return errors.New("TimeFrom and TimeTo cannot be empty both")
	}
	if msg.TimeFrom != "" && !pkg.ValidateTime(msg.TimeFrom) {
		return errors.New("TimeFrom incorrect")
	}
	if msg.TimeTo != "" && !pkg.ValidateTime(msg.TimeTo) {
		return errors.New("TimeTo incorrect")
	}

	return nil
}
