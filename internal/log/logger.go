package log

import (
	"sync"

	"github.com/K4rian/dslogger"
)

var (
	Logger *dslogger.Logger
	once   sync.Once
)

func Init(
	level string,
	file string,
	fileFormat string,
	maxSize int,
	maxBackups int,
	maxAge int,
	logToFile bool,
) {
	once.Do(func() {
		loggerConfig := &dslogger.Config{
			LogFile:       file,
			LogFileFormat: dslogger.LogFormat(fileFormat),
			MaxSize:       maxSize,
			MaxBackups:    maxBackups,
			MaxAge:        maxAge,
			Level:         level,
		}

		if logToFile {
			Logger = dslogger.NewLogger(level, loggerConfig)
		} else {
			Logger = dslogger.NewConsoleLogger(level, loggerConfig)
		}
	})
}
