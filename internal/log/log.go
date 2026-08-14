package log

import (
	"fmt"
	"time"
)

var DebugMode bool

func basePrint(color, msg string) {
	fmt.Printf("%s[%s] %s\x1B[0m\n", color, time.Now().Format("2006-01-02 15:04:05.000"), msg)
}

func Red(msg string)    { basePrint("\x1B[31m", msg) }
func Green(msg string)  { basePrint("\x1B[32m", msg) }
func Yellow(msg string) { basePrint("\x1B[33m", msg) }
func Blue(msg string)   { basePrint("\x1B[34m", msg) }
func Magenta(msg string){ basePrint("\x1B[35m", msg) }
func Grey(msg string)   { basePrint("\x1B[2m", msg) }

func Debugf(format string, args ...interface{}) {
	if DebugMode {
		basePrint("\x1B[2m", fmt.Sprintf(format, args...))
	}
}
