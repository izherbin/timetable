package repository

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sergrom/timetable/internal/ds"
)

const (
	DataDir = "data"

	CoachesFile   = "Тренеры.xlsx"
	DivisionsFile = "Дивизионы.xlsx"
	GamesFile     = "Игры.xlsx"
	StadiumsFile  = "Стадионы.xlsx"
	TeamsFile     = "Команды.xlsx"
	WishesFile    = "Пожелания.xlsx"
	UploadedFile  = "uploaded.xlsx"
)

type Repo struct{}

// NewRepo ...
func NewRepo() *Repo {
	return &Repo{}
}

// GetDivisions ...
func (r *Repo) GetDivisions() ([]ds.Division, error) {
	data, err := r.readFile(filepath.Join(DataDir, DivisionsFile))
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("empty Division's list")
	}

	divisions := make([]ds.Division, 0, len(data))
	for _, row := range data {
		if len(row) < 3 || strings.ToLower(row[0]) == "id" {
			continue
		}
		idStr, dName, formatStr := strings.TrimSpace(row[0]), strings.TrimSpace(row[1]), strings.TrimSpace(row[2])
		if formatStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		format, err := strconv.Atoi(formatStr)
		if err != nil {
			continue
		}
		divisions = append(divisions, ds.Division{
			ID:     id,
			Name:   dName,
			Format: format,
		})
	}

	return divisions, nil
}

func (r *Repo) GetDivisionsMap() (map[int]ds.Division, error) {
	divs, err := r.GetDivisions()
	if err != nil {
		return nil, err
	}

	dMap := make(map[int]ds.Division, len(divs))
	for _, d := range divs {
		dMap[d.ID] = d
	}

	return dMap, nil
}

// GetCoaches ...
func (r *Repo) GetCoaches() ([]ds.Coach, error) {
	data, err := r.readFile(filepath.Join(DataDir, CoachesFile))
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("empty Coach's list")
	}

	coaches := make([]ds.Coach, 0, len(data))
	for _, row := range data {
		if len(row) < 2 || strings.ToLower(row[0]) == "id" {
			continue
		}
		idStr, cName := strings.TrimSpace(row[0]), strings.TrimSpace(row[1])
		if idStr == "" || cName == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		coaches = append(coaches, ds.Coach{
			ID:   id,
			Name: cName,
		})
	}

	return coaches, nil
}

func (r *Repo) GetCoachesMap() (map[int]ds.Coach, error) {
	coaches, err := r.GetCoaches()
	if err != nil {
		return nil, err
	}

	cMap := make(map[int]ds.Coach, len(coaches))
	for _, c := range coaches {
		cMap[c.ID] = c
	}

	return cMap, nil
}

// GetStadiums ...
func (r *Repo) GetStadiums() ([]ds.Stadium, error) {
	data, err := r.readFile(filepath.Join(DataDir, StadiumsFile))
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("empty Stadium's list")
	}

	stads := make([]ds.Stadium, 0, len(data))
	for _, row := range data {
		if len(row) < 7 || strings.ToLower(row[0]) == "id" {
			continue
		}
		st, err := r.getStadium(row)
		if err != nil {
			continue
		}
		stads = append(stads, st)
	}

	return stads, nil
}

func (r *Repo) GetTeams() ([]ds.Team, error) {
	data, err := r.readFile(filepath.Join(DataDir, TeamsFile))
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("empty Team's list")
	}

	teams := make([]ds.Team, 0, len(data))
	for _, row := range data {
		if len(row) < 4 || strings.ToLower(row[0]) == "id" {
			continue
		}
		t, err := r.getTeam(row)
		if err != nil {
			continue
		}
		teams = append(teams, t)
	}

	return teams, nil
}

func (r *Repo) GetTeamsMap() (map[int]ds.Team, error) {
	teams, err := r.GetTeams()
	if err != nil {
		return nil, err
	}

	tMap := make(map[int]ds.Team, len(teams))
	for _, t := range teams {
		tMap[t.ID] = t
	}

	return tMap, nil
}

