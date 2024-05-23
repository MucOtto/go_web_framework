package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

type LoggerLevel int

type LoggingFormatter interface {
	Format(param *LoggingFormatterParam) string
}

type LoggingFormatterParam struct {
	Color bool
	Level LoggerLevel
	Msg   any
}

type LoggerFormatter struct {
	Color bool
	Level LoggerLevel
}

type Logger struct {
	Outs      []io.Writer
	Level     LoggerLevel
	Formatter LoggingFormatter
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
	logger.Formatter = &TextFormatter{}
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
	param := &LoggingFormatterParam{
		Level: l.Level,
		Msg:   msg,
	}
	formatter := l.Formatter.Format(param)
	for _, out := range l.Outs {
		if out == os.Stdout {
			param.Color = true
			formatter = l.Formatter.Format(param)
		}
		fmt.Fprint(out, formatter)
	}
}

func (f *LoggerFormatter) formatter(msg any) string {
	now := time.Now()
	if f.Color {
		//要带颜色  error的颜色 为红色 info为绿色 debug为蓝色
		levelColor := f.LevelColor()
		msgColor := f.MsgColor()
		return fmt.Sprintf("%s [otto] %s %s%v%s | level= %s %s %s | msg=%s %#v %s \n",
			yellow, reset, blue, now.Format("2006/01/02 - 15:04:05"), reset,
			levelColor, f.Level.Level(), reset, msgColor, msg, reset,
		)
	}
	return fmt.Sprintf("[otto] %v | level=%s | msg= %#v \n",
		now.Format("2006/01/02 - 15:04:05"),
		f.Level.Level(), msg,
	)
}

func (f *LoggerFormatter) LevelColor() string {
	switch f.Level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return cyan
	}
}

func (f *LoggerFormatter) MsgColor() string {
	switch f.Level {
	case LevelDebug:
		return ""
	case LevelInfo:
		return ""
	case LevelError:
		return red
	default:
		return cyan
	}
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
