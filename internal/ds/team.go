package ds

import "fmt"

type Team struct {
	ID         int
	Name       string
	CoachID    int
	DivisionID int
}

type TeamPair struct {
	Team1 *Team
	Team2 *Team
}

func (tp TeamPair) String() string {
	if tp.Team1.ID < tp.Team2.ID {
		return fmt.Sprintf("[%s - %s]", tp.Team1.Name, tp.Team2.Name)
	}
	return fmt.Sprintf("[%s - %s]", tp.Team2.Name, tp.Team1.Name)
}
