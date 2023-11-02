package searcher

import (
	"sort"
	"time"
)

type slot struct {
	field int
	nth   int
}

func IntersectSlots(slots1, slots2 []int) []int {
	out := make([]int, 0, len(slots1))
	m := make(map[int]bool, len(slots1))
	for _, slot := range slots1 {
		m[slot] = true
	}
	for _, slot := range slots2 {
		if slot >= 0 && m[slot] {
			out = append(out, slot)
		}
	}
	return out
}

func UniqueSlots(slots []int) []int {
	out := make([]int, 0, len(slots))
	keys := make(map[int]bool)
	for _, slot := range slots {
		if _, ok := keys[slot]; !ok {
			out = append(out, slot)
			keys[slot] = true
		}
	}
	sort.Ints(out)
	return out
}

func GetFromTo(start time.Time, dur time.Duration, slot int) (time.Time, time.Time) {
	return start.Add(dur * time.Duration(slot)), start.Add(dur * time.Duration(slot+1))
}
