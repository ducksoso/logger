package logger

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

type LogRecorder struct {
	Level     Level
	Formatter Formatter
	Records   map[string][]*Record
	Closed    bool
}

func NewLogRecorder() *LogRecorder {
	return &LogRecorder{
		Records: make(map[string][]*Record),
	}
}

func (b *LogRecorder) SetLevel(l Level) {
	b.Level = l
}

func (b *LogRecorder) SetFormatter(f Formatter) {
	b.Formatter = f
}

func (b *LogRecorder) Handle(rec *Record) {
	v, ok := b.Records[rec.LoggerName]
	if !ok {
		v = []*Record{}
	}
	b.Records[rec.LoggerName] = append(v, rec)
}

func (b *LogRecorder) Close() {
	b.Closed = true
}

func TestSinkHandler_Handle(t *testing.T) {
	loggers := 4 * runtime.NumCPU()
	logEntries := 100

	r := NewLogRecorder()
	b := NewSinkHandler(r, loggers)

	wg := sync.WaitGroup{}
	for i := 0; i < loggers; i++ {
		l := NewLogger(fmt.Sprint("logger ", i))
		l.SetHandler(b)
		wg.Add(1)
		go doLog(l, logEntries, &wg)
	}
	wg.Wait()

	b.Close()
	if !r.Closed {
		t.Errorf("Not closed")
	}

	for i := 0; i < loggers; i++ {
		if v, ok := r.Records[fmt.Sprint("logger ", i)]; !ok || len(v) != logEntries {
			t.Errorf("Missing log records expected %d got %d", logEntries, r.Records[fmt.Sprint("logger ", i)])
		}
	}
}

func doLog(l Logger, n int, wg *sync.WaitGroup) {
	for i := 0; i < n; i++ {
		l.Info("test %d", i)
		time.Sleep(time.Millisecond)
	}
	wg.Done()
}
