package util

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func PrintErrors(pref string, errs []error) {
	for _, err := range errs {
		messagePref := pref
		if len(messagePref) > 0 {
			messagePref = fmt.Sprintf("%s: ", messagePref)
		}
		logrus.Errorf("%s%s", messagePref, err.Error())
	}
}
