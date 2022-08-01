package config

import (
	"fmt"
	"monkaos/pkg/victims"
)

var (
	excludedNamespaces = []string{"kube-system", "kube-public", "kube-node-lease"}
)

const (
	victimsPerSchedule            = 1
	terminationGracePeriodSeconds = 10
	tickPeriodSeconds             = 10
	maxLatencySeconds             = 5
	enableRandomLatency           = true
	deadlineSeconds               = 1
	strategy                      = victims.RandomPodRandomNamespaceStrategy
)

type Config struct {
	VictimsPerSchedule            int
	ExcludedNamespaces            []string
	TerminationGracePeriodSeconds int
	TickPeriodSeconds             int
	MaxLatencySeconds             int
	EnableRandomLatency           bool
	DeadlineSeconds               int
	Strategy                      victims.Strategy
}

func NewFromDefault() Config {
	return getDefaultConfig()
}

func getDefaultConfig() Config {
	return Config{
		VictimsPerSchedule:            victimsPerSchedule,
		ExcludedNamespaces:            excludedNamespaces,
		TerminationGracePeriodSeconds: terminationGracePeriodSeconds,
		TickPeriodSeconds:             tickPeriodSeconds,
		MaxLatencySeconds:             maxLatencySeconds,
		EnableRandomLatency:           enableRandomLatency,
		DeadlineSeconds:               deadlineSeconds,
		Strategy:                      strategy,
	}
}

func (c *Config) Print() string {
	return fmt.Sprintf(
		"Configuration initialized with:\ndeadline: %d seconds, max latency: %d seconds, tick period: %d seconds, random latency enabled: %t, termination grace period: %d seconds, pods per schedule: %d, excluded namespaces: %s, strategy: %s",
		c.DeadlineSeconds,
		c.MaxLatencySeconds,
		c.TickPeriodSeconds,
		c.EnableRandomLatency,
		c.TerminationGracePeriodSeconds,
		c.VictimsPerSchedule,
		c.ExcludedNamespaces,
		c.Strategy)
}
