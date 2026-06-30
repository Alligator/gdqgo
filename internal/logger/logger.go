package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var Verbose bool

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
}

func Debugf(area string, format string, args ...any) {
	s := fmtLog(area, format, args...)
	if !Verbose {
		logFile(s)
		return
	}

	logStdout(s)
	logFile(s)
}

func Logf(area string, format string, args ...any) {
	s := fmtLog(area, format, args...)
	logStdout(s)
	logFile(s)
}

func fmtLog(area string, format string, args ...any) string {
	logStr := fmt.Sprintf(format, args...)
	return fmt.Sprintf("[%s] %s", area, logStr)
}

func logStdout(msg string) {
	log.Print(msg)
}

func logFile(msg string) {
	f, err := os.OpenFile("gdq.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o700)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	t := time.Now().UTC()
	fmt.Fprintf(f, "%s %s\n", t.Format("2006-01-02 15:04:05"), msg)
}
