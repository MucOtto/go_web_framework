package web

import (
	"log"
	"net"
	"strings"
	"time"
)

type LoggingConfig struct {
}

func LoggingWithConfig(conf LoggingConfig, next HandlerFunc) HandlerFunc {
	return func(ctx *Context) {
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

		log.Printf("[msgo] %v | %3d | %13v | %15s |%-7s %#v",
			stop.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency, clientIP, method, path,
		)
	}
}

func Logging(handlerFunc HandlerFunc) HandlerFunc {
	return LoggingWithConfig(LoggingConfig{}, handlerFunc)
}
