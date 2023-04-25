package log

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var (
	root = &logger{[]interface{}{}, new(swapHandler)}
)

func init() {
	//默认
	var logger *GlogHandler
	logger = NewGlogHandler(StreamHandler(os.Stderr, TerminalFormat(true)))
	logger.Verbosity(LvlInfo)
	root.SetHandler(logger)
}

// New returns a new logger with the given context.
// New is a convenient alias for Root().New
func New(ctx ...interface{}) Logger {
	return root.New(ctx...)
}

// Root returns the root logger
func Root() Logger {
	return root
}

// The following functions bypass the exported logger methods (logger.Debug,
// etc.) to keep the call depth the same for all paths to logger.write so
// runtime.Caller(2) always refers to the call site in client code.

// Trace is a convenient alias for Root().Trace
func Trace(msg string, ctx ...interface{}) {
	extension := extra(ctx...)
	root.write(msg, LvlTrace, extension, skipLevel)
}

// Debug is a convenient alias for Root().Debug
func Debug(msg string, ctx ...interface{}) {
	extension := extra(ctx...)
	root.write(msg, LvlDebug, extension, skipLevel)
}

// Info is a convenient alias for Root().Info
func Info(msg string, ctx ...interface{}) {
	extension := extra(ctx...)
	root.write(msg, LvlInfo, extension, skipLevel)
}

// Warn is a convenient alias for Root().Warn
func Warn(msg string, ctx ...interface{}) {
	extension := extra(ctx...)
	root.write(msg, LvlWarn, extension, skipLevel)
}

// Error is a convenient alias for Root().Error
func Error(msg string, ctx ...interface{}) {
	extension := extra(ctx...)
	root.write(msg, LvlError, extension, skipLevel)
}

// Crit is a convenient alias for Root().Crit
func Crit(msg string, ctx ...interface{}) {
	extension := extra(ctx...)
	root.write(msg, LvlCrit, extension, skipLevel)
	os.Exit(1)
}

// Output is a convenient alias for write, allowing for the modification of
// the calldepth (number of stack frames to skip).
// calldepth influences the reported line number of the log message.
// A calldepth of zero reports the immediate caller of Output.
// Non-zero calldepth skips as many stack frames.
func Output(msg string, lvl Lvl, callDepth int, ctx ...interface{}) {
	root.write(msg, lvl, ctx, callDepth+skipLevel)
}
func WelcomeLog(name, version string) {
	Info(fmt.Sprintf("name: %s; version: %s", name, version))
}

func ExitLog(name, version string) {
	Info(fmt.Sprintf("name: %s; version: %s, exit", name, version))
}
func GetFileAndLine() (string, string) {
	var code string
	var funcName string
	for skip := 1; true; skip++ {
		pc, codePath, codeLine, ok := runtime.Caller(skip)
		if !ok {
			return code, funcName
		} else {
			code = fmt.Sprintf("%s:%d", codePath, codeLine)
			funcName = runtime.FuncForPC(pc).Name()
			if !strings.Contains(funcName, "log") {
				return code, funcName
			}
		}
	}
	return code, funcName
}
func GetCurrentGoroutineId() int {
	buf := make([]byte, 128)
	buf = buf[:runtime.Stack(buf, false)]
	stackInfo := string(buf)
	firsts := strings.Split(stackInfo, "[running]")
	first := firsts[0]
	contents := strings.Split(first, "goroutine")
	if len(contents) >= 2 {
		goIdStr := strings.TrimSpace(contents[1])
		goId, err := strconv.Atoi(goIdStr)
		if err != nil {
			fmt.Println("err=", err)
			return 0
		}
		return goId
	} else {
		return 0
	}
}
func extra(ctx ...interface{}) []interface{} {
	extension := make([]interface{}, 0)
	extension = append(extension, ctx...)
	realPath, funcName := GetFileAndLine()
	extension = append(extension, "path")
	extension = append(extension, realPath)
	extension = append(extension, "func")
	extension = append(extension, funcName)
	goroutineId := GetCurrentGoroutineId()
	extension = append(extension, "goroutine")
	extension = append(extension, goroutineId)
	return extension
}
