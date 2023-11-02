package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sergrom/timetable/internal/services/searcher"
	"github.com/xuri/excelize/v2"
)

func (tt *TimetableAPI) downloadSolution(c *gin.Context) {
	q := c.Request.URL.Query()
	hash := q.Get("hash")

	if len(hash) == 0 {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var theSolution *searcher.Solution

	sols, _ := tt.searcher.GetSolutions()
	for _, s := range sols {
		if s.HashStr == hash {
			theSolution = &s
			break
		}
	}

	if theSolution == nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	teamsMap, err := tt.repo.GetTeamsMap()
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	divsMap, err := tt.repo.GetDivisionsMap()
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
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
		fmt.Println(err)
		return
	}

	styleTopHead, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 20,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	styleSecondHead, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
	})
	styleWrap, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			WrapText: true,
		},
	})

	fields, preparedGames := prepareGames(theSolution.Games)

	fieldID := 1
	f.SetRowHeight("Sheet1", 1, 32.0)
	for _, field := range fields {
		games := preparedGames[field]
		cell, _ := excelize.CoordinatesToCellName(fieldID, 1)
		cellTo, _ := excelize.CoordinatesToCellName(fieldID+2, 1)
		f.SetCellValue("Sheet1", cell, strings.Trim(field, "[]"))
		f.SetCellStyle("Sheet1", cell, cell, styleTopHead)

		f.MergeCell("Sheet1", cell, cellTo)
		cn, _ := excelize.ColumnNumberToName(fieldID + 2)
		f.SetColWidth("Sheet1", cn, cn, 40)

		ch1, _ := excelize.CoordinatesToCellName(fieldID, 2)
		ch2, _ := excelize.CoordinatesToCellName(fieldID+1, 2)
		ch3, _ := excelize.CoordinatesToCellName(fieldID+2, 2)
		f.SetCellValue("Sheet1", ch1, "Время")
		f.SetCellValue("Sheet1", ch2, "Див.")
		f.SetCellValue("Sheet1", ch3, "Команды")
		f.SetCellStyle("Sheet1", ch1, ch3, styleSecondHead)

		rowID := 3
		for _, g := range games {
			if g.FixRowIdx > 0 {
				rowID = g.FixRowIdx + 3
			}

			teamsStr := fmt.Sprintf("%s\r\n%s", teamsMap[g.TeamID1].Name, teamsMap[g.TeamID2].Name)

			cell1, _ := excelize.CoordinatesToCellName(fieldID, rowID)
			f.SetCellValue("Sheet1", cell1, g.Start.Format("15:04"))
			cell2, _ := excelize.CoordinatesToCellName(fieldID+1, rowID)
			cell3, _ := excelize.CoordinatesToCellName(fieldID+2, rowID)

			if g.ExtraInfo == "" {
				f.SetCellValue("Sheet1", cell2, divsMap[teamsMap[g.TeamID1].DivisionID].Name)
				f.SetCellStr("Sheet1", cell3, teamsStr)
			} else {
				f.SetCellStr("Sheet1", cell3, g.ExtraInfo)
			}

			f.SetCellStyle("Sheet1", cell3, cell3, styleWrap)
			f.SetRowHeight("Sheet1", rowID, 32.0)
			rowID++
		}
		fieldID += 3
	}

	// Set active sheet of the workbook.
	f.SetActiveSheet(index)

	c.Header("Content-Disposition", "attachment; filename=Турнир.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Pragma", "public")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "must-revalidate")

	f.WriteTo(c.Writer)
}

func prepareGames(games map[string][]searcher.SolutioGame) ([]string, map[string][]searcher.SolutioGame) {
	singleFileds := make([]string, 0, len(games))
	doubleFields := make([]string, 0, len(games))
	for field := range games {
		if len(strings.Split(field, ",")) == 1 {
			singleFileds = append(singleFileds, field)
			continue
		}
		doubleFields = append(doubleFields, field)
	}

	output := make(map[string][]searcher.SolutioGame, len(singleFileds))
	for _, f := range singleFileds {
		games := games[f]
		for _, g := range games {
			output[f] = append(output[f], g)
		}
	}
	for _, f := range doubleFields {
		fields := strings.Split(f, ",")
		f1, f2 := strings.Trim(fields[0], "[]"), strings.Trim(fields[1], "[]")
		extraInfo := fmt.Sprintf("игра %s, %s", f1, f2)
		f1, f2 = "["+f1+"]", "["+f2+"]"
		games := games[f]
		for _, g := range games {
			output[f1] = append(output[f1], g)
			output[f2] = append(output[f2], searcher.SolutioGame{
				Start:     g.Start,
				ExtraInfo: extraInfo,
			})
		}
	}

	for f := range output {
		sort.Slice(output[f], func(i, j int) bool {
			return output[f][i].Start.Before(output[f][j].Start)
		})
	}
	sort.Slice(singleFileds, func(i, j int) bool {
		return singleFileds[i] < singleFileds[j]
	})

	timeCntMap := make(map[string]int, 500)
	idxTimeMap := make(map[string]int, 500)
	for f := range output {
		games := output[f]
		for i, g := range games {
			tStr := g.Start.Format("15:04")
			if _, ok := timeCntMap[tStr]; !ok {
				timeCntMap[tStr] = 0
				idxTimeMap[tStr] = 0
			}
			timeCntMap[tStr]++
			if idxTimeMap[tStr] < i {
				idxTimeMap[tStr] = i
			}
		}
	}

	for f := range output {
		games := output[f]
		for i := range games {
			tStr := games[i].Start.Format("15:04")
			if timeCntMap[tStr] > 1 {
				games[i].FixRowIdx = idxTimeMap[tStr]
			}
		}
	}

	return singleFileds, output
}
