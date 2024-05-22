package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

type LoggerLevel int

type LoggerFormatter struct {
	Color bool
	Level LoggerLevel
}

type Logger struct {
	Outs      []io.Writer
	Level     LoggerLevel
	Formatter LoggerFormatter
}

const (
	LevelDebug LoggerLevel = iota
	LevelInfo
	LevelError
)

func New() *Logger {
	return &Logger{}
}

func Default() *Logger {
	logger := New()
	logger.Outs = append(logger.Outs, os.Stdout)
	logger.Level = LevelDebug
	return logger
}

func (l *Logger) Info(msg any) {
	l.Print(LevelInfo, msg)
}

func (l *Logger) Debug(msg any) {
	l.Print(LevelDebug, msg)
}

func (l *Logger) Error(msg any) {
	l.Print(LevelError, msg)
}

func (l *Logger) Print(level LoggerLevel, msg any) {
	if l.Level > level {
		//级别不满足 不打印日志
		return
	}
	l.Formatter.Level = level
	formatter := l.Formatter.formatter(msg)
	for _, out := range l.Outs {
		fmt.Fprint(out, formatter)
	}
}

func (f *LoggerFormatter) formatter(msg any) string {
	now := time.Now()
	return fmt.Sprintf("[otto] %v | level=%s | msg=%#v \n",
		now.Format("2006/01/02 - 15:04:05"),
		f.Level.Level(), msg,
	)
}

func (level LoggerLevel) Level() string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	default:
		return ""
	}
}
