package main

import (
	"flag"
	"fmt"
	"github.com/metrilyx/annolityx/annolityx"
	"github.com/metrilyx/annolityx/annolityx/config"
	"github.com/metrilyx/annolityx/annolityx/logging"
	"os"
)

var (
	configFile = flag.String("c", "/etc/annolityx/annolityx.toml", "Configuration file")
	webroot    = flag.String("webroot", "", "Path to web ui directory.")
	logLevel   = flag.String("l", "info", "Log level")
	version    = flag.Bool("version", false, "Version")
)

func main() {

	flag.Parse()

	if *version {
		fmt.Printf("%s-%s\n", Version, PreReleaseVersion)
		os.Exit(0)
	}

	logger := logging.NewLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	logger.SetLogLevel(*logLevel)

	cfg, err := config.LoadConfigFromFile(*configFile)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		os.Exit(1)
	}
	if *webroot != "" {
		cfg.Http.Webroot = *webroot
	}

	annoSvc, err := annolityx.NewEventAnnoService(cfg, logger)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		os.Exit(1)
	}
	err = annoSvc.Start()
	if err != nil {
		logger.Error.Printf("%s\n", err)
		os.Exit(1)
	}
}
