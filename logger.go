// Package logger is an alternative to log package in standard library.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type (
	color int // color represents log level colors
	level int // level represents severity of logs
)

// Logger levels.
const (
	CRITICAL level = iota
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

// Colors for different log levels.
const (
	BLACK color = iota + 30
	RED
	GREEN
	YELLOW
	BLUE
	MAGENTA
	CYAN
	WHITE
)

// levelNames provides mapping for log levels.
var levelNames = map[level]string{
	CRITICAL: "CRITICAL",
	ERROR:    "ERROR",
	WARNING:  "WARNING",
	NOTICE:   "NOTICE",
	INFO:     "INFO",
	DEBUG:    "DEBUG",
}

// levelColors provides mapping for log colors.
var levelColors = map[level]color{
	CRITICAL: MAGENTA,
	ERROR:    RED,
	WARNING:  YELLOW,
	NOTICE:   GREEN,
	INFO:     WHITE,
	DEBUG:    CYAN,
}

var (
	// DefaultLogger holds default logger
	DefaultLogger Logger = NewLogger(procName())

	// DefaultLevel holds default value for loggers
	DefaultLevel level = INFO

	// DefaultFormatter holds default formatter for loggers
	DefaultFormatter Formatter = &defaultFormatter{}

	// DefaultHandler holds default handler for loggers
	DefaultHandler Handler = stderrHandler

	// stdoutHandler holds a handler with outputting to stdout
	stdoutHandler = NewWriterHandler(os.Stdout)

	// stderrHandler holds a handler with outputting to stderr
	stderrHandler = NewWriterHandler(os.Stderr)
)

// Logger is the interface for output log messages in different levels.
// A new Logger can be created with NewLogger() function.
// You can changed the output handler with SetHandler() function.
type Logger interface {
	// SetLevel changes the level of the logger. Default is logging.Info.
	SetLevel(level)

	// SetHandler replaces the current handler for output. Default is logger.stderrHandler.
	SetHandler(Handler)

	// SetCallDepth sets the parameter passed to runtime.Caller().
	// It is used to get the file name from call stack.
	// For example you need to set it to 1 if you are using a wrapper around
	// the Logger. Default value is zero.
	SetCallDepth(int)

	// New creates a new inerhited context logger with given prefixes.
	New(prefixes ...interface{}) Logger

	// Fatal is equivalent to l.Critical followed by a call to os.Exit(1).
	Fatal(format string, args ...interface{})

	// Panic is equivalent to l.Critical followed by a call to panic().
	Panic(format string, args ...interface{})

	// Critical logs a message using CRITICAL as log level.
	Critical(format string, args ...interface{})

	// Error logs a message using ERROR as log level.
	Error(format string, args ...interface{})

	// Warning logs a message using WARNING as log level.
	Warning(format string, args ...interface{})

	// Notice logs a message using NOTICE as log level.
	Notice(format string, args ...interface{})

	// Info logs a message using INFO as log level.
	Info(format string, args ...interface{})

	// Debug logs a message using DEBUG as log level.
	Debug(format string, args ...interface{})
}

// Handler handles the output.
type Handler interface {
	SetFormatter(Formatter)
	SetLevel(level)

	// Handle single log record.
	Handle(*Record)

	// Close the handler.
	Close()
}

// Record contains all of the information about a single log message.
type Record struct {
	Format      string        // Format string
	Args        []interface{} // Arguments to format string
	LoggerName  string        // Name of the logger module
	Level       level         // Level of the record
	Time        time.Time     // Time of the record (local time)
	Filename    string        // File name of the log call (absolute path)
	Line        int           // Lint number in file
	ProcessID   int           // PID
	ProcessName string        // Name of the process
}

// Formatter formats a record.
type Formatter interface {
	// Format the record and return a message.
	Format(*Record) (message string)
}

// /////////////////////
//                   //
// Default Formatter //
//                   //
// /////////////////////

type defaultFormatter struct {
}

func (df *defaultFormatter) Format(rec *Record) string {
	paths := strings.Split(rec.Filename, string(os.PathSeparator))

	filePath := strings.Join(paths[len(paths)-2:], string(os.PathSeparator))

	return fmt.Sprintf("%s %-8s[%s:%d] %s", fmt.Sprint(rec.Time)[:19],
		levelNames[rec.Level], filePath, rec.Line, fmt.Sprintf(rec.Format, rec.Args...))
}

// /////////////////////////
//                       //
// Logger implementation //
//                       //
// /////////////////////////

// logger is the default Logger implementation.
type logger struct {
	Name      string
	Level     level
	Handler   Handler
	calldepth int
}

func NewLogger(name string) Logger {
	return &logger{
		Name:    name,
		Level:   DefaultLevel,
		Handler: DefaultHandler,
	}
}

// New creates a new inerhited logger with the given prefixes.
func (l *logger) New(prefixes ...interface{}) Logger {
	return newContext(*l, "", prefixes...)
}

func (l *logger) SetLevel(level level) {
	l.Level = level
}

func (l *logger) SetHandler(b Handler) {
	l.Handler = b
}

func (l *logger) SetCallDepth(d int) {
	l.calldepth = d
}

// Fatal is equivalent to l.Critical followed by a call to os.Exit(1).
func (l *logger) Fatal(format string, args ...interface{}) {
	l.Critical(format, args...)
	l.Handler.Close()
	os.Exit(1)
}

// Panic is equivalent to Critical() followed by a call to panic().
func (l *logger) Panic(format string, args ...interface{}) {
	l.Critical(format, args...)
	panic(fmt.Sprintf(format, args...))
}

// Critical sends a critical level log message to the handler. Arguments are handled in the manner of fmt.Printf.
func (l *logger) Critical(format string, args ...interface{}) {
	if l.Level >= CRITICAL {
		l.log(CRITICAL, format, args...)
	}
}

// Error sends a error level log message to the handler. Arguments are handled in the manner of fmt.Printf.
func (l *logger) Error(format string, args ...interface{}) {
	if l.Level >= ERROR {
		l.log(ERROR, format, args...)
	}
}

// Warning sends a warning level log message to the handler. Arguments are handled in the manner of fmt.Printf.
func (l *logger) Warning(format string, args ...interface{}) {
	if l.Level >= WARNING {
		l.log(WARNING, format, args...)
	}
}

// Notice sends a notice level log message to the handler. Arguments are handled in the manner of fmt.Printf.
func (l *logger) Notice(format string, args ...interface{}) {
	if l.Level >= NOTICE {
		l.log(NOTICE, format, args...)
	}
}

// Info sends a info level log message to the handler. Arguments are handled in the manner of fmt.Printf.
func (l *logger) Info(format string, args ...interface{}) {
	if l.Level >= INFO {
		l.log(INFO, format, args...)
	}
}

// Debug sends a debug level log message to the handler. Arguments are handled in the manner of fmt.Printf.
func (l *logger) Debug(format string, args ...interface{}) {
	if l.Level >= DEBUG {
		l.log(DEBUG, format, args...)
	}
}

func (l *logger) log(level level, format string, args ...interface{}) {
	// Add missing newline at the end.
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}

	_, file, line, ok := runtime.Caller(l.calldepth + 3)
	if !ok {
		file = "???"
		line = 0
	}

	rec := &Record{
		Format:      format,
		Args:        args,
		LoggerName:  l.Name,
		Level:       level,
		Time:        time.Now(),
		Filename:    file,
		Line:        line,
		ProcessID:   os.Getpid(),
		ProcessName: procName(),
	}

	l.Handler.Handle(rec)
}

