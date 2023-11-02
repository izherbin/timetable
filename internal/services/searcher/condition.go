package searcher

import (
	"fmt"
	"time"

	"github.com/sergrom/timetable/internal/ds"
)

type Condition struct {
	TourName       string
	Fields         []ds.Field
	Divisions      []ds.Division
	Coaches        []ds.Coach
	Teams          []ds.Team
	Wishes         []ds.Wish
	Games          []ds.Game
	TeamsPrettyMap map[int]string
	DayStart       time.Time
	DayEnd         time.Time

	teamsByDivs        map[int][]*ds.Team
	teamsByIDs         map[int]*ds.Team
	teamFormats        map[int]int
	divMap             map[int]*ds.Division
	coachMap           map[int]*ds.Coach
	wishMap            map[int]*ds.Wish
	teamSlots          map[int][]int
	teamPairsMap       map[int][]*ds.TeamPair
	teamPairsByTeamMap map[int][]*ds.TeamPair
	fieldNodes         []*ds.FieldNode
	coachGameCnt       map[int]int
	stadSlotsCnt       int
}

func NewCondition(tourName string, fields []ds.Field, divisions []ds.Division, coaches []ds.Coach, teams []ds.Team, wishes []ds.Wish, games []ds.Game) *Condition {
	cond := &Condition{
		TourName:  tourName,
		Fields:    fields,
		Divisions: divisions,
		Coaches:   coaches,
		Teams:     teams,
		Wishes:    wishes,
		Games:     games,
	}

	dayStart := fields[0].TimeFrom
	dayEnd := fields[0].TimeTo
	for i := range fields {
		t1 := fields[i].TimeFrom
		t2 := fields[i].TimeTo
		if t1.Before(dayStart) {
			dayStart = t1
		}
		if t2.After(dayEnd) {
			dayEnd = t2
		}
	}
	cond.DayStart = dayStart
	cond.DayEnd = dayEnd

	divMap := make(map[int]*ds.Division, len(cond.Divisions))
	for i := range cond.Divisions {
		divMap[cond.Divisions[i].ID] = &cond.Divisions[i]
	}

	teamsByIDs := make(map[int]*ds.Team)
	teamsByDivs := make(map[int][]*ds.Team)
	teamFormats := make(map[int]int)
	for i := range teams {
		teamsByIDs[teams[i].ID] = &teams[i]
		teamsByDivs[teams[i].DivisionID] = append(teamsByDivs[teams[i].DivisionID], &teams[i])
		teamFormats[teams[i].ID] = divMap[teams[i].DivisionID].Format
	}

	teamsPrettyMap := make(map[int]string)
	for tID, team := range teamsByIDs {
		div := divMap[team.DivisionID]
		teamsPrettyMap[tID] = fmt.Sprintf("%s (%s)", team.Name, div.Name)
	}

	rematchRestricts := make(map[int]map[int]string)
	for _, g := range games {
		id1, id2 := g.TeamID1, g.TeamID2
		if id1 > id2 {
			id1, id2 = id2, id1
		}
		if _, ok := rematchRestricts[id1]; !ok {
			rematchRestricts[id1] = make(map[int]string)
		}
		rematchRestricts[id1][id2] = g.Tour
	}

	// Собираем пары команд, которые могут между собой играть
	teamPairsMap := make(map[int][]*ds.TeamPair, len(teamsByDivs)) // Пары команд
	teamPairsByTeamMap := make(map[int][]*ds.TeamPair)
	coachGameCnt := make(map[int]int, len(coaches)) // Игры тренеров
	for divID, tt := range teamsByDivs {
		if _, ok := teamPairsMap[divID]; !ok {
			teamPairsMap[divID] = make([]*ds.TeamPair, 0, len(tt)*len(tt)/2+1)
		}

		for i1 := 0; i1 < len(tt); i1++ {
			for i2 := i1 + 1; i2 < len(tt); i2++ {
				if tt[i1].CoachID == tt[i2].CoachID {
					// команды одного тренера не могут играть между собой (уловие из тз)
					continue
				}

				if !canRematch(tt[i1].ID, tt[i2].ID, rematchRestricts) {
					// ранее была игра между этими командами и снова играть нельзя
					continue
				}

				// все ок, добавляем пару команд
				tp := &ds.TeamPair{Team1: tt[i1], Team2: tt[i2]}
				teamPairsMap[divID] = append(teamPairsMap[divID], tp)

				if _, ok := teamPairsByTeamMap[tt[i1].ID]; !ok {
					teamPairsByTeamMap[tt[i1].ID] = make([]*ds.TeamPair, 0, 10)
				}
				if _, ok := teamPairsByTeamMap[tt[i2].ID]; !ok {
					teamPairsByTeamMap[tt[i2].ID] = make([]*ds.TeamPair, 0, 10)
				}
				teamPairsByTeamMap[tt[i1].ID] = append(teamPairsByTeamMap[tt[i1].ID], tp)
				teamPairsByTeamMap[tt[i2].ID] = append(teamPairsByTeamMap[tt[i2].ID], tp)

				coachGameCnt[tt[i1].CoachID]++
				coachGameCnt[tt[i2].CoachID]++
			}
		}
	}

	// Собираем возможные поля и пары полей
	stadSlots := 0
	fieldNodes := make([]*ds.FieldNode, 0, len(cond.Fields))
	for i := range cond.Fields {
		fieldNodes = append(fieldNodes, &ds.FieldNode{Field1: &cond.Fields[i]})
		if cnt := int(cond.Fields[i].TimeTo.Sub(cond.Fields[i].TimeFrom).Minutes() / cond.Fields[i].GameDur.Minutes()); cnt > stadSlots {
			stadSlots = cnt
		}
	}
	has7 := false
	for i := range cond.Divisions {
		if cond.Divisions[i].Format == 7 && len(teamsByDivs[cond.Divisions[i].ID]) > 0 {
			has7 = true
			break
		}
	}
	if has7 {
		// если есть команды которые играют по формату 7, то добавляем пары полей форматов меньше 7
		if len(cond.Fields) > 1 {
			fieldNodes = append(fieldNodes, &ds.FieldNode{
				Field1: &cond.Fields[0],
				Field2: &cond.Fields[1],
			})
		}
		if len(cond.Fields) > 3 {
			fieldNodes = append(fieldNodes, &ds.FieldNode{
				Field1: &cond.Fields[2],
				Field2: &cond.Fields[3],
			})
		}
	}

	coachMap := make(map[int]*ds.Coach, len(cond.Coaches))
	for i := range cond.Coaches {
		coachMap[cond.Coaches[i].ID] = &cond.Coaches[i]
	}

	wishMap := make(map[int]*ds.Wish, len(cond.Wishes))
	for i := range cond.Wishes {
		wishMap[cond.Wishes[i].TeamID] = &cond.Wishes[i]
	}

	fieldStart, gameDur := cond.Fields[0].TimeFrom, cond.Fields[0].GameDur
	teamSlots := make(map[int][]int)
	for tID := range teamsByIDs {
		slots := make([]int, 0, stadSlots)
		w, ok := wishMap[tID]
		for i := 0; i < stadSlots; i++ {
			if !ok {
				slots = append(slots, i)
				continue
			}
			if w.IsSlotOk(fieldStart, gameDur, i) {
				slots = append(slots, i)
			}
		}
		teamSlots[tID] = slots
	}

	cond.teamsByDivs = teamsByDivs
	cond.teamsByIDs = teamsByIDs
	cond.teamFormats = teamFormats
	cond.TeamsPrettyMap = teamsPrettyMap
	cond.divMap = divMap
	cond.coachMap = coachMap
	cond.wishMap = wishMap
	cond.teamSlots = teamSlots
	cond.teamPairsMap = teamPairsMap
	cond.teamPairsByTeamMap = teamPairsByTeamMap
	cond.fieldNodes = fieldNodes
	cond.coachGameCnt = coachGameCnt
	cond.stadSlotsCnt = stadSlots

	return cond
}

