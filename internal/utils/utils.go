package utils

import (
	"errors"
	"fmt"
	"io"
	"monkaos/pkg/config"
	"monkaos/pkg/victims"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Initialize the logging stack.
func GetLogger(out io.Writer, level string) (*log.Logger, error) {
	logger := log.New()
	logger.Out = out

	lvl, err := log.ParseLevel(level)
	if err != nil {
		return &log.Logger{}, err
	}

	logger.Level = lvl

	return logger, nil
}

func GetViper(configFile string, configName string) (*viper.Viper, error) {
	if configFile != "" {

		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {

		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		// Search config yaml files in $HOME/.monkaos directory.
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName("." + configName)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return viper.New(), fmt.Errorf("fatal error config file: %w", err)
		}
	}

	return viper.GetViper(), nil
}

// Initialize the configuration stack.
func GetConfigFromViper(viper *viper.Viper) config.Config {
	config := config.NewFromDefault()

	schedulerConfig := viper.Sub("scheduler")
	if schedulerConfig != nil {
		deadlineSeconds := schedulerConfig.GetInt("deadlineSeconds")
		if deadlineSeconds > -1 {
			config.DeadlineSeconds = deadlineSeconds
		}

		config.MaxLatencySeconds = schedulerConfig.GetInt("maxLatencySeconds")

		tickPeriodSeconds := schedulerConfig.GetInt("tickPeriodSeconds")
		if tickPeriodSeconds > -1 {
			config.TickPeriodSeconds = tickPeriodSeconds
		}

		config.EnableRandomLatency = schedulerConfig.GetBool("enableRandomLatency")
	}

	chaosConfig := viper.Sub("chaos")
	if chaosConfig != nil {
		config.TerminationGracePeriodSeconds = chaosConfig.GetInt("terminationGracePeriodSeconds")
		config.VictimsPerSchedule = chaosConfig.GetInt("victimsPerSchedule")
		config.ExcludedNamespaces = chaosConfig.GetStringSlice("excludeNamespaces")
		config.Strategy = victims.Strategy(chaosConfig.GetString("strategy"))
	}

	return config
}
