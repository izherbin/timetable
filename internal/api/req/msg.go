package req

type SearchStartRequest struct {
	TourName   string  `json:"tour_name"`
	StaduiumID int     `json:"stadium_id"`
	Fields     []Field `json:"fields"`
	Teams      []int   `json:"teams"`
	Wishes     []Wish  `json:"wishes"`
	Games      []Game  `json:"games"`
}

type Field struct {
	Format int    `json:"format"`
	From   string `json:"from"`
	To     string `json:"to"`
	Dur    int    `json:"dur"`
}

type Wish struct {
	TeamID int    `json:"team_id"`
	From   string `json:"from"`
	To     string `json:"to"`
}

type Game struct {
	TeamID1 int `json:"team_id_1"`
	TeamID2 int `json:"team_id_2"`
}

func (r SearchStartRequest) GetTemsMap() map[int]bool {
	m := make(map[int]bool, len(r.Teams))
	for _, t := range r.Teams {
		m[t] = true
	}
	return m
}

type CheckAccessRequest struct {
	Msg string `json:"msg"`
}

type SaveStadiumRequest struct {
	ID       string `json:"id"`
	Tag      string `json:"tag"`
	Name     string `json:"name"`
	Fields   string `json:"fields"`
	Format   string `json:"format"`
	TimeFrom string `json:"time_from"`
	TimeTo   string `json:"time_to"`
	GameDur  string `json:"game_dur"`
}

type SaveDivisionRequest struct {
	ID     string `json:"id"`
	Tag    string `json:"tag"`
	Name   string `json:"name"`
	Format string `json:"format"`
}

type SaveCoachRequest struct {
	ID   string `json:"id"`
	Tag  string `json:"tag"`
	Name string `json:"name"`
}

type SaveTeamRequest struct {
	ID         string `json:"id"`
	Tag        string `json:"tag"`
	Name       string `json:"name"`
	DivisionID string `json:"division_id"`
	CoachID    string `json:"coach_id"`
}

type SaveWishRequest struct {
	ID       string `json:"id"`
	TeamID   string `json:"team_id"`
	TimeFrom string `json:"time_from"`
	TimeTo   string `json:"time_to"`
}

type SaveGameRequest struct {
	ID         string `json:"id"`
	Tour       string `json:"tour"`
	TeamID1    string `json:"team_id_1"`
	TeamID2    string `json:"team_id_2"`
	CanRematch string `json:"can_rematch"`
}
