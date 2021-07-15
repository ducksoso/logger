// +build darwin freebsd linux netbsd openbsd

package logger

func init() {
	stdoutHandler.Colorize = true
	stderrHandler.Colorize = true
}
