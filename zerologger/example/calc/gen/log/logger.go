// Code generated by goa v3.11.0, DO NOT EDIT.
//
// Zerodriver logger implementation
//
// Command:
// $ goa gen calc/design

package log

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hirosassa/zerodriver"
	"github.com/rs/zerolog"
	httpmdlwr "goa.design/goa/v3/http/middleware"
	"goa.design/goa/v3/middleware"
)

// Logger is an adapted zerodriver logger
type Logger struct {
	*zerodriver.Logger
}

// New creates a new zerodriver logger
func New(serviceName string, isDebug bool) *Logger {
	logger := zerodriver.NewProductionLogger()
	if isDebug {
		logger = zerodriver.NewDevelopmentLogger()
	}
	return &Logger{logger}
}

// Log is called by the log middleware to log HTTP requests key values
func (logger *Logger) Log(keyvals ...interface{}) error {
	fields := FormatFields(keyvals)
	logger.Info().Fields(fields).Msgf("HTTP Request")
	return nil
}

// FormatFields formats input keyvals
// ref: https://github.com/goadesign/goa/blob/v1/logging/logrus/adapter.go#L64
func FormatFields(keyvals []interface{}) map[string]interface{} {
	n := (len(keyvals) + 1) / 2
	res := make(map[string]interface{}, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		var v interface{}
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}
		res[fmt.Sprintf("%v", k)] = v
	}
	return res
}

// ZerodriverHttpMiddleware extracts and formats http request and response information into
// GCP Cloud Logging optimized format.
// If logger is not *Logger, it returns goa default middleware.
func ZerodriverHttpMiddleware(logger middleware.Logger) func(h http.Handler) http.Handler {
	switch logr := logger.(type) {
	case *Logger:
		return zerodriverHttpMiddleware(logr)
	default:
		return httpmdlwr.Log(logger)
	}
}

func zerodriverHttpMiddleware(logger *Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := httpmdlwr.CaptureResponse(w)
			h.ServeHTTP(rw, r)

			var res http.Response
			res.StatusCode = rw.StatusCode
			res.ContentLength = int64(rw.ContentLength)

			p := zerodriver.NewHTTP(r, &res)
			p.Latency = time.Since(start).String()

			var level zerolog.Level
			switch {
			case rw.StatusCode < 400:
				level = zerolog.InfoLevel
			case rw.StatusCode < 500:
				level = zerolog.WarnLevel
			default:
				level = zerolog.ErrorLevel
			}

			logger.WithLevel(level).
				HTTP(p).
				Msg("request finished")
		})
	}
}
