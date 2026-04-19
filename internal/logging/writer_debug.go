//go:build debug

package logging

import (
	"io"
	"os"
)

func logWriter(w io.Writer) io.Writer {
	return io.MultiWriter(w, os.Stdout)
}