func (r *Repo) GetWishes() ([]ds.Wish, error) {
	data, err := r.readFile(filepath.Join(DataDir, WishesFile))
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("empty Wish's list")
	}

	wishes := make([]ds.Wish, 0, len(data))
	for _, row := range data {
		if len(row) < 3 || strings.HasPrefix(strings.ToLower(row[0]), "id") {
			continue
		}
		w, err := r.getWish(row)
		if err != nil {
			fmt.Println(err)
			continue
		}
		wishes = append(wishes, w)
	}

	return wishes, nil
}

func (r *Repo) GetGames() ([]ds.Game, error) {
	data, err := r.readFile(filepath.Join(DataDir, GamesFile))
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("empty Game's list")
	}

	games := make([]ds.Game, 0, len(data))
	for _, row := range data {
		if len(row) < 5 || strings.HasPrefix(strings.ToLower(row[0]), "id") {
			continue
		}
		w, err := r.getGame(row)
		if err != nil {
			fmt.Println(err)
			continue
		}
		games = append(games, w)
	}

	return games, nil
}

// func (r *Repo) UplodStadiums() error {
// 	data, err := r.readFile(filepath.Join(DataDir, UploadedFile))
// 	if err != nil {
// 		return err
// 	}

// 	stadMap := make(map[string][]string)
// 	for i, row := range data {
// 		if i == 0 {
// 			continue
// 		}
// 		if len(row) < 6 {
// 			return fmt.Errorf("В строке %d недостаточно данных", i)
// 		}
// 		name := strings.ToLower(strings.TrimSpace(row[0]))
// 		if _, ok := stadMap[name]; ok {
// 			return fmt.Errorf("Название стадиона %s дублируется", row[0])
// 		}
// 		stadMap[name] = row
// 	}
// 	if len(stadMap) == 0 {
// 		return errors.New("Пустой список стадионов")
// 	}

// 	f := excelize.NewFile()
// 	defer func() {
// 		if err := f.Close(); err != nil {
// 			fmt.Println(err)
// 		}
// 	}()

// 	// Create a new sheet.
// 	index, _ := f.NewSheet("Sheet1")

// 	f.SetCellStr("Sheet1", "A1", "ID")
// 	f.SetCellStr("Sheet1", "B1", "Название")
// 	f.SetCellStr("Sheet1", "C1", "Полей")
// 	f.SetCellStr("Sheet1", "D1", "Формат")
// 	f.SetCellStr("Sheet1", "E1", "Работает с")
// 	f.SetCellStr("Sheet1", "F1", "Работает по")
// 	f.SetCellStr("Sheet1", "G1", "Игра (минут)")

// 	for i, row := range data {
// 		if i == 0 {
// 			continue
// 		}
// 		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+1), i)
// 		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+1), strings.TrimSpace(row[0]))
// 		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+1), strings.TrimSpace(row[1]))
// 		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+1), strings.TrimSpace(row[2]))
// 		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+1), strings.TrimSpace(row[3]))
// 		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", i+1), strings.TrimSpace(row[4]))
// 		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", i+1), strings.TrimSpace(row[5]))
// 	}

// 	// Set active sheet of the workbook.
// 	f.SetActiveSheet(index)

// 	err = f.SaveAs(filepath.Join(DataDir, StadiumsFile))
// 	if err != nil {
// 		return errors.New("Ошибка сохранения файла")
// 	}

// 	return nil
// }

// func (r *Repo) UplodDivisions() error {
// 	data, err := r.readFile(filepath.Join(DataDir, UploadedFile))
// 	if err != nil {
// 		return err
// 	}

// 	divsMap := make(map[string][]string)
// 	for i, row := range data {
// 		if i == 0 {
// 			continue
// 		}
// 		if len(row) < 2 {
// 			return fmt.Errorf("В строке %d недостаточно данных", i)
// 		}
// 		format, err := strconv.Atoi(row[1])
// 		if err != nil {
// 			return fmt.Errorf("В строке %d не распознается формат", i)
// 		}
// 		if format < 3 || format > 7 {
// 			return fmt.Errorf("В строке %d формат должен быть от 3 до 7", i)
// 		}
// 		name := strings.ToLower(strings.TrimSpace(row[0]))
// 		if _, ok := divsMap[name]; ok {
// 			return fmt.Errorf("Название дивизиона %s дублируется", row[0])
// 		}
// 		divsMap[name] = row
// 	}

