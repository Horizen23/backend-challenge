package logger

import (
	"context"
	"os"
	"time"

	"backend-challenge/apps/common/pkg/ctxutils"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field = zapcore.Field

func String(key, val string) Field                 { return zap.String(key, val) }
func Int(key string, val int) Field                { return zap.Int(key, val) }
func Int64(key string, val int64) Field            { return zap.Int64(key, val) }
func Float64(key string, val float64) Field        { return zap.Float64(key, val) }
func Bool(key string, val bool) Field              { return zap.Bool(key, val) }
func Any(key string, val interface{}) Field        { return zap.Any(key, val) }
func Err(err error) Field                          { return zap.Error(err) }
func Duration(key string, val time.Duration) Field { return zap.Duration(key, val) }

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithComponent(component string) Logger
	WithContext(ctx context.Context) Logger
	Sync() error
}

type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) Debug(msg string, fields ...Field) { l.logger.Debug(msg, fields...) }
func (l *zapLogger) Info(msg string, fields ...Field)  { l.logger.Info(msg, fields...) }
func (l *zapLogger) Warn(msg string, fields ...Field)  { l.logger.Warn(msg, fields...) }
func (l *zapLogger) Error(msg string, fields ...Field) { l.logger.Error(msg, fields...) }
func (l *zapLogger) Fatal(msg string, fields ...Field) { l.logger.Fatal(msg, fields...) }

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

func (l *zapLogger) WithComponent(component string) Logger {
	return &zapLogger{logger: l.logger.With(zap.String("component", component))}
}

func (l *zapLogger) WithContext(ctx context.Context) Logger {
	var fields []Field
	if reqID := ctxutils.GetRequestID(ctx); reqID != "" {
		fields = append(fields, String("request_id", reqID))
	}
	if userID, email := ctxutils.GetAuthUser(ctx); userID != "" {
		fields = append(fields, String("user_id", userID), String("email", email))
	}
	if len(fields) > 0 {
		return l.With(fields...)
	}
	return l
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

func NewLogger() Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	zapLog := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return &zapLogger{logger: zapLog}
}
