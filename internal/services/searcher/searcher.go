package searcher

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/sergrom/timetable/internal/ds"
)

const (
	StatusInit      = 1
	StatusInProcess = 2
	StatusStopped   = 3
)

var (
	SearchCounter int = 1
)

type Searcher struct {
	lock        sync.RWMutex
	status      int
	Condition   *Condition
	tree        *searchTree
	nextNodeIdx int
	nextNode    *node

	solutions []Solution
	attempts  int
	solHashes map[string]struct{}
	stop      chan struct{}
	now       time.Time
}

type tInterval struct {
	from, to time.Time
}

func NewSearcher() *Searcher {
	return &Searcher{
		status:    StatusInit,
		tree:      nil,
		solutions: make([]Solution, 0, 2000),
		solHashes: make(map[string]struct{}),
		stop:      make(chan struct{}),
		now:       time.Now(),
	}
}

func (s *Searcher) Mem() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func (s *Searcher) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.status == StatusStopped || s.status == StatusInit {
		return
	}

	close(s.stop)
	time.Sleep(time.Second)
	s.status = StatusStopped
	s.tree = nil
	runtime.GC()
}

func (s *Searcher) Search(cond *Condition) ([]Solution, int, error) {
	s.lock.Lock()
	if s.status == StatusInProcess {
		return []Solution{}, 0, errors.New("Searcher::Search() error: searcher must be 'init' status")
	}

	SearchCounter++
	s.stop = make(chan struct{})
	s.status = StatusInProcess
	s.Condition = cond
	s.solutions = make([]Solution, 0, 2000)
	s.attempts = 0
	if s.tree != nil {
		s.tree.empty()
	}
	s.tree = newSearchTree()
	s.nextNode = nil
	runtime.GC()

	firstNodes, err := s.genFirstNodes()
	if err != nil {
		return nil, 0, err
	}
	s.tree.setFirstNodes(firstNodes)

	if len(s.tree.firstNodes) == 0 {
		return []Solution{}, 0, errors.New("couldn't generate first nodes")
	}

	// Устанавливаем указатель nextNode на первый элемент
	s.nextNode = s.tree.firstNodes[0]

	go s.rotateNextNode()

	s.lock.Unlock()

	go func() {
		defer func() {
			s.lock.Lock()
			s.status = StatusStopped
			if s.tree != nil {
				s.tree.empty()
			}
			runtime.GC()
			s.lock.Unlock()
		}()

		isSingleCPU := runtime.NumCPU() == 1
		for {
			select {
			case <-s.stop:
				return
			default:
				if s.Mem() > 14*1024*1024*1024 {
					fmt.Println("break on max memory exceeded")
					return
				}

				n := s.getNextNode()
				if n == nil {
					fmt.Println("search finished")
					return
				}

				s.drillNode(n)

				if isSingleCPU {
					time.Sleep(50 * time.Millisecond)
				}
			}
		}
	}()

	time.Sleep(2 * time.Second)

	sol, att := s.GetSolutions()

	return sol, att, nil
}

func (s *Searcher) drillNode(theNode *node) {
	teamsCnt := len(s.Condition.Teams)

	curNode := theNode
	for curNode.depth <= teamsCnt {
		if curNode.next == nil {
			curNode.next = s.genNodes(curNode)
		}
		if curNode.nextIdx >= len(curNode.next) {
			break
		}
		curNode = curNode.next[curNode.nextIdx]
	}

	if curNode.depth >= teamsCnt {
		s.addSolution(curNode)
	}

	s.lock.Lock()
	s.attempts++
	s.lock.Unlock()

	curNode = curNode.parent
	for curNode != nil {
		curNode.next[curNode.nextIdx] = nil
		curNode.nextIdx++
		if curNode.nextIdx < len(curNode.next) {
			break
		}
		curNode = curNode.parent
	}
}