// procName returns the name of the current process.
func procName() string {
	return filepath.Base(os.Args[0])
}

// /////////////////
//               //
// DefaultLogger //
//               //
// /////////////////

// Fatal is equivalent to Critical() followed by a call to os.Exit(1).
func Fatal(format string, args ...interface{}) {
	DefaultLogger.Fatal(format, args...)
}

// Panic is equivalent to Critical() followed by a call to panic().
func Panic(format string, args ...interface{}) {
	DefaultLogger.Panic(format, args...)
}

// Critical prints a critical level log message to the stderr. Arguments are handled in the manner of fmt.Printf.
func Critical(format string, args ...interface{}) {
	DefaultLogger.Critical(format, args...)
}

// Error prints a error level log message to the stderr. Arguments are handled in the manner of fmt.Printf.
func Error(format string, args ...interface{}) {
	DefaultLogger.Error(format, args...)
}

// Warning prints a warning level log message to the stderr. Arguments are handled in the manner of fmt.Printf.
func Warning(format string, args ...interface{}) {
	DefaultLogger.Warning(format, args...)
}

// Notice prints a notice level log message to the stderr. Arguments are handled in the manner of fmt.Printf.
func Notice(format string, args ...interface{}) {
	DefaultLogger.Notice(format, args...)
}

