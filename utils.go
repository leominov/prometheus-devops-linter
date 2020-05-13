package main

import (
	"strings"

	"github.com/sirupsen/logrus"
)

func ConfigureLogging(logLevel, formatter string) error {
	level := logrus.InfoLevel
	if len(logLevel) > 0 {
		levelParsed, err := logrus.ParseLevel(logLevel)
		if err != nil {
			return err
		}
		level = levelParsed
	}
	logrus.SetLevel(level)

	switch strings.ToLower(formatter) {
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	return nil
}
