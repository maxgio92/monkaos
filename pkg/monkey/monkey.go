package monkey

import (
	"time"

	log "github.com/sirupsen/logrus"

	"monkaos/pkg/config"
	"monkaos/pkg/schedule"
)

type Monkey struct {
	logger *log.Logger
	config *config.Config
}

func SetupWithLogger(config *config.Config, logger *log.Logger) (*Monkey, error) {
	return &Monkey{
		config: config,
		logger: logger,
	}, nil
}

func (m *Monkey) Run() error {
	scheduler, err := schedule.SetupSchedulerWithLogger(
		m.logger,
		m.config.VictimsPerSchedule,
		m.config.ExcludedNamespaces,
		m.config.TerminationGracePeriodSeconds,
		m.config.MaxLatencySeconds,
		m.config.EnableRandomLatency,
		m.config.DeadlineSeconds)
	if err != nil {
		return err
	}
	for {
		scheduler.Tick()
		time.Sleep(time.Duration(m.config.TickPeriodSeconds) * time.Second)
	}
}