// Info prints a info level log message to the stderr. Arguments are handled in the manner of fmt.Printf.
func Info(format string, args ...interface{}) {
	DefaultLogger.Info(format, args...)
}

// Debug prints a debug level log message to the stderr. Arguments are handled in the manner of fmt.Printf.
func Debug(format string, args ...interface{}) {
	DefaultLogger.Debug(format, args...)
}

// ///////////////
//             //
// BaseHandler //
//             //
// ///////////////

// BaseHandler provides basic functionality for handler
type BaseHandler struct {
	Level     level
	Formatter Formatter
}

// NewBaseHandler creates a newBaseHandler with default values
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{
		Level:     DefaultLevel,
		Formatter: DefaultFormatter,
	}
}

// SetLevel sets logger level for handler
func (h *BaseHandler) SetLevel(l level) {
	h.Level = l
}

// SetFormatter sets logger formatter for handler
func (h *BaseHandler) SetFormatter(f Formatter) {
	h.Formatter = f
}

// FilterAndFormat filters any record according to logger level
func (h *BaseHandler) FilterAndFormat(rec *Record) string {
	if h.Level >= rec.Level {
		return h.Formatter.Format(rec)
	}
	return ""
}

// /////////////////
//               //
// WriterHandler //
//               //
// /////////////////

// WriterHandler is a handler implementation that writes the logger output to a io.Writer.
type WriterHandler struct {
	*BaseHandler
	w        io.Writer
	Colorize bool
}

// NewWriterHandler creates a new writer handler with given io.Writer
func NewWriterHandler(w io.Writer) *WriterHandler {
	return &WriterHandler{
		BaseHandler: NewBaseHandler(),
		w:           w,
	}
}

func (b *WriterHandler) Handle(rec *Record) {
	message := b.BaseHandler.FilterAndFormat(rec)
	if message == "" {
		return
	}
	if b.Colorize {
		b.w.Write([]byte(fmt.Sprintf("\033[%dm", levelColors[rec.Level])))
	}
	fmt.Fprint(b.w, message)
	if b.Colorize {
		b.w.Write([]byte("\033[0m")) // reset color
	}
}

// Close closes WriterHandler
func (b *WriterHandler) Close() {}

// ////////////////
//              //
// MultiHandler //
//              //
// ////////////////

// MultiHandler sends the log output to multiple handlers concurrently.
type MultiHandler struct {
	handlers []Handler
}

// NewMultiHandler creates a new handler with given handlers
func NewMultiHandler(handlers ...Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

// SetFormatter sets formatter for all handlers
func (b *MultiHandler) SetFormatter(f Formatter) {
	for _, h := range b.handlers {
		h.SetFormatter(f)
	}
}

// SetLevel sets level for all handlers
func (b *MultiHandler) SetLevel(l level) {
	for _, h := range b.handlers {
		h.SetLevel(l)
	}
}

// Handle handles given record with all handlers concurrently
func (b *MultiHandler) Handle(rec *Record) {
	wg := sync.WaitGroup{}
	wg.Add(len(b.handlers))
	for _, handler := range b.handlers {
		go func(handler Handler) {
			handler.Handle(rec)
			wg.Done()
		}(handler)
	}
	wg.Wait()
}

// Close closes all handlers concurrently
func (b *MultiHandler) Close() {
	wg := sync.WaitGroup{}
	wg.Add(len(b.handlers))
	for _, handler := range b.handlers {
		go func(handler Handler) {
			handler.Close()
			wg.Done()
		}(handler)
	}
	wg.Wait()
}
