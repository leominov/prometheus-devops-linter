package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/leominov/prometheus-devops-linter/linter"
	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
)

const (
	DefaultConfigFilename = ".prometheus-linter.yaml"
)

func DiscoverConfigFile() string {
	configEnv := os.Getenv("PROM_LINTER_CONFIG")
	if len(configEnv) > 0 {
		return configEnv
	}
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

func ParseArgs() (string, []string, error) {
	args := os.Args
	if len(args) < 1 {
		return "", []string{}, errors.New("Linter type must be specified")
	}
	if len(args) < 2 {
		return "", []string{}, errors.New("Directory must be specified")
	}
	linterType := strings.ToLower(args[1])
	paths := args[2:]
	return linterType, paths, nil
}

func HasVersionFlag() bool {
	for _, arg := range os.Args {
		if arg == "--version" || arg == "-v" {
			return true
		}
	}
	return false
}

func HasHelpFlag() bool {
	for _, arg := range os.Args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func main() {
	if HasVersionFlag() {
		fmt.Println(version.Info())
		fmt.Println(version.BuildContext())
		os.Exit(0)
	}
	if HasHelpFlag() {
		fmt.Println("Usage:")
		fmt.Println("prometheus-devops-linter jobs jobs/*.*")
		fmt.Println("prometheus-devops-linter rules rules/*.*")
		os.Exit(0)
	}
	configFile := DiscoverConfigFile()
	logrus.Infof("Starting prometheus-devops-linter %s...", version.Info())
	if len(configFile) > 0 {
		logrus.Infof("Configuration path: %s", configFile)
	} else {
		logrus.Warn("Configuration file is not defined, default configuration will be used")
	}
	ml, err := linter.NewMetaLinter(configFile)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	linterType, paths, err := ParseArgs()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	err = ml.LintFilesAs(linterType, paths)
	if err != nil {
		logrus.Error(err)
		os.Exit(2)
	}
	logrus.Info("Done without errors")
}
