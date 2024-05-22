package web

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
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

var DefaultWriter io.Writer = os.Stdout

type LoggerConfig struct {
	Formatter LoggerFormatter
	out       io.Writer
}

type LoggerFormatter func(params *LogFormatterParams) string

type LogFormatterParams struct {
	Request      *http.Request
	TimeStamp    time.Time
	StatusCode   int
	Latency      time.Duration
	ClientIP     net.IP
	Method       string
	Path         string
	DisplayColor bool
}

func (p *LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode
	switch code {
	case http.StatusOK:
		return green
	default:
		return red
	}
}

func (p *LogFormatterParams) ResetColor() string {
	return reset
}

var defaultLogFormatter = func(params *LogFormatterParams) string {
	statusCodeColor := params.StatusCodeColor()
	resetColor := params.ResetColor()
	if params.Latency > time.Minute {
		params.Latency = params.Latency.Truncate(time.Second)
	}
	// 不开启颜色显示 用于向文件输出日志
	if params.DisplayColor == false {
		return fmt.Sprintf("[msgo] %v | %3d | %13v | %15s |%-7s %#v",
			params.TimeStamp.Format("2006/01/02 - 15:04:05"),
			params.StatusCode,
			params.Latency, params.ClientIP, params.Method, params.Path,
		)
	}
	return fmt.Sprintf("%s [otto] %s |%s %v %s| %s %3d %s |%s %13v %s| %15s  |%s %-7s %s %s %#v %s",
		yellow, resetColor, blue, params.TimeStamp.Format("2006/01/02 - 15:04:05"), resetColor,
		statusCodeColor, params.StatusCode, resetColor,
		red, params.Latency, resetColor,
		params.ClientIP,
		magenta, params.Method, resetColor,
		cyan, params.Path, resetColor,
	)
}

func LoggingWithConfig(conf *LoggerConfig, next HandlerFunc) HandlerFunc {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultLogFormatter
	}
	out := conf.out
	if out == nil {
		out = DefaultWriter
	}
	return func(ctx *Context) {
		param := &LogFormatterParams{
			Request: ctx.R,
		}
		log.Println("log....")
		// Start timer
		start := time.Now()
		path := ctx.R.URL.Path
		raw := ctx.R.URL.RawQuery

		next(ctx)

		// stop timer
		stop := time.Now()
		latency := stop.Sub(start)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(ctx.R.RemoteAddr))
		clientIP := net.ParseIP(ip)
		method := ctx.R.Method
		statusCode := ctx.StatusCode

		if raw != "" {
			path = path + "?" + raw
		}

		param.ClientIP = clientIP
		param.TimeStamp = stop
		param.Latency = latency
		param.StatusCode = statusCode
		param.Method = method
		param.Path = path
		param.DisplayColor = true
		fmt.Fprint(out, formatter(param))
	}
}

func Logging(handlerFunc HandlerFunc) HandlerFunc {
	return LoggingWithConfig(&LoggerConfig{}, handlerFunc)
}