// 	dbDivsMap, err := r.GetDivisionsMap()
// 	if err != nil {
// 		return err
// 	}
// 	dbTeams, err := r.GetTeams()
// 	if err != nil {
// 		return err
// 	}

// 	for _, dbTeam := range dbTeams {
// 		div := strings.ToLower(strings.TrimSpace(dbDivsMap[dbTeam.DivisionID].Name))
// 		if _, ok := divsMap[div]; !ok {
// 			return fmt.Errorf("В загружаемом файле отсутствует дивизион %s, но существуют команды с таким дивизионом", dbDivsMap[dbTeam.DivisionID].Name)
// 		}
// 	}

// 	divIdsMap := make(map[string]int, len(dbDivsMap))
// 	maxDivId := 0
// 	for _, div := range dbDivsMap {
// 		divIdsMap[strings.ToLower(strings.TrimSpace(div.Name))] = div.ID
// 		if maxDivId < div.ID {
// 			maxDivId = div.ID
// 		}
// 	}

// 	f := excelize.NewFile()
// 	defer func() {
// 		if err := f.Close(); err != nil {
// 			fmt.Println(err)
// 		}
// 	}()

// 	// Create a new sheet.
// 	index, _ := f.NewSheet("Sheet1")

// 	f.SetCellStr("Sheet1", "A1", "ID")
// 	f.SetCellStr("Sheet1", "B1", "Название")
// 	f.SetCellStr("Sheet1", "C1", "Формат")

// 	for i, row := range data {
// 		divName := strings.ToLower(strings.TrimSpace(row[0]))
// 		var divID int
// 		if _, ok := divIdsMap[divName]; ok {
// 			divID = divIdsMap[divName]
// 		} else {
// 			divID = maxDivId + 1
// 			maxDivId++
// 		}
// 		if i == 0 {
// 			continue
// 		}
// 		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+1), divID)
// 		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+1), strings.TrimSpace(row[0]))
// 		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+1), strings.TrimSpace(row[1]))
// 	}

// 	// Set active sheet of the workbook.
// 	f.SetActiveSheet(index)

// 	err = f.SaveAs(filepath.Join(DataDir, DivisionsFile))
// 	if err != nil {
// 		return errors.New("Ошибка сохранения файла")
// 	}

// 	return nil
// }

// func (r *Repo) UplodCoaches() error {
// 	data, err := r.readFile(filepath.Join(DataDir, UploadedFile))
// 	if err != nil {
// 		return err
// 	}

// 	coachMap := make(map[string]bool)
// 	for i, row := range data {
// 		if i == 0 {
// 			continue
// 		}
// 		if len(row) == 0 {
// 			continue
// 		}

// 		name := strings.ToLower(strings.TrimSpace(row[0]))
// 		if _, ok := coachMap[name]; ok {
// 			return fmt.Errorf("Строки дублируются: %s", row[0])
// 		}
// 		coachMap[name] = true
// 	}

// 	if len(coachMap) == 0 {
// 		return fmt.Errorf("Пустой список тренеров")
// 	}

// 	dbCoaches, err := r.GetCoaches()
// 	if err != nil {
// 		return err
// 	}

// 	dbCoachesMap := make(map[string]int)
// 	dbCoachesByIDMap := make(map[int]string)
// 	maxCoachID := 0
// 	for _, c := range dbCoaches {
// 		name := strings.ToLower(strings.TrimSpace(c.Name))
// 		dbCoachesMap[name] = c.ID
// 		dbCoachesByIDMap[c.ID] = name
// 		if maxCoachID < c.ID {
// 			maxCoachID = c.ID
// 		}
// 	}

// 	teams, err := r.GetTeams()
// 	if err != nil {
// 		return err
// 	}
// 	for _, t := range teams {
// 		cName := dbCoachesByIDMap[t.CoachID]
// 		if !coachMap[cName] {
// 			return fmt.Errorf("В загружаемом файле отсутствует тренер %s, но существуют команды с таким тренером", cName)
// 		}
// 	}

// 	f := excelize.NewFile()
// 	defer func() {
// 		if err := f.Close(); err != nil {
// 			fmt.Println(err)
// 		}
// 	}()

// 	// Create a new sheet.
// 	index, _ := f.NewSheet("Sheet1")

