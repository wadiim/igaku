package utils

import (
	"io"
	"strings"
)

type PrefixedWriter struct {
	Out	io.Writer
	Prefix	string
}

func (w *PrefixedWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	msg = whitespaceRE.ReplaceAllString(msg, " ")
	msg = w.Prefix + msg + "\n"
	return w.Out.Write([]byte(msg))
}

