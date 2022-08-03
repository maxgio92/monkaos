package schedule

import (
	"context"
	"crypto/rand"
	"math/big"
	"monkaos/pkg/chaos"
	"monkaos/pkg/victims"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Scheduler struct {
	logger *log.Logger

	// The pod count for each chaos
	victimsPerScheduleCount int

	// The list of namespace to exclude
	chaosExcludedNamespaces []string

	// Grace period before terminating each chaos
	chaosGracePeriodSeconds int64

	// Max latency for each chaos.
	// This is useful to add latency per individual schedule chaos.
	maxLatencySeconds int64

	// Whether to calculate chaos latency randomly.
	// `maxLatencySeconds` puts the highest value
	// of latency.
	enableRandomLatency bool

	// Schedule deadline
	scheduleDeadlineSeconds int64
}

func SetupSchedulerWithLogger(
	logger *log.Logger,
	victimsPerScheduleCount int,
	chaosExcludedNamespaces []string,
	chaosGracePeriodSeconds int,
	maxLatencySeconds int,
	enableRandomLatency bool,
	scheduleDeadlineSeconds int) (*Scheduler, error) {

	return &Scheduler{
		logger:                  logger,
		victimsPerScheduleCount: victimsPerScheduleCount,
		chaosExcludedNamespaces: chaosExcludedNamespaces,
		chaosGracePeriodSeconds: int64(chaosGracePeriodSeconds),
		maxLatencySeconds:       int64(maxLatencySeconds),
		enableRandomLatency:     enableRandomLatency,
		scheduleDeadlineSeconds: int64(scheduleDeadlineSeconds),
	}, nil
}

func (s *Scheduler) Tick() {
	schedule, err := s.Next()
	if err != nil {
		s.logger.Fatal(err.Error())
	}

	s.Schedule(schedule)
}

func (s *Scheduler) Next() (*Schedule, error) {
	s.logger.Info("Status Update: Generating schedule for terminations")

	victims, err := victims.GetPodVictims(s.victimsPerScheduleCount, s.chaosExcludedNamespaces, victims.RandomPodRandomNamespaceStrategy)
	if err != nil {
		return nil, err
	}

	schedule := &Schedule{
		Entries: []ScheduleEntry{},
	}

	// Add one entry per victim to the schedule.
	for _, victim := range victims {
		execTime, err := s.calculateExecTime()
		if err != nil {
			return nil, err
		}

		schedule.Add(chaos.New(execTime, victim))
	}

	return schedule, nil
}

func (s *Scheduler) Schedule(schedule *Schedule) {
	s.scheduleSafe(schedule)
}

func (s *Scheduler) scheduleSafe(schedule *Schedule) {

	// Senders wait group (chaos workers group).
	chaosWG := sync.WaitGroup{}
	chaosWG.Add(len(schedule.Entries))

	// Results data channel.
	dataCh := make(chan *chaos.Result)
	defer close(dataCh)

	// Context with stop channel.
	deadline := time.Now().Add(time.Duration(s.scheduleDeadlineSeconds) * time.Second)
	ctx, cancelCtx := context.WithDeadline(context.Background(), deadline)
	defer cancelCtx()

	// Run data senders (chaos workers).

	for _, chaos := range schedule.Entries {
		go chaos.Run(ctx, dataCh, &chaosWG, s.chaosGracePeriodSeconds)
	}

	// Receive data from senders (chaos workers).

	chaosCount := len(schedule.Entries)
	completedChaosCount := 0
	var result *chaos.Result

	s.logger.Info("Status Update: Waiting to run scheduled chaos.")

	// Run data receiver
	go func() {
		for completedChaosCount < chaosCount {
			select {
			case result = <-dataCh:
				if result.Err != nil {
					s.logger.Errorf("Failed to execute chaos for %s/%s. Error: %v", result.Chaos.Victim.Namespace, result.Chaos.Victim.Name, result.Err)
				} else {
					s.logger.Infof("Chaos successfully executed for %s/%s", result.Chaos.Victim.Namespace, result.Chaos.Victim.Name)
				}
				completedChaosCount++

				s.logger.Infof("Status Update: %d scheduled chaos left", chaosCount-completedChaosCount)
			case <-ctx.Done():
				if err := ctx.Err(); err != nil {
					s.logger.Errorf("Schedule has been terminated for: %s", err)
				}

				return
			}
		}
	}()

	chaosWG.Wait()
}

func (s *Scheduler) calculateExecTime() (time.Time, error) {
	switch s.enableRandomLatency {
	case true:
		return s.calculateRandomDelayedExecTime()
	default:
		return s.calculateDelayedExecTime(), nil
	}
}

func (s *Scheduler) calculateDelayedExecTime() time.Time {
	return time.Now().Add(time.Duration(s.maxLatencySeconds) * time.Second)
}

func (s *Scheduler) calculateRandomDelayedExecTime() (time.Time, error) {
	randomLatencySeconds, err := getRandomInt(int(s.maxLatencySeconds))
	if err != nil {
		return time.Time{}, err
	}

	return time.Now().Add(time.Duration(randomLatencySeconds) * time.Second), nil
}

func getRandomInt(max int) (int, error) {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}

	return int(r.Int64()), nil
}
