package logger

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Embiggenerd/articles/config"
	"github.com/google/uuid"
	slogmulti "github.com/samber/slog-multi"
	"github.com/urfave/negroni"
)

const (
	logFatal = slog.Level(13)
)

// Loger is an extension of log.slog that includes fatal
type Logger interface {
	Fatal(msg string)
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
	Info(msg string, args ...any)
	LoggingMW(next http.Handler) http.Handler
	LogAPIRequest(id, ip, path, port, method string, timeRecieved time.Time, nanoSeconds int64, statusCode int)
	LogRequestError(requestID, errorMessage string, statusCode int)
	logEvent(requestID, metadata, direction, event, data string)
	LogEventSent(requestID, metadata string, message any)
	LogEventReceived(requestID, metadata string, message any)
}

// NewLoggerService creates and returns a new Logger instance
func NewLoggerService(ctx context.Context, cfg *config.Config) Logger {
	dirPath := "../logs" // Path to the directory you want to create

	// os.ModePerm is 0777, granting full permissions.
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		log.Fatal(err.Error())
	}

	file, err := os.OpenFile(dirPath+"/"+cfg.Get(cfg.LogFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		log.Fatal(err.Error())
	}
	stderr := os.Stderr

	slogger := slog.New(
		slogmulti.Fanout(
			slog.NewJSONHandler(file, &slog.HandlerOptions{}),
			slog.NewTextHandler(stderr, &slog.HandlerOptions{}),
		),
	)
	logger := &CustomLogger{Logger: slogger}
	logger.Info("logging service Up")
	return logger
}

// CustomLogger implements slog.Handler with custom behavior
type CustomLogger struct {
	*slog.Logger
}

// Fatal logs a message and exits
func (l *CustomLogger) Fatal(msg string) {
	l.Log(context.TODO(), logFatal, "msg")
	os.Exit(1)
}

func (l *CustomLogger) LoggingMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(WithMetadata(context.Background()))
		defer cancel()

		method := r.Method
		path := r.URL.EscapedPath()
		ip, port, _ := net.SplitHostPort(r.RemoteAddr)
		rw := negroni.NewResponseWriter(w)
		newUUID := uuid.New()

		ExposeContextMetadata(ctx).Set("requestID", newUUID.String())

		next.ServeHTTP(w, r.WithContext(ctx))

		statusCode := rw.Status()

		defer func(begin time.Time) {
			tookMs := time.Since(begin).Nanoseconds()
			l.LogAPIRequest(newUUID.String(), ip, path, port, method, time.Now(), tookMs, statusCode)
		}(time.Now())
	})
}

func (l *CustomLogger) LogAPIRequest(id, ip, path, port, method string, timeRecieved time.Time, nanoSeconds int64, statusCode int) {
	level := slog.LevelInfo
	if statusCode >= 400 {
		level = slog.LevelError
	}
	l.Log(context.TODO(), level, "API Request",
		slog.String("requestID", id),
		slog.Int("statusCode", statusCode),
		slog.String("ip", ip),
		slog.String("path", path),
		slog.String("port", port),
		slog.String("method", method),
		slog.Int64("nanoSeconds", nanoSeconds),
		slog.Time("timeReceived", timeRecieved))
}

func (l *CustomLogger) LogRequestError(requestID, errorMessage string, statusCode int) {
	l.Log(
		context.TODO(),
		slog.LevelError,
		"Request Error",
		slog.String("requestID", requestID),
		slog.String("errorMessage", errorMessage),
		slog.Int("statusCode", statusCode),
	)
}

func (l *CustomLogger) LogEventSent(requestID, metadata string, message any) {
	d, err := json.Marshal(message)
	if err != nil {
		l.Error(err.Error())
		return
	}

	l.logEvent(requestID, metadata, "Sent", message.(string), string(d))
}

func (l *CustomLogger) LogEventReceived(requestID, metadata string, message any) {
	d, err := json.Marshal(message)
	if err != nil {
		l.Error(err.Error())
		return
	}

	l.logEvent(requestID, metadata, "Received", message.(string), string(d))
}

func (l *CustomLogger) logEvent(requestID, metadata, direction, event, data string) {
	l.Log(
		context.TODO(),
		slog.LevelInfo,
		"Event Message "+direction,
		slog.String("requestID", requestID),
		slog.String("event", event),
		slog.String("data", data),
		slog.String("metadata", metadata),
	)
}
