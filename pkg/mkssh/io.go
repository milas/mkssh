package mkssh

import (
	"io"
	"os"
)

var OpenTruncate = func(name string) (io.WriteCloser, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
}