// genFirstNodes генерировать первые ноды
func (s *Searcher) genFirstNodes() ([]*node, error) {
	// для каждой команды, для каждого поля содержит слоты, на которых возможна игра
	places := make(map[int]map[*ds.FieldNode][]int, len(s.Condition.Teams))
	for tID, team := range s.Condition.teamsByIDs {
		div := s.Condition.divMap[team.DivisionID]
		slots := s.Condition.teamSlots[tID]
		places[tID] = make(map[*ds.FieldNode][]int, len(slots))
		for _, fNode := range s.Condition.fieldNodes {
			if div.Format < 7 && fNode.Format() == 7 {
				continue
			}
			if div.Format == 7 && fNode.Format() != 7 {
				continue
			}
			for _, slot := range slots {
				if fNode.IsSlotOk(slot) {
					places[tID][fNode] = append(places[tID][fNode], slot)
				}
			}
		}
	}

	// choose the most specific team
	var theTeam *ds.Team
	theNodeTeamPairSlots := make(map[*ds.FieldNode]map[*ds.TeamPair][]int, len(s.Condition.fieldNodes))

	curCnt := 0
	for tID, fieldSlots := range places {
		slotsCnt := 0
		fnKeys := make(map[string]bool, len(s.Condition.fieldNodes))
		for fNode, slots := range fieldSlots {
			key := fNodeKey(fNode)
			if _, ok := fnKeys[key]; !ok {
				fnKeys[key] = true
				slotsCnt += len(slots)
			}
		}
		if slotsCnt == 0 {
			team := s.Condition.teamsByIDs[tID]
			div := s.Condition.divMap[team.DivisionID]
			return nil, fmt.Errorf("Невозможно разместить команду %s %s", team.Name, div.Name)
		}
		if theTeam == nil || curCnt > slotsCnt {
			theTeam, curCnt = s.Condition.teamsByIDs[tID], slotsCnt
		}
	}

	// Собираем все возможные пары с этой командой
	teamPairs := make([]*ds.TeamPair, 0, len(s.Condition.teamsByIDs))
	for _, pair := range s.Condition.teamPairsMap[theTeam.DivisionID] {
		if theTeam.ID == pair.Team1.ID || theTeam.ID == pair.Team2.ID {
			teamPairs = append(teamPairs, pair)
		}
	}

	fnKeys := make(map[string]bool, len(s.Condition.fieldNodes))
	nodesLen := 0
	for _, fNode := range s.Condition.fieldNodes {
		key := fNodeKey(fNode)
		if _, ok := fnKeys[key]; ok {
			continue
		}
		fnKeys[key] = true

		theNodeTeamPairSlots[fNode] = make(map[*ds.TeamPair][]int, len(teamPairs))
		for _, pair := range teamPairs {
			slots1, ok1 := places[pair.Team1.ID][fNode]
			slots2, ok2 := places[pair.Team2.ID][fNode]
			if !ok1 || !ok2 {
				continue
			}

			slots := IntersectSlots(slots1, slots2)
			theNodeTeamPairSlots[fNode][pair] = slots
			nodesLen += len(slots)
		}
	}

	fieldStart, gameDur := s.Condition.Fields[0].TimeFrom, s.Condition.Fields[0].GameDur
	firstNodes := make([]*node, 0, nodesLen)
	for fNode, pairSlots := range theNodeTeamPairSlots {
		for pair, slots := range pairSlots {
			for _, slot := range slots {
				from, to := GetFromTo(fieldStart, gameDur, slot)
				firstNodes = append(firstNodes, &node{
					field:    fNode,
					teamPair: pair,
					parent:   nil,
					depth:    1,
					slot:     slot,
					timeFrom: from,
					timeTo:   to,
				})
			}
		}
	}

	sort.Slice(firstNodes, func(i, j int) bool {
		return firstNodes[i].timeFrom.Before(firstNodes[j].timeFrom)
	})

	theDiv := s.Condition.divMap[theTeam.DivisionID]
	fmt.Println("The most specific team: ", theTeam.Name, theDiv.Name)

	return firstNodes, nil
}

