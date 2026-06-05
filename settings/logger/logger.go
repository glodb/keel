package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/glodb/keel/settings/configinterface"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides structured logging capabilities
type Logger struct {
	zapLogger *zap.Logger
	sugar     *zap.SugaredLogger
	mu        sync.RWMutex
	config    configinterface.ConfigInterface
}

var getLoggerInstance = sync.OnceValue(func() *Logger {
	logger := &Logger{}
	logger.initLogger("", "")
	return logger
})

func NewLogger() *Logger {
	return &Logger{}
}

func Log() *Logger {
	return getLoggerInstance()
}

func LogInit(config configinterface.ConfigInterface) *Logger {
	getLoggerInstance().config = config
	return getLoggerInstance()
}

func (l *Logger) Init(timekey string, format string) {
	l.initLogger(timekey, format)
}

func (l *Logger) SetSugaredLogger(zapLogger *zap.Logger, sugar *zap.SugaredLogger) {
	l.zapLogger = zapLogger
	l.sugar = sugar
}

func (l *Logger) SetLogger(zapLogger *zap.Logger) {
	l.zapLogger = zapLogger
	l.sugar = l.zapLogger.Sugar()
}

// initLogger initializes the Zap logger with appropriate configuration
func (l *Logger) initLogger(timekey string, format string) {
	// Get environment configuration

	env := "info"

	// Parse log level
	var level zapcore.Level
	switch env {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	if timekey == "" {
		timekey = "timestamp"
	}

	if format == "" {
		format = "json"
	}

	// Configure encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = timekey
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Determine output format

	var encoder zapcore.Encoder
	if format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else if format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Create logger
	l.zapLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	l.sugar = l.zapLogger.Sugar()
}

// SetLogLevel dynamically changes the log level
func (l *Logger) SetLogLevel(level string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	case "fatal":
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Create new logger with updated level
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	encoder := zapcore.NewJSONEncoder(encoderConfig)
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	l.zapLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	l.sugar = l.zapLogger.Sugar()
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}

// Debug logs a debug message with structured fields
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	if l.config != nil && l.config.GetPrintInfo() {
		l.zapLogger.Info(msg, fields...)
	}
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(template string, args ...interface{}) {
	if l.config != nil && l.config.GetPrintInfo() {
		l.sugar.Debugf(template, args...)
	}
}

func (l *Logger) DebugProd(template string, fields ...zap.Field) {
	l.zapLogger.Debug(template, fields...)
}

// Info logs an info message with structured fields
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zapLogger.Info(msg, fields...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

// Warn logs a warning message with structured fields
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zapLogger.Warn(msg, fields...)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

// Error logs an error message with structured fields
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zapLogger.Error(msg, fields...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

// Fatal logs a fatal message with structured fields and exits
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zapLogger.Fatal(msg, fields...)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

// With creates a child logger with the given fields
func (l *Logger) With(fields ...zap.Field) *zap.Logger {
	return l.zapLogger.With(fields...)
}

// WithSugar creates a child sugared logger with the given fields
func (l *Logger) WithSugar(args ...interface{}) *zap.SugaredLogger {
	return l.sugar.With(args...)
}

// LogRequest logs HTTP request information
func (l *Logger) LogRequest(method, path, remoteAddr string, statusCode int, duration time.Duration, userID string) {
	l.Info("HTTP Request",
		zap.String("method", method),
		zap.String("path", path),
		zap.String("remote_addr", remoteAddr),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("user_id", userID),
	)
}

// LogDatabase logs database operation information
func (l *Logger) LogDatabase(operation, collection string, duration time.Duration, error error) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("collection", collection),
		zap.Duration("duration", duration),
	}

	if error != nil {
		fields = append(fields, zap.Error(error))
		l.Error("Database operation failed", fields...)
	} else {
		l.Info("Database operation completed", fields...)
	}
}

// LogConfig logs configuration changes
func (l *Logger) LogConfig(action, key string, value interface{}) {
	l.Info("Configuration change",
		zap.String("action", action),
		zap.String("key", key),
		zap.Any("value", value),
	)
}

// LogSecurity logs security-related events
func (l *Logger) LogSecurity(event, userID, ipAddress string, success bool) {
	l.Info("Security event",
		zap.String("event", event),
		zap.String("user_id", userID),
		zap.String("ip_address", ipAddress),
		zap.Bool("success", success),
	)
}

// LogPerformance logs performance metrics
func (l *Logger) LogPerformance(operation string, duration time.Duration, memoryUsage int64) {
	l.Info("Performance metric",
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.Int64("memory_usage_bytes", memoryUsage),
	)
}

// LogBusiness logs business logic events
func (l *Logger) LogBusiness(event, entityType, entityID string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event", event),
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
	}

	for key, value := range metadata {
		fields = append(fields, zap.Any(key, value))
	}

	l.Info("Business event", fields...)
}

// Legacy compatibility methods
type CustomLogWriter struct {
	Prefix string
}

// Write formats the log output with the prefix after the date (legacy compatibility)
func (clw *CustomLogWriter) Write(p []byte) (n int, err error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("%s %s%s", timestamp, clw.Prefix, p)
	return os.Stdout.Write([]byte(logLine))
}

// Field wrappers to abstract zap field creation
// This allows easy migration to different logging libraries in the future

// StringField creates a string field for logging
func StringField(key, value string) zap.Field {
	return zap.String(key, value)
}

// IntField creates an int field for logging
func IntField(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Int64Field creates an int64 field for logging
func Int64Field(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

// Float64Field creates a float64 field for logging
func Float64Field(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

// BoolField creates a bool field for logging
func BoolField(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

// AnyField creates an interface{} field for logging
func AnyField(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// ErrorField creates an error field for logging
func ErrorField(key string, err error) zap.Field {
	return zap.Error(err)
}

// DurationField creates a duration field for logging
func DurationField(key string, duration time.Duration) zap.Field {
	return zap.Duration(key, duration)
}

// TimeField creates a time field for logging
func TimeField(key string, time time.Time) zap.Field {
	return zap.Time(key, time)
}

// StringSliceField creates a string slice field for logging
func StringSliceField(key string, values []string) zap.Field {
	return zap.Strings(key, values)
}

// IntSliceField creates an int slice field for logging
func IntSliceField(key string, values []int) zap.Field {
	return zap.Ints(key, values)
}

// Int64SliceField creates an int64 slice field for logging
func Int64SliceField(key string, values []int64) zap.Field {
	return zap.Int64s(key, values)
}

// Float64SliceField creates a float64 slice field for logging
func Float64SliceField(key string, values []float64) zap.Field {
	return zap.Float64s(key, values)
}

// BoolSliceField creates a bool slice field for logging
func BoolSliceField(key string, values []bool) zap.Field {
	return zap.Bools(key, values)
}

// ObjectField creates an object field for logging
func ObjectField(key string, obj interface{}) zap.Field {
	return zap.Any(key, obj)
}

// MapField creates a map field for logging
func MapField(key string, value map[string]interface{}) zap.Field {
	return zap.Any(key, value)
}

// ArrayField creates an array field for logging
func ArrayField(key string, value []interface{}) zap.Field {
	return zap.Any(key, value)
}
