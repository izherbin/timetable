package api

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sergrom/timetable/internal/ds"
)

var (
	mainTmpl, _ = template.New(`mainTemplate`).Parse(`
		<div class="container">
			<form>
				<div class="row" style="padding-bottom:20px">
					<div class="col-6">
						<div class="form-group">
							<label>Название тура</label>
							<input id="TourName" class="form-control form-control-sm" type="text" value="Тур_1">
						</div>
						<div class="form-group">
							<label>Стадион</label>
							<select id="StadID" class="select2 form-control form-control-sm">
								{{range $i, $stad := .stads }}
								<option value="{{$stad.ID}}" data-fields="{{$stad.Fields}}" data-format="{{$stad.Format}}" data-from="{{$stad.TimeFrom.Format "15:04"}}" data-to="{{$stad.TimeTo.Format "15:04"}}" data-game-dur="{{$stad.GameDur.Minutes}}" {{ if eq $i 0}}selected{{end}}>{{$stad.Name}}</option>
								{{end}}
							</select>
						</div>
						<div class="form-group">
							<table id="FieldsTable" class="table table-sm">
								<thead class="thead-light">
									<tr>
										<th scope="col">Поле</th>
										<th scope="col">Формат</th>
										<th scope="col" colspan="2">Работает (с/по)</th>
										<th scope="col">Игра (мин.)</th>
										<th scope="col"></th>
										<th scope="col"></th>
									</tr>
								</thead>
								<tbody></tbody>
							</table>
							<button id="AddField" type="button" class="btn btn-sm btn-success">＋ добавить</button>
						</div>
					</div>
					<div class="col-6">
						<div id="TeamsSelects" class="form-group">
							<label>Команды</label>
							<span class="pull-right font-weight-bold">всего: <span id="TeamsCnt">_</span></span>
							{{range $div, $teams := .teamsByDiv }}
								<h6 style="margin:5px 0 0 0;">{{$div}} <span class="pull-right" style="font-size:0.95rem;">выбрано: <span class="teams-div-selected">_</span></span></h6>
								<select multiple="multiple" class="select2 form-control form-control-sm" style="width:100%">
									{{range $i, $team := $teams }}
									<option value="{{$team.ID}}" selected>{{$team.Name}}</option>
									{{end}}
								</select>
							{{end}}
						</div>
					</div>
				</div>
				<div class="row" style="padding-bottom:20px">
					<div class="col-6">
						<div class="form-group">
							<label>Предыдущие игры</label>
							<table id="GamesTable" class="table table-sm">
								<thead class="thead-light">
									<tr>
										<th scope="col">Тур</th>
										<th scope="col">Команда1</th>
										<th scope="col">Команда2</th>
										<th scope="col" title="Возможна переигровка" style="text-align:center">Переигр.</th>
									</tr>
								</thead>
								<tbody>
									{{range $i, $game := .gamesData }}
									<tr data-team-id-1="{{index $game 1}}" data-team-id-2="{{index $game 3}}">
										<th>{{index $game 0}}</th>
										<td>{{index $game 2}} ({{index $game 6}})</td>
										<td>{{index $game 4}} ({{index $game 7}})</td>
										<td style="text-align:center"><input type="checkbox" class="form-control form-control-sm" {{index $game 5}}></td>
									</tr>
									{{end}}
								<tbody>
							</table>
						</div>
					</div>
					<div class="col-6">
						<div class="form-group">
							<label>Пожелания</label>
							<table id="WishTable" class="table table-sm">
								<thead class="thead-light">
									<tr>
										<th scope="col">Команда</th>
										<th scope="col" style="width:100px">С</th>
										<th scope="col" style="width:100px">По</th>
										<th scope="col" style="width:10px"></th>
									</tr>
								</thead>
								<tbody>
									{{range $i, $wish := .wishesData }}
									<tr data-team-id="{{index $wish 0}}">
										<th>{{index $wish 1}} ({{index $wish 4}})</th>
										<td><input type="text" class="form-control form-control-sm w-from" value="{{index $wish 2}}"></td>
										<td><input type="text" class="form-control form-control-sm w-to" value="{{index $wish 3}}"></td>
										<td><button type="button" class="btn btn-sm btn-warning x-wish-btn">✕</button></td>
									</tr>
									{{end}}
								<tbody>
							</table>
						</div>
					</div>
				</div>
				<div class="row" style="padding-bottom:20px">
					<div class="col-12">
						<button id="GO" type="button" class="btn btn-success btn-lg" style="display:block;width:300px;margin:0 auto;">Пуск</button>
					</div>
				</div>
			</form>
			<div id="Results" class="row">
				<div id="SolutionDetailsArea" class="col-8">
					<h4>Выберите решение из списка</h4>
				</div>
				<div class="col-4">
					<div id="SolutionsList">
						<div class="card">
							<div class="card-header">
								Найдено <span id="SolCnt">0</span><small>/</small><small id="AttCnt">0</small>
								<button id="LoadSolutions" type="button" class="btn btn-success btn-sm" style="float:right" title="Подгрузить новые">⟳</button>
							</div>
							<ul class="list-group">
							</ul>
						</div>
					</div>
				</div>
			</div>
		</div>
	`)
)

