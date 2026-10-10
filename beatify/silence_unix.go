//go:build unix

package beatify

import (
	"os"
	"syscall"
)

// redirectStdio duplicates the real stdout, then points fds 1 and 2 at
// os.DevNull so the testing package cannot leak buffered failure output.
func redirectStdio() (*os.File, error) {
	fd, err := syscall.Dup(syscall.Stdout)
	if err != nil {
		return nil, err
	}

	saved := os.NewFile(uintptr(fd), "stdout")

	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		_ = saved.Close()

		return nil, err
	}

	if err := syscall.Dup2(int(null.Fd()), syscall.Stdout); err != nil {
		_ = null.Close()
		_ = saved.Close()

		return nil, err
	}

	if err := syscall.Dup2(int(null.Fd()), syscall.Stderr); err != nil {
		_ = null.Close()
		_ = saved.Close()

		return nil, err
	}

	os.Stdout = null
	os.Stderr = null

	return saved, nil
}
