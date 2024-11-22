package logger

type Logger interface {
	Log(format string, v ...interface{})
	SetPrefix(prefix string)
}
