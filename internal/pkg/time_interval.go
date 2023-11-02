package pkg

import (
	"fmt"
	"strings"
	"time"
)

// IvGroup ...
type IvGroup struct {
	// base      tInterval
	intervals []tInterval
}

type tInterval struct {
	from time.Time
	to   time.Time
}

// NewIvGroup ...
func NewIvGroup(baseFrom, baseTo time.Time) *IvGroup {
	ig := &IvGroup{
		intervals: make([]tInterval, 0, 10),
	}
	ig.intervals = append(ig.intervals, tInterval{
		from: baseFrom,
		to:   baseTo,
	})
	return ig
}

// String ...
func (ig *IvGroup) String() string {
	var sb strings.Builder
	for i := range ig.intervals {
		if ig.intervals[i].from.IsZero() && ig.intervals[i].to.IsZero() {
			continue
		}
		f, t := "I", "I"
		if !ig.intervals[i].from.IsZero() {
			f = ig.intervals[i].from.Format("15:04")
		}
		if !ig.intervals[i].to.IsZero() {
			t = ig.intervals[i].to.Format("15:04")
		}
		sb.WriteString(fmt.Sprintf("[%s-%s]", f, t))
	}
	return sb.String()
}

// ApplyWish применить пожелание
func (ig *IvGroup) ApplyWish(from, to time.Time) {
	for i := range ig.intervals {
		f := ig.intervals[i].from
		if !from.IsZero() && from.After(f.Add(100*time.Millisecond)) {
			f = from
		}
		t := ig.intervals[i].to
		if !to.IsZero() && to.Before(t.Add(-100*time.Millisecond)) {
			t = to
		}

		if f.After(t.Add(-100 * time.Millisecond)) {
			ig.intervals[i].from, ig.intervals[i].to = time.Time{}, time.Time{}
			continue
		}
		ig.intervals[i].from, ig.intervals[i].to = f, t
	}
}

// Exclude исключить интервал
func (ig *IvGroup) Exclude(from, to time.Time) {
	var newIntervals []tInterval
	for i := range ig.intervals {
		if ig.intervals[i].from.Add(100*time.Millisecond).After(to) || ig.intervals[i].to.Add(-100*time.Millisecond).Before(from) {
			continue
		}
		if from.After(ig.intervals[i].from.Add(-100*time.Millisecond)) && to.Before(ig.intervals[i].to.Add(100*time.Millisecond)) {
			// выкалываем из середины
			newIntervals = []tInterval{
				{from: ig.intervals[i].from, to: from},
				{from: to, to: ig.intervals[i].to},
			}
			ig.intervals[i].from, ig.intervals[i].to = time.Time{}, time.Time{}
			continue
		}
		if from.Before(ig.intervals[i].from.Add(100 * time.Millisecond)) {
			ig.intervals[i].from = to
		} else {
			ig.intervals[i].to = from
		}
	}
	if len(newIntervals) > 0 {
		ig.intervals = append(ig.intervals, newIntervals...)
	}
}

// PlaceDurationNear разместить промежуток времени dur наиболее близко к точке point
func (ig *IvGroup) PlaceDurationNear(dur time.Duration, point time.Time) (time.Time, int) {
	bestPlace := time.Time{}
	dist := 1000 * time.Hour

	for i := range ig.intervals {
		if ig.intervals[i].from.IsZero() || ig.intervals[i].to.IsZero() || ig.intervals[i].to.Sub(ig.intervals[i].from) < dur {
			continue
		}
		if point.After(ig.intervals[i].to) {
			if newDist := point.Sub(ig.intervals[i].to); newDist < dist {
				bestPlace = ig.intervals[i].to.Add(-dur)
				dist = newDist
			}
		} else if point.Before(ig.intervals[i].from) {
			if newDist := ig.intervals[i].from.Sub(point); newDist < dist {
				bestPlace = ig.intervals[i].from
				dist = newDist
			}
		} else {
			half := dur / 2
			if point.Add(-half).Before(ig.intervals[i].from) {
				return ig.intervals[i].from, 0
			} else if point.Add(half).After(ig.intervals[i].from) {
				return ig.intervals[i].to.Add(-dur), 0
			} else {
				return point.Add(-half), 0
			}
		}
	}

	return bestPlace, int(dist.Minutes())
}

func (ig *IvGroup) Intervals() string {
	sb := strings.Builder{}
	for i := range ig.intervals {
		sb.WriteString("[" + ig.intervals[i].from.String() + ig.intervals[i].to.String() + "]")
	}
	return sb.String()
}

// CompactCount количество интервалов продолжительностью spanMin минут, которое можно уместить в этоу интервал-группу
func (ig *IvGroup) CompactCount(spanMin int) int {
	cnt := 0
	for i := range ig.intervals {
		mins := int(ig.intervals[i].to.Sub(ig.intervals[i].from).Abs().Minutes())
		cnt += mins / spanMin
	}
	// if cnt > 16 {
	// 	fmt.Println(ig.Intervals())
	// }
	return cnt
}

// IsOk ...
func (ig *IvGroup) IsOk(from, to time.Time) bool {
	for i := range ig.intervals {
		if ig.intervals[i].from.Add(-time.Second).Before(from) && ig.intervals[i].to.Add(time.Second).After(to) {
			return true
		}
	}
	return false
}
