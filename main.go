package main

import (
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

const (
	DefaultConfigFilename = ".prometheus-linter.yaml"
)

var (
	AlertsPath = flag.String("path", "", "Directory with alert files")
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
	flag.Parse()
	configFile := DiscoverConfigFile()
	linter, err := NewLinter(configFile)
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
