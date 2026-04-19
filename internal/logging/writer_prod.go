//go:build !debug

package logging

import "io"

func logWriter(w io.Writer) io.Writer {
	return w
}
