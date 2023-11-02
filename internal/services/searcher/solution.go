package searcher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Solution struct {
	Sum     int                      `json:"sum"`
	Games   map[string][]SolutioGame `json:"games"`
	HashStr string                   `json:"hash"`
}

type SolutioGame struct {
	TeamID1   int       `json:"team_id_1"`
	TeamID2   int       `json:"team_id_2"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	ExtraInfo string
	FixRowIdx int
}

func (g SolutioGame) String() string {
	id1, id2 := g.TeamID1, g.TeamID2
	if id1 > id2 {
		id1, id2 = id2, id1
	}
	return fmt.Sprintf("%d_%d_%s_%s", id1, id1, g.Start.Format("15:04"), g.End.Format("15:04"))
}

func (s Solution) Hash() string {
	fields := make([]string, 0, len(s.Games))
	for field := range s.Games {
		fields = append(fields, field)
	}

	sort.Strings(fields)

	sb := strings.Builder{}
	for _, f := range fields {
		sb.WriteString(f + ",")
		for _, g := range s.Games[f] {
			sb.WriteString(g.String() + ",")
		}
	}

	hash256 := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash256[:])
}
