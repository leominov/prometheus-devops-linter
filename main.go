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
	linter := NewLinter()
	err := linter.ProcessFilesFromPath(*AlertsPath)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	logrus.Info("Done without errors")
}
