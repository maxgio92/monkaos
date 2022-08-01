package main

import (
	"flag"
	"fmt"
	"os"

	"monkaos/internal/utils"
	"monkaos/pkg/monkey"

	log "github.com/sirupsen/logrus"
)

const (
	appName = "monkaos"
)

var (
	configFile = flag.String("c", "/etc/"+appName+".yaml", "The path to the config file")
	logLevel   = flag.String("v", log.InfoLevel.String(), "The log level between debug, info, warn, error, fatal, panic")
)

func main() {
	flag.Usage = printUsage
	flag.Parse()

	logger, err := utils.GetLogger(os.Stdout, *logLevel)
	if err != nil {
		log.Fatal(err.Error())
	}

	viper, err := utils.GetViper(*configFile, appName)
	if err != nil {
		logger.Warning(err.Error())
	}

	config := utils.GetConfigFromViper(viper)
	logger.Info(config.Print())

	monkey, err := monkey.SetupWithLogger(&config, logger)
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := monkey.Run(); err != nil {
		log.Fatal(err.Error())
	}
}

// printUsage() prints to stderr the CLI usage.
func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: "+appName+" [-v=verbosity]\n")
	flag.PrintDefaults()
	os.Exit(2)
}
