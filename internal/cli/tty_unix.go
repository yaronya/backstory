//go:build !windows

package cli

import "os"

func openTTY() (*os.File, error) {
	return os.Open("/dev/tty")
}