// 	f.SetCellStr("Sheet1", "A1", "ID")
// 	f.SetCellStr("Sheet1", "B1", "Имя")

// 	for i, row := range data {
// 		name := row[0]
// 		id := 0
// 		if idd, ok := dbCoachesMap[strings.ToLower(strings.TrimSpace(name))]; ok {
// 			id = idd
// 		} else {
// 			id = maxCoachID + 1
// 			maxCoachID++
// 		}
// 		if i == 0 {
// 			continue
// 		}
// 		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+1), id)
// 		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+1), name)
// 	}

// 	// Set active sheet of the workbook.
// 	f.SetActiveSheet(index)

// 	err = f.SaveAs(filepath.Join(DataDir, CoachesFile))
// 	if err != nil {
// 		return errors.New("Ошибка сохранения файла")
// 	}

// 	return nil
// }

// func (r *Repo) UplodTeams() error {
// 	data, err := r.readFile(filepath.Join(DataDir, UploadedFile))
// 	if err != nil {
// 		return err
// 	}

// 	teamMap := make(map[string]bool)
// 	teamDivMap := make(map[string]map[string]bool)
// 	for i, row := range data {
// 		if i == 0 {
// 			continue
// 		}
// 		if len(row) < 3 {
// 			return fmt.Errorf("В строке %d недостаточно данных", i)
// 		}

// 		name := strings.ToLower(strings.TrimSpace(row[0]))
// 		div := strings.ToLower(strings.TrimSpace(row[1]))

// 		if _, ok := teamDivMap[name]; !ok {
// 			teamDivMap[name] = make(map[string]bool)
// 		}
// 		teamDivMap[name][div] = true

// 		if _, ok := teamMap[name+div]; ok {
// 			return fmt.Errorf("Строки дублируются: %s %s", name, div)
// 		}
// 		teamMap[name+div] = true
// 	}

// 	if len(teamMap) == 0 {
// 		return fmt.Errorf("Пустой список команд")
// 	}

// 	dbTeams, err := r.GetTeams()
// 	if err != nil {
// 		return nil
// 	}

// 	dbTeamDivMap := make(map[string]ds.Team, len(dbTeams))
// 	maxTeamID := 0
// 	for _, t := range dbTeams {
// 		if t.ID > maxTeamID {
// 			maxTeamID = t.ID
// 		}
// 		teamdiv := strings.ToLower(strings.TrimSpace(t.Name + di))
// 	}

// 	// ---

// 	f := excelize.NewFile()
// 	defer func() {
// 		if err := f.Close(); err != nil {
// 			fmt.Println(err)
// 		}
// 	}()

// 	// Create a new sheet.
// 	index, _ := f.NewSheet("Sheet1")

// 	f.SetCellStr("Sheet1", "A1", "ID")
// 	f.SetCellStr("Sheet1", "B1", "Название")
// 	f.SetCellStr("Sheet1", "C1", "ID тренера")
// 	f.SetCellStr("Sheet1", "D1", "ID дивизиона")

// 	for i, row := range data {
// 		name := row[0]
// 		div := row[1]
// 		coach := row[2]

// 		teamID := 0
// 		coachID := 0
// 		divID := 0

// 		teamdiv := strings.ToLower(strings.TrimSpace(name + div))
// 		if _, ok := dbTeamDivMap[teamdiv]; ok {
// 			teamID = dbTeamDivMap[teamdiv]
// 		} else {
// 			teamID = maxTeamID + 1
// 			maxTeamID++
// 		}

// 		// id := 0
// 		// if idd, ok := dbCoachesMap[strings.ToLower(strings.TrimSpace(name))]; ok {
// 		// 	id = idd
// 		// } else {
// 		// 	id = maxCoachID + 1
// 		// 	maxCoachID++
// 		// }

// 		if i == 0 {
// 			continue
// 		}
// 		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+1), teamID)
// 		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+1), name)
// 		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+1), coachID)
// 		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+1), divID)
// 	}

// 	// Set active sheet of the workbook.
// 	f.SetActiveSheet(index)

// 	err = f.SaveAs(filepath.Join(DataDir, CoachesFile))
// 	if err != nil {
// 		return errors.New("Ошибка сохранения файла")
// 	}

// 	return nil
// }
