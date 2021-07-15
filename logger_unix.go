// +build darwin freebsd linux netbsd openbsd

package logger

func init() {
	StdoutHandler.Colorize = true
	StderrHandler.Colorize = true
}
