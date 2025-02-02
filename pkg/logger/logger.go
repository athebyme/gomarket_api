package logger

type Logger interface {
	Log(format string, v ...interface{})
	SetPrefix(prefix string)
	Write(p []byte) (n int, err error)
	FatalLog(format string, v ...interface{})
	WithPrefix(extraPrefix string) Logger
	Close() error
}