// index ...
func (tt *TimetableAPI) index(c *gin.Context) {
	errs := make([]string, 0)
	divsMap, err := tt.repo.GetDivisionsMap()
	if err != nil {
		errs = append(errs, err.Error())
	}
	teams, err := tt.repo.GetTeams()
	if err != nil {
		errs = append(errs, err.Error())
	}
	stads, err := tt.repo.GetStadiums()
	if err != nil {
		errs = append(errs, err.Error())
	}
	wishes, err := tt.repo.GetWishes()
	if err != nil {
		errs = append(errs, err.Error())
	}
	games, err := tt.repo.GetGames()
	if err != nil {
		errs = append(errs, err.Error())
	}

	teamsByDiv := make(map[string][]ds.Team)
	teamsByID := make(map[int]ds.Team)
	for _, t := range teams {
		if d, ok := divsMap[t.DivisionID]; ok {
			key := fmt.Sprintf("%s (формат %d)", d.Name, d.Format)
			teamsByDiv[key] = append(teamsByDiv[key], t)
		}
		teamsByID[t.ID] = t
	}
	sortTeamsByDiv(teamsByDiv)

	wishesData := make([][]string, 0, len(wishes))
	for _, w := range wishes {
		t, tOk := teamsByID[w.TeamID]
		if !tOk {
			continue
		}
		from, to := "", ""
		if !w.TimeFrom.IsZero() {
			from = w.TimeFrom.Format("15:04")
		}
		if !w.TimeTo.IsZero() {
			to = w.TimeTo.Format("15:04")
		}
		wishesData = append(wishesData, []string{strconv.Itoa(w.TeamID), t.Name, from, to, divsMap[t.DivisionID].Name})
	}

	gamesData := make([][]string, 0, len(games))
	for _, g := range games {
		team1, ok1 := teamsByID[g.TeamID1]
		team2, ok2 := teamsByID[g.TeamID2]
		if !ok1 || !ok2 {
			continue
		}
		checked := ""
		if g.CanRematch == 1 {
			checked = "checked"
		}
		gamesData = append(gamesData, []string{g.Tour, strconv.Itoa(g.TeamID1), team1.Name, strconv.Itoa(g.TeamID2), team2.Name, checked,
			divsMap[team1.DivisionID].Name, divsMap[team2.DivisionID].Name})
	}

	body := tt.renderTemplate(mainTmpl, map[string]interface{}{
		"teamsByDiv": teamsByDiv,
		"stads":      stads,
		"wishesData": wishesData,
		"gamesData":  gamesData,
	})

	c.HTML(http.StatusOK, "tmpl.html", gin.H{
		"title":    "Конструктор турниров",
		"subtitle": "Сформировать тур",
		"errors":   errs,
		"body":     template.HTML(body),
		"page":     "index",
	})
}

func sortTeamsByDiv(teamsByDiv map[string][]ds.Team) {
	for d := range teamsByDiv {
		sort.Slice(teamsByDiv[d], func(i, j int) bool {
			return teamsByDiv[d][i].Name < teamsByDiv[d][j].Name
		})
	}
}
