package logging

/*
   potentially use: https://github.com/op/go-logging
*/

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type Logger struct {
	Trace   *log.Logger
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

func NewLogger(
	traceHandle io.Writer,
	debugHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) *Logger {

	trace := log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	debug := log.New(traceHandle, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	info := log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warning := log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	err := log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	return &Logger{trace, debug, info, warning, err}
}

func NewStdLogger() *Logger {
	return NewLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stderr)
}

func (l *Logger) SetLogLevel(level string) error {
	switch level {
	case "trace":
		break
	case "debug":
		l.Trace = log.New(ioutil.Discard, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
		break
	case "info":
		l.Trace = log.New(ioutil.Discard, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
		l.Debug = log.New(ioutil.Discard, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		break
	case "warning":
		l.Trace = log.New(ioutil.Discard, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
		l.Debug = log.New(ioutil.Discard, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		l.Info = log.New(ioutil.Discard, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
		break
	case "error":
		l.Trace = log.New(ioutil.Discard, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
		l.Debug = log.New(ioutil.Discard, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		l.Info = log.New(ioutil.Discard, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
		l.Warning = log.New(ioutil.Discard, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
		break
	default:
		return errors.New(fmt.Sprintf("Invalid log level: %s", level))
	}
	return nil
}
