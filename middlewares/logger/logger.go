package logger

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type loggerContextKey int

const (
	request_id loggerContextKey = iota
	request_time
)

func GetRequestId(req *http.Request) string {
	if req == nil {
		return ""
	}
	if reqID, ok := req.Context().Value(request_id).(string); ok {
		return reqID
	}
	return ""
}

func GetRequestTime(req *http.Request) time.Time {
	if req == nil {
		return time.Time{}
	}
	if reqTime, ok := req.Context().Value(request_time).(time.Time); ok {
		return reqTime
	}
	return time.Time{}
}

type LogFunc func(log *slog.Logger, ctx context.Context, level slog.Level, requestId string, status int, size int, duration time.Duration, req *http.Request)

type requestLogger struct {
	logFunc  LogFunc
	matchers []func(*http.Request, int) *slog.Level
}

type LoggerOption func(*requestLogger)

const LevelNone slog.Level = -10

func WithLogFunc(logFunc LogFunc) LoggerOption {
	return func(logger *requestLogger) {
		logger.logFunc = logFunc
	}
}

func WithRequestMatcher(matcher func(*http.Request) *slog.Level) LoggerOption {
	return func(logger *requestLogger) {
		logger.matchers = append(logger.matchers, func(req *http.Request, status int) *slog.Level {
			return matcher(req)
		})
	}
}

func WithStatusMatcher(matcher func(int) *slog.Level) LoggerOption {
	return func(logger *requestLogger) {
		logger.matchers = append(logger.matchers, func(req *http.Request, status int) *slog.Level {
			return matcher(status)
		})
	}
}

func RegexPath(regexPath string, level slog.Level) LoggerOption {
	regex := regexp.MustCompile(regexPath)
	return WithRequestMatcher(func(req *http.Request) *slog.Level {
		if regex.MatchString(req.URL.Path) {
			return &level
		}
		return nil
	})
}

func Ok(level slog.Level) LoggerOption {
	return rangeStatusMatcher(200, 299, level)
}

func Redirect(level slog.Level) LoggerOption {
	return rangeStatusMatcher(300, 399, level)
}

func BadRequest(level slog.Level) LoggerOption {
	return rangeStatusMatcher(400, 499, level)
}

func InternalServerError(level slog.Level) LoggerOption {
	return rangeStatusMatcher(500, 599, level)
}

func rangeStatusMatcher(minIncl, maxIncl int, level slog.Level) LoggerOption {
	return WithStatusMatcher(func(status int) *slog.Level {
		if status >= minIncl && status <= maxIncl {
			return &level
		}
		return nil
	})
}

var defaultLogFunc LogFunc = func(log *slog.Logger, ctx context.Context, level slog.Level, requestId string, status int, size int, duration time.Duration, req *http.Request) {
	log.Log(ctx, level, "request", "requestId", requestId, "method", req.Method, "path", req.URL.Path, "query", req.URL.RawQuery, "addr", req.RemoteAddr, "status", status, "size", size, "duration", duration)
}

func New(log *slog.Logger, opts ...LoggerOption) func(http.HandlerFunc) http.HandlerFunc {
	logger := &requestLogger{
		logFunc:  defaultLogFunc,
		matchers: make([]func(*http.Request, int) *slog.Level, 0),
	}
	for _, o := range opts {
		o(logger)
	}
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			requestId := uuid.NewString()
			now := time.Now()
			req.Context()
			ctx := context.WithValue(req.Context(), request_id, requestId)
			ctx = context.WithValue(ctx, request_time, now)
			ww := newResponseLogger(w)
			defer func() {
				duration := time.Since(now)
				size := ww.Size()
				status := ww.Status()
				level := slog.LevelInfo
				for _, matcher := range logger.matchers {
					if lvl := matcher(req, status); lvl != nil {
						level = *lvl
						break
					}
				}
				if level == LevelNone {
					return
				}
				logger.logFunc(log, ctx, level, requestId, status, size, duration, req)
			}()
			next.ServeHTTP(ww, req.WithContext(ctx))
		}
	}
}

func newResponseLogger(rw http.ResponseWriter) *responseLogger {
	nrw := &responseLogger{
		ResponseWriter: rw,
	}
	return nrw
}

type responseLogger struct {
	http.ResponseWriter
	pendingStatus int
	status        int
	size          int
}

func (rw *responseLogger) WriteHeader(s int) {
	rw.pendingStatus = s
	if rw.Written() {
		return
	}
	rw.status = s
	rw.ResponseWriter.WriteHeader(s)
}

func (rw *responseLogger) Write(b []byte) (int, error) {
	if !rw.Written() {
		rw.WriteHeader(http.StatusOK)
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func (rw *responseLogger) Status() int {
	if rw.Written() {
		return rw.status
	}

	return rw.pendingStatus
}

func (rw *responseLogger) Size() int {
	return rw.size
}

func (rw *responseLogger) Written() bool {
	return rw.status != 0
}