func resolveTeamSlots(teamSlots []int, teamPrevSlots []int, coachPrevSlots []int, coachTeamsCnt int) []int {
	if len(teamPrevSlots) > 0 {
		slot := teamPrevSlots[0]
		teamSlots = IntersectSlots(teamSlots, []int{slot - 2, slot - 1, slot + 1, slot + 2})
	}

	teamSlotsMap := make(map[int]bool, len(teamSlots)+4)
	for _, slot := range teamSlots {
		if slot >= 0 {
			teamSlotsMap[slot] = true
		}
	}

	if len(coachPrevSlots) > 0 {
		coachSlotsCnt := coachTeamsCnt * 3
		slotMin, slotMax := 100000, -100000
		for _, slot := range coachPrevSlots {
			if slot < slotMin {
				slotMin = slot
			}
			if slot > slotMax {
				slotMax = slot
			}
		}
		diff := coachSlotsCnt - (slotMax - slotMin)
		if diff < 0 {
			return []int{}
		}

		for slot := range teamSlotsMap {
			if slot <= slotMin-diff || slot >= slotMax+diff {
				delete(teamSlotsMap, slot)
			}
		}
		for _, slot := range coachPrevSlots {
			delete(teamSlotsMap, slot)
		}
	}

	out := make([]int, 0, len(teamSlotsMap))
	for slot := range teamSlotsMap {
		out = append(out, slot)
	}

	sort.Ints(out)

	return out
}

