package log

import (
	"fmt"
	"time"
)

type TextFormatter struct {
}

func (f *TextFormatter) Format(param *LoggingFormatterParam) string {
	now := time.Now()
	if param.Color {
		//要带颜色  error的颜色 为红色 info为绿色 debug为蓝色
		levelColor := f.LevelColor(param.Level)
		msgColor := f.MsgColor(param.Level)
		msgInfo := "| mgs="
		if param.Level == LevelError {
			msgInfo = "\nError caused by:"
		}
		return fmt.Sprintf("%s [otto] %s %s%v%s | level= %s %s %s %s%s %v %s \n",
			yellow, reset, blue, now.Format("2006/01/02 - 15:04:05"), reset,
			levelColor, param.Level.Level(), reset, msgInfo, msgColor, param.Msg, reset,
		)
	}
	return fmt.Sprintf("[otto] %v | level=%s | msg= %v \n",
		now.Format("2006/01/02 - 15:04:05"),
		param.Level.Level(), param.Msg,
	)
}

func (f *TextFormatter) LevelColor(level LoggerLevel) string {
	switch level {
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

func (f *TextFormatter) MsgColor(level LoggerLevel) string {
	switch level {
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
