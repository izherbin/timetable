package repository

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/sergrom/timetable/internal/ds"
	"github.com/sergrom/timetable/internal/pkg"
)

func (r *Repo) getStadium(row []string) (ds.Stadium, error) {
	idStr, sName, fieldsStr, formatStr, timeFromStr, timeToStr, gameDurStr :=
		strings.TrimSpace(row[0]), strings.TrimSpace(row[1]), strings.TrimSpace(row[2]), strings.TrimSpace(row[3]), strings.TrimSpace(row[4]), strings.TrimSpace(row[5]), strings.TrimSpace(row[6])

	if idStr == "" || sName == "" || fieldsStr == "" || formatStr == "" || timeFromStr == "" || timeToStr == "" || gameDurStr == "" {
		return ds.Stadium{}, errors.New("empty col")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ds.Stadium{}, errors.New("id is not integer")
	}
	fields, err := strconv.Atoi(fieldsStr)
	if err != nil {
		return ds.Stadium{}, errors.New("fields is not integer")
	}
	format, err := strconv.Atoi(formatStr)
	if err != nil {
		return ds.Stadium{}, errors.New("format is not integer")
	}
	timeFrom, err := pkg.ParseHM(timeFromStr)
	if err != nil {
		return ds.Stadium{}, errors.New("finvalid timeFrom")
	}
	timeTo, err := pkg.ParseHM(timeToStr)
	if err != nil {
		return ds.Stadium{}, errors.New("finvalid timeTo")
	}
	gameDur, err := strconv.Atoi(gameDurStr)
	if err != nil {
		return ds.Stadium{}, errors.New("gameDuration is not integer")
	}
	if gameDur <= 0 || gameDur > 150 {
		return ds.Stadium{}, errors.New("gameDuration must be (0;150]")
	}

	return ds.Stadium{
		ID:       id,
		Name:     sName,
		Fields:   fields,
		Format:   format,
		TimeFrom: timeFrom,
		TimeTo:   timeTo,
		GameDur:  time.Duration(gameDur) * time.Minute,
	}, nil
}

func (r *Repo) getTeam(row []string) (ds.Team, error) {
	idStr, sName, coachIDStr, divIDStr :=
		strings.TrimSpace(row[0]), strings.TrimSpace(row[1]), strings.TrimSpace(row[2]), strings.TrimSpace(row[3])

	if idStr == "" || sName == "" || coachIDStr == "" || divIDStr == "" {
		return ds.Team{}, errors.New("empty col")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ds.Team{}, errors.New("id is not integer")
	}
	coachID, err := strconv.Atoi(coachIDStr)
	if err != nil {
		return ds.Team{}, errors.New("coachID is not integer")
	}
	divID, err := strconv.Atoi(divIDStr)
	if err != nil {
		return ds.Team{}, errors.New("divisionID is not integer")
	}

	return ds.Team{
		ID:         id,
		Name:       sName,
		CoachID:    coachID,
		DivisionID: divID,
	}, nil
}

func (r *Repo) getWish(row []string) (ds.Wish, error) {
	idStr, tIdStr, timeFrom :=
		strings.TrimSpace(row[0]), strings.TrimSpace(row[1]), strings.TrimSpace(row[2])
	timeTo := ""
	if len(row) > 3 {
		timeTo = strings.TrimSpace(row[3])
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ds.Wish{}, errors.New("id is not integer")
	}

	tId, err := strconv.Atoi(tIdStr)
	if err != nil {
		return ds.Wish{}, errors.New("teamId is not integer")
	}

	if timeFrom == "" && timeTo == "" {
		return ds.Wish{}, errors.New("'timeFrom' and 'timeTo' cannot be both empty")
	}

	from, to := time.Time{}, time.Time{}

	if timeFrom != "" {
		var err error
		from, err = pkg.ParseHM(timeFrom)
		if err != nil {
			return ds.Wish{}, errors.New("finvalid timeFrom")
		}
	}
	if timeTo != "" {
		var err error
		to, err = pkg.ParseHM(timeTo)
		if err != nil {
			return ds.Wish{}, errors.New("finvalid timeTo")
		}
	}

	return ds.Wish{
		ID:       id,
		TeamID:   tId,
		TimeFrom: from,
		TimeTo:   to,
	}, nil
}

func (r *Repo) getGame(row []string) (ds.Game, error) {
	idStr, tourName, team1, team2, rematchStr :=
		strings.TrimSpace(row[0]), strings.TrimSpace(row[1]), strings.TrimSpace(row[2]), strings.TrimSpace(row[3]), strings.TrimSpace(row[4])

	gameID, err := strconv.Atoi(idStr)
	if err != nil {
		return ds.Game{}, errors.New("ID is not integer")
	}
	id1, err := strconv.Atoi(team1)
	if err != nil {
		return ds.Game{}, errors.New("teamID1 is not integer")
	}
	id2, err := strconv.Atoi(team2)
	if err != nil {
		return ds.Game{}, errors.New("teamID2 is not integer")
	}
	rematch, err := strconv.Atoi(rematchStr)
	if err != nil {
		return ds.Game{}, errors.New("rematch is not integer")
	}
	return ds.Game{
		ID:         gameID,
		Tour:       tourName,
		TeamID1:    id1,
		TeamID2:    id2,
		CanRematch: rematch,
	}, nil
}