func (s *Searcher) genNodes(theNode *node) []*node {
	teamGames := make(map[int]map[int]bool)
	teamGamesCnt := make(map[int]int)
	teamPrevSlots := make(map[int][]int)
	coachPrevSlots := make(map[int][]int)
	coachTeamsCnt := make(map[int]int)
	fieldsPrevSlots := make(map[int]map[int]bool)
	fieldSlotsMap := make(map[int][]int)
	prevPairsByDiv := make(map[int]map[*ds.TeamPair]bool)

	curNode := theNode
	for curNode != nil {
		id1, id2 := curNode.teamPair.Team1.ID, curNode.teamPair.Team2.ID
		team1, team2 := s.Condition.teamsByIDs[id1], s.Condition.teamsByIDs[id2]

		teamGames[id1] = map[int]bool{id2: true}
		teamGamesCnt[id1]++
		teamGamesCnt[id2]++
		teamPrevSlots[id1] = append(teamPrevSlots[id1], curNode.slot)
		teamPrevSlots[id2] = append(teamPrevSlots[id2], curNode.slot)

		coachPrevSlots[team1.CoachID] = append(coachPrevSlots[team1.CoachID], curNode.slot)
		coachPrevSlots[team2.CoachID] = append(coachPrevSlots[team2.CoachID], curNode.slot)

		if fieldsPrevSlots[curNode.field.Field1.ID] == nil {
			fieldsPrevSlots[curNode.field.Field1.ID] = make(map[int]bool)
		}
		fieldsPrevSlots[curNode.field.Field1.ID][curNode.slot] = true

		if curNode.field.Field2 != nil {
			if fieldsPrevSlots[curNode.field.Field2.ID] == nil {
				fieldsPrevSlots[curNode.field.Field2.ID] = make(map[int]bool)
			}
			fieldsPrevSlots[curNode.field.Field2.ID][curNode.slot] = true
		}

		divID := team1.DivisionID
		if _, ok := prevPairsByDiv[divID]; !ok {
			prevPairsByDiv[divID] = make(map[*ds.TeamPair]bool)
		}
		prevPairsByDiv[divID][curNode.teamPair] = true

		curNode = curNode.parent
	}

	for _, team := range s.Condition.teamsByIDs {
		coachTeamsCnt[team.CoachID]++
	}

	for i := range s.Condition.Fields {
		field := s.Condition.Fields[i]
		fID := field.ID
		prevSlots := fieldsPrevSlots[fID]
		for _, slot := range field.GetSlots() {
			if !prevSlots[slot] {
				fieldSlotsMap[fID] = append(fieldSlotsMap[fID], slot)
			}
		}
		fieldSlotsMap[fID] = UniqueSlots(fieldSlotsMap[fID])
	}

	teamNodeSlotsMap := make(map[int]map[*ds.FieldNode][]int)
	for tID, team := range s.Condition.teamsByIDs {
		if teamGamesCnt[tID] == 2 {
			continue // команда сыграла обе игры
		}

		div := s.Condition.divMap[team.DivisionID]
		cID := team.CoachID

		teamSlots := resolveTeamSlots(s.Condition.teamSlots[tID], teamPrevSlots[tID], coachPrevSlots[cID], coachTeamsCnt[cID])
		for _, fNode := range s.Condition.fieldNodes {
			if fNode.Format() == 7 && div.Format < 7 || div.Format > fNode.Format() {
				continue
			}

			fieldSlots := fieldSlotsMap[fNode.Field1.ID]
			if fNode.Field2 != nil {
				fieldSlots = IntersectSlots(fieldSlots, fieldSlotsMap[fNode.Field2.ID])
			}

			availableSlots := IntersectSlots(teamSlots, fieldSlots)
			if _, ok := teamNodeSlotsMap[tID]; !ok {
				teamNodeSlotsMap[tID] = make(map[*ds.FieldNode][]int)
			}
			teamNodeSlotsMap[tID][fNode] = availableSlots
		}
	}
	if len(teamNodeSlotsMap) == 0 {
		return nil
	}

	teamSlotsCnt := make(map[int]int, len(teamNodeSlotsMap))
	for tID, nodeSlots := range teamNodeSlotsMap {
		for _, slots := range nodeSlots {
			teamSlotsCnt[tID] += len(slots)
		}
		if teamSlotsCnt[tID] == 0 {
			return nil
		}
	}

	var theMostProblemTeam *ds.Team
	minCnt := 1000000
	for tID, cnt := range teamSlotsCnt {
		if theMostProblemTeam == nil || minCnt > cnt {
			theMostProblemTeam = s.Condition.teamsByIDs[tID]
			minCnt = cnt
		}
	}

	if minCnt == 0 {
		return nil
	}

	nodes := make([]*node, 0, len(s.Condition.teamPairsMap)*3)
	fieldStart, gameDur := s.Condition.Fields[0].TimeFrom, s.Condition.Fields[0].GameDur

	for _, pairs := range s.Condition.teamPairsMap {
		for _, pair := range pairs {
			if pair.Team1.ID != theMostProblemTeam.ID && pair.Team2.ID != theMostProblemTeam.ID {
				continue
			}
			if teamGamesCnt[pair.Team1.ID] == 2 || teamGamesCnt[pair.Team2.ID] == 2 {
				continue // оманда сыграла обе игры
			}
			if teamGames[pair.Team1.ID][pair.Team2.ID] || teamGames[pair.Team2.ID][pair.Team1.ID] {
				continue
			}
			if !s.checkRestDivTeams(prevPairsByDiv, pair) {
				continue
			}

			for _, fNode := range s.Condition.fieldNodes {
				slots1 := teamNodeSlotsMap[pair.Team1.ID][fNode]
				slots2 := teamNodeSlotsMap[pair.Team2.ID][fNode]
				pairSlots := IntersectSlots(slots1, slots2)

				fieldSlots := fieldSlotsMap[fNode.Field1.ID]
				if fNode.Field2 != nil {
					fieldSlots = IntersectSlots(fieldSlots, fieldSlotsMap[fNode.Field2.ID])
				}

				availableSlots := IntersectSlots(pairSlots, fieldSlots)
				if len(availableSlots) == 0 {
					continue
				}

				for _, slot := range availableSlots {
					// количество размещений по команде

					valueSum :=
						calcTeamSumValue(teamPrevSlots, pair, slot) +
							calcCoachSumValue(coachPrevSlots, pair, slot) +
							calcFieldSumValue(fieldsPrevSlots, slot)

					from, to := GetFromTo(fieldStart, gameDur, slot)
					nodes = append(nodes, &node{
						value:    0,
						valueSum: valueSum,
						field:    fNode,
						teamPair: pair,
						slot:     slot,
						parent:   theNode,
						next:     nil,
						nextIdx:  0,
						depth:    theNode.depth + 1,
						priority: 0,
						timeFrom: from,
						timeTo:   to,
					})
				}
			}
		}
	}

	if len(nodes) == 0 {
		return nil
	}

	if theNode.depth >= len(s.Condition.Teams)-1 {
		tPrevSum := calcTeamPrevSum(teamPrevSlots)
		cMinMaxCnt := calcCoachMinMaxCnt(coachPrevSlots)
		for _, node := range nodes {
			node.score = tPrevSum + calcSumValue(node, teamPrevSlots, cMinMaxCnt)
		}
	}

	// sort nodes by priority
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].valueSum < nodes[j].valueSum
	})

	return limitNodes(nodes, theNode.depth)
}

