package ds

import "time"

type Wish struct {
	ID       int
	TeamID   int
	TimeFrom time.Time
	TimeTo   time.Time
}

func (w *Wish) IsSlotOk(fieldStart time.Time, gameDur time.Duration, slotIdx int) bool {
	from, to := fieldStart.Add(gameDur*time.Duration(slotIdx)), fieldStart.Add(gameDur*time.Duration(slotIdx+1))
	if !w.TimeFrom.IsZero() && w.TimeFrom.After(from) {
		return false
	}
	if !w.TimeTo.IsZero() && w.TimeTo.Before(to) {
		return false
	}
	return true
}
