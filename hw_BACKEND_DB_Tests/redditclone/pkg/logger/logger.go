package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type CustomLogger struct {
	zapLog *zap.Logger
}

func NewCustomLogger(config zap.Config) (*CustomLogger, error) {
	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("can't create logger")
	}
	return &CustomLogger{zapLog: logger}, nil
}

func (cl *CustomLogger) LogW(level string, msg string, keyValues map[string]interface{}) {
	zlevel := convertLevel(level)
	zapFields := make([]zap.Field, len(keyValues))
	i := 0
	for key, value := range keyValues {
		zapFields[i] = zap.Any(key, value)
		i++
	}
	cl.zapLog.With(zapFields...).Log(zlevel, msg)
}
func (cl *CustomLogger) Log(level string, msg string) {
	zlevel := convertLevel(level)
	cl.zapLog.Log(zlevel, msg)
}

func (cl *CustomLogger) WithFields(keyValues map[string]interface{}) Logger {
	zapFields := make([]zap.Field, len(keyValues))
	i := 0
	for key, value := range keyValues {
		zapFields[i] = zap.Any(key, value)
		i++
	}
	return &CustomLogger{zapLog: cl.zapLog.With(zapFields...)}
}

func (cl *CustomLogger) WithError(err error) Logger {
	return &CustomLogger{zapLog: cl.zapLog.With(zap.Error(err))}
}

func convertLevel(level string) zapcore.Level {
	switch level {
	case "Info":
		return zapcore.InfoLevel
	case "Debug":
		return zapcore.DebugLevel
	case "Error":
		return zapcore.ErrorLevel
	case "Panic":
		return zap.PanicLevel
	case "Fatal":
		return zap.FatalLevel
	case "Warn":
		return zap.WarnLevel
	default:
		return zapcore.InfoLevel
	}

}
