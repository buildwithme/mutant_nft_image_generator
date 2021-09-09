package utils

import (
	"log"
	"os"
	"runtime"
	"strings"
)

var debug bool

func init() {
	debug = os.Getenv("DEBUG") != ""
	debug = true
}

// Info example:
//
// Info("timezone %s", timezone)
//
func Info(msg string, vars ...interface{}) {
	log.Printf(strings.Join([]string{"[INFO ]", msg}, " "), vars...)
}

// Debug example:
//
// Debug("timezone %s", timezone)
//
func Debug(msg string, vars ...interface{}) {
	if debug {
		log.Printf(strings.Join([]string{"[DEBUG]", msg}, " "), vars...)
	}
}

// Fatal example:
//
// Fatal(errors.New("db timezone must be UTC"))
//
func Fatal(err error) {
	pc, fn, line, _ := runtime.Caller(1)
	// Include function name if debugging
	if debug {
		log.Fatalf("[FATAL] %s [%s:%s:%d]", err, runtime.FuncForPC(pc).Name(), fn, line)
	} else {
		log.Fatalf("[FATAL] %s [%s:%d]", err, fn, line)
	}
}

// Fatal example:
//
// Fatal(errors.New("db timezone must be UTC"))
//
func Fatalf(err string, values ...interface{}) {
	pc, fn, line, _ := runtime.Caller(1)
	// Include function name if debugging
	if debug {
		log.Fatalf("[FATAL] %s [%s:%s:%d] %v", err, runtime.FuncForPC(pc).Name(), fn, line, values)
	} else {
		log.Fatalf("[FATAL] %s [%s:%d] %v", err, fn, line,values)
	}
}

// Error example:
//
// Error(errors.Errorf("Invalid timezone %s", timezone))
//
func Error(err error) {
	pc, fn, line, _ := runtime.Caller(1)
	// Include function name if debugging
	if debug {
		log.Printf("[ERROR] %s [%s:%s:%d]", err, runtime.FuncForPC(pc).Name(), fn, line)
	} else {
		log.Printf("[ERROR] %s [%s:%d]", err, fn, line)
	}
}
