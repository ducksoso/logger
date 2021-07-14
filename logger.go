// Package logger is an alternative to log package in standard library.
package logger

import (
	"fmt"
	"time"
)

type (
	// Color represents log level colors
	Color int
	// Level represents severity of logs
	Level int
)

// Colors for different log levels.
const (
	BLACK Color = iota + 30
	RED
	GREEN
	YELLOW
	BLUE
	MAGENTA
	CYAN
	WHITE
)

// Logger levels.
const (
	CRITICAL Level = iota
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

// LevelNames provides mapping for log levels.
var LevelNames = map[Level]string{
	CRITICAL: "CRITICAL",
	ERROR:    "ERROR",
	WARNING:  "WARNING",
	NOTICE:   "NOTICE",
	INFO:     "INFO",
	DEBUG:    "DEBUG",
}

// LevelColors provides mapping for log colors.
var LevelColors = map[Level]Color{
	CRITICAL: MAGENTA,
	ERROR:    RED,
	WARNING:  YELLOW,
	NOTICE:   GREEN,
	INFO:     WHITE,
	DEBUG:    CYAN,
}

var (
	// DefaultLogger holds default logger
	DefaultLogger Logger = NewLogger()

	DefaultLevel Level = INFO

	DefaultHandler Handler = StderrHandler
)

// Logger is the interface for output log messages in different levels.
// A new Logger can be created with NewLogger() function.
// You can changed the output handler with SetHandler() function.
type Logger interface {
	// SetLevel changes the level of the logger. Default is logging.Info.
	SetLevel(Level)

	// SetHandler replaces the current handler for output. Default is logger.StderrHandler.
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
	SetLevel(Level)

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
	Level       Level         // Level of the record
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

// #####################
// Default Formatter
// #####################

type defaultFormatter struct {
}

func (df *defaultFormatter) Format(rec *Record) string {
	return fmt.Sprintf("%s [%s] %-8s %s", fmt.Sprint(rec.Time)[:19], rec.LoggerName, LevelNames[rec.Level], fmt.Sprintf(rec.Format, rec.Args...))
}

// ########################
// Logger implementation
// ########################

// logger is the default Logger implementation.
type logger struct {
	Name      string
	Level     Level
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
	return nil
}

func (l *logger) SetLevel(level Level) {
	l.Level = level
}

func (l *logger) SetHandler(b Handler) {
	l.Handler = b
}

func (l logger) SetCallDepth(d int) {
	l.calldepth = d
}