func (c *Condition) Dump() {
	fmt.Printf("== Расчет матчей тура \"%s\" == \n", c.TourName)

	fmt.Printf("Стадион: %s (полей: %d):\n", "c.Stadium.Name todo", len(c.Fields))
	for _, f := range c.Fields {
		fmt.Printf("	%d: %s-%s (формат поля:%d, игра: %.0fmin)\n", f.ID, f.TimeFrom, f.TimeTo, f.Format, f.GameDur.Minutes())
	}

	fmt.Println()

	fmt.Printf("Команд: %dшт\n", len(c.Teams))
	fmt.Println("Команды по дивизионам:")
	for divID, tt := range c.teamsByDivs {
		fmt.Printf("Дивизион: %s (формат: %d)\n", c.divMap[divID].Name, c.divMap[divID].Format)
		for _, t := range tt {
			fmt.Printf("	%d: %s, тренер:%s\n", t.ID, t.Name, c.coachMap[t.CoachID].Name)
		}
	}

	fmt.Println()
	fmt.Println("Возможные пары команд по дивизионам:")
	for divID, pairs := range c.teamPairsMap {
		fmt.Printf("Дивизион: %s (формат: %d)\n", c.divMap[divID].Name, c.divMap[divID].Format)
		for _, p := range pairs {
			fmt.Printf("	%s\n", p)
		}
	}

	fmt.Println()
	fmt.Println("Возможные поля и пары полей (для формата 7):")
	for _, n := range c.fieldNodes {
		fmt.Printf("	%s\n", n)
	}
}

func (c *Condition) getTeamPairsMap() map[int][]*ds.TeamPair {
	return c.teamPairsMap
}

func (c *Condition) getFieldNodes() []*ds.FieldNode {
	return c.fieldNodes
}

func canRematch(teamID1, teamID2 int, rematchRestricts map[int]map[int]string) bool {
	id1, id2 := teamID1, teamID2
	if id1 > id2 {
		id1, id2 = id2, id1
	}

	if _, ok := rematchRestricts[id1][id2]; ok {
		return false
	}
	return true
}
