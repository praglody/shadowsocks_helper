package slog

import "log"

const (
	LOG_EMERGENCY = 1 << 7
	LOG_ALERT     = 1 << 6
	LOG_CRITICAL  = 1 << 5
	LOG_ERROR     = 1 << 4
	LOG_WARNING   = 1 << 3
	LOG_NOTICE    = 1 << 2
	LOG_INFO      = 1 << 1
	LOG_DEBUG     = 1 << 0
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