func (s *Searcher) checkRestDivTeams(prevPairsByDiv map[int]map[*ds.TeamPair]bool, pair *ds.TeamPair) bool {
	gamesRest := make(map[int]int, len(s.Condition.teamsByDivs[pair.Team1.DivisionID]))
	for _, team := range s.Condition.teamsByDivs[pair.Team1.DivisionID] {
		gamesRest[team.ID] = 2
	}

	for p := range prevPairsByDiv[pair.Team1.DivisionID] {
		gamesRest[p.Team1.ID]--
		gamesRest[p.Team2.ID]--
	}
	gamesRest[pair.Team1.ID]--
	gamesRest[pair.Team2.ID]--

	for tID, cnt := range gamesRest {
		cntRest := 0
		for _, p := range s.Condition.teamPairsByTeamMap[tID] {
			if gamesRest[p.Team1.ID] > 0 && gamesRest[p.Team2.ID] > 0 {
				cntRest++
			}
		}
		if cnt > cntRest {
			return false
		}
	}

	return true
}

func limitNodes(nodes []*node, depth int) []*node {
	cnt := len(nodes)
	switch {
	case depth > 30 && cnt >= 30:
		return nodes[:2]
	case depth > 20 && cnt >= 20:
		return nodes[:3]
	case depth > 10 && cnt >= 10:
		return nodes[:4]
	case depth > 5 && cnt >= 5:
		return nodes[:5]
	case cnt >= 10:
		return nodes[:10]
	}
	return nodes
}

func calcTeamSumValue(teamPrevSlots map[int][]int, newPair *ds.TeamPair, newSlot int) int {
	val := 0

	for tID, slots := range teamPrevSlots {
		if len(slots) == 2 {
			val += abs(slots[0]-slots[1]) - 1

		} else if len(slots) == 1 {
			if newPair.Team1.ID == tID || newPair.Team2.ID == tID {
				val += abs(slots[0]-newSlot) - 1
			}
		}
	}
	return val
}

func calcSumValue(curNode *node, teamPrevSlots map[int][]int, cMinMaxCnt map[int][3]int) int {
	sum := 0

	if slots := teamPrevSlots[curNode.teamPair.Team1.ID]; len(slots) > 0 {
		sum += abs(slots[0]-curNode.slot) - 1
	}
	if slots := teamPrevSlots[curNode.teamPair.Team2.ID]; len(slots) > 0 {
		sum += abs(slots[0]-curNode.slot) - 1
	}

	for _, cID := range []int{curNode.teamPair.Team1.CoachID, curNode.teamPair.Team2.CoachID} {
		if mmc, ok := cMinMaxCnt[cID]; ok {
			min, max, cnt := mmc[0], mmc[1], mmc[2]
			if curNode.slot < min {
				min = curNode.slot
			}
			if curNode.slot > max {
				max = curNode.slot
			}
			cnt++
			sum += cnt - (max - min + 1)
		}
	}

	return sum
}

func calcTeamPrevSum(teamPrevSlots map[int][]int) int {
	sum := 0
	for _, slots := range teamPrevSlots {
		if len(slots) == 2 {
			sum += abs(slots[0]-slots[1]) - 1
		}
	}
	return sum
}

