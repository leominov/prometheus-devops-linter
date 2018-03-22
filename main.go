package main

import (
	"flag"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	AlertsPath = flag.String("path", "", "Directory with alert files")
)

func main() {
	flag.Parse()
	linter, err := NewLinter()
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
