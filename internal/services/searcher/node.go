package searcher

import (
	"time"

	"github.com/sergrom/timetable/internal/ds"
)

const (
	StateInit      = 1
	StateProcessed = 2
)

var nID = 0

func nodeGenID() int {
	nID++
	return nID
}

type node struct {
	id       				int
	value    				int // стоимость ноды
	valueSum 				int // стоимость ноды + стоимость родительской
	score    				int

	teamGames				map[int]map[int]bool
	teamGamesCnt		map[int]int
	teamPrevSlots		map[int][]int
	coachPrevSlots	map[int][]int
	coachTeamsCnt		map[int]int
	fieldsPrevSlots	map[int]map[int]bool
	prevPairsByDiv	map[int]map[*ds.TeamPair]bool	

	field    				*ds.FieldNode // поле или пара полей
	teamPair 				*ds.TeamPair  // пара команд
	slot     				int
	parent   				*node   // предыдущая нода
	next     				[]*node // следующие ноды
	nextIdx  				int
	depth    				int
	priority 				int

	timeFrom time.Time // начало игры
	timeTo   time.Time // конец игры
}
