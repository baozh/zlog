package zlog

import (
	"os"
	"path/filepath"
)

var (
	LEVEL_FLAGS = [...]string{"DEBUG", " INFO", " WARN", "ERROR", "FATAL"}
	pid      = os.Getpid()
	baseName = filepath.Base(os.Args[0])
	hostName, _ = os.Hostname()
)

type LogLevel uint

const (
	DebugLevel 	LogLevel = 0
	InfoLevel 	LogLevel = 1
	WarnLevel 	LogLevel = 2
	ErrorLevel 	LogLevel = 3
	FatalLevel 	LogLevel = 4
)