func calcCoachMinMaxCnt(coachPrevSlots map[int][]int) map[int][3]int {
	minMax := make(map[int][3]int, 30)
	for cID, slots := range coachPrevSlots {
		min, max := 10000, -10000
		for _, slot := range slots {
			if min > slot {
				min = slot
			}
			if max < slot {
				max = slot
			}
		}
		minMax[cID] = [3]int{min, max, len(slots)}
	}
	return minMax
}

func calcCoachSumValue(coachPrevSlots map[int][]int, newPair *ds.TeamPair, newSlot int) int {
	val := 0
	coachSlots := make(map[int][]int)
	for cID, slots := range coachPrevSlots {
		coachSlots[cID] = append(coachSlots[cID], slots...)
	}

	c1, c2 := newPair.Team1.CoachID, newPair.Team2.CoachID
	coachSlots[c1] = append(coachSlots[c1], newSlot)
	coachSlots[c2] = append(coachSlots[c2], newSlot)

	for _, slots := range coachSlots {
		min, max := 10000, -10000
		for _, slot := range slots {
			if max < slot {
				max = slot
			}
			if min > slot {
				min = slot
			}
		}
		val += len(slots) - (max - min)

	}
	return val
}

func calcFieldSumValue(fieldsPrevSlots map[int]map[int]bool, slot int) int {
	val := 0
	for _, slots := range fieldsPrevSlots {
		max := -10000
		for slot := range slots {
			if max < slot {
				max = slot
			}
		}
		val += max
	}
	return val / len(fieldsPrevSlots)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func fNodeKey(fNode *ds.FieldNode) string {
	return fmt.Sprintf("%d_%s_%s", fNode.Format(), fNode.GetTimeFrom().Format("15:04"), fNode.GetTimeTo().Format("15:04"))
}

func (s *Searcher) getNextNode() *node {
	// rotate nextNode pointer
	// todo ...

	return s.nextNode
}

func (s *Searcher) GetSolutions() ([]Solution, int) {
	s.lock.RLock()
	sol := s.solutions
	att := s.attempts
	s.lock.RUnlock()

	sort.Slice(sol, func(i, j int) bool {
		return sol[i].Sum < sol[j].Sum
	})

	return sol, att
}

func (s *Searcher) Status() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.status
}

func (s *Searcher) StatusLabel() string {
	return map[int]string{
		StatusInit:      "init",
		StatusInProcess: "process",
		StatusStopped:   "stopped",
	}[s.Status()]
}

func (s *Searcher) SolutionsCnt() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.status == StateInit {
		return 0
	}

	return len(s.solutions)
}

func (s *Searcher) AttemptsCnt() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.status == StateInit {
		return 0
	}

	return s.attempts
}

func (s *Searcher) addSolution(theNode *node) {
	games := make(map[string][]SolutioGame, len(s.Condition.Fields))

	for cur := theNode; cur != nil; cur = cur.parent {
		field := cur.field.String()
		games[field] = append(games[field], SolutioGame{
			TeamID1: cur.teamPair.Team1.ID,
			TeamID2: cur.teamPair.Team2.ID,
			Start:   cur.timeFrom,
			End:     cur.timeTo,
		})
	}

	for fld := range games {
		sort.Slice(games[fld], func(i, j int) bool {
			return games[fld][i].Start.Before(games[fld][j].Start)
		})
	}

	sl := Solution{
		Sum:   theNode.score,
		Games: games,
	}

	sl.HashStr = fmt.Sprintf("%d_%s", SearchCounter, sl.Hash())

	if _, ok := s.solHashes[sl.HashStr]; !ok {
		s.lock.Lock()
		s.solHashes[sl.HashStr] = struct{}{}
		s.solutions = append(s.solutions, sl)
		s.lock.Unlock()
	}
}

func (s *Searcher) rotateNextNode() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		default:
		}
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.lock.Lock()
			fmt.Println("switch node", s.nextNodeIdx)
			s.nextNodeIdx++
			idx, nLen := s.nextNodeIdx, len(s.tree.firstNodes)
			idx = idx % nLen
			s.nextNodeIdx = idx
			s.nextNode = s.tree.firstNodes[idx]
			s.lock.Unlock()
		}
	}
}
