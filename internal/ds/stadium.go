package ds

import "time"

type Stadium struct {
	ID       int
	Name     string
	Fields   int
	Format   int
	TimeFrom time.Time
	TimeTo   time.Time
	GameDur  time.Duration
}
