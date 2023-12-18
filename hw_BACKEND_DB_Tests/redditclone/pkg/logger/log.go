package logger

type Logger interface {
	LogW(level string, mesg string, keyValues map[string]interface{})
	Log(level string, mesg string)
	WithFields(keyValues map[string]interface{}) Logger
	WithError(err error) Logger
}
