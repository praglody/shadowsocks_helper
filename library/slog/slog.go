package slog

import "log"

const (
	LOG_DEBUG = 1 << iota
	LOG_INFO
	LOG_NOTICE
	LOG_WARNING
	LOG_ERROR
	LOG_CRITICAL
	LOG_ALERT
	LOG_EMERGENCY
)

var LogLevel int8 = LOG_INFO

func Debug(v ...interface{}) {
	if LogLevel <= LOG_DEBUG {
		log.Println(v)
	}
}

func Debugf(format string, v ...interface{}) {
	if LogLevel <= LOG_DEBUG {
		log.Printf(format, v)
	}
}

func Info(v ...interface{}) {
	log.Println(v)
}

func Infof(format string, v ...interface{}) {
	log.Printf(format, v)
}

func Error(v ...interface{}) {
	log.Println(v)
}

func Errorf(format string, v ...interface{}) {
	log.Printf(format, v)
}

func Panic(v ...interface{}) {
	log.Panicln(v)
}

func Panicf(format string, v ...interface{}) {
	log.Panicf(format, v)
}
