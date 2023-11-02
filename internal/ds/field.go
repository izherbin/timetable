package ds

import (
	"fmt"
	"time"

	"github.com/sergrom/timetable/internal/pkg"
)

type Field struct {
	ID       int
	Format   int
	GameDur  time.Duration
	TimeFrom time.Time // format example 9:00
	TimeTo   time.Time // format example 21:00
	slotsCnt int
	slots    []int
}

func NewField(id, format int, dur time.Duration, from, to string) Field {
	var timeFrom, timeTo time.Time
	var err error
	timeFrom, err = pkg.ParseHM(from)
	if err != nil {
		panic(err.Error())
	}
	timeTo, err = pkg.ParseHM(to)
	if err != nil {
		panic(err.Error())
	}

	slotsCnt := int(timeTo.Sub(timeFrom).Minutes() / dur.Minutes())
	if slotsCnt < 1 || slotsCnt > 40 {
		panic(fmt.Sprintf("slotsCnt(%d) error. slots count must be from 1 to 40", slotsCnt))
	}

	slots := make([]int, 0, slotsCnt)
	for i := 0; i < slotsCnt; i++ {
		slots = append(slots, i)
	}

	return Field{
		ID:       id,
		Format:   format,
		GameDur:  dur,
		TimeFrom: timeFrom,
		TimeTo:   timeTo,
		slotsCnt: slotsCnt,
		slots:    slots,
	}
}

func (f *Field) SlotsCnt() int {
	return f.slotsCnt
}

func (f *Field) GetSlots() []int {
	return f.slots
}

type FieldNode struct {
	Field1           *Field
	Field2           *Field
	timeFrom, timeTo time.Time
}

func (fn *FieldNode) GameDur() time.Duration {
	if fn.Field2 == nil || fn.Field1.GameDur > fn.Field2.GameDur {
		return fn.Field1.GameDur
	}
	return fn.Field2.GameDur
}

func (fn *FieldNode) GetTimeFrom() time.Time {
	if !fn.timeFrom.IsZero() {
		return fn.timeFrom
	}

	from := fn.Field1.TimeFrom
	if fn.Field2 != nil {
		if from2 := fn.Field2.TimeFrom; from.Before(from2) {
			from = from2
		}
	}

	fn.timeFrom = from
	return fn.timeFrom
}

func (fn *FieldNode) Format() int {
	if fn.Field2 != nil {
		return 7
	}
	return fn.Field1.Format
}

func (fn *FieldNode) GetTimeTo() time.Time {
	if !fn.timeTo.IsZero() {
		return fn.timeTo
	}

	to := fn.Field1.TimeTo
	if fn.Field2 != nil {
		if to2 := fn.Field2.TimeTo; to.After(to2) {
			to = to2
		}
	}

	fn.timeTo = to
	return fn.timeTo
}

func (f *FieldNode) SlotsCnt() int {
	cnt := f.Field1.SlotsCnt()
	if f.Field2 != nil && f.Field2.SlotsCnt() < cnt {
		return f.Field2.SlotsCnt()
	}
	return cnt
}

func (fn *FieldNode) IsSlotOk(slot int) bool {
	if slot >= fn.Field1.SlotsCnt() {
		return false
	}
	if fn.Field2 != nil && slot >= fn.Field2.SlotsCnt() {
		return false
	}
	return true
}

func (fn *FieldNode) String() string {
	if fn.Field2 == nil {
		return fmt.Sprintf("[поле%d]", fn.Field1.ID)
	}
	return fmt.Sprintf("[поле%d,поле%d]", fn.Field1.ID, fn.Field2.ID)
}
