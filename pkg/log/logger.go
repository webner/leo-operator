package log

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
)

type StdLoggerAdapter struct {
	ReqLogger logr.Logger
}

func (a StdLoggerAdapter) Fatal(args ...interface{}) {
	a.ReqLogger.Info("[FATAL] " + fmt.Sprint(args...))
	os.Exit(1)
}

func (a StdLoggerAdapter) Fatalln(args ...interface{}) {
	a.ReqLogger.Info("[FATAL] " + fmt.Sprintln(args...))
	os.Exit(1)
}

func (a StdLoggerAdapter) Fatalf(format string, args ...interface{}) {
	a.ReqLogger.Info("[FATAL] " + fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (a StdLoggerAdapter) Print(args ...interface{}) {
	a.ReqLogger.Info(fmt.Sprint(args...))
}

func (a StdLoggerAdapter) Println(args ...interface{}) {
	a.ReqLogger.Info(fmt.Sprintln(args...))
}

func (a StdLoggerAdapter) Printf(format string, args ...interface{}) {
	a.ReqLogger.Info(fmt.Sprintf(format, args...))

}
