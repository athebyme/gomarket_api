package logger

import (
	"fmt"
	"io"
	"log"
	"sync"
)

type BaseLogger struct {
	mu     sync.Mutex
	prefix string
	writer io.Writer
}

func NewLogger(writer io.Writer, prefix string) *BaseLogger {
	return &BaseLogger{
		writer: writer,
		prefix: prefix,
	}
}

func (l *BaseLogger) Log(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	message := fmt.Sprintf(l.prefix+" "+format, v...)
	if l.writer != nil {
		fmt.Fprintln(l.writer, message) // пишем в writer
	}
	log.Print(message) // дублируем в консоль
}

func (l *BaseLogger) WithPrefix(extraPrefix string) *BaseLogger {
	return &BaseLogger{
		writer: l.writer,
		prefix: l.prefix + " " + extraPrefix,
	}
}

func (l *BaseLogger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}
func (l *BaseLogger) SetWriter(writer io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writer = writer
}
