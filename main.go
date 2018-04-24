package main

import (
	"flag"
	"os"
	"path"

	"github.com/leominov/prometheus-devops-linter/linter"
	"github.com/sirupsen/logrus"
)

const (
	DefaultConfigFilename = ".prometheus-linter.yaml"
)

var (
	AlertsPath = flag.String("path", "", "Directory with alert files")
	ConfigFile = flag.String("config-file", "", "Configuration file")
)

func DiscoverConfigFile() string {
	dir, err := os.Getwd()
	if err != nil {
		dir = "./"
	}
	filename := path.Join(dir, DefaultConfigFilename)
	if _, err := os.Stat(filename); err == nil {
		return filename
	}
	return ""
}

func main() {
	var configFile string
	flag.Parse()
	configFile = *ConfigFile
	if len(configFile) == 0 {
		configFile = DiscoverConfigFile()
	}
	linter, err := linter.NewLinter(configFile)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	err = linter.ProcessFilesFromPath(*AlertsPath)
	if err != nil {
		logrus.Error(err)
		os.Exit(2)
	}
	logrus.Info("Done without errors")
}
