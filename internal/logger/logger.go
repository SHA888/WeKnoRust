package logger

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"sort"
	"strings"

	"github.com/Tencent/WeKnowRust/internal/types"
	"github.com/sirupsen/logrus"
)

// LogLevel represents the log level type
type LogLevel string

// Log level constants
const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelFatal LogLevel = "fatal"
)

// ANSI color codes
const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorReset  = "\033[0m"
)

type CustomFormatter struct {
	ForceColor bool // Whether to force colors even in non-TTY environments
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	level := strings.ToUpper(entry.Level.String())

	// Set color based on log level
	var levelColor, resetColor string
	if f.ForceColor {
		switch entry.Level {
		case logrus.DebugLevel:
			levelColor = colorCyan
		case logrus.InfoLevel:
			levelColor = colorGreen
		case logrus.WarnLevel:
			levelColor = colorYellow
		case logrus.ErrorLevel:
			levelColor = colorRed
		case logrus.FatalLevel:
			levelColor = colorPurple
		default:
			levelColor = colorReset
		}
		resetColor = colorReset
	}

	// Extract caller field from entry data
	caller := ""
	if val, ok := entry.Data["caller"]; ok {
		caller = fmt.Sprintf("%v", val)
	}

	// Build fields string: request_id first, then others sorted
	fields := ""

	// request_id first
	if v, ok := entry.Data["request_id"]; ok {
		fields += fmt.Sprintf("request_id=%v ", v)
	}

	// Sort and append remaining fields
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		if k != "caller" && k != "request_id" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		fields += fmt.Sprintf("%s=%v ", k, entry.Data[k])
	}

	fields = strings.TrimSpace(fields)

	// Compose final output with optional colors
	return []byte(fmt.Sprintf("%s%-5s%s[%s] [%s] %-20s | %s\n",
		levelColor, level, resetColor, timestamp, fields, caller, entry.Message)), nil
}

// Initialize global logger settings
func init() {
	// Set log formatter (without altering global timezone)
	logrus.SetFormatter(&CustomFormatter{ForceColor: true})
	logrus.SetReportCaller(false)
}

// GetLogger returns a logger entry from context or a new one
func GetLogger(c context.Context) *logrus.Entry {
	if logger := c.Value(types.LoggerContextKey); logger != nil {
		return logger.(*logrus.Entry)
	}
	newLogger := logrus.New()
	newLogger.SetFormatter(&CustomFormatter{ForceColor: true})
	// Set default log level
	newLogger.SetLevel(logrus.DebugLevel)
	// Caller info controlled by formatter; entry returned here
	return logrus.NewEntry(newLogger)
}

// SetLogLevel sets the global log level
func SetLogLevel(level LogLevel) {
	var logLevel logrus.Level

	switch level {
	case LevelDebug:
		logLevel = logrus.DebugLevel
	case LevelInfo:
		logLevel = logrus.InfoLevel
	case LevelWarn:
		logLevel = logrus.WarnLevel
	case LevelError:
		logLevel = logrus.ErrorLevel
	case LevelFatal:
		logLevel = logrus.FatalLevel
	default:
		logLevel = logrus.InfoLevel
	}

	logrus.SetLevel(logLevel)
}

// addCaller adds caller information to the log entry
func addCaller(entry *logrus.Entry, skip int) *logrus.Entry {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return entry
	}
	shortFile := path.Base(file)
	funcName := "unknown"
	if fn := runtime.FuncForPC(pc); fn != nil {
		// Keep only the function name without package path (e.g., doSomething)
		fullName := path.Base(fn.Name())
		parts := strings.Split(fullName, ".")
		funcName = parts[len(parts)-1]
	}
	return entry.WithField("caller", fmt.Sprintf("%s:%d[%s]", shortFile, line, funcName))
}

// WithRequestID attaches the request ID to the logger in context
func WithRequestID(c context.Context, requestID string) context.Context {
	return WithField(c, "request_id", requestID)
}

// WithField attaches a single field to the logger in context
func WithField(c context.Context, key string, value interface{}) context.Context {
	logger := GetLogger(c).WithField(key, value)
	return context.WithValue(c, types.LoggerContextKey, logger)
}

// WithFields attaches multiple fields to the logger in context
func WithFields(c context.Context, fields logrus.Fields) context.Context {
	logger := GetLogger(c).WithFields(fields)
	return context.WithValue(c, types.LoggerContextKey, logger)
}

// Debug logs at debug level
func Debug(c context.Context, args ...interface{}) {
	addCaller(GetLogger(c), 2).Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(c context.Context, format string, args ...interface{}) {
	addCaller(GetLogger(c), 2).Debugf(format, args...)
}

// Info logs at info level
func Info(c context.Context, args ...interface{}) {
	addCaller(GetLogger(c), 2).Info(args...)
}

// Infof logs a formatted info message
func Infof(c context.Context, format string, args ...interface{}) {
	addCaller(GetLogger(c), 2).Infof(format, args...)
}

// Warn logs at warn level
func Warn(c context.Context, args ...interface{}) {
	addCaller(GetLogger(c), 2).Warn(args...)
}

// Warnf logs a formatted warn message
func Warnf(c context.Context, format string, args ...interface{}) {
	addCaller(GetLogger(c), 2).Warnf(format, args...)
}

// Error logs at error level
func Error(c context.Context, args ...interface{}) {
	addCaller(GetLogger(c), 2).Error(args...)
}

// Errorf logs a formatted error message
func Errorf(c context.Context, format string, args ...interface{}) {
	addCaller(GetLogger(c), 2).Errorf(format, args...)
}

// ErrorWithFields logs an error with additional fields
func ErrorWithFields(c context.Context, err error, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	addCaller(GetLogger(c), 2).WithFields(fields).Error("An error occurred")
}

// Fatal logs at fatal level and exits
func Fatal(c context.Context, args ...interface{}) {
	addCaller(GetLogger(c), 2).Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(c context.Context, format string, args ...interface{}) {
	addCaller(GetLogger(c), 2).Fatalf(format, args...)
}

// CloneContext copies key values from one context to a new background context
func CloneContext(ctx context.Context) context.Context {
	newCtx := context.Background()

	for _, k := range []types.ContextKey{
		types.LoggerContextKey,
		types.TenantIDContextKey,
		types.RequestIDContextKey,
		types.TenantInfoContextKey,
	} {
		if v := ctx.Value(k); v != nil {
			newCtx = context.WithValue(newCtx, k, v)
		}
	}

	return newCtx
}
