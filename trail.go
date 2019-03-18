package uadmin

import (
	"fmt"
	"log"
	"strings"

	"github.com/dekunt/uadmin/colors"
)

// Reporting Levels
const (
	DEBUG   = 0
	WORKING = 1
	INFO    = 2
	OK      = 3
	WARNING = 4
	ERROR   = 5
)

var trailTag = map[int]string{
	DEBUG:   colors.Debug,
	WORKING: colors.Working,
	INFO:    colors.Info,
	OK:      colors.OK,
	WARNING: colors.Warning,
	ERROR:   colors.Error,
}

// Trail prints to the log
func Trail(level int, msg interface{}, i ...interface{}) {
	if level >= ReportingLevel {
		message := fmt.Sprint(msg)
		if level != WORKING && !strings.HasSuffix(message, "\n") {
			message += "\n"
		} else if level == WORKING && !strings.HasPrefix(message, "\r") {
			message = message + "\r"
		}
		if ReportTimeStamp {
			log.Printf(trailTag[level]+message, i...)
		} else {
			fmt.Printf(trailTag[level]+message, i...)
		}
	}
}
