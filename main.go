package main

import (
	"fmt"
	nFormatter "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
)

func init() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&nFormatter.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		CallerFirst:     true,
		CustomCallerFormatter: func(frame *runtime.Frame) string {
			filename := ""
			slash := strings.LastIndex(frame.File, "/")
			if slash >= 0 {
				filename = frame.File[slash+1:]
			}
			return fmt.Sprintf("  %s:%d", filename, frame.Line)
		},
	})
}

func main() {
	u := NewUpdater()
	u.Start()
}
