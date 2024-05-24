package log

import (
	"fmt"
	"github.com/MucOtto/web/internel/mystrings"
	"io"
	"os"
	"path"
	"strings"
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
	Outs        []LoggerWriter
	Level       LoggerLevel
	Formatter   LoggingFormatter
	logPath     string
	LogFileSize int64
}

type LoggerWriter struct {
	Level LoggerLevel
	Out   io.Writer
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
	w := LoggerWriter{
		Level: LevelDebug,
		Out:   os.Stdout,
	}
	logger.Outs = append(logger.Outs, w)
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

func (l *Logger) SetFilePath(logPath string) {
	l.logPath = logPath
	l.Outs = append(l.Outs, LoggerWriter{
		Level: -1,
		Out:   fileWriter(path.Join(logPath, "all.log")),
	})
	l.Outs = append(l.Outs, LoggerWriter{
		Level: LevelDebug,
		Out:   fileWriter(path.Join(logPath, "debug.log")),
	})
	l.Outs = append(l.Outs, LoggerWriter{
		Level: LevelInfo,
		Out:   fileWriter(path.Join(logPath, "info.log")),
	})
	l.Outs = append(l.Outs, LoggerWriter{
		Level: LevelError,
		Out:   fileWriter(path.Join(logPath, "error.log")),
	})
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
		if out.Out == os.Stdout {
			param.Color = true
			formatter = l.Formatter.Format(param)
			fmt.Fprint(out.Out, formatter)
		} else if out.Level == -1 || out.Level == level {
			param.Color = false
			formatter = l.Formatter.Format(param)
			// 检查日志大小 适当进行切分
			l.checkFileSize(out)
			fmt.Fprint(out.Out, formatter)
		}
	}
}

func (l *Logger) checkFileSize(writer LoggerWriter) {
	file := writer.Out.(*os.File)
	if file != nil {
		stat, _ := file.Stat()
		size := stat.Size()
		if l.LogFileSize <= 0 {
			l.LogFileSize = 64 << 20
		}
		if size > l.LogFileSize {
			_, filename := path.Split(file.Name())
			name := filename[0:strings.Index(filename, ".")]
			filename = path.Join(l.logPath, mystrings.ConnectAnyStr(name, ".", time.Now().Format("2006-01-02 15:04:05"), ".log"))
			io := fileWriter(filename)
			writer.Out = io
		}
	}
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

func fileWriter(name string) io.Writer {
	w, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return w
}
