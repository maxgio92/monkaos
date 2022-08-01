package schedule

import (
	"context"
	"sync"

	"monkaos/pkg/chaos"
)

type Schedule struct {
	Entries []ScheduleEntry
}

type ScheduleEntry interface {
	Run(context.Context, chan<- *chaos.Result, *sync.WaitGroup, int64)
}

type ScheduleResult interface{}

func (s *Schedule) Add(entry ScheduleEntry) {
	s.Entries = append(s.Entries, entry)
}
